package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type ChatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) *ChatRepo {
	return &ChatRepo{pool: pool}
}

func (r *ChatRepo) Insert(ctx context.Context, m *domain.ChatMessage) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO chat_messages (id, author_id, content) VALUES ($1, $2, $3)`,
		m.ID, m.AuthorID, m.Content,
	)
	return pgErr(err)
}

func (r *ChatRepo) List(ctx context.Context, limit int) ([]domain.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, author_id, content, created_at
		 FROM chat_messages
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ChatMessage{}
	for rows.Next() {
		var m domain.ChatMessage
		if err = rows.Scan(&m.ID, &m.AuthorID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
