package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type EntityLookupRepo struct {
	pool *pgxpool.Pool
}

func NewEntityLookupRepo(pool *pgxpool.Pool) *EntityLookupRepo {
	return &EntityLookupRepo{pool: pool}
}

func (r *EntityLookupRepo) GetPostAuthorID(ctx context.Context, postID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT author_id FROM posts WHERE id = $1`, postID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, pgErr(err)
}

func (r *EntityLookupRepo) GetCommunityOwnerID(ctx context.Context, communityID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT owner_id FROM communities WHERE id = $1`, communityID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, pgErr(err)
}
