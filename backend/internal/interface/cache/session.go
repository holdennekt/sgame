package cache

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
)

type Session interface {
	Get(ctx context.Context, sessionId string) (*domain.User, error)
	GetKey(ctx context.Context, userId string) (string, error)
	Set(ctx context.Context, sessionId string, user *domain.User) error
	Delete(ctx context.Context, sessionId string) error
}
