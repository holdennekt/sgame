package repository

import (
	"context"

	"github.com/holdennekt/sgame/internal/domain"
	"github.com/holdennekt/sgame/internal/dto"
)

type Pack interface {
	Create(ctx context.Context, pack *domain.Pack) (string, error)
	GetById(ctx context.Context, id string) (*domain.Pack, error)
	GetByRoundsChecksum(ctx context.Context, dto dto.GetPackByRoundsChecksumDTO) (*domain.Pack, error)
	GetPreviews(ctx context.Context, dto dto.GetPacksDTO) ([]domain.PackPreview, error)
	GetHiddens(ctx context.Context, dto dto.GetPacksDTO) ([]domain.HiddenPack, error)
	GetCount(ctx context.Context, dto dto.GetPacksDTO) (int, error)
	Update(ctx context.Context, pack *domain.Pack) error
	Delete(ctx context.Context, id string) error
}
