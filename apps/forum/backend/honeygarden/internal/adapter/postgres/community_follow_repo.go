package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type CommunityFollowRepo struct {
	pool *pgxpool.Pool
}

func NewCommunityFollowRepo(pool *pgxpool.Pool) *CommunityFollowRepo {
	return &CommunityFollowRepo{pool: pool}
}

func (r *CommunityFollowRepo) Follow(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO community_follows (user_id, community_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, communityID,
	)
	return tag.RowsAffected() > 0, pgErr(err)
}

func (r *CommunityFollowRepo) Unfollow(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM community_follows WHERE user_id = $1 AND community_id = $2`,
		userID, communityID,
	)
	return tag.RowsAffected() > 0, pgErr(err)
}

func (r *CommunityFollowRepo) IsFollowing(ctx context.Context, userID, communityID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM community_follows WHERE user_id = $1 AND community_id = $2)`,
		userID, communityID,
	).Scan(&exists)
	return exists, err
}

func (r *CommunityFollowRepo) GetFollowers(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.auth_id, u.username, u.display_name, u.avatar_url, u.banner_url, u.bio,
		        u.reputation, u.privacy, u.onboarding_done, u.created_at, u.updated_at
		 FROM community_follows f
		 JOIN users u ON u.id = f.user_id
		 WHERE f.community_id = $1
		 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
		communityID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(&u.ID, &u.AuthID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.BannerURL,
			&u.Bio, &u.Reputation, &u.Privacy, &u.OnboardingDone, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *CommunityFollowRepo) GetFollowed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Community, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.owner_id, c.slug, c.name, c.description, c.avatar_url, c.banner_url,
		        c.followers_count,
		        (SELECT COUNT(*) FROM posts WHERE community_id = c.id)::int AS posts_count,
		        c.stars_count, c.created_at, c.updated_at,
		        COALESCE(array_agg(t.tag) FILTER (WHERE t.tag IS NOT NULL), '{}') AS tags
		 FROM communities c
		 JOIN community_follows f ON f.community_id = c.id
		 LEFT JOIN community_tags t ON t.community_id = c.id
		 WHERE f.user_id = $1
		 GROUP BY c.id, f.created_at
		 ORDER BY f.created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	communities := []domain.Community{}
	for rows.Next() {
		var c domain.Community
		if err = rows.Scan(&c.ID, &c.OwnerID, &c.Slug, &c.Name, &c.Description,
			&c.AvatarURL, &c.BannerURL, &c.FollowersCount, &c.PostsCount,
			&c.StarsCount, &c.CreatedAt, &c.UpdatedAt, &c.Tags); err != nil {
			return nil, err
		}
		communities = append(communities, c)
	}
	return communities, nil
}
