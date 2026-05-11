package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.scanUser(ctx,
		`SELECT id, auth_id, username, display_name, avatar_url, banner_url, bio,
		        reputation, privacy, onboarding_done, show_activity, last_seen_at, banned_at, created_at, updated_at
		 FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return r.loadRelations(ctx, u)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, err := r.scanUser(ctx,
		`SELECT id, auth_id, username, display_name, avatar_url, banner_url, bio,
		        reputation, privacy, onboarding_done, show_activity, last_seen_at, banned_at, created_at, updated_at
		 FROM users WHERE username = $1`, username)
	if err != nil {
		return nil, err
	}
	return r.loadRelations(ctx, u)
}

func (r *UserRepo) GetByAuthID(ctx context.Context, authID string) (*domain.User, error) {
	u, err := r.scanUser(ctx,
		`SELECT id, auth_id, username, display_name, avatar_url, banner_url, bio,
		        reputation, privacy, onboarding_done, show_activity, last_seen_at, banned_at, created_at, updated_at
		 FROM users WHERE auth_id = $1`, authID)
	if err != nil {
		return nil, err
	}
	return r.loadRelations(ctx, u)
}

func (r *UserRepo) Search(ctx context.Context, query string, limit int) ([]domain.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	pattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx,
		`SELECT id, auth_id, username, display_name, avatar_url, banner_url, bio,
		        reputation, privacy, onboarding_done, show_activity, last_seen_at, banned_at, created_at, updated_at
		 FROM users
		 WHERE (username ILIKE $1 OR display_name ILIKE $1) AND banned_at IS NULL
		 ORDER BY reputation DESC, username ASC
		 LIMIT $2`, pattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		u := domain.User{}
		if err := rows.Scan(&u.ID, &u.AuthID, &u.Username, &u.DisplayName,
			&u.AvatarURL, &u.BannerURL, &u.Bio, &u.Reputation, &u.Privacy,
			&u.OnboardingDone, &u.ShowActivity, &u.LastSeenAt, &u.BannedAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) SetBanned(ctx context.Context, id uuid.UUID, ban bool) error {
	var err error
	if ban {
		_, err = r.pool.Exec(ctx, `UPDATE users SET banned_at = NOW() WHERE id = $1`, id)
	} else {
		_, err = r.pool.Exec(ctx, `UPDATE users SET banned_at = NULL WHERE id = $1`, id)
	}
	return pgErr(err)
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, auth_id, username, display_name, avatar_url, banner_url, bio, reputation, privacy, onboarding_done, show_activity, last_seen_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		u.ID, u.AuthID, u.Username, u.DisplayName, u.AvatarURL, u.BannerURL, u.Bio, u.Reputation, u.Privacy, u.OnboardingDone, u.ShowActivity, u.LastSeenAt,
	)
	return pgErr(err)
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET username=$2, display_name=$3, avatar_url=$4, banner_url=$5, bio=$6, privacy=$7, onboarding_done=$8, show_activity=$9, last_seen_at=$10, updated_at=NOW()
		 WHERE id = $1`,
		u.ID, u.Username, u.DisplayName, u.AvatarURL, u.BannerURL, u.Bio, u.Privacy, u.OnboardingDone, u.ShowActivity, u.LastSeenAt,
	)
	return pgErr(err)
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return pgErr(err)
}

func (r *UserRepo) TouchLastSeen(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_seen_at = NOW() WHERE id = $1`, id)
	return pgErr(err)
}

func (r *UserRepo) SetTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM user_tags WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err = tx.Exec(ctx, `INSERT INTO user_tags (user_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, tag); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *UserRepo) SetPlatforms(ctx context.Context, userID uuid.UUID, platforms []domain.UserPlatform) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM user_platforms WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, p := range platforms {
		if _, err = tx.Exec(ctx,
			`INSERT INTO user_platforms (id, user_id, type, username, profile_url) VALUES ($1, $2, $3, $4, $5)`,
			p.ID, userID, p.Type, p.Username, p.ProfileURL,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *UserRepo) scanUser(ctx context.Context, query string, args ...any) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.AuthID, &u.Username, &u.DisplayName,
		&u.AvatarURL, &u.BannerURL, &u.Bio, &u.Reputation, &u.Privacy,
		&u.OnboardingDone, &u.ShowActivity, &u.LastSeenAt, &u.BannedAt, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, pgErr(err)
}

func (r *UserRepo) loadRelations(ctx context.Context, u *domain.User) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT tag FROM user_tags WHERE user_id = $1`, u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	u.Tags = []string{}
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, err
		}
		u.Tags = append(u.Tags, tag)
	}

	pRows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, username, profile_url, created_at FROM user_platforms WHERE user_id = $1`, u.ID)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()
	u.Platforms = []domain.UserPlatform{}
	for pRows.Next() {
		var p domain.UserPlatform
		if err = pRows.Scan(&p.ID, &p.UserID, &p.Type, &p.Username, &p.ProfileURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		u.Platforms = append(u.Platforms, p)
	}

	err = r.pool.QueryRow(ctx,
		`SELECT
		   (SELECT COUNT(*) FROM posts WHERE author_id = $1),
		   (SELECT COUNT(*) FROM user_follows WHERE following_id = $1),
		   (SELECT COUNT(*) FROM user_follows WHERE follower_id  = $1)`,
		u.ID,
	).Scan(&u.PostsCount, &u.FollowersCount, &u.FollowingCount)
	if err != nil {
		return nil, err
	}
	return u, nil
}
