package repository

import (
	"context"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/dto"
)

type PackDraft interface {
	Create(ctx context.Context, draft *domain.PackDraft) (string, error)
	GetById(ctx context.Context, id string) (*domain.PackDraft, error)
	GetByUser(ctx context.Context, userId string, search dto.SearchRequest) ([]domain.PackDraft, int, error)
	GetByUserAndLinkedPack(ctx context.Context, userId, packId string) (*domain.PackDraft, error)
	// GetReferencedKeys returns the subset of keys that appear in any of userId's drafts other than excludeId.
	GetReferencedKeys(ctx context.Context, userId string, keys []string, excludeId string) (map[string]struct{}, error)
	Update(ctx context.Context, draft *domain.PackDraft) error
	Delete(ctx context.Context, id string) error
}
