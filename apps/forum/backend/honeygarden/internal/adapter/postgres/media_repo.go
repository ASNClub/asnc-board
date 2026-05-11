package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type MediaRepo struct {
	pool *pgxpool.Pool
}

func NewMediaRepo(pool *pgxpool.Pool) *MediaRepo {
	return &MediaRepo{pool: pool}
}

func (r *MediaRepo) Create(ctx context.Context, m *domain.PostMedia) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO post_media (id, post_id, type, url, name, size) VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.PostID, m.Type, m.URL, m.Name, m.Size,
	)
	return pgErr(err)
}

func (r *MediaRepo) GetByPost(ctx context.Context, postID uuid.UUID) ([]domain.PostMedia, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, post_id, type, url, name, size, created_at FROM post_media WHERE post_id = $1 ORDER BY created_at`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	media := []domain.PostMedia{}
	for rows.Next() {
		var m domain.PostMedia
		if err = rows.Scan(&m.ID, &m.PostID, &m.Type, &m.URL, &m.Name, &m.Size, &m.CreatedAt); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

func (r *MediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM post_media WHERE id = $1`, id)
	return pgErr(err)
}
