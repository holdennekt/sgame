package cache

import (
	"context"
	"time"

	"github.com/holdennekt/sgame/internal/domain"
)

type Room interface {
	GetById(ctx context.Context, id string) (*domain.Room, error)
	Get(ctx context.Context) ([]domain.RoomLobby, error)
	Set(ctx context.Context, room *domain.Room) error
	SafeSet(ctx context.Context, roomId string, updateFunc func(room *domain.Room) error) (*domain.Room, error)
	Delete(ctx context.Context, roomId string) error
	Expire(ctx context.Context, roomId string, duration time.Duration) error
	Persist(ctx context.Context, roomId string) error
}
