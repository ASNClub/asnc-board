package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type UserResolver interface {
	ResolveUsers(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.UserBrief, error)
	ResolveUsernames(ctx context.Context, usernames []string) (map[string]uuid.UUID, error)
}

type CommunitySlugResolver interface {
	ResolveCommunitySlugs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
}
