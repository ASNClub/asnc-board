package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type ActivityRepo struct {
	pool *pgxpool.Pool
}

func NewActivityRepo(pool *pgxpool.Pool) *ActivityRepo {
	return &ActivityRepo{pool: pool}
}

func (r *ActivityRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.ActivityItem, error) {
	rows, err := r.pool.Query(ctx, `
		(
			SELECT 'post' AS type, id, COALESCE(title, '') AS title, LEFT(content, 200) AS content, id AS ref_id, created_at
			FROM posts WHERE author_id = $1
		)
		UNION ALL
		(
			SELECT 'comment' AS type, id, '' AS title, LEFT(content, 200) AS content, post_id AS ref_id, created_at
			FROM comments WHERE author_id = $1
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ActivityItem{}
	for rows.Next() {
		var item domain.ActivityItem
		if err = rows.Scan(&item.Type, &item.ID, &item.Title, &item.Content, &item.RefID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
