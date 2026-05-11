package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
}

type NotificationPreferenceRepository interface {
	Get(ctx context.Context, userID uuid.UUID) ([]domain.NotificationPreference, error)
	Set(ctx context.Context, userID uuid.UUID, notifType string, enabled bool) error
	IsEnabled(ctx context.Context, userID uuid.UUID, notifType string) (bool, error)
}

type EntityLookup interface {
	GetPostAuthorID(ctx context.Context, postID uuid.UUID) (uuid.UUID, error)
	GetCommunityOwnerID(ctx context.Context, communityID uuid.UUID) (uuid.UUID, error)
}
