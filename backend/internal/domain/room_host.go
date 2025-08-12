package domain

type RoomHost struct {
	Id                    string                `json:"id"`
	Name                  string                `json:"name"`
	PackPreview           PackPreview           `json:"packPreview"`
	Host                  *Host                 `json:"host"`
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
}

func NewHostRoom(room *Room) RoomHost {
	return RoomHost{
		Id:                    room.Id,
		Name:                  room.Name,
		PackPreview:           room.PackPreview,
		Host:                  room.Host,
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
	}
}
