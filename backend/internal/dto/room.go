package dto

import "github.com/holdennekt/sgame/backend/internal/domain"

type CreateRoomRequest struct {
	Name    string             `json:"name" binding:"min=1,max=50"`
	PackId  string             `json:"packId" binding:"required"`
	Options domain.RoomOptions `json:"options"`
}

type CreateRoomResponse struct {
	Id string `json:"id" example:"507f1f77bcf86cd799439011"`
}
