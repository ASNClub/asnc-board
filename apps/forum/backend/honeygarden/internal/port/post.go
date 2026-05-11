package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type PostRepository interface {
	Create(ctx context.Context, p *domain.Post) error
	UpsertExternal(ctx context.Context, e *domain.ExternalPost) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	GetByShortID(ctx context.Context, shortID string) (*domain.Post, error)
	GetByCommunity(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]domain.Post, error)
	GetByCommunityKind(ctx context.Context, communityID uuid.UUID, kind domain.PostKind, limit, offset int) ([]domain.Post, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error)
	Update(ctx context.Context, p *domain.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrViews(ctx context.Context, id uuid.UUID) error
	AddVote(ctx context.Context, userID, postID uuid.UUID) error
	RemoveVote(ctx context.Context, userID, postID uuid.UUID) error
	IsVoted(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	BatchIsVoted(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error
	GetTrending(ctx context.Context, since time.Time, limit int) ([]domain.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, c *domain.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	GetByPost(ctx context.Context, postID uuid.UUID, limit, offset int) ([]domain.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddVote(ctx context.Context, userID, commentID uuid.UUID) error
	RemoveVote(ctx context.Context, userID, commentID uuid.UUID) error
}

type MediaRepository interface {
	Create(ctx context.Context, m *domain.PostMedia) error
	GetByPost(ctx context.Context, postID uuid.UUID) ([]domain.PostMedia, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type BookmarkRepository interface {
	Add(ctx context.Context, userID, postID uuid.UUID) error
	Remove(ctx context.Context, userID, postID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error)
	IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	BatchIsBookmarked(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}

type ActivityRepository interface {
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ActivityItem, error)
}

type CommunityAccess interface {
	GetCommunityIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	IsOwner(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
	IsFollower(ctx context.Context, userID, communityID uuid.UUID) (bool, error)
}
