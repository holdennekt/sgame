package repository

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
)

type GameEvent interface {
	// InsertMany is idempotent: events already stored for the room (same seq) are skipped.
	InsertMany(ctx context.Context, roomId string, events []domain.GameEvent) error
	GetByRoom(ctx context.Context, roomId string) ([]domain.GameEvent, error)
}
