package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type SourceRepository interface {
	Create(ctx context.Context, s *domain.RSSSource) error
	List(ctx context.Context) ([]domain.RSSSource, error)
	UpdateLastFetched(ctx context.Context, id uuid.UUID, t time.Time) error
	ResolveSources(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.FeedSource, error)
}

type FeedRepository interface {
	GetFeed(ctx context.Context, userID *uuid.UUID, cursor time.Time, limit int) ([]domain.FeedItem, error)
}
