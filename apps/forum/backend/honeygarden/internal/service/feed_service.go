package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
	"honeygarden/internal/worker"
)

type FeedService struct {
	sources port.SourceRepository
	feed    port.FeedRepository
	log     zerolog.Logger
}

func NewFeedService(
	sources port.SourceRepository,
	feed port.FeedRepository,
	log zerolog.Logger,
) *FeedService {
	return &FeedService{sources: sources, feed: feed, log: log}
}

func (s *FeedService) GetFeed(ctx context.Context, userID *uuid.UUID, cursor *time.Time, limit int) ([]domain.FeedItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cur := time.Now().UTC()
	if cursor != nil {
		cur = *cursor
	}
	return s.feed.GetFeed(ctx, userID, cur, limit)
}

func (s *FeedService) AddSource(ctx context.Context, input domain.CreateSourceInput) (*domain.RSSSource, error) {
	if input.URL == "" || input.Name == "" {
		return nil, domain.ErrInvalidInput
	}
	if err := worker.ValidatePublicURL(input.URL); err != nil {
		s.log.Warn().Err(err).Str("url", input.URL).Msg("rss: rejected source url")
		return nil, domain.ErrInvalidInput
	}
	src := &domain.RSSSource{
		ID:         uuid.New(),
		URL:        input.URL,
		Name:       input.Name,
		SiteURL:    input.SiteURL,
		FaviconURL: input.FaviconURL,
		Tags:       input.Tags,
	}
	if err := s.sources.Create(ctx, src); err != nil {
		return nil, err
	}
	return src, nil
}

func (s *FeedService) ListSources(ctx context.Context) ([]domain.RSSSource, error) {
	return s.sources.List(ctx)
}
