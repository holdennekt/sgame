package repository

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
)

type Pack interface {
	Create(ctx context.Context, pack *domain.Pack) (string, error)
	GetById(ctx context.Context, id string) (*domain.Pack, error)
	GetHiddens(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.HiddenPack, int, error)
	GetCreatedBy(ctx context.Context, userId, createdBy string, search dto.SearchRequest) ([]domain.HiddenPack, int, error)
	GetPreviews(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackPreview, int, error)
	GetByChecksum(ctx context.Context, userId string, checksum []byte, ignoreId string) ([]*domain.Pack, error)
	Update(ctx context.Context, pack *domain.Pack) error
	Delete(ctx context.Context, id string) error
}
