package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type RoadmapRepository interface {
	List(ctx context.Context) ([]domain.RoadmapItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RoadmapItem, error)
	Create(ctx context.Context, item *domain.RoadmapItem) error
	Update(ctx context.Context, item *domain.RoadmapItem) error
	Delete(ctx context.Context, id uuid.UUID) error
}
