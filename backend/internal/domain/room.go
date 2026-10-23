package domain

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

const (
	LOBBY              = "lobby"
	ROOM_PREFIX        = "room:"
	STATE_POSTFIX      = ":state"
	LOCK_POSTFIX       = ":lock"
	OWNER_POSTFIX      = ":owner"
	INTERNAL_POSTFIX   = ":internal"
	SPECTATORS_POSTFIX = ":spectators"
	LOG_POSTFIX        = ":log"

	ExtraQuestionThinkingTime = time.Second
	MaxPauseDuration          = time.Hour
)

type Room struct {
	Id                    string                `json:"id" bson:"_id"`
	Name                  string                `json:"name" bson:"name"`
	CreatedBy             string                `json:"createdBy" bson:"createdBy"`
	Options               RoomOptions           `json:"options" bson:"options"`
	PackPreview           PackPreview           `json:"packPreview" bson:"packPreview"`
	Moderator             *Moderator            `json:"moderator" bson:"moderator"`
	Players               []Player              `json:"players" bson:"players"`
	BanList               []string              `json:"banList" bson:"banList"`
	State                 RoomState             `json:"state" bson:"state"`
	CurrentRoundName      *string               `json:"currentRoundName" bson:"currentRoundName"`
	CurrentRoundQuestions CurrentRoundQuestions `json:"currentRoundQuestions" bson:"currentRoundQuestions"`
	CurrentPlayer         *string               `json:"currentPlayer" bson:"currentPlayer"`
	CurrentQuestion       *CurrentQuestion      `json:"currentQuestion" bson:"currentQuestion"`
	AnsweringPlayer       *AnsweringPlayer      `json:"answeringPlayer" bson:"answeringPlayer"`
	AllowedToAnswer       []string              `json:"allowedToAnswer" bson:"allowedToAnswer"`
	FinalRoundState       *FinalRoundState      `json:"finalRoundState" bson:"finalRoundState"`
	PausedState           PausedState           `json:"pausedState" bson:"pausedState"`
	FinishedAt            *time.Time            `json:"finishedAt" bson:"finishedAt"`
	EventSeq              int                   `json:"eventSeq" bson:"eventSeq"`

	pendingEvents []GameEvent
}

type RoomOptions struct {
	MaxPlayers                int         `json:"maxPlayers" bson:"maxPlayers" binding:"min=1,max=10"`
	Type                      PrivacyType `json:"type" bson:"type" binding:"oneof=public private"`
	Password                  *string     `json:"password" bson:"password" binding:"omitnil,min=4,max=16"`
	ReadingSymbolsPerSecond   int         `json:"readingSymbolsPerSecond" bson:"readingSymbolsPerSecond" binding:"min=10,max=50"`
	QuestionThinkingTime      int         `json:"questionThinkingTime" bson:"questionThinkingTime" binding:"min=1,max=30"`
	AnswerThinkingTime        int         `json:"answerThinkingTime" bson:"answerThinkingTime" binding:"min=1,max=30"`
	QuestionThinkingTimeFinal int         `json:"questionThinkingTimeFinal" bson:"questionThinkingTimeFinal" binding:"min=1,max=120"`
	FalseStartEnabled         bool        `json:"falseStartEnabled" bson:"falseStartEnabled"`
	TimeToBet                 int         `json:"timeToBet,omitempty" bson:"timeToBet"`
	TimeToPass                int         `json:"timeToPass,omitempty" bson:"timeToPass"`
	AIHost                    bool        `json:"aiHost" bson:"aiHost"`
}

type PrivacyType string

const (
	Public  PrivacyType = "public"
	Private PrivacyType = "private"
)

type RoomState string

const (
	WaitingForStart             RoomState = "waiting_for_start"
	SelectingQuestion           RoomState = "selecting_question"
	RevealingQuestion           RoomState = "revealing_question"
	ShowingQuestion             RoomState = "showing_question"
	Answering                   RoomState = "answering"
	Betting                     RoomState = "betting"
	Passing                     RoomState = "passing"
	SelectingFinalRoundCategory RoomState = "selecting_final_round_category"
	FinalRoundBetting           RoomState = "final_round_betting"
	ShowingFinalRoundQuestion   RoomState = "showing_final_round_question"
	ValidatingFinalRoundAnswers RoomState = "validating_final_round_answers"
	GameOver                    RoomState = "game_over"
)

type CurrentRoundQuestions []CategoryQuestions
type CategoryQuestions struct {
	Category  string          `json:"category" bson:"category"`
	Questions []BoardQuestion `json:"questions" bson:"questions"`
}
type BoardQuestion struct {
	Value         int  `json:"value" bson:"value"`
	HasBeenPlayed bool `json:"hasBeenPlayed" bson:"hasBeenPlayed"`
}

func (crq CurrentRoundQuestions) findCategory(category string) *CategoryQuestions {
	for i := range crq {
		if crq[i].Category == category {
			return &crq[i]
		}
	}
	return nil
}

type CurrentQuestion struct {
	Question
	AttachmentRevealEndsAt       time.Time `json:"attachmentRevealEndsAt" bson:"attachmentRevealEndsAt"`
	AttachmentRevealLastProgress float64   `json:"attachmentRevealLastProgress" bson:"attachmentRevealLastProgress"`
	TextRevealLastProgress       float64   `json:"textRevealLastProgress" bson:"textRevealLastProgress"`
	TimerStartsAt                time.Time `json:"timerStartsAt" bson:"timerStartsAt"`
	TimerEndsAt                  time.Time `json:"timerEndsAt" bson:"timerEndsAt"`
	TimerLastProgress            float64   `json:"timerLastProgress" bson:"timerLastProgress"`
	BettingEndsAt                time.Time `json:"bettingEndsAt" bson:"bettingEndsAt"`
	PassingEndsAt                time.Time `json:"passingEndsAt" bson:"passingEndsAt"`
}

type AnsweringPlayer struct {
	Id            string    `json:"id" bson:"id"`
	TimerStartsAt time.Time `json:"timerStartsAt" bson:"timerStartsAt"`
	TimerEndsAt   time.Time `json:"timerEndsAt" bson:"timerEndsAt"`
	Answer        string    `json:"answer" bson:"answer"`
}

type FinalRoundState struct {
	AvailableCategories map[string]bool     `json:"availableCategories" bson:"availableCategories"`
	Question            *FinalRoundQuestion `json:"question" bson:"question"`
	Players             []string            `json:"players" bson:"players"`
	PlayersAnswers      map[string]string   `json:"playersAnswers" bson:"playersAnswers"`
	BettingEndsAt       *time.Time          `json:"bettingEndsAt" bson:"bettingEndsAt"`
	TimerEndsAt         *time.Time          `json:"timerEndsAt" bson:"timerEndsAt"`
}

type PausedState struct {
	Paused   bool       `json:"paused" bson:"paused"`
	PausedAt *time.Time `json:"pausedAt" bson:"pausedAt"`
}

func (r *Room) IsUserModerator(userId string) bool {
	if r.Moderator == nil {
		return false
	}
	return userId == r.Moderator.Id
}

func (r *Room) UsersPlayerIndex(userId string) int {
	return slices.IndexFunc(r.Players, func(p Player) bool {
		return userId == p.Id
	})
}

func (r *Room) IsUserIn(userId string) bool {
	return r.IsUserModerator(userId) || r.UsersPlayerIndex(userId) != -1
}

func (r *Room) IsUserBanned(userId string) bool {
	return slices.ContainsFunc(r.BanList, func(id string) bool {
		return userId == id
	})
}

func (r *Room) GetProjection(userId string, spectatorCount int) any {
	if r.IsUserModerator(userId) && !r.Options.AIHost {
		return NewModeratorRoom(r, spectatorCount)
	}
	return NewPlayerRoom(r, spectatorCount)
}

func (r *Room) StartGame(pack *Pack) {
	playerIds := make([]string, len(r.Players))
	for i, p := range r.Players {
		playerIds[i] = p.Id
	}
	r.emit(EventGameStarted, r.Moderator.Id, map[string]any{
		"pack":       r.PackPreview.Name,
		"players":    playerIds,
		"aiHost":     r.Options.AIHost,
		"falseStart": r.Options.FalseStartEnabled,
	})
	r.StartNextRegularRound(pack)
	r.CurrentPlayer = &r.Players[rand.Intn(len(r.Players))].Id
}

func (r *Room) StartNextRegularRound(pack *Pack) bool {
	var nextRoundIndex int
	if r.State == WaitingForStart {
		nextRoundIndex = 0
	} else {
		currentRoundIndex := slices.IndexFunc(pack.Rounds, func(round Round) bool {
			return *r.CurrentRoundName == round.Name
		})
		nextRoundIndex = currentRoundIndex + 1
	}
	if nextRoundIndex < len(pack.Rounds) {
		nextRound := pack.Rounds[nextRoundIndex]
		r.CurrentRoundName = &nextRound.Name
		r.CurrentRoundQuestions = nextRound.getCurrentRoundQuestions()
		r.State = SelectingQuestion
		r.emit(EventRoundStarted, SYSTEM, map[string]any{"round": nextRound.Name, "index": nextRoundIndex})
		return true
	}
	return false
}

func (r *Room) SelectQuestion(userId string, pack *Pack, category string, index int, getAttachmentUrl func(key string) (string, error)) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != SelectingQuestion {
		return custerr.NewConflictErr("can not select question now")
	}
	if userId != *r.CurrentPlayer && !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to select question")
	}
	catQuestions := r.CurrentRoundQuestions.findCategory(category)
	if catQuestions == nil || catQuestions.Questions[index].HasBeenPlayed {
		return custerr.NewConflictErr("question has already been played")
	}

	question, err := pack.GetQuestion(*r.CurrentRoundName, category, index)
	if err != nil {
		return err
	}

	if question.Attachment != nil {
		u, err := getAttachmentUrl(question.Attachment.Key)
		if err != nil {
			return err
		}
		question.Attachment.URL = u
	}

	r.CurrentQuestion = &CurrentQuestion{Question: *question}
	catQuestions.Questions[index].HasBeenPlayed = true
	r.emit(EventQuestionSelected, userId, map[string]any{
		"round":          question.Round,
		"category":       question.Category,
		"value":          question.Value,
		"type":           string(question.Type),
		"text":           question.Text,
		"attachmentType": attachmentTypeOf(question.Attachment),
	})

	switch question.Type {
	case Regular:
		allowedToAnswer := make([]string, len(r.Players))
		for i, player := range r.Players {
			allowedToAnswer[i] = player.Id
		}
		r.revealRegularQuestion(allowedToAnswer)
	case CatInBag:
		canPassTo := make([]Player, 0)
		for _, p := range r.Players {
			if p.Id != userId && p.IsConnected {
				canPassTo = append(canPassTo, p)
			}
		}
		if len(canPassTo) == 0 {
			r.passTo(userId, *r.CurrentPlayer, userId, true)
		} else if len(canPassTo) == 1 {
			r.passTo(userId, *r.CurrentPlayer, canPassTo[0].Id, true)
		} else {
			r.State = Passing
			r.CurrentQuestion.PassingEndsAt = time.Now().Add(time.Duration(r.Options.TimeToPass) * time.Second)
		}
	case Auction:
		canBet := make([]Player, 0)
		for _, p := range r.Players {
			if p.Score > 0 {
				canBet = append(canBet, p)
			}
		}
		if len(canBet) > 0 {
			r.State = Betting
			r.CurrentQuestion.BettingEndsAt = time.Now().Add(time.Duration(r.Options.TimeToBet) * time.Second)
		} else {
			allowedToAnswer := make([]string, len(r.Players))
			for i, player := range r.Players {
				allowedToAnswer[i] = player.Id
			}
			r.revealRegularQuestion(allowedToAnswer)
		}
	}

	return nil
}

func (r *Room) revealRegularQuestion(allowedToAnswer []string) {
	r.AllowedToAnswer = allowedToAnswer
	mediaRevealingDuration := r.CurrentQuestion.GetMediaRevealingDuration()
	if r.CurrentQuestion.Attachment != nil {
		r.CurrentQuestion.AttachmentRevealEndsAt = time.Now().Add(mediaRevealingDuration)
	}
	textRevealDuration := r.CurrentQuestion.GetTextRevealingDuration(r.Options.ReadingSymbolsPerSecond)
	r.CurrentQuestion.TimerStartsAt = time.Now().Add(mediaRevealingDuration).Add(textRevealDuration)
	r.State = RevealingQuestion
}

func (r *Room) StartRegularQuestion() {
	if r.CurrentQuestion.Attachment != nil {
		r.CurrentQuestion.AttachmentRevealLastProgress = 1
	}
	r.CurrentQuestion.TextRevealLastProgress = 1
	thinkingDuration := time.Duration(r.Options.QuestionThinkingTime) * time.Second
	r.CurrentQuestion.TimerEndsAt = r.CurrentQuestion.TimerStartsAt.Add(thinkingDuration)
	r.CurrentQuestion.TimerLastProgress = 1
	r.State = ShowingQuestion
}

func (r *Room) passTo(actor, from, to string, auto bool) {
	r.emit(EventPassed, actor, map[string]any{"from": from, "to": to, "auto": auto})
	r.startNonRegularQuestion(to)
}

func (r *Room) startNonRegularQuestion(allowedToAnswer string) {
	r.CurrentPlayer = &allowedToAnswer
	now := time.Now()
	revealingDuration := r.CurrentQuestion.GetMediaRevealingDuration() + r.CurrentQuestion.GetTextRevealingDuration(r.Options.ReadingSymbolsPerSecond)
	r.CurrentQuestion.TextRevealLastProgress = 1
	thinkingDuration := time.Duration(r.Options.AnswerThinkingTime) * time.Second
	r.AnsweringPlayer = &AnsweringPlayer{
		Id:            allowedToAnswer,
		TimerStartsAt: now,
		TimerEndsAt:   now.Add(revealingDuration + thinkingDuration),
	}
	r.State = Answering
}

func (r *Room) StartAnswer(userId string) error {
	now := time.Now()

	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != RevealingQuestion && r.State != ShowingQuestion {
		return custerr.NewConflictErr("can not start answer now")
	}
	if !slices.Contains(r.AllowedToAnswer, userId) {
		return custerr.NewConflictErr("not allowed to start answer")
	}
	if r.State == RevealingQuestion && r.Options.FalseStartEnabled {
		r.AllowedToAnswer = slices.DeleteFunc(r.AllowedToAnswer, func(playerId string) bool {
			return userId == playerId
		})
		r.emit(EventFalseStart, userId, map[string]any{"player": userId})
		if len(r.AllowedToAnswer) == 0 {
			r.EndQuestion(QuestionNobodyAnswered, SYSTEM)
		}
		return nil
	}

	phase := "showing"
	if r.State == RevealingQuestion {
		phase = "revealing"
	}
	r.emit(EventBuzzed, userId, map[string]any{
		"player":           userId,
		"phase":            phase,
		"msFromTimerStart": now.Sub(r.CurrentQuestion.TimerStartsAt).Milliseconds(),
	})

	switch r.State {
	case RevealingQuestion:
		if r.CurrentQuestion.Attachment != nil {
			fullDuration := float64(r.CurrentQuestion.GetMediaRevealingDuration())
			remainedDuration := math.Max(0, float64(r.CurrentQuestion.AttachmentRevealEndsAt.Sub(now)))
			r.CurrentQuestion.AttachmentRevealLastProgress = 1 - remainedDuration/fullDuration
		}
		if r.CurrentQuestion.Text != nil {
			fullDuration := float64(r.CurrentQuestion.GetTextRevealingDuration(r.Options.ReadingSymbolsPerSecond))
			remainedDuration := math.Min(fullDuration, float64(r.CurrentQuestion.TimerStartsAt.Sub(now)))
			r.CurrentQuestion.TextRevealLastProgress = 1 - remainedDuration/fullDuration
		}
	case ShowingQuestion:
		fullDuration := float64((time.Duration(r.Options.QuestionThinkingTime) * time.Second))
		remainedDuration := float64(r.CurrentQuestion.TimerEndsAt.Sub(now))
		r.CurrentQuestion.TimerLastProgress = remainedDuration / fullDuration
	}

	thinkingDuration := time.Duration(r.Options.AnswerThinkingTime) * time.Second
	r.AnsweringPlayer = &AnsweringPlayer{
		Id:            userId,
		TimerStartsAt: now,
		TimerEndsAt:   now.Add(thinkingDuration),
	}
	r.CurrentPlayer = &userId
	r.AllowedToAnswer = slices.DeleteFunc(r.AllowedToAnswer, func(playerId string) bool {
		return userId == playerId
	})
	r.State = Answering
	return nil
}

func (r *Room) SubmitAnswer(userId, answer string) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != Answering {
		return custerr.NewConflictErr("can not submit answer now")
	}
	if r.AnsweringPlayer == nil || r.AnsweringPlayer.Id != userId {
		return custerr.NewForbiddenErr("not your turn to answer")
	}
	if !r.Options.AIHost {
		return custerr.NewForbiddenErr("text answers are only used in AI host mode")
	}
	r.AnsweringPlayer.Answer = answer
	r.AnsweringPlayer.TimerEndsAt = time.Now()
	r.emit(EventAnswerSubmitted, userId, map[string]any{"player": userId, "answer": answer})
	return nil
}

func (r *Room) ValidateAnswer(userId string, isCorrect bool) error {
	return r.validateAnswer(userId, isCorrect, nil)
}

// ValidateAnswerSystem applies an AI host verdict.
func (r *Room) ValidateAnswerSystem(isCorrect bool, verdict Verdict) error {
	return r.validateAnswer(SYSTEM, isCorrect, &verdict)
}

func (r *Room) validateAnswer(userId string, isCorrect bool, verdict *Verdict) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != Answering {
		return custerr.NewConflictErr("can not validate answer now")
	}
	if r.Options.AIHost && userId != SYSTEM {
		return custerr.NewForbiddenErr("answer validation is handled by AI in this room")
	}
	if !r.IsUserModerator(userId) && userId != SYSTEM {
		return custerr.NewForbiddenErr("not allowed to validate answer")
	}

	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return r.AnsweringPlayer.Id == p.Id
	})
	questionValue := r.CurrentQuestion.Value
	if betAmount := r.Players[playerIndex].BetAmount; betAmount != nil {
		questionValue = *betAmount
	}
	delta := questionValue
	if !isCorrect {
		delta = -questionValue
	}
	r.Players[playerIndex].Score += delta

	validated := validationSource(userId, verdict, r.AnsweringPlayer.Answer)
	validated["player"] = r.AnsweringPlayer.Id
	validated["answer"] = r.AnsweringPlayer.Answer
	validated["correct"] = isCorrect
	validated["delta"] = delta
	validated["score"] = r.Players[playerIndex].Score
	r.emit(EventAnswerValidated, userId, validated)

	if isCorrect {
		r.EndQuestion(QuestionAnswered, SYSTEM)
	} else if len(r.AllowedToAnswer) == 0 {
		r.EndQuestion(QuestionNobodyAnswered, SYSTEM)
	} else {
		r.continueRegularQuestion()
	}
	return nil
}

func (r *Room) continueRegularQuestion() {
	now := time.Now()
	answerDuration := now.Sub(r.AnsweringPlayer.TimerStartsAt)
	if r.AnsweringPlayer.TimerStartsAt.Before(r.CurrentQuestion.TimerStartsAt) {
		if r.CurrentQuestion.Attachment != nil && r.AnsweringPlayer.TimerStartsAt.Before(r.CurrentQuestion.AttachmentRevealEndsAt) {
			r.CurrentQuestion.AttachmentRevealEndsAt = r.CurrentQuestion.AttachmentRevealEndsAt.Add(answerDuration)
		}
		r.CurrentQuestion.TimerStartsAt = r.CurrentQuestion.TimerStartsAt.Add(answerDuration)
		r.State = RevealingQuestion
	} else {
		r.CurrentQuestion.TimerEndsAt = r.CurrentQuestion.TimerEndsAt.Add(answerDuration).Add(ExtraQuestionThinkingTime)
		r.State = ShowingQuestion
	}
	r.AnsweringPlayer = nil
}

func (r *Room) PassQuestion(fromUserId string, toUserId string) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != Passing {
		return custerr.NewConflictErr("can not pass question now")
	}
	toUserIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return p.Id == toUserId
	})
	if toUserIndex == -1 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no player with id \"%s\" in room", toUserId))
	}
	if *r.CurrentPlayer != fromUserId && !r.IsUserModerator(fromUserId) {
		return custerr.NewConflictErr("not allowed to pass question")
	}
	if *r.CurrentPlayer == toUserId {
		return custerr.NewConflictErr("can not pass question to current player")
	}
	if !r.Players[toUserIndex].IsConnected {
		return custerr.NewConflictErr("can not pass question to disconnected player")
	}

	r.passTo(fromUserId, *r.CurrentPlayer, toUserId, false)
	return nil
}

func (r *Room) PassQuestionAuto() {
	var passTo string
	canPassTo := make([]Player, 0)
	for _, p := range r.Players {
		if p.Id != *r.CurrentPlayer && p.IsConnected {
			canPassTo = append(canPassTo, p)
		}
	}
	if len(canPassTo) == 0 {
		passTo = *r.CurrentPlayer
	} else {
		passTo = canPassTo[rand.Intn(len(canPassTo))].Id
	}
	r.passTo(SYSTEM, *r.CurrentPlayer, passTo, true)
}

func (r *Room) PlaceBet(userId string, amount int) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != Betting {
		return custerr.NewConflictErr("can not place bet now")
	}
	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return userId == p.Id
	})
	alreadyBet := r.Players[playerIndex].BetAmount != nil
	if alreadyBet {
		return custerr.NewConflictErr("can not place bet again")
	}
	insufficientScore := amount > r.Players[playerIndex].Score || amount < 0
	if insufficientScore {
		return custerr.NewConflictErr("insufficient bet size")
	}

	r.Players[playerIndex].BetAmount = &amount
	r.emit(EventBetPlaced, userId, map[string]any{"player": userId, "amount": amount, "auto": false})

	canBet := make([]Player, 0)
	for _, p := range r.Players {
		if p.Score > 0 {
			canBet = append(canBet, p)
		}
	}
	allBet := !slices.ContainsFunc(canBet, func(p Player) bool {
		return p.BetAmount == nil
	})
	if allBet {
		playerMaxBet := canBet[0]
		for _, p := range canBet[1:] {
			if *playerMaxBet.BetAmount < *p.BetAmount {
				playerMaxBet = p
			}
		}
		r.finishBetting(playerMaxBet)
	}
	return nil
}

func (r *Room) PlaceBetsAuto() {
	canBet := make([]*Player, 0)
	for pi, p := range r.Players {
		if p.Score > 0 {
			canBet = append(canBet, &r.Players[pi])
		}
	}
	for i, player := range canBet {
		if player.BetAmount == nil {
			zero := 0
			canBet[i].BetAmount = &zero
			r.emit(EventBetPlaced, SYSTEM, map[string]any{"player": player.Id, "amount": 0, "auto": true})
		}
	}
	playerMaxBet := canBet[0]
	for _, p := range canBet[1:] {
		if *playerMaxBet.BetAmount < *p.BetAmount {
			playerMaxBet = p
		}
	}
	r.finishBetting(*playerMaxBet)
}

func (r *Room) finishBetting(highest Player) {
	if *highest.BetAmount > 0 {
		r.emit(EventAuctionWon, SYSTEM, map[string]any{"player": highest.Id, "amount": *highest.BetAmount})
		r.startNonRegularQuestion(highest.Id)
	} else {
		r.EndQuestion(QuestionNoBets, SYSTEM)
	}
}

func (r *Room) EndQuestion(reason QuestionEndReason, actor string) {
	ended := map[string]any{"reason": string(reason)}
	if r.CurrentQuestion != nil {
		ended["answers"] = r.CurrentQuestion.Answers
	}
	if reason == QuestionAnswered && r.AnsweringPlayer != nil {
		ended["winner"] = r.AnsweringPlayer.Id
	}
	r.emit(EventQuestionEnded, actor, ended)
	r.CurrentQuestion = nil
	r.AnsweringPlayer = nil
	r.AllowedToAnswer = make([]string, 0)
	for i := range r.Players {
		r.Players[i].BetAmount = nil
	}
	r.State = SelectingQuestion
}

func (r *Room) AnyAvailableQuestions() bool {
	for _, cq := range r.CurrentRoundQuestions {
		notPlayedIndex := slices.IndexFunc(cq.Questions, func(bq BoardQuestion) bool {
			return !bq.HasBeenPlayed
		})
		if notPlayedIndex != -1 {
			return true
		}
	}
	return false
}

func (r *Room) StartFinalRound(pack *Pack, getAttachmentUrl func(key string) (string, error)) (bool, error) {
	r.CurrentRoundName = nil
	r.CurrentRoundQuestions = nil
	r.CurrentPlayer = nil
	r.CurrentQuestion = nil
	r.AnsweringPlayer = nil
	r.AllowedToAnswer = nil

	finalRoundPlayers := make([]string, 0)
	for _, player := range r.Players {
		if player.Score > 0 {
			finalRoundPlayers = append(finalRoundPlayers, player.Id)
		}
	}
	if len(finalRoundPlayers) == 0 {
		return false, nil
	}

	r.FinalRoundState = &FinalRoundState{
		Players:             finalRoundPlayers,
		AvailableCategories: pack.FinalRound.getAvailableCategories(),
	}
	r.CurrentPlayer = &r.FinalRoundState.Players[rand.Intn(len(r.FinalRoundState.Players))]
	r.State = SelectingFinalRoundCategory
	r.emit(EventFinalRoundStarted, SYSTEM, map[string]any{"players": finalRoundPlayers})

	availableFinalCategories := r.GetAvailableFinalRoundCategories()
	if len(availableFinalCategories) == 1 {
		if err := r.chooseFinalRoundCategory(pack, availableFinalCategories[0], getAttachmentUrl); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (r *Room) GetAvailableFinalRoundCategories() []string {
	availableCategories := make([]string, 0)
	for category, available := range r.FinalRoundState.AvailableCategories {
		if available {
			availableCategories = append(availableCategories, category)
		}
	}
	return availableCategories
}

func (r *Room) RemoveFinalRoundCategory(pack *Pack, userId string, category string, getAttachmentUrl func(key string) (string, error)) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != SelectingFinalRoundCategory {
		return custerr.NewConflictErr("cannot remove final round category now")
	}
	if userId != *r.CurrentPlayer && !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to remove final round category")
	}

	available, ok := r.FinalRoundState.AvailableCategories[category]
	if !ok {
		return custerr.NewNotFoundErr(fmt.Sprintf("no final round category \"%s\"", category))
	}
	if !available {
		return custerr.NewConflictErr(fmt.Sprintf("final round category \"%s\" has already been removed", category))
	}

	r.FinalRoundState.AvailableCategories[category] = false
	r.emit(EventFinalCategoryRemoved, userId, map[string]any{"category": category})
	playerIndex := slices.IndexFunc(r.FinalRoundState.Players, func(p string) bool {
		return p == userId
	})
	if playerIndex == len(r.FinalRoundState.Players)-1 {
		r.CurrentPlayer = &r.FinalRoundState.Players[0]
	} else {
		r.CurrentPlayer = &r.FinalRoundState.Players[playerIndex+1]
	}

	availableFinalCategories := r.GetAvailableFinalRoundCategories()
	if len(availableFinalCategories) == 1 {
		if err := r.chooseFinalRoundCategory(pack, availableFinalCategories[0], getAttachmentUrl); err != nil {
			return err
		}
	}
	return nil
}

func (r *Room) chooseFinalRoundCategory(pack *Pack, category string, getAttachmentUrl func(key string) (string, error)) error {
	categoryIndex := slices.IndexFunc(pack.FinalRound.Categories, func(frc FinalRoundCategory) bool {
		return frc.Name == category
	})
	r.FinalRoundState.Question = &pack.FinalRound.Categories[categoryIndex].Question
	if r.FinalRoundState.Question.Attachment != nil {
		u, err := getAttachmentUrl(r.FinalRoundState.Question.Attachment.Key)
		if err != nil {
			return err
		}
		r.FinalRoundState.Question.Attachment.URL = u
	}
	r.emit(EventFinalCategoryChosen, SYSTEM, map[string]any{
		"category":       category,
		"text":           r.FinalRoundState.Question.Text,
		"attachmentType": attachmentTypeOf(r.FinalRoundState.Question.Attachment),
		"answers":        r.FinalRoundState.Question.Answers,
	})
	r.CurrentPlayer = nil
	r.State = FinalRoundBetting
	bettingEndsAt := time.Now().Add(time.Duration(r.Options.TimeToBet) * time.Second)
	r.FinalRoundState.BettingEndsAt = &bettingEndsAt
	return nil
}

func (r *Room) PlaceFinalRoundBet(userId string, amount int) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != FinalRoundBetting {
		return custerr.NewConflictErr("can not place final round bet now")
	}
	inFinalRound := slices.Contains(r.FinalRoundState.Players, userId)
	if !inFinalRound {
		return custerr.NewForbiddenErr("not allowed to place bet")
	}
	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return userId == p.Id
	})
	alreadyBet := r.Players[playerIndex].BetAmount != nil
	if alreadyBet {
		return custerr.NewConflictErr("can not place bet again")
	}
	insufficientScore := amount > r.Players[playerIndex].Score || amount < 0
	if insufficientScore {
		return custerr.NewConflictErr("insufficient score")
	}

	r.Players[playerIndex].BetAmount = &amount
	r.emit(EventFinalBetPlaced, userId, map[string]any{"player": userId, "amount": amount, "auto": false})

	allBet := !slices.ContainsFunc(r.Players, func(p Player) bool {
		return slices.Contains(r.FinalRoundState.Players, p.Id) && p.BetAmount == nil
	})
	if allBet {
		r.startFinalRoundQuestion()
	}
	return nil
}

func (r *Room) PlaceFinalRoundBetsAuto() {
	for i := range r.Players {
		if r.Players[i].BetAmount == nil {
			zero := 0
			r.Players[i].BetAmount = &zero
			if slices.Contains(r.FinalRoundState.Players, r.Players[i].Id) {
				r.emit(EventFinalBetPlaced, SYSTEM, map[string]any{"player": r.Players[i].Id, "amount": 0, "auto": true})
			}
		}
	}
	r.startFinalRoundQuestion()
}

func (r *Room) startFinalRoundQuestion() {
	finalRoundPlayers := make([]string, 0)
	for _, p := range r.Players {
		if slices.Contains(r.FinalRoundState.Players, p.Id) && *p.BetAmount > 0 {
			finalRoundPlayers = append(finalRoundPlayers, p.Id)
		}
	}

	if len(finalRoundPlayers) == 0 {
		r.EndGame()
		return
	}

	r.FinalRoundState.Players = finalRoundPlayers
	r.AllowedToAnswer = finalRoundPlayers
	r.emit(EventFinalQuestionStarted, SYSTEM, map[string]any{"players": finalRoundPlayers})
	r.FinalRoundState.PlayersAnswers = make(map[string]string)

	duration := time.Duration(r.Options.QuestionThinkingTimeFinal) * time.Second
	timerEndsAt := time.Now().Add(duration)
	r.FinalRoundState.TimerEndsAt = &timerEndsAt
	r.State = ShowingFinalRoundQuestion
}

func (r *Room) SubmitFinalRoundAnswer(userId string, answer string) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != ShowingFinalRoundQuestion {
		return custerr.NewConflictErr("can not submit final round answer now")
	}
	if !slices.Contains(r.AllowedToAnswer, userId) {
		return custerr.NewConflictErr("not allowed to submit final round answer")
	}

	r.FinalRoundState.PlayersAnswers[userId] = answer
	r.emit(EventFinalAnswerSubmitted, userId, map[string]any{"player": userId, "answer": answer})
	r.AllowedToAnswer = slices.DeleteFunc(r.AllowedToAnswer, func(playerId string) bool {
		return userId == playerId
	})

	if len(r.AllowedToAnswer) == 0 {
		r.EndFinalRoundQuestion()
	}

	return nil
}

func (r *Room) EndFinalRoundQuestion() {
	r.CurrentPlayer = &r.FinalRoundState.Players[0]
	r.State = ValidatingFinalRoundAnswers
}

func (r *Room) ValidateFinalRoundAnswer(userId string, isCorrect bool) error {
	return r.validateFinalRoundAnswer(userId, isCorrect, nil)
}

// ValidateFinalRoundAnswerSystem applies an AI host verdict.
func (r *Room) ValidateFinalRoundAnswerSystem(isCorrect bool, verdict Verdict) error {
	return r.validateFinalRoundAnswer(SYSTEM, isCorrect, &verdict)
}

func (r *Room) validateFinalRoundAnswer(userId string, isCorrect bool, verdict *Verdict) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if r.State != ValidatingFinalRoundAnswers {
		return custerr.NewConflictErr("can not validate final round answer now")
	}
	if r.Options.AIHost && userId != SYSTEM {
		return custerr.NewForbiddenErr("answer validation is handled by AI in this room")
	}
	if !r.IsUserModerator(userId) && userId != SYSTEM {
		return custerr.NewForbiddenErr("not allowed to validate final round answer")
	}

	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return p.Id == *r.CurrentPlayer
	})
	delta := *r.Players[playerIndex].BetAmount
	if !isCorrect {
		delta = -delta
	}
	r.Players[playerIndex].Score += delta

	answer := r.FinalRoundState.PlayersAnswers[*r.CurrentPlayer]
	validated := validationSource(userId, verdict, answer)
	validated["player"] = *r.CurrentPlayer
	validated["answer"] = answer
	validated["correct"] = isCorrect
	validated["delta"] = delta
	validated["score"] = r.Players[playerIndex].Score
	r.emit(EventFinalAnswerValidated, userId, validated)

	playerIndex = slices.IndexFunc(r.FinalRoundState.Players, func(p string) bool {
		return p == *r.CurrentPlayer
	})
	if playerIndex < len(r.FinalRoundState.Players)-1 {
		r.CurrentPlayer = &r.FinalRoundState.Players[playerIndex+1]
	} else {
		r.EndGame()
	}

	return nil
}

func (r *Room) SkipQuestion(userId string) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	skippable := r.State == RevealingQuestion || r.State == ShowingQuestion ||
		r.State == Answering || r.State == Passing || r.State == Betting
	if !skippable {
		return custerr.NewConflictErr("can not skip question now")
	}
	if !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to skip question")
	}
	r.EndQuestion(QuestionSkipped, userId)
	return nil
}

func (r *Room) SkipRound(userId string) error {
	if r.PausedState.Paused {
		return custerr.NewConflictErr("game is paused")
	}
	if !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to skip round")
	}
	if r.State != SelectingQuestion {
		return custerr.NewConflictErr("can not skip round now")
	}
	r.emit(EventRoundSkipped, userId, map[string]any{"round": r.CurrentRoundName})
	return nil
}

func (r *Room) ChangeScore(userId string, playerId string, score int) error {
	if !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to change score")
	}
	playerIndex := slices.IndexFunc(r.Players, func(p Player) bool {
		return p.Id == playerId
	})
	if playerIndex == -1 {
		return custerr.NewNotFoundErr(fmt.Sprintf("no player with id \"%s\" in room", playerId))
	}
	r.emit(EventScoreChanged, userId, map[string]any{"player": playerId, "from": r.Players[playerIndex].Score, "to": score})
	r.Players[playerIndex].Score = score
	return nil
}

func (r *Room) BanPlayer(callerUserId, targetUserId string) error {
	if !r.IsUserModerator(callerUserId) {
		return custerr.NewForbiddenErr("not allowed to ban players")
	}
	if callerUserId == targetUserId {
		return custerr.NewBadRequestErr("cannot ban yourself")
	}
	if r.IsUserBanned(targetUserId) {
		return custerr.NewConflictErr("player is already banned")
	}
	r.emit(EventPlayerBanned, callerUserId, map[string]any{"player": targetUserId, "name": r.playerName(targetUserId)})
	r.BanList = append(r.BanList, targetUserId)
	r.Players = slices.DeleteFunc(r.Players, func(p Player) bool {
		return p.Id == targetUserId
	})
	return nil
}

// captureTimerProgress saves timer and animation progress before pausing so clients can resume where they left off.
func (r *Room) captureTimerProgress(now time.Time) {
	switch r.State {
	case RevealingQuestion:
		if r.CurrentQuestion.Attachment != nil {
			fullDuration := float64(r.CurrentQuestion.GetMediaRevealingDuration())
			remainedDuration := math.Max(0, float64(r.CurrentQuestion.AttachmentRevealEndsAt.Sub(now)))
			r.CurrentQuestion.AttachmentRevealLastProgress = 1 - remainedDuration/fullDuration
		}
		if r.CurrentQuestion.Text != nil {
			fullDuration := float64(r.CurrentQuestion.GetTextRevealingDuration(r.Options.ReadingSymbolsPerSecond))
			remainedDuration := math.Min(fullDuration, float64(r.CurrentQuestion.TimerStartsAt.Sub(now)))
			r.CurrentQuestion.TextRevealLastProgress = 1 - remainedDuration/fullDuration
		}
	case ShowingQuestion:
		fullDuration := float64(time.Duration(r.Options.QuestionThinkingTime) * time.Second)
		remainedDuration := float64(r.CurrentQuestion.TimerEndsAt.Sub(now))
		r.CurrentQuestion.TimerLastProgress = remainedDuration / fullDuration
	}
}

// shiftTimers moves active deadlines forward by the pause duration, preserving the remaining time.
func (r *Room) shiftTimers(elapsed time.Duration) {
	switch r.State {
	case RevealingQuestion:
		if r.CurrentQuestion.Attachment != nil {
			r.CurrentQuestion.AttachmentRevealEndsAt = r.CurrentQuestion.AttachmentRevealEndsAt.Add(elapsed)
		}
		r.CurrentQuestion.TimerStartsAt = r.CurrentQuestion.TimerStartsAt.Add(elapsed)
	case ShowingQuestion:
		r.CurrentQuestion.TimerEndsAt = r.CurrentQuestion.TimerEndsAt.Add(elapsed)
	case Answering:
		r.AnsweringPlayer.TimerStartsAt = r.AnsweringPlayer.TimerStartsAt.Add(elapsed)
		r.AnsweringPlayer.TimerEndsAt = r.AnsweringPlayer.TimerEndsAt.Add(elapsed)
	case Betting:
		r.CurrentQuestion.BettingEndsAt = r.CurrentQuestion.BettingEndsAt.Add(elapsed)
	case Passing:
		r.CurrentQuestion.PassingEndsAt = r.CurrentQuestion.PassingEndsAt.Add(elapsed)
	case FinalRoundBetting:
		newBettingEndsAt := r.FinalRoundState.BettingEndsAt.Add(elapsed)
		r.FinalRoundState.BettingEndsAt = &newBettingEndsAt
	case ShowingFinalRoundQuestion:
		newTimerEndsAt := r.FinalRoundState.TimerEndsAt.Add(elapsed)
		r.FinalRoundState.TimerEndsAt = &newTimerEndsAt
	}
}

func (r *Room) Pause(userId string) error {
	now := time.Now()

	if !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to pause")
	}
	if r.PausedState.Paused {
		return custerr.NewConflictErr("already paused")
	}
	if !slices.Contains(
		[]RoomState{
			SelectingQuestion, RevealingQuestion, ShowingQuestion,
			Answering, Betting, Passing, SelectingFinalRoundCategory,
			FinalRoundBetting, ShowingFinalRoundQuestion,
		},
		r.State,
	) {
		return custerr.NewConflictErr("can not pause now")
	}

	r.captureTimerProgress(now)

	r.PausedState.Paused = true
	r.PausedState.PausedAt = &now
	r.emit(EventPaused, userId, nil)
	return nil
}

func (r *Room) Unpause(userId string) error {
	if !r.IsUserModerator(userId) {
		return custerr.NewForbiddenErr("not allowed to unpause")
	}
	if !r.PausedState.Paused {
		return custerr.NewConflictErr("not paused")
	}
	elapsed := time.Since(*r.PausedState.PausedAt)
	r.PausedState.Paused = false
	r.PausedState.PausedAt = nil

	r.shiftTimers(elapsed)
	r.emit(EventUnpaused, userId, map[string]any{"auto": false})
	return nil
}

func (r *Room) UnpauseSystem(pausedAt time.Time) error {
	if !r.PausedState.Paused || !r.PausedState.PausedAt.Equal(pausedAt) {
		return custerr.NewConflictErr("pause session changed")
	}
	elapsed := time.Since(*r.PausedState.PausedAt)
	r.PausedState.Paused = false
	r.PausedState.PausedAt = nil

	r.shiftTimers(elapsed)
	r.emit(EventUnpaused, SYSTEM, map[string]any{"auto": true})
	return nil
}

func (r *Room) EndGame() {
	r.CurrentPlayer = nil
	if r.FinalRoundState != nil && r.FinalRoundState.TimerEndsAt != nil {
		r.FinalRoundState.TimerEndsAt = nil
	}
	for pi := range r.Players {
		newBetAmount := 0
		r.Players[pi].BetAmount = &newBetAmount
	}
	r.State = GameOver
	r.emit(EventGameEnded, SYSTEM, map[string]any{"standings": r.standings()})
}
