package port

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
)

type FeedbackRepository interface {
	Create(ctx context.Context, f *domain.Feedback) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Feedback, error)
	List(ctx context.Context, sort string, status string, limit, offset int) ([]domain.Feedback, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FeedbackStatus) error
	Delete(ctx context.Context, id uuid.UUID) error

	Vote(ctx context.Context, userID, feedbackID uuid.UUID) (bool, error)
	Unvote(ctx context.Context, userID, feedbackID uuid.UUID) (bool, error)
	BatchIsVoted(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error)
	CountByAuthorSince(ctx context.Context, authorID uuid.UUID, sinceMinutes int) (int, error)
}
