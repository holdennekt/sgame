package domain

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func eventTypes(events []GameEvent) []GameEventType {
	types := make([]GameEventType, len(events))
	for i, e := range events {
		types[i] = e.Type
	}
	return types
}

func TestScriptedGame_EmitsOrderedEvents(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.State = WaitingForStart
		r.PackPreview = PackPreview{Id: pack.Id, Name: pack.Name}
		for i := range r.Players {
			r.Players[i].Score = 0
		}
		r.Players[0].Score = 500
		r.Players[1].Score = 500
	})
	selectQ := func(category string, index int) {
		require.NoError(t, r.SelectQuestion("host1", pack, category, index, noopAttachmentUrl))
	}

	r.StartGame(pack)

	selectQ("Geography", 0)
	require.NoError(t, r.StartAnswer("p1"))
	require.NoError(t, r.ValidateAnswer("host1", true))

	selectQ("Geography", 1)
	require.NoError(t, r.PlaceBet("p1", 100))
	require.NoError(t, r.PlaceBet("p2", 50))
	require.NoError(t, r.ValidateAnswer("host1", false))

	selectQ("Geography", 2)
	passTo := "p1"
	if *r.CurrentPlayer == "p1" {
		passTo = "p2"
	}
	require.NoError(t, r.PassQuestion("host1", passTo))
	require.NoError(t, r.ValidateAnswer("host1", true))

	selectQ("Science", 0)
	require.NoError(t, r.SkipQuestion("host1"))

	require.False(t, r.AnyAvailableQuestions())
	require.True(t, r.StartNextRegularRound(pack))
	selectQ("History", 0)
	require.NoError(t, r.StartAnswer("p2"))
	require.NoError(t, r.ValidateAnswer("host1", true))

	require.False(t, r.StartNextRegularRound(pack))
	started, err := r.StartFinalRound(pack, noopAttachmentUrl)
	require.NoError(t, err)
	require.True(t, started)
	require.NoError(t, r.RemoveFinalRoundCategory(pack, "host1", "Final Cat A", noopAttachmentUrl))
	require.NoError(t, r.PlaceFinalRoundBet("p1", 100))
	require.NoError(t, r.PlaceFinalRoundBet("p2", 200))
	r.AllowedToAnswer = slices.Clone(r.AllowedToAnswer) // production reloads the room on each update, which un-aliases this slice
	require.NoError(t, r.SubmitFinalRoundAnswer("p1", "foo"))
	require.NoError(t, r.SubmitFinalRoundAnswer("p2", "bar"))
	require.NoError(t, r.ValidateFinalRoundAnswer("host1", true))
	require.NoError(t, r.ValidateFinalRoundAnswer("host1", false))
	require.Equal(t, GameOver, r.State)

	events := r.DrainEvents()
	want := []GameEventType{
		EventGameStarted, EventRoundStarted,
		EventQuestionSelected, EventBuzzed, EventAnswerValidated, EventQuestionEnded,
		EventQuestionSelected, EventBetPlaced, EventBetPlaced, EventAuctionWon, EventAnswerValidated, EventQuestionEnded,
		EventQuestionSelected, EventPassed, EventAnswerValidated, EventQuestionEnded,
		EventQuestionSelected, EventQuestionEnded,
		EventRoundStarted,
		EventQuestionSelected, EventBuzzed, EventAnswerValidated, EventQuestionEnded,
		EventFinalRoundStarted, EventFinalCategoryRemoved, EventFinalCategoryChosen,
		EventFinalBetPlaced, EventFinalBetPlaced, EventFinalQuestionStarted,
		EventFinalAnswerSubmitted, EventFinalAnswerSubmitted,
		EventFinalAnswerValidated, EventFinalAnswerValidated, EventGameEnded,
	}
	assert.Equal(t, want, eventTypes(events))

	for i, e := range events {
		assert.Equal(t, i+1, e.Seq)
		assert.False(t, e.At.IsZero())
	}
	assert.Equal(t, len(events), r.EventSeq)
	assert.Empty(t, r.DrainEvents())
}

func TestEvents_Emitted(t *testing.T) {
	tests := []struct {
		name      string
		room      func() Room
		act       func(r *Room) error
		wantTypes []GameEventType
		check     func(t *testing.T, events []GameEvent)
	}{
		{
			name:      "buzz during showing records phase",
			room:      func() Room { return buildRoom(withShowingQuestion(100)) },
			act:       func(r *Room) error { return r.StartAnswer("p1") },
			wantTypes: []GameEventType{EventBuzzed},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, "showing", events[0].Data["phase"])
				assert.Equal(t, "p1", events[0].Actor)
			},
		},
		{
			name:      "wrong answer by moderator carries negative delta and source",
			room:      func() Room { return buildRoom(withAnsweringP1(200)) },
			act:       func(r *Room) error { return r.ValidateAnswer("host1", false) },
			wantTypes: []GameEventType{EventAnswerValidated},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, -200, events[0].Data["delta"])
				assert.Equal(t, 800, events[0].Data["score"])
				assert.Equal(t, "moderator", events[0].Data["source"])
			},
		},
		{
			name:      "system validation without verdict is a timeout",
			room:      func() Room { return buildRoom(withAnsweringP1(200)) },
			act:       func(r *Room) error { return r.ValidateAnswer(SYSTEM, false) },
			wantTypes: []GameEventType{EventAnswerValidated},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, "timeout", events[0].Data["source"])
			},
		},
		{
			name: "ai verdict is logged with method and confidence",
			room: func() Room { return buildRoom(withAnsweringP1(200)) },
			act: func(r *Room) error {
				return r.ValidateAnswerSystem(true, Verdict{Method: VerdictAI, Confidence: 0.8})
			},
			wantTypes: []GameEventType{EventAnswerValidated, EventQuestionEnded},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, "ai", events[0].Data["source"])
				assert.Equal(t, "ai", events[0].Data["method"])
				assert.Equal(t, 0.8, events[0].Data["confidence"])
				assert.Equal(t, "answered", events[1].Data["reason"])
				assert.Equal(t, "p1", events[1].Data["winner"])
			},
		},
		{
			name:      "score change logs old and new value",
			room:      func() Room { return buildRoom() },
			act:       func(r *Room) error { return r.ChangeScore("host1", "p1", 42) },
			wantTypes: []GameEventType{EventScoreChanged},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, 1000, events[0].Data["from"])
				assert.Equal(t, 42, events[0].Data["to"])
			},
		},
		{
			name: "ban snapshots the player's name",
			room: func() Room {
				return buildRoom(func(r *Room) { r.Players[0].Name = "Alice" })
			},
			act:       func(r *Room) error { return r.BanPlayer("host1", "p1") },
			wantTypes: []GameEventType{EventPlayerBanned},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, "Alice", events[0].Data["name"])
			},
		},
		{
			name: "pause and unpause",
			room: func() Room { return buildRoom() },
			act: func(r *Room) error {
				if err := r.Pause("host1"); err != nil {
					return err
				}
				return r.Unpause("host1")
			},
			wantTypes: []GameEventType{EventPaused, EventUnpaused},
		},
		{
			name: "auto bets are logged for players who did not bet",
			room: func() Room {
				return buildRoom(func(r *Room) {
					r.State = Betting
					r.CurrentQuestion = &CurrentQuestion{}
					r.Players[0].BetAmount = ptr(100)
				})
			},
			act:       func(r *Room) error { r.PlaceBetsAuto(); return nil },
			wantTypes: []GameEventType{EventBetPlaced, EventAuctionWon},
			check: func(t *testing.T, events []GameEvent) {
				assert.Equal(t, true, events[0].Data["auto"])
				assert.Equal(t, "p2", events[0].Data["player"])
			},
		},
		{
			name: "connection changes are logged only during a game",
			room: func() Room { return buildRoom() },
			act: func(r *Room) error {
				r.SetConnected("p1", false)
				r.SetConnected("p1", false)
				r.SetConnected("p1", true)
				return nil
			},
			wantTypes: []GameEventType{EventPlayerDisconnected, EventPlayerReconnected},
		},
		{
			name:      "connection changes in the lobby are not logged",
			room:      func() Room { return buildRoom(func(r *Room) { r.State = WaitingForStart }) },
			act:       func(r *Room) error { r.SetConnected("p1", false); return nil },
			wantTypes: []GameEventType{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.room()
			require.NoError(t, tt.act(&r))
			events := r.DrainEvents()
			assert.Equal(t, tt.wantTypes, eventTypes(events))
			if tt.check != nil {
				tt.check(t, events)
			}
		})
	}
}

func TestEvents_FailedActionsEmitNothing(t *testing.T) {
	tests := []struct {
		name string
		room func() Room
		act  func(r *Room) error
	}{
		{"validate by player", func() Room { return buildRoom(withAnsweringP1(100)) }, func(r *Room) error { return r.ValidateAnswer("p1", true) }},
		{"buzz when not allowed", func() Room { return buildRoom(withShowingQuestion(100)) }, func(r *Room) error { return r.StartAnswer("nobody") }},
		{"bet in wrong state", func() Room { return buildRoom() }, func(r *Room) error { return r.PlaceBet("p1", 10) }},
		{"score change by player", func() Room { return buildRoom() }, func(r *Room) error { return r.ChangeScore("p1", "p2", 1) }},
		{"ban self", func() Room { return buildRoom() }, func(r *Room) error { return r.BanPlayer("host1", "host1") }},
		{"skip round by player", func() Room { return buildRoom() }, func(r *Room) error { return r.SkipRound("p1") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.room()
			assert.Error(t, tt.act(&r))
			assert.Empty(t, r.DrainEvents())
			assert.Zero(t, r.EventSeq)
		})
	}
}

func TestEmit_CapsLogWithTruncationMarker(t *testing.T) {
	r := buildRoom(func(r *Room) { r.EventSeq = MaxGameEvents - 2 })

	r.emit(EventPaused, "host1", nil)
	r.emit(EventUnpaused, "host1", nil)
	r.emit(EventPaused, "host1", nil)

	events := r.DrainEvents()
	assert.Equal(t, []GameEventType{EventPaused, EventLogTruncated}, eventTypes(events))
	assert.Equal(t, MaxGameEvents, r.EventSeq)
	assert.Equal(t, MaxGameEvents, lastEventSeq(events))
}

func lastEventSeq(events []GameEvent) int { return events[len(events)-1].Seq }
