package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StarRepo struct {
	pool *pgxpool.Pool
}

func NewStarRepo(pool *pgxpool.Pool) *StarRepo {
	return &StarRepo{pool: pool}
}

func (r *StarRepo) Star(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO community_stars (user_id, community_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, communityID,
	)
	return tag.RowsAffected() > 0, pgErr(err)
}

func (r *StarRepo) Unstar(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM community_stars WHERE user_id = $1 AND community_id = $2`,
		userID, communityID,
	)
	return tag.RowsAffected() > 0, pgErr(err)
}

func (r *StarRepo) IsStarred(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM community_stars WHERE user_id = $1 AND community_id = $2)`,
		userID, communityID,
	).Scan(&exists)
	return exists, err
}
