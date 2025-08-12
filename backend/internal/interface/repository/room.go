package repository

import (
	"context"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
)

type Room interface {
	Create(ctx context.Context, room *domain.Room) error
	GetById(ctx context.Context, id string) (*domain.Room, error)
	GetByCreatedBy(ctx context.Context, dto dto.GetRoomsByCreatedByDTO) ([]domain.Room, error)
}
