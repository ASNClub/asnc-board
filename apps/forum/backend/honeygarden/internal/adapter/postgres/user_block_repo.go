package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type UserBlockRepo struct {
	pool *pgxpool.Pool
}

func NewUserBlockRepo(pool *pgxpool.Pool) *UserBlockRepo {
	return &UserBlockRepo{pool: pool}
}

func (r *UserBlockRepo) Block(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_blocks (blocker_id, blocked_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		blockerID, blockedID,
	)
	return pgErr(err)
}

func (r *UserBlockRepo) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2`,
		blockerID, blockedID,
	)
	return pgErr(err)
}

func (r *UserBlockRepo) IsBlockedEither(ctx context.Context, a, b uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_blocks
		   WHERE (blocker_id = $1 AND blocked_id = $2)
		      OR (blocker_id = $2 AND blocked_id = $1))`,
		a, b,
	).Scan(&exists)
	return exists, pgErr(err)
}

func (r *UserBlockRepo) ListBlockSet(ctx context.Context, viewer uuid.UUID) (map[uuid.UUID]bool, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT blocked_id FROM user_blocks WHERE blocker_id = $1
		 UNION
		 SELECT blocker_id FROM user_blocks WHERE blocked_id = $1`,
		viewer,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	out := map[uuid.UUID]bool{}
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// ListBlocks — кого я заблокировал(кого?)
func (r *UserBlockRepo) ListBlocks(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.auth_id, u.username, u.display_name, u.avatar_url, u.banner_url, u.bio,
		        u.reputation, u.privacy, u.created_at, u.updated_at
		   FROM users u
		   JOIN user_blocks b ON b.blocked_id = u.id
		  WHERE b.blocker_id = $1
		  ORDER BY b.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(&u.ID, &u.AuthID, &u.Username, &u.DisplayName,
			&u.AvatarURL, &u.BannerURL, &u.Bio, &u.Reputation, &u.Privacy,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Tags = []string{}
		u.Platforms = []domain.UserPlatform{}
		users = append(users, u)
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, nil
}
