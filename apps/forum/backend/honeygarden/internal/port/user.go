package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByAuthID(ctx context.Context, authID string) (*domain.User, error)
	Search(ctx context.Context, query string, limit int) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetTags(ctx context.Context, userID uuid.UUID, tags []string) error
	SetPlatforms(ctx context.Context, userID uuid.UUID, platforms []domain.UserPlatform) error
	TouchLastSeen(ctx context.Context, id uuid.UUID) error
	SetBanned(ctx context.Context, id uuid.UUID, ban bool) error
}

type UserFollowRepository interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error)
}

type UserBlockRepository interface {
	Block(ctx context.Context, blockerID, blockedID uuid.UUID) error
	Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlockedEither(ctx context.Context, a, b uuid.UUID) (bool, error)
	ListBlocks(ctx context.Context, userID uuid.UUID) ([]domain.User, error)
	ListBlockSet(ctx context.Context, viewer uuid.UUID) (map[uuid.UUID]bool, error)
}

type FriendshipRepository interface {
	Create(ctx context.Context, f *domain.Friendship) error
	UpdateStatus(ctx context.Context, requesterID, addresseeID uuid.UUID, status domain.FriendshipStatus) error
	Get(ctx context.Context, userA, userB uuid.UUID) (*domain.Friendship, error)
	GetFriends(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error)
	GetPending(ctx context.Context, addresseeID uuid.UUID) ([]domain.Friendship, error)
}

type GitAccountRepository interface {
	Upsert(ctx context.Context, a *domain.GitAccount) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.GitAccount, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GitAccount, error)
	GetByUserProvider(ctx context.Context, userID uuid.UUID, provider domain.GitProvider, instanceURL *string) (*domain.GitAccount, error)
	UpdateTokens(ctx context.Context, id uuid.UUID, accessToken string, refreshToken *string, expiresAt *time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PinnedRepoRepository interface {
	Upsert(ctx context.Context, r *domain.PinnedRepo) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.PinnedRepo, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	UpdateOrder(ctx context.Context, userID uuid.UUID, repoIDs []uuid.UUID) error
}

type RepoData struct {
	ExternalID  string   `json:"externalId"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	URL         string   `json:"url"`
	Language    *string  `json:"language"`
	StarsCount  int      `json:"starsCount"`
	ForksCount  int      `json:"forksCount"`
	IsFork      bool     `json:"isFork"`
	Topics      []string `json:"topics"`
}

type GitProviderClient interface {
	FetchRepos(ctx context.Context, token string, instanceURL *string) ([]RepoData, error)
	GetUsername(ctx context.Context, token string, instanceURL *string) (string, error)
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time // nil = бессрочный
}

type OAuthProvider interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*OAuthToken, error)
	Refresh(ctx context.Context, refreshToken string) (*OAuthToken, error)
}
