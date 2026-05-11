package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type BannedWordRepository interface {
	List(ctx context.Context) ([]domain.BannedWord, error)
	Create(ctx context.Context, word *domain.BannedWord) error
	Delete(ctx context.Context, id uuid.UUID) error
	IsWordBanned(ctx context.Context, word string, scope domain.BannedWordScope) (bool, error)
}
