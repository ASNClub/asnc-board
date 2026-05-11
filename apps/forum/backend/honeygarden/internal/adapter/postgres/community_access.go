package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type CommunityAccessRepo struct {
	pool *pgxpool.Pool
}

func NewCommunityAccessRepo(pool *pgxpool.Pool) *CommunityAccessRepo {
	return &CommunityAccessRepo{pool: pool}
}

func (r *CommunityAccessRepo) GetCommunityIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM communities WHERE slug = $1`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, pgErr(err)
}

func (r *CommunityAccessRepo) IsOwner(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM communities WHERE id = $1 AND owner_id = $2)`,
		communityID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *CommunityAccessRepo) IsFollower(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM community_follows WHERE community_id = $1 AND user_id = $2)`,
		communityID, userID,
	).Scan(&exists)
	return exists, err
}
