package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type CommunityRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Community, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Community, error)
	GetByOwner(ctx context.Context, ownerID uuid.UUID) (*domain.Community, error)
	Create(ctx context.Context, c *domain.Community) error
	Update(ctx context.Context, c *domain.Community) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetTags(ctx context.Context, communityID uuid.UUID, tags []string) error
	IncrFollowers(ctx context.Context, communityID uuid.UUID, delta int) error
	IncrStars(ctx context.Context, communityID uuid.UUID, delta int) error
	List(ctx context.Context, sort string, limit, offset int) ([]domain.Community, error)
}

type CommunityFollowRepository interface {
	Follow(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	Unfollow(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	IsFollowing(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	GetFollowed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Community, error)
	GetFollowers(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]domain.User, error)
}

type StarRepository interface {
	Star(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	Unstar(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	IsStarred(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
}

type ModeratorRepository interface {
	Add(ctx context.Context, communityID, userID uuid.UUID, role string) error
	Remove(ctx context.Context, communityID, userID uuid.UUID) error
	GetByCommunity(ctx context.Context, communityID uuid.UUID) ([]domain.CommunityModerator, error)
	IsModerator(ctx context.Context, communityID, userID uuid.UUID) (bool, error)
}

type BanRepository interface {
	Ban(ctx context.Context, b *domain.CommunityBan) error
	Unban(ctx context.Context, communityID, userID uuid.UUID) error
	IsBanned(ctx context.Context, communityID, userID uuid.UUID) (bool, error)
	GetBans(ctx context.Context, communityID uuid.UUID) ([]domain.CommunityBan, error)
}
