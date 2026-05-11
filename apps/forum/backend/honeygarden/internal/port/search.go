package port

import (
	"context"

	"honeygarden/internal/domain"
)

type SearchIndex interface {
	SearchPosts(ctx context.Context, q string, limit, offset int) (*domain.SearchResult[domain.PostDocument], error)
	SearchCommunities(ctx context.Context, q string, limit, offset int) (*domain.SearchResult[domain.CommunityDocument], error)
	IndexPost(ctx context.Context, doc domain.PostDocument) error
	IndexCommunity(ctx context.Context, doc domain.CommunityDocument) error
	DeletePost(ctx context.Context, id string) error
	DeleteCommunity(ctx context.Context, id string) error
}
