package domain

import (
	"time"

	"github.com/google/uuid"
)

type Community struct {
	ID             uuid.UUID `json:"id"`
	OwnerID        uuid.UUID `json:"ownerId"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	AvatarURL      *string   `json:"avatarUrl"`
	BannerURL      *string   `json:"bannerUrl"`
	FollowersCount int       `json:"followersCount"`
	PostsCount     int       `json:"postsCount"`
	StarsCount     int       `json:"starsCount"`
	Tags           []string  `json:"tags"`
	Rules          []string  `json:"rules"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Followed       bool      `json:"followed,omitempty"`
	Starred        bool      `json:"starred,omitempty"`
}

type CreateCommunityInput struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Rules       []string `json:"rules"`
	BannerURL   *string  `json:"bannerUrl"`
	AvatarURL   *string  `json:"avatarUrl"`
}

type UpdateCommunityInput struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	AvatarURL   *string  `json:"avatarUrl"`
	BannerURL   *string  `json:"bannerUrl"`
	Rules       []string `json:"rules"`
}

type CommunityBan struct {
	CommunityID uuid.UUID  `json:"communityId"`
	UserID      uuid.UUID  `json:"userId"`
	Type        BanType    `json:"type"`
	Reason      *string    `json:"reason"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type CommunityModerator struct {
	CommunityID uuid.UUID `json:"communityId"`
	UserID      uuid.UUID `json:"userId"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
}

type BanType string

const (
	BanTypeBan  BanType = "ban"
	BanTypeMute BanType = "mute"
)
