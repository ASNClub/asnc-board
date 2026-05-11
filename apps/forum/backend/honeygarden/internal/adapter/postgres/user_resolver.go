package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type UserResolverRepo struct {
	pool *pgxpool.Pool
}

func NewUserResolverRepo(pool *pgxpool.Pool) *UserResolverRepo {
	return &UserResolverRepo{pool: pool}
}

func (r *UserResolverRepo) ResolveUsers(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.UserBrief, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]domain.UserBrief{}, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, display_name, avatar_url, reputation FROM users WHERE id = ANY($1) AND banned_at IS NULL`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]domain.UserBrief, len(ids))
	for rows.Next() {
		var u domain.UserBrief
		if err = rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.Reputation); err != nil {
			return nil, err
		}
		result[u.ID] = u
	}
	return result, rows.Err()
}

func (r *UserResolverRepo) ResolveUsernames(ctx context.Context, usernames []string) (map[string]uuid.UUID, error) {
	if len(usernames) == 0 {
		return map[string]uuid.UUID{}, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, username FROM users WHERE username = ANY($1)`,
		usernames,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]uuid.UUID, len(usernames))
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err = rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, rows.Err()
}

func (r *UserResolverRepo) ResolveCommunitySlugs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, slug FROM communities WHERE id = ANY($1)`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]string, len(ids))
	for rows.Next() {
		var id uuid.UUID
		var slug string
		if err = rows.Scan(&id, &slug); err != nil {
			return nil, err
		}
		result[id] = slug
	}
	return result, rows.Err()
}
