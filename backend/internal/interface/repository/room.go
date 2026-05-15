package repository

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
)

type Room interface {
	Create(ctx context.Context, room *domain.Room) error
	GetById(ctx context.Context, id string) (*domain.Room, error)
	GetByCreatedBy(ctx context.Context, id string, search dto.SearchRequest) ([]domain.Room, error)
}
