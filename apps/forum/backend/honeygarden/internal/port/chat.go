package port

import (
	"context"

	"honeygarden/internal/domain"
)

type ChatRepository interface {
	Insert(ctx context.Context, m *domain.ChatMessage) error
	List(ctx context.Context, limit int) ([]domain.ChatMessage, error)
}
