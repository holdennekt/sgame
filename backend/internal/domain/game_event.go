package domain

import (
	"slices"
	"time"
)

// MaxGameEvents caps the per-room log; the last slot is a log_truncated marker.
const MaxGameEvents = 5000

type GameEventType string

const (
	EventPlayerJoined         GameEventType = "player_joined"
	EventPlayerLeft           GameEventType = "player_left"
	EventPlayerDisconnected   GameEventType = "player_disconnected"
	EventPlayerReconnected    GameEventType = "player_reconnected"
	EventGameStarted          GameEventType = "game_started"
	EventRoundStarted         GameEventType = "round_started"
	EventRoundSkipped         GameEventType = "round_skipped"
	EventQuestionSelected     GameEventType = "question_selected"
	EventPassed               GameEventType = "passed"
	EventBetPlaced            GameEventType = "bet_placed"
	EventAuctionWon           GameEventType = "auction_won"
	EventBuzzed               GameEventType = "buzzed"
	EventFalseStart           GameEventType = "false_start"
	EventAnswerSubmitted      GameEventType = "answer_submitted"
	EventAnswerValidated      GameEventType = "answer_validated"
	EventQuestionEnded        GameEventType = "question_ended"
	EventFinalRoundStarted    GameEventType = "final_round_started"
	EventFinalCategoryRemoved GameEventType = "final_category_removed"
	EventFinalCategoryChosen  GameEventType = "final_category_chosen"
	EventFinalBetPlaced       GameEventType = "final_bet_placed"
	EventFinalQuestionStarted GameEventType = "final_question_started"
	EventFinalAnswerSubmitted GameEventType = "final_answer_submitted"
	EventFinalAnswerValidated GameEventType = "final_answer_validated"
	EventScoreChanged         GameEventType = "score_changed"
	EventPlayerBanned         GameEventType = "player_banned"
	EventPaused               GameEventType = "paused"
	EventUnpaused             GameEventType = "unpaused"
	EventGameEnded            GameEventType = "game_ended"
	EventLogTruncated         GameEventType = "log_truncated"
)

type QuestionEndReason string

const (
	QuestionAnswered       QuestionEndReason = "answered"
	QuestionNobodyAnswered QuestionEndReason = "nobody_answered"
	QuestionNoBets         QuestionEndReason = "no_bets"
	QuestionSkipped        QuestionEndReason = "skipped"
)

type VerdictMethod string

const (
	VerdictExact VerdictMethod = "exact"
	VerdictAI    VerdictMethod = "ai"
	VerdictError VerdictMethod = "error"
)

// Verdict is how the system decided an answer (AI host mode).
type Verdict struct {
	Method     VerdictMethod
	Confidence float64
}

type GameEvent struct {
	Seq   int            `json:"seq" bson:"seq"`
	At    time.Time      `json:"at" bson:"at"`
	Type  GameEventType  `json:"type" bson:"type"`
	Actor string         `json:"actor,omitempty" bson:"actor,omitempty"`
	Data  map[string]any `json:"data,omitempty" bson:"data,omitempty"`
}

func (r *Room) emit(eventType GameEventType, actor string, data map[string]any) {
	if r.EventSeq >= MaxGameEvents {
		return
	}
	if r.EventSeq == MaxGameEvents-1 {
		eventType, actor, data = EventLogTruncated, SYSTEM, nil
	}
	r.EventSeq++
	r.pendingEvents = append(r.pendingEvents, GameEvent{
		Seq:   r.EventSeq,
		At:    time.Now(),
		Type:  eventType,
		Actor: actor,
		Data:  data,
	})
}

// DrainEvents returns the events emitted since the last drain.
func (r *Room) DrainEvents() []GameEvent {
	events := r.pendingEvents
	r.pendingEvents = nil
	return events
}

func attachmentTypeOf(a *Attachment) any {
	if a == nil {
		return nil
	}
	return string(a.Type)
}

func (r *Room) playerName(userId string) string {
	if i := r.UsersPlayerIndex(userId); i != -1 {
		return r.Players[i].Name
	}
	if r.IsUserModerator(userId) {
		return r.Moderator.Name
	}
	return ""
}

func (r *Room) standings() []map[string]any {
	sorted := slices.Clone(r.Players)
	slices.SortStableFunc(sorted, func(a, b Player) int { return b.Score - a.Score })
	standings := make([]map[string]any, len(sorted))
	for i, p := range sorted {
		standings[i] = map[string]any{"player": p.Id, "score": p.Score}
	}
	return standings
}

func validationSource(actor string, verdict *Verdict, answer string) map[string]any {
	switch {
	case actor != SYSTEM:
		return map[string]any{"source": "moderator"}
	case verdict != nil:
		return map[string]any{"source": "ai", "method": string(verdict.Method), "confidence": verdict.Confidence}
	case answer == "":
		return map[string]any{"source": "timeout"}
	default:
		return map[string]any{"source": "system"}
	}
}

// RecordJoined logs a user taking a seat in the room.
func (r *Room) RecordJoined(user User, role string) {
	r.emit(EventPlayerJoined, user.Id, map[string]any{
		"player":  user.Id,
		"name":    user.Name,
		"avatar":  user.Avatar,
		"isGuest": user.IsGuest,
		"role":    role,
	})
}

// RecordLeft logs a user leaving before the game starts.
func (r *Room) RecordLeft(userId string) {
	if r.State != WaitingForStart {
		return
	}
	r.emit(EventPlayerLeft, userId, map[string]any{"player": userId})
}

// SetConnected updates the user's connection flag and logs changes while a game is running.
func (r *Room) SetConnected(userId string, connected bool) {
	var was bool
	if i := r.UsersPlayerIndex(userId); i != -1 {
		was = r.Players[i].IsConnected
		r.Players[i].IsConnected = connected
	}
	if r.IsUserModerator(userId) {
		if i := r.UsersPlayerIndex(userId); i == -1 {
			was = r.Moderator.IsConnected
		}
		r.Moderator.IsConnected = connected
	}
	if was == connected || r.State == WaitingForStart || r.State == GameOver {
		return
	}
	eventType := EventPlayerReconnected
	if !connected {
		eventType = EventPlayerDisconnected
	}
	r.emit(eventType, userId, map[string]any{"player": userId})
}
