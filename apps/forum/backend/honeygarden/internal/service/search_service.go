package service

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type SearchService struct {
	index port.SearchIndex
	log   zerolog.Logger
}

func NewSearchService(index port.SearchIndex, log zerolog.Logger) *SearchService {
	return &SearchService{index: index, log: log}
}

func (s *SearchService) SearchPosts(ctx context.Context, q string, limit, offset int) (*domain.SearchResult[domain.PostDocument], error) {
	return s.index.SearchPosts(ctx, q, limit, offset)
}

func (s *SearchService) SearchCommunities(ctx context.Context, q string, limit, offset int) (*domain.SearchResult[domain.CommunityDocument], error) {
	return s.index.SearchCommunities(ctx, q, limit, offset)
}

func (s *SearchService) HandleEvent(ctx context.Context, subject string, data []byte) error {
	switch subject {
	case "post.created":
		var p struct {
			PostID      string  `json:"post_id"`
			CommunityID string  `json:"community_id"`
			AuthorID    string  `json:"author_id"`
			Title       *string `json:"title"`
			Content     string  `json:"content"`
			CreatedAt   int64   `json:"created_at"`
		}
		if err := json.Unmarshal(data, &p); err != nil {
			s.log.Warn().Err(err).Msg("search: bad post.created payload")
			return nil
		}
		doc := domain.PostDocument{
			ID:          p.PostID,
			CommunityID: p.CommunityID,
			AuthorID:    p.AuthorID,
			Title:       p.Title,
			Content:     p.Content,
			CreatedAt:   p.CreatedAt,
		}
		if err := s.index.IndexPost(ctx, doc); err != nil {
			s.log.Error().Err(err).Str("post_id", p.PostID).Msg("search: index post failed")
			return err
		}
	case "community.created":
		var c struct {
			CommunityID string  `json:"community_id"`
			Slug        string  `json:"slug"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
		}
		if err := json.Unmarshal(data, &c); err != nil {
			s.log.Warn().Err(err).Msg("search: bad community.created payload")
			return nil
		}
		doc := domain.CommunityDocument{
			ID:          c.CommunityID,
			Slug:        c.Slug,
			Name:        c.Name,
			Description: c.Description,
		}
		if err := s.index.IndexCommunity(ctx, doc); err != nil {
			s.log.Error().Err(err).Str("community_id", c.CommunityID).Msg("search: index community failed")
			return err
		}
	}
	return nil
}
