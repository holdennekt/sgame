package domain

type RoomLobby struct {
	Id          string      `json:"id"`
	Name        string      `json:"name"`
	PackPreview PackPreview `json:"packPreview"`
	Host        *Host       `json:"host"`
	Players     []Player    `json:"players"`
	MaxPlayers  int         `json:"maxPlayers"`
	Type        PrivacyType `json:"type"`
	Status      string      `json:"status"`
}

func NewRoomLobby(room *Room) RoomLobby {
	var status string
	switch room.State {
	case WaitingForStart:
		status = "Waiting for start"
	case GameOver:
		status = "Game ended"
	default:
		status = "Playing"
	}

	return RoomLobby{
		Id:          room.Id,
		Name:        room.Name,
		PackPreview: room.PackPreview,
		Host:        room.Host,
		Players:     room.Players,
		MaxPlayers:  room.Options.MaxPlayers,
		Type:        room.Options.Type,
		Status:      status,
	}
}
