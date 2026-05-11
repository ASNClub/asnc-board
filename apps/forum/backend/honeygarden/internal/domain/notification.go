package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"userId"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	IsRead    bool            `json:"isRead"`
	CreatedAt time.Time       `json:"createdAt"`

	Actor     *NotifActor    `json:"actor,omitempty"`
	Post      *NotifPostRef  `json:"post,omitempty"`
	Community *NotifCommRef  `json:"community,omitempty"`
	Snippet   string         `json:"snippet,omitempty"`
}

type NotifActor struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl,omitempty"`
}

type NotifPostRef struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	CommunitySlug *string   `json:"communitySlug,omitempty"`
}

type NotifCommRef struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatarUrl,omitempty"`
}

type NotificationPreference struct {
	UserID  uuid.UUID `json:"userId"`
	Type    string    `json:"type"`
	Enabled bool      `json:"enabled"`
}
