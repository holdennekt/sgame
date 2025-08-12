package dto

import "github.com/holdennekt/sgame/internal/domain"

type CreateRoomDTO struct {
	UserId  string
	RoomDTO *domain.RoomDTO
}

type GetRoomProjectionDTO struct {
	UserId   string
	Id       string
	Password string
}

type ConnectRoomDTO struct {
	UserId string
	Id     string
}

type GetRoomsByCreatedByDTO struct {
	Id string
	SearchDTO
}
