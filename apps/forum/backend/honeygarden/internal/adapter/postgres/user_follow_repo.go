package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type UserFollowRepo struct {
	pool *pgxpool.Pool
}

func NewUserFollowRepo(pool *pgxpool.Pool) *UserFollowRepo {
	return &UserFollowRepo{pool: pool}
}

func (r *UserFollowRepo) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		followerID, followingID,
	)
	return pgErr(err)
}

func (r *UserFollowRepo) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM user_follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	return pgErr(err)
}

func (r *UserFollowRepo) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&exists)
	return exists, pgErr(err)
}

func (r *UserFollowRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error) {
	return r.queryUsers(ctx,
		`SELECT u.id, u.auth_id, u.username, u.display_name, u.avatar_url, u.banner_url, u.bio,
		        u.reputation, u.privacy, u.created_at, u.updated_at
		 FROM users u JOIN user_follows f ON f.follower_id = u.id
		 WHERE f.following_id = $1 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
}

func (r *UserFollowRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error) {
	return r.queryUsers(ctx,
		`SELECT u.id, u.auth_id, u.username, u.display_name, u.avatar_url, u.banner_url, u.bio,
		        u.reputation, u.privacy, u.created_at, u.updated_at
		 FROM users u JOIN user_follows f ON f.following_id = u.id
		 WHERE f.follower_id = $1 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
}

func (r *UserFollowRepo) queryUsers(ctx context.Context, query string, args ...any) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, query, args...)
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
