package domain

import (
	"time"

	"github.com/google/uuid"
)

type Privacy string

const (
	PrivacyPublic  Privacy = "public"
	PrivacyPrivate Privacy = "private"
)

type User struct {
	ID             uuid.UUID      `json:"id"`
	AuthID         string         `json:"authId"`
	Username       string         `json:"username"`
	DisplayName    string         `json:"displayName"`
	AvatarURL      *string        `json:"avatarUrl"`
	BannerURL      *string        `json:"bannerUrl"`
	Bio            *string        `json:"bio"`
	Reputation     int            `json:"reputation"`
	Privacy        Privacy        `json:"privacy"`
	OnboardingDone bool           `json:"onboardingDone"`
	ShowActivity   bool           `json:"showActivity"`
	LastSeenAt     *time.Time     `json:"lastSeenAt"`
	BannedAt       *time.Time     `json:"bannedAt"`
	PostsCount     int            `json:"postsCount"`
	FollowersCount int            `json:"followersCount"`
	FollowingCount int            `json:"followingCount"`
	Tags           []string       `json:"tags"`
	Platforms      []UserPlatform `json:"platforms"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

type UserPlatform struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	Type       string    `json:"type"`
	Username   *string   `json:"username"`
	ProfileURL string    `json:"profileUrl"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UpdateUserInput struct {
	Username       *string  `json:"username"`
	DisplayName    *string  `json:"displayName"`
	Bio            *string  `json:"bio"`
	AvatarURL      *string  `json:"avatarUrl"`
	BannerURL      *string  `json:"bannerUrl"`
	Privacy        *Privacy `json:"privacy"`
	OnboardingDone *bool    `json:"onboardingDone"`
	ShowActivity   *bool    `json:"showActivity"`
}

type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipRejected FriendshipStatus = "rejected"
)

type Friendship struct {
	ID          uuid.UUID        `json:"id"`
	RequesterID uuid.UUID        `json:"requesterId"`
	AddresseeID uuid.UUID        `json:"addresseeId"`
	Status      FriendshipStatus `json:"status"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type ActivityItem struct {
	Type      string    `json:"type"` // "post", "comment"
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	RefID     uuid.UUID `json:"refId"`
	CreatedAt time.Time `json:"createdAt"`
}

type GitProvider string

const (
	GitProviderGitHub   GitProvider = "github"
	GitProviderGitLab   GitProvider = "gitlab"
	GitProviderCodeberg GitProvider = "codeberg"
)

type GitAccount struct {
	ID           uuid.UUID   `json:"id"`
	UserID       uuid.UUID   `json:"userId"`
	Provider     GitProvider `json:"provider"`
	AccessToken  string      `json:"-"`
	RefreshToken *string     `json:"-"`
	ExpiresAt    *time.Time  `json:"-"`
	Username     string      `json:"username"`
	InstanceURL  *string     `json:"instanceUrl"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

type PinnedRepo struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"userId"`
	GitAccountID uuid.UUID `json:"gitAccountId"`
	ExternalID   string    `json:"externalId"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	URL          string    `json:"url"`
	Language     *string   `json:"language"`
	StarsCount   int       `json:"starsCount"`
	ForksCount   int       `json:"forksCount"`
	IsFork       bool      `json:"isFork"`
	Topics       []string  `json:"topics"`
	SortOrder    int       `json:"sortOrder"`
	SyncedAt     time.Time `json:"syncedAt"`
}
