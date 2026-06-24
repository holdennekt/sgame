package cache

import (
	"context"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
)

type Room interface {
	GetById(ctx context.Context, id string) (*domain.Room, error)
	Get(ctx context.Context) ([]domain.RoomLobby, error)
	Set(ctx context.Context, room *domain.Room) error
	SafeUpdate(ctx context.Context, roomId string, updateFunc func(room *domain.Room) error) (*domain.Room, error)
	Delete(ctx context.Context, roomId string) error
	Expire(ctx context.Context, roomId string, duration time.Duration) error
	Persist(ctx context.Context, roomId string) error
	TrySetOwner(ctx context.Context, roomId string, ttl time.Duration) (bool, error)
	UpdateOwner(ctx context.Context, roomId string, ttl time.Duration) error
	ListenForExpiredOwners(ctx context.Context, handleExpiredOwner func(roomId string))
	IncrSpectators(ctx context.Context, roomId string) (int, error)
	DecrSpectators(ctx context.Context, roomId string) (int, error)
	GetSpectatorCount(ctx context.Context, roomId string) (int, error)
}
