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
	CurrentQuestion       *HiddenCurrentQuestion `json:"currentQuestion"`
	AnsweringPlayer       *AnsweringPlayer       `json:"answeringPlayer"`
	AllowedToAnswer       []string               `json:"allowedToAnswer"`
	FinalRoundState       *HiddenFinalRoundState `json:"finalRoundState"`
	PausedState           PausedState            `json:"pausedState"`
}

type HiddenCurrentQuestion struct {
	HiddenQuestion
	Type                         QuestionType `json:"type"`
	Text                         *string      `json:"text"`
	Attachment                   *Attachment  `json:"attachment"`
	AttachmentRevealEndsAt       time.Time    `json:"attachmentRevealEndsAt"`
	AttachmentRevealLastProgress float64      `json:"attachmentRevealLastProgress"`
	TextRevealLastProgress       float64      `json:"textRevealLastProgress"`
	TimerStartsAt                time.Time    `json:"timerStartsAt"`
	TimerEndsAt                  time.Time    `json:"timerEndsAt"`
	TimerLastProgress            float64      `json:"timerLastProgress"`
	BettingEndsAt                time.Time    `json:"bettingEndsAt"`
	PassingEndsAt                time.Time    `json:"passingEndsAt"`
}

type HiddenFinalRoundState struct {
	AvailableCategories map[string]bool           `json:"availableCategories"`
	Question            *HiddenFinalRoundQuestion `json:"question"`
	Players             []string                  `json:"players"`
	PlayersAnswers      map[string]bool           `json:"playersAnswers"`
	BettingEndsAt       *time.Time                `json:"bettingEndsAt"`
	TimerEndsAt         *time.Time                `json:"timerEndsAt"`
}

func NewPlayerRoom(room *Room) RoomPlayer {
	var currentQuestion *HiddenCurrentQuestion
	if room.CurrentQuestion != nil {
		currentQuestion = &HiddenCurrentQuestion{
			HiddenQuestion: HiddenQuestion{
				Round:    room.CurrentQuestion.Round,
				Category: room.CurrentQuestion.Category,
				Index:    room.CurrentQuestion.Index,
				Value:    room.CurrentQuestion.Value,
			},
			Type:                         room.CurrentQuestion.Type,
			Text:                         room.CurrentQuestion.Text,
			Attachment:                   room.CurrentQuestion.Attachment,
			AttachmentRevealEndsAt:       room.CurrentQuestion.AttachmentRevealEndsAt,
			AttachmentRevealLastProgress: room.CurrentQuestion.AttachmentRevealLastProgress,
			TextRevealLastProgress:       room.CurrentQuestion.TextRevealLastProgress,
			TimerStartsAt:                room.CurrentQuestion.TimerStartsAt,
			TimerEndsAt:                  room.CurrentQuestion.TimerEndsAt,
			TimerLastProgress:            room.CurrentQuestion.TimerLastProgress,
			BettingEndsAt:                room.CurrentQuestion.BettingEndsAt,
			PassingEndsAt:                room.CurrentQuestion.PassingEndsAt,
		}
	}
	var finalRoundState *HiddenFinalRoundState
	if room.FinalRoundState != nil {
		var finalRoundQuestion *HiddenFinalRoundQuestion
		if room.FinalRoundState.Question != nil {
			finalRoundQuestion = &HiddenFinalRoundQuestion{
				Category:   room.FinalRoundState.Question.Category,
				Text:       room.FinalRoundState.Question.Text,
				Attachment: room.FinalRoundState.Question.Attachment,
			}
		}
		finalRoundPlayersAnswers := make(map[string]bool)
		for id := range room.FinalRoundState.PlayersAnswers {
			finalRoundPlayersAnswers[id] = true
		}
		finalRoundState = &HiddenFinalRoundState{
			AvailableCategories: room.FinalRoundState.AvailableCategories,
			Question:            finalRoundQuestion,
			Players:             room.FinalRoundState.Players,
			PlayersAnswers:      finalRoundPlayersAnswers,
			BettingEndsAt:       room.FinalRoundState.BettingEndsAt,
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
