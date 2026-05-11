package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type BadgeRepository interface {
	ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error)
	GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error)
	Award(ctx context.Context, userID uuid.UUID, badgeID string) error
	HasBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error)
}

type BadgeStatsProvider interface {
	UserPostCount(ctx context.Context, userID uuid.UUID) (int, error)
	UserCommentCount(ctx context.Context, userID uuid.UUID) (int, error)
	UserFollowedCommunitiesCount(ctx context.Context, userID uuid.UUID) (int, error)
	UserMaxPostVotes(ctx context.Context, userID uuid.UUID) (int, error)
	UserReputation(ctx context.Context, userID uuid.UUID) (int, error)
	UserCommunityFollowers(ctx context.Context, userID uuid.UUID) (int, error)
	UserCommunityStars(ctx context.Context, userID uuid.UUID) (int, error)
	HasCommunity(ctx context.Context, userID uuid.UUID) (bool, error)
}
