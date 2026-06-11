package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/stretchr/testify/assert"
)

// ---- helpers ----

func ptr[T any](v T) *T { return &v }

func noopAttachmentUrl(key string) (string, error) { return "http://cdn/" + key, nil }

func buildRoom(opts ...func(*Room)) Room {
	r := Room{
		Id:        "room1",
		CreatedBy: "host1",
		Host:      &Host{User: User{Id: "host1"}, IsConnected: true},
		Players: []Player{
			{User: User{Id: "p1"}, Score: 1000, IsConnected: true},
			{User: User{Id: "p2"}, Score: 1000, IsConnected: true},
		},
		Options: RoomOptions{
			ReadingSymbolsPerSecond:   20,
			QuestionThinkingTime:      10,
			AnswerThinkingTime:        10,
			QuestionThinkingTimeFinal: 60,
		},
		State:           SelectingQuestion,
		AllowedToAnswer: []string{},
	}
	for _, o := range opts {
		o(&r)
	}
	return r
}

func buildPack() *Pack {
	text := "Sample question text"
	return &Pack{
		Id:   "pack1",
		Name: "Test Pack",
		Rounds: []Round{
			{
				Name: "Round 1",
				Categories: []Category{
					{
						Name: "Geography",
						Questions: []Question{
							{HiddenQuestion: HiddenQuestion{Round: "Round 1", Category: "Geography", Index: 0, Value: 100}, Type: Regular, Text: &text, Answers: []string{"Paris"}},
							{HiddenQuestion: HiddenQuestion{Round: "Round 1", Category: "Geography", Index: 1, Value: 200}, Type: Auction, Text: &text, Answers: []string{"Berlin"}},
							{HiddenQuestion: HiddenQuestion{Round: "Round 1", Category: "Geography", Index: 2, Value: 300}, Type: CatInBag, Text: &text, Answers: []string{"Rome"}},
						},
					},
					{
						Name: "Science",
						Questions: []Question{
							{HiddenQuestion: HiddenQuestion{Round: "Round 1", Category: "Science", Index: 0, Value: 100}, Type: Regular, Text: &text, Answers: []string{"Newton"}},
						},
					},
				},
			},
			{
				Name: "Round 2",
				Categories: []Category{
					{
						Name: "History",
						Questions: []Question{
							{HiddenQuestion: HiddenQuestion{Round: "Round 2", Category: "History", Index: 0, Value: 200}, Type: Regular, Text: &text, Answers: []string{"Caesar"}},
						},
					},
				},
			},
		},
		FinalRound: FinalRound{
			Categories: []FinalRoundCategory{
				{
					HiddenFinalRoundCategory: HiddenFinalRoundCategory{Name: "Final Cat A"},
					Question: FinalRoundQuestion{
						HiddenFinalRoundQuestion: HiddenFinalRoundQuestion{Category: "Final Cat A", Text: &text},
						Answers:                  []string{"Answer A"},
					},
				},
				{
					HiddenFinalRoundCategory: HiddenFinalRoundCategory{Name: "Final Cat B"},
					Question: FinalRoundQuestion{
						HiddenFinalRoundQuestion: HiddenFinalRoundQuestion{Category: "Final Cat B", Text: &text},
						Answers:                  []string{"Answer B"},
					},
				},
			},
		},
	}
}

func buildPackSingleFinalCat() *Pack {
	p := buildPack()
	p.FinalRound.Categories = p.FinalRound.Categories[:1]
	return p
}

func withRound1(pack *Pack) func(*Room) {
	return func(r *Room) {
		r.State = WaitingForStart
		r.StartNextRegularRound(pack)
		r.CurrentPlayer = ptr("p1")
	}
}

func withShowingQuestion(value int) func(*Room) {
	return func(r *Room) {
		r.State = ShowingQuestion
		r.CurrentQuestion = &CurrentQuestion{
			Question:    Question{HiddenQuestion: HiddenQuestion{Value: value}},
			TimerEndsAt: time.Now().Add(5 * time.Second),
		}
		r.AllowedToAnswer = []string{"p1", "p2"}
	}
}

func withAnsweringP1(value int) func(*Room) {
	return func(r *Room) {
		r.State = Answering
		r.CurrentQuestion = &CurrentQuestion{
			Question: Question{HiddenQuestion: HiddenQuestion{Value: value}},
		}
		r.AnsweringPlayer = &AnsweringPlayer{
			Id:            "p1",
			TimerStartsAt: time.Now().Add(-2 * time.Second),
			TimerEndsAt:   time.Now().Add(8 * time.Second),
		}
		r.AllowedToAnswer = []string{"p2"}
	}
}

func withPaused(state RoomState) func(*Room) {
	return func(r *Room) {
		r.State = state
		r.PausedState = PausedState{
			Paused:   true,
			PausedAt: ptr(time.Now().Add(-100 * time.Millisecond)),
		}
	}
}

// ---- 1. SubmitAnswer ----

func TestSubmitAnswer_WrongState_ReturnsError(t *testing.T) {
	r := buildRoom(func(r *Room) { r.State = Betting })
	err := r.SubmitAnswer("p1")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
	assert.Equal(t, Betting, r.State)
}

func TestSubmitAnswer_WrongState_Answering(t *testing.T) {
	r := buildRoom(func(r *Room) { r.State = Answering })
	err := r.SubmitAnswer("p1")
	assert.Error(t, err)
}

func TestSubmitAnswer_PlayerNotInAllowed_ReturnsError(t *testing.T) {
	r := buildRoom(withShowingQuestion(200), func(r *Room) {
		r.AllowedToAnswer = []string{"p2"}
	})
	err := r.SubmitAnswer("p1")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
	assert.Equal(t, ShowingQuestion, r.State)
}

func TestSubmitAnswer_FalseStartNotAllowed_DuringReveal(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.Options.FalseStartAllowed = false
		r.CurrentQuestion = &CurrentQuestion{}
		r.AllowedToAnswer = []string{"p1"}
	})
	err := r.SubmitAnswer("p1")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestSubmitAnswer_FalseStartAllowed_DuringReveal(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.Options.FalseStartAllowed = true
		r.CurrentQuestion = &CurrentQuestion{
			Question:      Question{HiddenQuestion: HiddenQuestion{Value: 200}},
			TimerStartsAt: time.Now().Add(3 * time.Second),
		}
		r.AllowedToAnswer = []string{"p1"}
	})
	err := r.SubmitAnswer("p1")
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
}

func TestSubmitAnswer_DuringShowingQuestion_SetsAnsweringState(t *testing.T) {
	r := buildRoom(withShowingQuestion(200))
	err := r.SubmitAnswer("p1")
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p1", r.AnsweringPlayer.Id)
	assert.Equal(t, "p1", *r.CurrentPlayer)
	assert.Equal(t, []string{"p2"}, r.AllowedToAnswer)
	assert.True(t, r.CurrentQuestion.TimerLastProgress >= 0 && r.CurrentQuestion.TimerLastProgress <= 1)
	assert.WithinDuration(t, time.Now().Add(time.Duration(r.Options.AnswerThinkingTime)*time.Second), r.AnsweringPlayer.TimerEndsAt, 200*time.Millisecond)
}

func TestSubmitAnswer_DuringRevealingQuestion_SnapshotsAttachmentProgress(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.Options.FalseStartAllowed = true
		r.CurrentQuestion = &CurrentQuestion{
			Question: Question{
				HiddenQuestion: HiddenQuestion{Value: 200},
				Attachment:     &Attachment{Duration: 2.0},
			},
			AttachmentRevealEndsAt: time.Now().Add(1 * time.Second),
			TimerStartsAt:          time.Now().Add(4 * time.Second),
		}
		r.AllowedToAnswer = []string{"p1"}
	})
	err := r.SubmitAnswer("p1")
	assert.NoError(t, err)
	assert.True(t, r.CurrentQuestion.AttachmentRevealLastProgress >= 0 && r.CurrentQuestion.AttachmentRevealLastProgress <= 1)
}

func TestSubmitAnswer_NilTextAndAttachment_DoesNotPanic(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.Options.FalseStartAllowed = true
		r.CurrentQuestion = &CurrentQuestion{
			Question:      Question{HiddenQuestion: HiddenQuestion{Value: 100}},
			TimerStartsAt: time.Now().Add(3 * time.Second),
		}
		r.AllowedToAnswer = []string{"p1"}
	})
	assert.NotPanics(t, func() { _ = r.SubmitAnswer("p1") })
}

// ---- 2. ValidateAnswer ----

func TestValidateAnswer_WrongState(t *testing.T) {
	r := buildRoom()
	err := r.ValidateAnswer("host1", true)
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestValidateAnswer_CallerIsPlayer(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	err := r.ValidateAnswer("p1", true)
	var fe custerr.ForbiddenErr
	assert.ErrorAs(t, err, &fe)
}

func TestValidateAnswer_CorrectAnswer_IncreasesScore(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	r.Players[0].Score = 100
	err := r.ValidateAnswer("host1", true)
	assert.NoError(t, err)
	assert.Equal(t, 300, r.Players[0].Score)
	assert.Equal(t, SelectingQuestion, r.State)
}

func TestValidateAnswer_WrongAnswer_DecreasesScore(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	r.Players[0].Score = 100
	r.CurrentQuestion.TimerStartsAt = time.Now().Add(5 * time.Second)
	err := r.ValidateAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, -100, r.Players[0].Score)
	// question continues — p2 still allowed to answer
	assert.NotEqual(t, SelectingQuestion, r.State)
}

func TestValidateAnswer_WrongAnswer_NobodyLeft_EndsQuestion(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	r.AllowedToAnswer = []string{} // nobody left
	err := r.ValidateAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, SelectingQuestion, r.State)
}

func TestValidateAnswer_UsesBetAmountNotQuestionValue(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	r.Players[0].Score = 0
	r.Players[0].BetAmount = ptr(500)
	err := r.ValidateAnswer("host1", true)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.Players[0].Score)
}

func TestValidateAnswer_SystemCanValidate(t *testing.T) {
	r := buildRoom(withAnsweringP1(200))
	err := r.ValidateAnswer(SYSTEM, true)
	assert.NoError(t, err)
}

// ---- 3. SelectQuestion ----

func TestSelectQuestion_Guards(t *testing.T) {
	pack := buildPack()

	tests := []struct {
		name     string
		setup    func(*Room)
		userId   string
		category string
		index    int
		errType  any
	}{
		{
			name:     "WrongState",
			setup:    func(r *Room) { r.State = Answering; r.CurrentPlayer = ptr("p1") },
			userId:   "p1",
			category: "Geography",
			index:    0,
			errType:  &custerr.ConflictErr{},
		},
		{
			name: "NonCurrentPlayer",
			setup: func(r *Room) {
				r.State = WaitingForStart
				r.StartNextRegularRound(pack)
				r.CurrentPlayer = ptr("p2")
			},
			userId:   "p1",
			category: "Geography",
			index:    0,
			errType:  &custerr.ForbiddenErr{},
		},
		{
			name: "AlreadyPlayed",
			setup: func(r *Room) {
				r.State = WaitingForStart
				r.StartNextRegularRound(pack)
				r.CurrentPlayer = ptr("p1")
				r.CurrentRoundQuestions[0].Questions[0].HasBeenPlayed = true
			},
			userId:   "p1",
			category: "Geography",
			index:    0,
			errType:  &custerr.ConflictErr{},
		},
		{
			name: "UnknownCategory",
			setup: func(r *Room) {
				r.State = WaitingForStart
				r.StartNextRegularRound(pack)
				r.CurrentPlayer = ptr("p1")
			},
			userId:   "p1",
			category: "DoesNotExist",
			index:    0,
			errType:  &custerr.ConflictErr{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.SelectQuestion(tc.userId, pack, tc.category, tc.index, noopAttachmentUrl)
			assert.True(t, errors.As(err, tc.errType), "expected error type %T, got %v", tc.errType, err)
		})
	}
}

func TestSelectQuestion_HostCanAlwaysSelect(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack))
	r.CurrentPlayer = ptr("p2")
	err := r.SelectQuestion("host1", pack, "Geography", 0, noopAttachmentUrl)
	assert.NoError(t, err)
}

func TestSelectQuestion_Regular_StartsReveal(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack))
	err := r.SelectQuestion("p1", pack, "Geography", 0, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, RevealingQuestion, r.State)
	assert.Contains(t, r.AllowedToAnswer, "p1")
	assert.Contains(t, r.AllowedToAnswer, "p2")
	assert.True(t, r.CurrentRoundQuestions[0].Questions[0].HasBeenPlayed)
}

func TestSelectQuestion_Auction_WithEligiblePlayers_StartsBetting(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack))
	// both players have Score > 0
	err := r.SelectQuestion("p1", pack, "Geography", 1, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, Betting, r.State)
	assert.True(t, r.CurrentQuestion.BettingEndsAt.After(time.Now()))
}

func TestSelectQuestion_Auction_NoEligiblePlayers_SkipsBetting(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack), func(r *Room) {
		r.Players[0].Score = 0
		r.Players[1].Score = 0
	})
	err := r.SelectQuestion("p1", pack, "Geography", 1, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, RevealingQuestion, r.State)
}

func TestSelectQuestion_CatInBag_TwoEligibleRecipients_StartsPassingState(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack), func(r *Room) {
		r.Players = append(r.Players, Player{User: User{Id: "p3"}, Score: 500, IsConnected: true})
	})
	err := r.SelectQuestion("p1", pack, "Geography", 2, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, Passing, r.State)
	assert.True(t, r.CurrentQuestion.PassingEndsAt.After(time.Now()))
}

func TestSelectQuestion_CatInBag_OneEligibleRecipient_AutoPasses(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack))
	// p1 is current player, p2 is only other connected → auto-pass to p2
	err := r.SelectQuestion("p1", pack, "Geography", 2, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p2", r.AnsweringPlayer.Id)
}

func TestSelectQuestion_CatInBag_NoEligibleRecipient_AssignsToSelf(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack), func(r *Room) {
		r.Players[1].IsConnected = false
	})
	err := r.SelectQuestion("p1", pack, "Geography", 2, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p1", r.AnsweringPlayer.Id)
}

// ---- 4. Pause ----

func TestPause_Guards(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Room)
		userId  string
		errType any
	}{
		{
			name:    "NotHost",
			setup:   func(r *Room) {},
			userId:  "p1",
			errType: &custerr.ForbiddenErr{},
		},
		{
			name:    "AlreadyPaused",
			setup:   func(r *Room) { r.PausedState = PausedState{Paused: true, PausedAt: ptr(time.Now())} },
			userId:  "host1",
			errType: &custerr.ConflictErr{},
		},
		{
			name:    "GameOver",
			setup:   func(r *Room) { r.State = GameOver },
			userId:  "host1",
			errType: &custerr.ConflictErr{},
		},
		{
			name:    "WaitingForStart",
			setup:   func(r *Room) { r.State = WaitingForStart },
			userId:  "host1",
			errType: &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.Pause(tc.userId)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestPause_SelectingQuestion_SetsPausedFlag(t *testing.T) {
	r := buildRoom()
	err := r.Pause("host1")
	assert.NoError(t, err)
	assert.True(t, r.PausedState.Paused)
	assert.NotNil(t, r.PausedState.PausedAt)
}

func TestPause_RevealingQuestion_CapturesAttachmentProgress(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.CurrentQuestion = &CurrentQuestion{
			Question:               Question{Attachment: &Attachment{Duration: 4.0}},
			AttachmentRevealEndsAt: time.Now().Add(2 * time.Second),
			TimerStartsAt:          time.Now().Add(6 * time.Second),
		}
	})
	err := r.Pause("host1")
	assert.NoError(t, err)
	assert.InDelta(t, 0.5, r.CurrentQuestion.AttachmentRevealLastProgress, 0.1)
}

func TestPause_ShowingQuestion_CapturesTimerProgress(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = ShowingQuestion
		r.CurrentQuestion = &CurrentQuestion{
			TimerEndsAt: time.Now().Add(5 * time.Second),
		}
	})
	err := r.Pause("host1")
	assert.NoError(t, err)
	assert.InDelta(t, 0.5, r.CurrentQuestion.TimerLastProgress, 0.1)
}

// ---- 5. Unpause ----

func TestUnpause_Guards(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Room)
		userId  string
		errType any
	}{
		{
			name:    "NotHost",
			setup:   withPaused(SelectingQuestion),
			userId:  "p1",
			errType: &custerr.ForbiddenErr{},
		},
		{
			name:    "NotPaused",
			setup:   func(r *Room) {},
			userId:  "host1",
			errType: &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.Unpause(tc.userId)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestUnpause_RevealingQuestion_ShiftsTimers(t *testing.T) {
	origAttach := time.Now().Add(3 * time.Second)
	origTimer := time.Now().Add(5 * time.Second)
	r := buildRoom(withPaused(RevealingQuestion), func(r *Room) {
		r.CurrentQuestion = &CurrentQuestion{
			Question:               Question{Attachment: &Attachment{Duration: 2.0}},
			AttachmentRevealEndsAt: origAttach,
			TimerStartsAt:          origTimer,
		}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.CurrentQuestion.AttachmentRevealEndsAt.After(origAttach))
	assert.True(t, r.CurrentQuestion.TimerStartsAt.After(origTimer))
}

func TestUnpause_ShowingQuestion_ShiftsTimerEndsAt(t *testing.T) {
	orig := time.Now().Add(5 * time.Second)
	r := buildRoom(withPaused(ShowingQuestion), func(r *Room) {
		r.CurrentQuestion = &CurrentQuestion{TimerEndsAt: orig}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.CurrentQuestion.TimerEndsAt.After(orig))
}

func TestUnpause_Answering_ShiftsAnsweringPlayerTimers(t *testing.T) {
	origStart := time.Now().Add(-1 * time.Second)
	origEnd := time.Now().Add(9 * time.Second)
	r := buildRoom(withPaused(Answering), func(r *Room) {
		r.CurrentQuestion = &CurrentQuestion{}
		r.AnsweringPlayer = &AnsweringPlayer{
			Id:            "p1",
			TimerStartsAt: origStart,
			TimerEndsAt:   origEnd,
		}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.AnsweringPlayer.TimerStartsAt.After(origStart))
	assert.True(t, r.AnsweringPlayer.TimerEndsAt.After(origEnd))
}

func TestUnpause_Betting_ShiftsBettingEndsAt(t *testing.T) {
	orig := time.Now().Add(30 * time.Second)
	r := buildRoom(withPaused(Betting), func(r *Room) {
		r.CurrentQuestion = &CurrentQuestion{BettingEndsAt: orig}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.CurrentQuestion.BettingEndsAt.After(orig))
}

func TestUnpause_Passing_ShiftsPassingEndsAt(t *testing.T) {
	orig := time.Now().Add(30 * time.Second)
	r := buildRoom(withPaused(Passing), func(r *Room) {
		r.CurrentQuestion = &CurrentQuestion{PassingEndsAt: orig}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.CurrentQuestion.PassingEndsAt.After(orig))
}

func TestUnpause_FinalRoundBetting_ShiftsBettingEndsAt(t *testing.T) {
	orig := time.Now().Add(30 * time.Second)
	r := buildRoom(withPaused(FinalRoundBetting), func(r *Room) {
		r.FinalRoundState = &FinalRoundState{BettingEndsAt: &orig}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.FinalRoundState.BettingEndsAt.After(orig))
}

func TestUnpause_ShowingFinalRoundQuestion_ShiftsTimerEndsAt(t *testing.T) {
	orig := time.Now().Add(30 * time.Second)
	r := buildRoom(withPaused(ShowingFinalRoundQuestion), func(r *Room) {
		r.FinalRoundState = &FinalRoundState{TimerEndsAt: &orig}
	})
	err := r.Unpause("host1")
	assert.NoError(t, err)
	assert.True(t, r.FinalRoundState.TimerEndsAt.After(orig))
}

func TestUnpause_ClearsPausedState(t *testing.T) {
	r := buildRoom(withPaused(SelectingQuestion))
	_ = r.Unpause("host1")
	assert.False(t, r.PausedState.Paused)
	assert.Nil(t, r.PausedState.PausedAt)
}

// ---- 6. UnpauseSystem ----

func TestUnpauseSystem_StaleTimestamp_ReturnsError(t *testing.T) {
	pausedAt := time.Now().Add(-100 * time.Millisecond)
	r := buildRoom(func(r *Room) {
		r.State = SelectingQuestion
		r.PausedState = PausedState{Paused: true, PausedAt: &pausedAt}
	})
	stale := pausedAt.Add(time.Second)
	err := r.UnpauseSystem(stale)
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestUnpauseSystem_MatchingTimestamp_Unpauses(t *testing.T) {
	pausedAt := time.Now().Add(-100 * time.Millisecond)
	orig := time.Now().Add(5 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = ShowingQuestion
		r.PausedState = PausedState{Paused: true, PausedAt: &pausedAt}
		r.CurrentQuestion = &CurrentQuestion{TimerEndsAt: orig}
	})
	err := r.UnpauseSystem(pausedAt)
	assert.NoError(t, err)
	assert.False(t, r.PausedState.Paused)
	assert.True(t, r.CurrentQuestion.TimerEndsAt.After(orig))
}

// ---- 7. PlaceBet ----

func TestPlaceBet_Guards(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Room)
		userId  string
		amount  int
		errType any
	}{
		{
			name:    "WrongState",
			setup:   func(r *Room) {},
			userId:  "p1",
			amount:  100,
			errType: &custerr.ConflictErr{},
		},
		{
			name: "AlreadyBet",
			setup: func(r *Room) {
				r.State = Betting
				r.CurrentQuestion = &CurrentQuestion{}
				r.Players[0].BetAmount = ptr(100)
			},
			userId:  "p1",
			amount:  200,
			errType: &custerr.ConflictErr{},
		},
		{
			name: "NegativeAmount",
			setup: func(r *Room) {
				r.State = Betting
				r.CurrentQuestion = &CurrentQuestion{}
			},
			userId:  "p1",
			amount:  -1,
			errType: &custerr.ConflictErr{},
		},
		{
			name: "ExceedsScore",
			setup: func(r *Room) {
				r.State = Betting
				r.CurrentQuestion = &CurrentQuestion{}
				r.Players[0].Score = 100
			},
			userId:  "p1",
			amount:  101,
			errType: &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.PlaceBet(tc.userId, tc.amount)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestPlaceBet_ValidBet_SetsBetAmount(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
	})
	err := r.PlaceBet("p1", 300)
	assert.NoError(t, err)
	assert.Equal(t, ptr(300), r.Players[0].BetAmount)
}

func TestPlaceBet_AllPlayersBet_StartsQuestion(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
	})
	_ = r.PlaceBet("p1", 300)
	_ = r.PlaceBet("p2", 500)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p2", r.AnsweringPlayer.Id) // p2 has higher bet
}

func TestPlaceBet_AllPlayersBetZero_EndsQuestion(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
	})
	_ = r.PlaceBet("p1", 0)
	_ = r.PlaceBet("p2", 0)
	assert.Equal(t, SelectingQuestion, r.State)
}

func TestPlaceBet_OnlyOnePlayerEligible_SingleBetTriggers(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
		r.Players[1].Score = 0 // p2 not eligible
	})
	err := r.PlaceBet("p1", 300)
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p1", r.AnsweringPlayer.Id)
}

// ---- 8. PlaceBetsAuto (bug proof) ----

// TestPlaceBetsAuto_SetsZeroBetOnRoomPlayers proves the bug fix.
// Setup: p1 has a positive bet already placed, p2 has nil bet (needs auto-zero).
// After PlaceBetsAuto, p1 wins (highest bet), state = Answering.
// EndQuestion is NOT called, so p2's BetAmount is still visible.
// Bug: old code wrote canBet[i].BetAmount (a copy), leaving r.Players[1].BetAmount nil.
// Fix: write directly to r.Players[i].BetAmount.
func TestPlaceBetsAuto_SetsZeroBetOnRoomPlayers(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
		r.Players[0].BetAmount = ptr(300) // p1 already bet positively
		// p2 BetAmount is nil → should be auto-zeroed on r.Players, not a local copy
	})
	r.PlaceBetsAuto()
	assert.Equal(t, Answering, r.State) // p1 wins, EndQuestion not called
	assert.NotNil(t, r.Players[1].BetAmount, "r.Players[1].BetAmount must be written on r.Players, not the canBet copy")
	assert.Equal(t, 0, *r.Players[1].BetAmount)
}

func TestPlaceBetsAuto_PlayerWithPositiveBet_GetsQuestion(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
		r.Players[0].BetAmount = ptr(300)
		// p2 gets auto-zero
	})
	r.PlaceBetsAuto()
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p1", r.AnsweringPlayer.Id)
}

func TestPlaceBetsAuto_AllZeroBets_EndsQuestion(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Betting
		r.CurrentQuestion = &CurrentQuestion{}
		r.Players[0].BetAmount = ptr(0)
		r.Players[1].BetAmount = ptr(0)
	})
	r.PlaceBetsAuto()
	assert.Equal(t, SelectingQuestion, r.State)
}

// ---- 9. PlaceFinalRoundBet ----

func withFinalRoundBetting() func(*Room) {
	bettingEndsAt := time.Now().Add(TimeToBet)
	return func(r *Room) {
		r.State = FinalRoundBetting
		r.FinalRoundState = &FinalRoundState{
			Players:       []string{"p1", "p2"},
			BettingEndsAt: &bettingEndsAt,
		}
	}
}

func TestPlaceFinalRoundBet_Guards(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Room)
		userId  string
		amount  int
		errType any
	}{
		{
			name:    "WrongState",
			setup:   func(r *Room) {},
			userId:  "p1",
			amount:  100,
			errType: &custerr.ConflictErr{},
		},
		{
			name:    "PlayerNotInFinal",
			setup:   withFinalRoundBetting(),
			userId:  "p3",
			amount:  100,
			errType: &custerr.ForbiddenErr{},
		},
		{
			name: "AlreadyBet",
			setup: func(r *Room) {
				withFinalRoundBetting()(r)
				r.Players[0].BetAmount = ptr(100)
			},
			userId:  "p1",
			amount:  200,
			errType: &custerr.ConflictErr{},
		},
		{
			name:    "ExceedsScore",
			setup:   withFinalRoundBetting(),
			userId:  "p1",
			amount:  9999,
			errType: &custerr.ConflictErr{},
		},
		{
			name:    "NegativeAmount",
			setup:   withFinalRoundBetting(),
			userId:  "p1",
			amount:  -1,
			errType: &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.PlaceFinalRoundBet(tc.userId, tc.amount)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestPlaceFinalRoundBet_ValidBet_SetsBetAmount(t *testing.T) {
	r := buildRoom(withFinalRoundBetting())
	err := r.PlaceFinalRoundBet("p1", 300)
	assert.NoError(t, err)
	assert.Equal(t, ptr(300), r.Players[0].BetAmount)
	assert.Equal(t, FinalRoundBetting, r.State) // not all have bet yet
}

func TestPlaceFinalRoundBet_AllPlayersInFinalBet_StartsFinalQuestion(t *testing.T) {
	r := buildRoom(withFinalRoundBetting())
	_ = r.PlaceFinalRoundBet("p1", 300)
	_ = r.PlaceFinalRoundBet("p2", 500)
	assert.Equal(t, ShowingFinalRoundQuestion, r.State)
	assert.NotNil(t, r.FinalRoundState.TimerEndsAt)
}

func TestPlaceFinalRoundBet_PartialBets_DoesNotAdvanceState(t *testing.T) {
	r := buildRoom(withFinalRoundBetting())
	_ = r.PlaceFinalRoundBet("p1", 300)
	assert.Equal(t, FinalRoundBetting, r.State)
}

// ---- 10. StartNextRegularRound ----

func TestStartNextRegularRound_FirstRound_FromWaiting(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) { r.State = WaitingForStart })
	ok := r.StartNextRegularRound(pack)
	assert.True(t, ok)
	assert.Equal(t, "Round 1", *r.CurrentRoundName)
	assert.Equal(t, SelectingQuestion, r.State)
}

func TestStartNextRegularRound_AdvancesToNextRound(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.State = SelectingQuestion
		r.CurrentRoundName = ptr("Round 1")
	})
	ok := r.StartNextRegularRound(pack)
	assert.True(t, ok)
	assert.Equal(t, "Round 2", *r.CurrentRoundName)
}

func TestStartNextRegularRound_NoRoundsLeft_ReturnsFalse(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.State = SelectingQuestion
		r.CurrentRoundName = ptr("Round 2") // last round
	})
	ok := r.StartNextRegularRound(pack)
	assert.False(t, ok)
}

func TestStartNextRegularRound_PopulatesCurrentRoundQuestions(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) { r.State = WaitingForStart })
	r.StartNextRegularRound(pack)
	for _, cq := range r.CurrentRoundQuestions {
		for _, bq := range cq.Questions {
			assert.False(t, bq.HasBeenPlayed)
		}
	}
	assert.Len(t, r.CurrentRoundQuestions, len(pack.Rounds[0].Categories))
}

// ---- 11. AnyAvailableQuestions ----

func TestAnyAvailableQuestions_AllPlayed_ReturnsFalse(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.CurrentRoundQuestions = CurrentRoundQuestions{
			{Category: "A", Questions: []BoardQuestion{{HasBeenPlayed: true}}},
		}
	})
	assert.False(t, r.AnyAvailableQuestions())
}

func TestAnyAvailableQuestions_OneAvailable_ReturnsTrue(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.CurrentRoundQuestions = CurrentRoundQuestions{
			{Category: "A", Questions: []BoardQuestion{
				{HasBeenPlayed: true},
				{HasBeenPlayed: false},
			}},
		}
	})
	assert.True(t, r.AnyAvailableQuestions())
}

func TestAnyAvailableQuestions_EmptyBoard_ReturnsFalse(t *testing.T) {
	r := buildRoom(func(r *Room) { r.CurrentRoundQuestions = CurrentRoundQuestions{} })
	assert.False(t, r.AnyAvailableQuestions())
}

// ---- 12. PassQuestion ----

func withPassingState() func(*Room) {
	return func(r *Room) {
		r.State = Passing
		r.CurrentPlayer = ptr("p1")
		r.CurrentQuestion = &CurrentQuestion{PassingEndsAt: time.Now().Add(TimeToPass)}
	}
}

func TestPassQuestion_Guards(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Room)
		fromUserId string
		toUserId   string
		errType    any
	}{
		{
			name:       "WrongState",
			setup:      func(r *Room) { r.CurrentPlayer = ptr("p1") },
			fromUserId: "p1",
			toUserId:   "p2",
			errType:    &custerr.ConflictErr{},
		},
		{
			name:       "UnknownTarget",
			setup:      withPassingState(),
			fromUserId: "p1",
			toUserId:   "unknown",
			errType:    &custerr.NotFoundErr{},
		},
		{
			name:       "NotCurrentPlayerOrHost",
			setup:      withPassingState(),
			fromUserId: "p2",
			toUserId:   "p1",
			errType:    &custerr.ConflictErr{},
		},
		{
			name:       "PassToSelf",
			setup:      withPassingState(),
			fromUserId: "p1",
			toUserId:   "p1",
			errType:    &custerr.ConflictErr{},
		},
		{
			name: "PassToDisconnected",
			setup: func(r *Room) {
				withPassingState()(r)
				r.Players[1].IsConnected = false
			},
			fromUserId: "p1",
			toUserId:   "p2",
			errType:    &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.PassQuestion(tc.fromUserId, tc.toUserId)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestPassQuestion_ValidPass_StartsNonRegularQuestion(t *testing.T) {
	r := buildRoom(withPassingState())
	err := r.PassQuestion("p1", "p2")
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
	assert.Equal(t, "p2", r.AnsweringPlayer.Id)
	assert.Equal(t, "p2", *r.CurrentPlayer)
}

func TestPassQuestion_HostCanPass(t *testing.T) {
	r := buildRoom(withPassingState())
	err := r.PassQuestion("host1", "p2")
	assert.NoError(t, err)
	assert.Equal(t, Answering, r.State)
}

// ---- 13. PassQuestionAuto ----

func TestPassQuestionAuto_ConnectedOtherPlayers_PicksOne(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Passing
		r.CurrentPlayer = ptr("p1")
		r.CurrentQuestion = &CurrentQuestion{}
	})
	r.PassQuestionAuto()
	assert.Equal(t, Answering, r.State)
	assert.NotNil(t, r.AnsweringPlayer)
	assert.NotEqual(t, "p1", r.AnsweringPlayer.Id)
}

func TestPassQuestionAuto_NoOtherConnected_KeepsCurrent(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = Passing
		r.CurrentPlayer = ptr("p1")
		r.CurrentQuestion = &CurrentQuestion{}
		r.Players[1].IsConnected = false
	})
	r.PassQuestionAuto()
	assert.Equal(t, "p1", r.AnsweringPlayer.Id)
}

// ---- 14. SkipQuestion ----

func TestSkipQuestion_WrongState(t *testing.T) {
	r := buildRoom()
	err := r.SkipQuestion("host1")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestSkipQuestion_NotHost(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = ShowingQuestion
		r.CurrentQuestion = &CurrentQuestion{}
	})
	err := r.SkipQuestion("p1")
	var fe custerr.ForbiddenErr
	assert.ErrorAs(t, err, &fe)
}

func TestSkipQuestion_ValidStates_EndsQuestion(t *testing.T) {
	states := []RoomState{RevealingQuestion, ShowingQuestion, Answering, Passing, Betting}
	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			r := buildRoom(func(r *Room) {
				r.State = state
				r.CurrentQuestion = &CurrentQuestion{}
				if state == Answering {
					r.AnsweringPlayer = &AnsweringPlayer{Id: "p1"}
				}
			})
			err := r.SkipQuestion("host1")
			assert.NoError(t, err)
			assert.Equal(t, SelectingQuestion, r.State)
		})
	}
}

// ---- 15. ChangeScore ----

func TestChangeScore_NotHost(t *testing.T) {
	r := buildRoom()
	err := r.ChangeScore("p1", "p2", 500)
	var fe custerr.ForbiddenErr
	assert.ErrorAs(t, err, &fe)
}

func TestChangeScore_UnknownPlayer(t *testing.T) {
	r := buildRoom()
	err := r.ChangeScore("host1", "unknown", 500)
	var ne custerr.NotFoundErr
	assert.ErrorAs(t, err, &ne)
}

func TestChangeScore_ValidChange(t *testing.T) {
	r := buildRoom()
	err := r.ChangeScore("host1", "p1", 9999)
	assert.NoError(t, err)
	assert.Equal(t, 9999, r.Players[0].Score)
}

// ---- 16. EndQuestion ----

func TestEndQuestion_ClearsAllFields(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = ShowingQuestion
		r.CurrentQuestion = &CurrentQuestion{}
		r.AnsweringPlayer = &AnsweringPlayer{Id: "p1"}
		r.AllowedToAnswer = []string{"p1", "p2"}
		r.Players[0].BetAmount = ptr(300)
		r.Players[1].BetAmount = ptr(200)
	})
	r.EndQuestion()
	assert.Equal(t, SelectingQuestion, r.State)
	assert.Nil(t, r.CurrentQuestion)
	assert.Nil(t, r.AnsweringPlayer)
	assert.Empty(t, r.AllowedToAnswer)
	assert.Nil(t, r.Players[0].BetAmount)
	assert.Nil(t, r.Players[1].BetAmount)
}

// ---- 17. StartFinalRound ----

func TestStartFinalRound_NoPositiveScorePlayers_ReturnsFalse(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.Players[0].Score = 0
		r.Players[1].Score = 0
	})
	ok, err := r.StartFinalRound(pack, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestStartFinalRound_EligiblePlayers_InitializesFinalState(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withRound1(pack))
	ok, err := r.StartFinalRound(pack, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, SelectingFinalRoundCategory, r.State)
	assert.NotNil(t, r.FinalRoundState)
	assert.Equal(t, []string{"p1", "p2"}, r.FinalRoundState.Players)
	assert.Nil(t, r.CurrentRoundName)
	assert.Nil(t, r.CurrentRoundQuestions)
	assert.Nil(t, r.CurrentQuestion)
	assert.Nil(t, r.AnsweringPlayer)
}

func TestStartFinalRound_SingleCategory_SkipsToFinalRoundBetting(t *testing.T) {
	pack := buildPackSingleFinalCat()
	r := buildRoom()
	ok, err := r.StartFinalRound(pack, noopAttachmentUrl)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, FinalRoundBetting, r.State)
	assert.NotNil(t, r.FinalRoundState.BettingEndsAt)
}

// ---- 18. SubmitFinalRoundAnswer ----

func withShowingFinalRound() func(*Room) {
	timerEndsAt := time.Now().Add(60 * time.Second)
	return func(r *Room) {
		r.State = ShowingFinalRoundQuestion
		r.AllowedToAnswer = []string{"p1", "p2"}
		r.FinalRoundState = &FinalRoundState{
			Players:        []string{"p1", "p2"},
			PlayersAnswers: make(map[string]string),
			TimerEndsAt:    &timerEndsAt,
		}
	}
}

func TestSubmitFinalRoundAnswer_WrongState(t *testing.T) {
	r := buildRoom()
	err := r.SubmitFinalRoundAnswer("p1", "answer")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestSubmitFinalRoundAnswer_NotAllowed(t *testing.T) {
	r := buildRoom(withShowingFinalRound(), func(r *Room) {
		r.AllowedToAnswer = []string{"p2"}
	})
	err := r.SubmitFinalRoundAnswer("p1", "answer")
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestSubmitFinalRoundAnswer_Recorded(t *testing.T) {
	r := buildRoom(withShowingFinalRound())
	err := r.SubmitFinalRoundAnswer("p1", "my answer")
	assert.NoError(t, err)
	assert.Equal(t, "my answer", r.FinalRoundState.PlayersAnswers["p1"])
	assert.NotContains(t, r.AllowedToAnswer, "p1")
	assert.Equal(t, ShowingFinalRoundQuestion, r.State)
}

func TestSubmitFinalRoundAnswer_LastAnswer_EndsQuestion(t *testing.T) {
	r := buildRoom(withShowingFinalRound(), func(r *Room) {
		r.AllowedToAnswer = []string{"p1"}
	})
	err := r.SubmitFinalRoundAnswer("p1", "last answer")
	assert.NoError(t, err)
	assert.Equal(t, ValidatingFinalRoundAnswers, r.State)
}

// ---- 19. ValidateFinalRoundAnswer ----

func withValidatingFinalRound() func(*Room) {
	return func(r *Room) {
		r.State = ValidatingFinalRoundAnswers
		r.FinalRoundState = &FinalRoundState{
			Players:        []string{"p1", "p2"},
			PlayersAnswers: map[string]string{"p1": "answer1", "p2": "answer2"},
		}
		r.CurrentPlayer = ptr("p1")
		r.Players[0].BetAmount = ptr(300)
		r.Players[1].BetAmount = ptr(500)
	}
}

func TestValidateFinalRoundAnswer_WrongState(t *testing.T) {
	r := buildRoom()
	err := r.ValidateFinalRoundAnswer("host1", true)
	var ce custerr.ConflictErr
	assert.ErrorAs(t, err, &ce)
}

func TestValidateFinalRoundAnswer_NotHost(t *testing.T) {
	r := buildRoom(withValidatingFinalRound())
	err := r.ValidateFinalRoundAnswer("p1", true)
	var fe custerr.ForbiddenErr
	assert.ErrorAs(t, err, &fe)
}

func TestValidateFinalRoundAnswer_Correct(t *testing.T) {
	r := buildRoom(withValidatingFinalRound())
	r.Players[0].Score = 1000
	err := r.ValidateFinalRoundAnswer("host1", true)
	assert.NoError(t, err)
	assert.Equal(t, 1300, r.Players[0].Score)
}

func TestValidateFinalRoundAnswer_Incorrect(t *testing.T) {
	r := buildRoom(withValidatingFinalRound())
	r.Players[0].Score = 1000
	err := r.ValidateFinalRoundAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, 700, r.Players[0].Score)
}

func TestValidateFinalRoundAnswer_NotLastPlayer_AdvancesCurrent(t *testing.T) {
	r := buildRoom(withValidatingFinalRound())
	err := r.ValidateFinalRoundAnswer("host1", true)
	assert.NoError(t, err)
	assert.Equal(t, "p2", *r.CurrentPlayer)
	assert.Equal(t, ValidatingFinalRoundAnswers, r.State)
}

func TestValidateFinalRoundAnswer_LastPlayer_EndsGame(t *testing.T) {
	r := buildRoom(withValidatingFinalRound(), func(r *Room) {
		r.FinalRoundState.Players = []string{"p1"}
		r.CurrentPlayer = ptr("p1")
	})
	err := r.ValidateFinalRoundAnswer("host1", true)
	assert.NoError(t, err)
	assert.Equal(t, GameOver, r.State)
}

// ---- 20. EndGame ----

func TestEndGame_SetsGameOverAndClearsTimers(t *testing.T) {
	timerEndsAt := time.Now().Add(30 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = ValidatingFinalRoundAnswers
		r.CurrentPlayer = ptr("p1")
		r.FinalRoundState = &FinalRoundState{TimerEndsAt: &timerEndsAt}
	})
	r.EndGame()
	assert.Equal(t, GameOver, r.State)
	assert.Nil(t, r.CurrentPlayer)
	assert.Nil(t, r.FinalRoundState.TimerEndsAt)
	for _, p := range r.Players {
		assert.NotNil(t, p.BetAmount)
		assert.Equal(t, 0, *p.BetAmount)
	}
}

// ---- 21. Role helpers ----

func TestIsUserHost(t *testing.T) {
	r := buildRoom()
	assert.True(t, r.IsUserHost("host1"))
	assert.False(t, r.IsUserHost("p1"))
	assert.False(t, r.IsUserHost("nobody"))

	rNoHost := buildRoom(func(r *Room) { r.Host = nil })
	assert.False(t, rNoHost.IsUserHost("host1"))
}

func TestIsUserPlayer(t *testing.T) {
	r := buildRoom()
	assert.True(t, r.IsUserPlayer("p1"))
	assert.True(t, r.IsUserPlayer("p2"))
	assert.False(t, r.IsUserPlayer("host1"))
	assert.False(t, r.IsUserPlayer("nobody"))
}

func TestIsUserIn(t *testing.T) {
	r := buildRoom()
	assert.True(t, r.IsUserIn("host1"))
	assert.True(t, r.IsUserIn("p1"))
	assert.False(t, r.IsUserIn("nobody"))
}

func TestIsUserBanned(t *testing.T) {
	r := buildRoom(func(r *Room) { r.BanList = []string{"banned1"} })
	assert.True(t, r.IsUserBanned("banned1"))
	assert.False(t, r.IsUserBanned("p1"))
	assert.False(t, r.IsUserBanned("nobody"))
}

// ---- 22. continueRegularQuestion (via ValidateAnswer) ----

func TestContinueRegularQuestion_DuringReveal_ReturnsToRevealingQuestion(t *testing.T) {
	origTimerStartsAt := time.Now().Add(5 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = Answering
		r.CurrentQuestion = &CurrentQuestion{
			Question:      Question{HiddenQuestion: HiddenQuestion{Value: 100}},
			TimerStartsAt: origTimerStartsAt,
		}
		r.AnsweringPlayer = &AnsweringPlayer{
			Id:            "p1",
			TimerStartsAt: time.Now().Add(-2 * time.Second),
			TimerEndsAt:   time.Now().Add(8 * time.Second),
		}
		r.AllowedToAnswer = []string{"p2"}
	})
	err := r.ValidateAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, RevealingQuestion, r.State)
	assert.True(t, r.CurrentQuestion.TimerStartsAt.After(origTimerStartsAt))
	assert.Nil(t, r.AnsweringPlayer)
}

func TestContinueRegularQuestion_DuringShowing_ReturnsToShowingQuestion(t *testing.T) {
	origTimerEndsAt := time.Now().Add(5 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = Answering
		r.CurrentQuestion = &CurrentQuestion{
			Question:      Question{HiddenQuestion: HiddenQuestion{Value: 100}},
			TimerStartsAt: time.Now().Add(-3 * time.Second),
			TimerEndsAt:   origTimerEndsAt,
		}
		r.AnsweringPlayer = &AnsweringPlayer{
			Id:            "p1",
			TimerStartsAt: time.Now().Add(-2 * time.Second),
			TimerEndsAt:   time.Now().Add(8 * time.Second),
		}
		r.AllowedToAnswer = []string{"p2"}
	})
	err := r.ValidateAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, ShowingQuestion, r.State)
	assert.True(t, r.CurrentQuestion.TimerEndsAt.After(origTimerEndsAt))
	assert.Nil(t, r.AnsweringPlayer)
}

func TestContinueRegularQuestion_DuringReveal_WithAttachment_ShiftsAttachmentTimer(t *testing.T) {
	origAttachEnd := time.Now().Add(3 * time.Second)
	origTimerStart := time.Now().Add(5 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = Answering
		r.CurrentQuestion = &CurrentQuestion{
			Question: Question{
				HiddenQuestion: HiddenQuestion{Value: 100},
				Attachment:     &Attachment{Duration: 3.0},
			},
			AttachmentRevealEndsAt: origAttachEnd,
			TimerStartsAt:          origTimerStart,
		}
		r.AnsweringPlayer = &AnsweringPlayer{
			Id:            "p1",
			TimerStartsAt: time.Now().Add(-1 * time.Second),
			TimerEndsAt:   time.Now().Add(9 * time.Second),
		}
		r.AllowedToAnswer = []string{"p2"}
	})
	err := r.ValidateAnswer("host1", false)
	assert.NoError(t, err)
	assert.Equal(t, RevealingQuestion, r.State)
	assert.True(t, r.CurrentQuestion.AttachmentRevealEndsAt.After(origAttachEnd))
	assert.True(t, r.CurrentQuestion.TimerStartsAt.After(origTimerStart))
}

// ---- 23. StartRegularQuestion ----

func TestStartRegularQuestion_NoAttachment_SetsShowingState(t *testing.T) {
	timerStartsAt := time.Now().Add(1 * time.Second)
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.Options.QuestionThinkingTime = 10
		r.CurrentQuestion = &CurrentQuestion{
			TimerStartsAt: timerStartsAt,
		}
	})
	r.StartRegularQuestion()
	assert.Equal(t, ShowingQuestion, r.State)
	assert.Equal(t, 1.0, r.CurrentQuestion.TextRevealLastProgress)
	assert.Equal(t, 1.0, r.CurrentQuestion.TimerLastProgress)
	expected := timerStartsAt.Add(10 * time.Second)
	assert.WithinDuration(t, expected, r.CurrentQuestion.TimerEndsAt, 50*time.Millisecond)
}

func TestStartRegularQuestion_WithAttachment_SetsAttachmentProgress(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = RevealingQuestion
		r.CurrentQuestion = &CurrentQuestion{
			Question:      Question{Attachment: &Attachment{Duration: 3.0}},
			TimerStartsAt: time.Now().Add(5 * time.Second),
		}
	})
	r.StartRegularQuestion()
	assert.Equal(t, 1.0, r.CurrentQuestion.AttachmentRevealLastProgress)
	assert.Equal(t, ShowingQuestion, r.State)
}

// ---- 24. StartGame ----

func TestStartGame_SetsCurrentPlayerAndStartsRound(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) { r.State = WaitingForStart })
	r.StartGame(pack)
	assert.NotNil(t, r.CurrentPlayer)
	assert.Contains(t, []string{"p1", "p2"}, *r.CurrentPlayer)
	assert.Equal(t, SelectingQuestion, r.State)
	assert.NotNil(t, r.CurrentRoundName)
	assert.Equal(t, "Round 1", *r.CurrentRoundName)
}

// ---- 25. GetProjection ----

func TestGetProjection_Host_ReturnsRoomHost(t *testing.T) {
	r := buildRoom()
	_, ok := r.GetProjection("host1").(RoomHost)
	assert.True(t, ok)
}

func TestGetProjection_Player_ReturnsRoomPlayer(t *testing.T) {
	r := buildRoom()
	_, ok := r.GetProjection("p1").(RoomPlayer)
	assert.True(t, ok)
}

func TestGetProjection_Outsider_ReturnsRoomLobby(t *testing.T) {
	r := buildRoom()
	_, ok := r.GetProjection("nobody").(RoomLobby)
	assert.True(t, ok)
}

// ---- 26. GetAvailableFinalRoundCategories ----

func TestGetAvailableFinalRoundCategories_ReturnsOnlyAvailable(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.FinalRoundState = &FinalRoundState{
			AvailableCategories: map[string]bool{
				"Cat A": true,
				"Cat B": false,
				"Cat C": true,
			},
		}
	})
	available := r.GetAvailableFinalRoundCategories()
	assert.Len(t, available, 2)
	assert.Contains(t, available, "Cat A")
	assert.Contains(t, available, "Cat C")
	assert.NotContains(t, available, "Cat B")
}

func TestGetAvailableFinalRoundCategories_AllRemoved_ReturnsEmpty(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.FinalRoundState = &FinalRoundState{
			AvailableCategories: map[string]bool{"Cat A": false},
		}
	})
	assert.Empty(t, r.GetAvailableFinalRoundCategories())
}

// ---- 27. RemoveFinalRoundCategory ----

func withSelectingFinalRoundCategory() func(*Room) {
	return func(r *Room) {
		r.State = SelectingFinalRoundCategory
		r.CurrentPlayer = ptr("p1")
		r.FinalRoundState = &FinalRoundState{
			Players: []string{"p1", "p2"},
			AvailableCategories: map[string]bool{
				"Final Cat A": true,
				"Final Cat B": true,
				"Final Cat C": true,
			},
		}
	}
}

func TestRemoveFinalRoundCategory_Guards(t *testing.T) {
	pack := buildPack()
	tests := []struct {
		name     string
		setup    func(*Room)
		userId   string
		category string
		errType  any
	}{
		{
			name:     "WrongState",
			setup:    func(r *Room) { r.State = FinalRoundBetting },
			userId:   "p1",
			category: "Final Cat A",
			errType:  &custerr.ConflictErr{},
		},
		{
			name:     "NotCurrentPlayerOrHost",
			setup:    withSelectingFinalRoundCategory(),
			userId:   "p2",
			category: "Final Cat A",
			errType:  &custerr.ForbiddenErr{},
		},
		{
			name:     "UnknownCategory",
			setup:    withSelectingFinalRoundCategory(),
			userId:   "p1",
			category: "DoesNotExist",
			errType:  &custerr.NotFoundErr{},
		},
		{
			name: "AlreadyRemoved",
			setup: func(r *Room) {
				withSelectingFinalRoundCategory()(r)
				r.FinalRoundState.AvailableCategories["Final Cat A"] = false
			},
			userId:   "p1",
			category: "Final Cat A",
			errType:  &custerr.ConflictErr{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := buildRoom(tc.setup)
			err := r.RemoveFinalRoundCategory(pack, tc.userId, tc.category, noopAttachmentUrl)
			assert.True(t, errors.As(err, tc.errType))
		})
	}
}

func TestRemoveFinalRoundCategory_MarksUnavailableAndAdvancesPlayer(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withSelectingFinalRoundCategory()) // 3 categories, p1 current
	err := r.RemoveFinalRoundCategory(pack, "p1", "Final Cat A", noopAttachmentUrl)
	assert.NoError(t, err)
	assert.False(t, r.FinalRoundState.AvailableCategories["Final Cat A"])
	assert.Equal(t, "p2", *r.CurrentPlayer)
	assert.Equal(t, SelectingFinalRoundCategory, r.State)
}

func TestRemoveFinalRoundCategory_LastPlayerWrapsToFirst(t *testing.T) {
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.State = SelectingFinalRoundCategory
		r.CurrentPlayer = ptr("p2")
		r.FinalRoundState = &FinalRoundState{
			Players: []string{"p1", "p2"},
			AvailableCategories: map[string]bool{
				"Final Cat A": true,
				"Final Cat B": true,
				"Final Cat C": true,
			},
		}
	})
	err := r.RemoveFinalRoundCategory(pack, "p2", "Final Cat A", noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, "p1", *r.CurrentPlayer)
}

func TestRemoveFinalRoundCategory_HostCanRemove(t *testing.T) {
	pack := buildPack()
	r := buildRoom(withSelectingFinalRoundCategory())
	err := r.RemoveFinalRoundCategory(pack, "host1", "Final Cat A", noopAttachmentUrl)
	assert.NoError(t, err)
	assert.False(t, r.FinalRoundState.AvailableCategories["Final Cat A"])
}

func TestRemoveFinalRoundCategory_OneCategoryLeft_AdvancesToBetting(t *testing.T) {
	// buildPack has "Final Cat A" and "Final Cat B" — removing A leaves B → auto-advance
	pack := buildPack()
	r := buildRoom(func(r *Room) {
		r.State = SelectingFinalRoundCategory
		r.CurrentPlayer = ptr("p1")
		r.FinalRoundState = &FinalRoundState{
			Players: []string{"p1", "p2"},
			AvailableCategories: map[string]bool{
				"Final Cat A": true,
				"Final Cat B": true,
			},
		}
	})
	err := r.RemoveFinalRoundCategory(pack, "p1", "Final Cat A", noopAttachmentUrl)
	assert.NoError(t, err)
	assert.Equal(t, FinalRoundBetting, r.State)
	assert.NotNil(t, r.FinalRoundState.BettingEndsAt)
	assert.NotNil(t, r.FinalRoundState.Question)
}

// ---- 28. PlaceFinalRoundBetsAuto ----

func TestPlaceFinalRoundBetsAuto_AllNilBets_AutoZerosAndEndsGame(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = FinalRoundBetting
		r.FinalRoundState = &FinalRoundState{Players: []string{"p1", "p2"}}
		// both BetAmounts nil → auto-zeroed → all 0 → EndGame
	})
	r.PlaceFinalRoundBetsAuto()
	assert.Equal(t, GameOver, r.State)
	for _, p := range r.Players {
		assert.NotNil(t, p.BetAmount)
		assert.Equal(t, 0, *p.BetAmount)
	}
}

func TestPlaceFinalRoundBetsAuto_OnePositiveBet_StartsFinalQuestion(t *testing.T) {
	r := buildRoom(func(r *Room) {
		r.State = FinalRoundBetting
		r.FinalRoundState = &FinalRoundState{Players: []string{"p1", "p2"}}
		r.Players[0].BetAmount = ptr(500) // p1 has bet, p2 nil → auto-zero
	})
	r.PlaceFinalRoundBetsAuto()
	assert.Equal(t, ShowingFinalRoundQuestion, r.State)
	assert.NotNil(t, r.FinalRoundState.TimerEndsAt)
	assert.Equal(t, 0, *r.Players[1].BetAmount)
}

// ---- 29. Pause — remaining allowed states ----

func TestPause_AllowedStates_SetsPausedFlag(t *testing.T) {
	states := []struct {
		state RoomState
		setup func(*Room)
	}{
		{Answering, func(r *Room) {
			r.CurrentQuestion = &CurrentQuestion{}
			r.AnsweringPlayer = &AnsweringPlayer{Id: "p1"}
		}},
		{Betting, func(r *Room) {
			r.CurrentQuestion = &CurrentQuestion{BettingEndsAt: time.Now().Add(30 * time.Second)}
		}},
		{Passing, func(r *Room) {
			r.CurrentQuestion = &CurrentQuestion{PassingEndsAt: time.Now().Add(30 * time.Second)}
		}},
		{SelectingFinalRoundCategory, func(r *Room) {}},
		{FinalRoundBetting, func(r *Room) {
			bettingEndsAt := time.Now().Add(30 * time.Second)
			r.FinalRoundState = &FinalRoundState{BettingEndsAt: &bettingEndsAt}
		}},
		{ShowingFinalRoundQuestion, func(r *Room) {
			timerEndsAt := time.Now().Add(60 * time.Second)
			r.FinalRoundState = &FinalRoundState{TimerEndsAt: &timerEndsAt}
		}},
	}
	for _, tc := range states {
		t.Run(string(tc.state), func(t *testing.T) {
			r := buildRoom(func(r *Room) {
				r.State = tc.state
				tc.setup(r)
			})
			err := r.Pause("host1")
			assert.NoError(t, err)
			assert.True(t, r.PausedState.Paused)
			assert.NotNil(t, r.PausedState.PausedAt)
		})
	}
}
