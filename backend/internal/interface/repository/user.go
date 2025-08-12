package repository

import (
	"context"

	"github.com/holdennekt/sgame/internal/domain"
)

type User interface {
	Create(ctx context.Context, dbUser *domain.DbUser) (string, error)
	GetById(ctx context.Context, id string) (*domain.DbUser, error)
	GetByLogin(ctx context.Context, login string) (*domain.DbUser, error)
	Update(ctx context.Context, dbUser *domain.DbUser) error
	Delete(ctx context.Context, id string) error
}
