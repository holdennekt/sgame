package domain

import "time"

type RoomPlayer struct {
	Id                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	PackPreview           PackPreview            `json:"packPreview"`
	Host                  *Host                  `json:"host"`
	Players               []Player               `json:"players"`
	State                 RoomState              `json:"state"`
	CurrentRoundName      *string                `json:"currentRoundName"`
	CurrentRoundQuestions CurrentRoundQuestions  `json:"currentRoundQuestions"`
	CurrentPlayer         *string                `json:"currentPlayer"`
	CurrentQuestion       *hiddenCurrentQuestion `json:"currentQuestion"`
	AnsweringPlayer       *AnsweringPlayer       `json:"answeringPlayer"`
	AllowedToAnswer       []string               `json:"allowedToAnswer"`
	FinalRoundState       *hiddenFinalRoundState `json:"finalRoundState"`
	PausedState           PausedState            `json:"pausedState"`
}

type hiddenCurrentQuestion struct {
	HiddenQuestion
	Type              QuestionType `json:"type"`
	Text              string       `json:"text"`
	TimerLastProgress float64      `json:"timerLastProgress" bson:"timerLastProgress"`
	TimerStartsAt     time.Time    `json:"timerStartsAt"`
	TimerEndsAt       time.Time    `json:"timerEndsAt"`
}

type hiddenFinalRoundState struct {
	AvailableCategories map[string]bool           `json:"availableCategories"`
	Question            *HiddenFinalRoundQuestion `json:"question"`
	Players             []string                  `json:"players"`
	PlayersAnswers      map[string]bool           `json:"playersAnswers" bson:"playersAnswers"`
	TimerEndsAt         *time.Time                `json:"timerEndsAt"`
}

func NewPlayerRoom(room *Room) RoomPlayer {
	var currentQuestion *hiddenCurrentQuestion
	if room.CurrentQuestion != nil {
		currentQuestion = &hiddenCurrentQuestion{
			HiddenQuestion: HiddenQuestion{
				Index:      room.CurrentQuestion.Index,
				Value:      room.CurrentQuestion.Value,
				Attachment: room.CurrentQuestion.Attachment,
			},
			Type:              room.CurrentQuestion.Type,
			Text:              room.CurrentQuestion.Text,
			TimerLastProgress: room.CurrentQuestion.TimerLastProgress,
			TimerStartsAt:     room.CurrentQuestion.TimerStartsAt,
			TimerEndsAt:       room.CurrentQuestion.TimerEndsAt,
		}
	}
	var finalRoundState *hiddenFinalRoundState
	if room.FinalRoundState != nil {
		var finalRoundQuestion *HiddenFinalRoundQuestion
		if room.FinalRoundState.Question != nil {
			finalRoundQuestion = &HiddenFinalRoundQuestion{
				Text:       room.FinalRoundState.Question.Text,
				Attachment: room.FinalRoundState.Question.Attachment,
			}
		}
		finalRoundPlayersAnswers := make(map[string]bool)
		for id := range room.FinalRoundState.PlayersAnswers {
			finalRoundPlayersAnswers[id] = true
		}
		finalRoundState = &hiddenFinalRoundState{
			AvailableCategories: room.FinalRoundState.AvailableCategories,
			Question:            finalRoundQuestion,
			Players:             room.FinalRoundState.Players,
			PlayersAnswers:      finalRoundPlayersAnswers,
			TimerEndsAt:         room.FinalRoundState.TimerEndsAt,
		}
	}
	return RoomPlayer{
		Id:                    room.Id,
		Name:                  room.Name,
		PackPreview:           room.PackPreview,
		Host:                  room.Host,
		Players:               room.Players,
		State:                 room.State,
		CurrentRoundName:      room.CurrentRoundName,
		CurrentRoundQuestions: room.CurrentRoundQuestions,
		CurrentPlayer:         room.CurrentPlayer,
		CurrentQuestion:       currentQuestion,
		AnsweringPlayer:       room.AnsweringPlayer,
		AllowedToAnswer:       room.AllowedToAnswer,
		FinalRoundState:       finalRoundState,
		PausedState:           room.PausedState,
	}
}
