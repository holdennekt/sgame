package domain

type RoomModerator struct {
	Id                    string                `json:"id"`
	Name                  string                `json:"name"`
	PackPreview           PackPreview           `json:"packPreview"`
	Options               RoomOptions           `json:"options"`
	Moderator             *Moderator            `json:"moderator"`
	Players               []Player              `json:"players"`
	State                 RoomState             `json:"state"`
	CurrentRoundName      *string               `json:"currentRoundName"`
	CurrentRoundQuestions CurrentRoundQuestions `json:"currentRoundQuestions"`
	CurrentPlayer         *string               `json:"currentPlayer"`
	CurrentQuestion       *CurrentQuestion      `json:"currentQuestion"`
	AnsweringPlayer       *AnsweringPlayer      `json:"answeringPlayer"`
	AllowedToAnswer       []string              `json:"allowedToAnswer"`
	FinalRoundState       *FinalRoundState      `json:"finalRoundState"`
	PausedState           PausedState           `json:"pausedState"`
	SpectatorCount        int                   `json:"spectatorCount"`
}

func NewModeratorRoom(room *Room, spectatorCount int) RoomModerator {
	return RoomModerator{
		Id:                    room.Id,
		Name:                  room.Name,
		PackPreview:           room.PackPreview,
		Options:               room.Options,
		Moderator:             room.Moderator,
		Players:               room.Players,
		State:                 room.State,
		CurrentRoundName:      room.CurrentRoundName,
		CurrentRoundQuestions: room.CurrentRoundQuestions,
		CurrentPlayer:         room.CurrentPlayer,
		CurrentQuestion:       room.CurrentQuestion,
		AnsweringPlayer:       room.AnsweringPlayer,
		AllowedToAnswer:       room.AllowedToAnswer,
		FinalRoundState:       room.FinalRoundState,
		PausedState:           room.PausedState,
		SpectatorCount:        spectatorCount,
	}
}
