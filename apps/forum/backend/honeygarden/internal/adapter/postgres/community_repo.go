package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type CommunityRepo struct {
	pool *pgxpool.Pool
}

func NewCommunityRepo(pool *pgxpool.Pool) *CommunityRepo {
	return &CommunityRepo{pool: pool}
}

func (r *CommunityRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Community, error) {
	c, err := r.scan(ctx,
		`SELECT id, owner_id, slug, name, description, avatar_url, banner_url,
		        followers_count,
		        (SELECT COUNT(*) FROM posts WHERE community_id = communities.id)::int AS posts_count,
		        stars_count, rules, created_at, updated_at
		 FROM communities WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return r.withTags(ctx, c)
}

func (r *CommunityRepo) GetBySlug(ctx context.Context, slug string) (*domain.Community, error) {
	c, err := r.scan(ctx,
		`SELECT id, owner_id, slug, name, description, avatar_url, banner_url,
		        followers_count,
		        (SELECT COUNT(*) FROM posts WHERE community_id = communities.id)::int AS posts_count,
		        stars_count, rules, created_at, updated_at
		 FROM communities WHERE slug = $1`, slug)
	if err != nil {
		return nil, err
	}
	return r.withTags(ctx, c)
}

func (r *CommunityRepo) GetByOwner(ctx context.Context, ownerID uuid.UUID) (*domain.Community, error) {
	c, err := r.scan(ctx,
		`SELECT id, owner_id, slug, name, description, avatar_url, banner_url,
		        followers_count,
		        (SELECT COUNT(*) FROM posts WHERE community_id = communities.id)::int AS posts_count,
		        stars_count, rules, created_at, updated_at
		 FROM communities WHERE owner_id = $1`, ownerID)
	if err != nil {
		return nil, err
	}
	return r.withTags(ctx, c)
}

func (r *CommunityRepo) Create(ctx context.Context, c *domain.Community) error {
	rules := c.Rules
	if rules == nil {
		rules = []string{}
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO communities (id, owner_id, slug, name, description, avatar_url, banner_url, rules)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.OwnerID, c.Slug, c.Name, c.Description, c.AvatarURL, c.BannerURL, rules,
	)
	return pgErr(err)
}

func (r *CommunityRepo) Update(ctx context.Context, c *domain.Community) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE communities SET name=$2, description=$3, avatar_url=$4, banner_url=$5, rules=$6, updated_at=NOW() WHERE id = $1`,
		c.ID, c.Name, c.Description, c.AvatarURL, c.BannerURL, c.Rules,
	)
	return pgErr(err)
}

func (r *CommunityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM communities WHERE id = $1`, id)
	return pgErr(err)
}

func (r *CommunityRepo) SetTags(ctx context.Context, communityID uuid.UUID, tags []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM community_tags WHERE community_id = $1`, communityID); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err = tx.Exec(ctx,
			`INSERT INTO community_tags (community_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			communityID, tag,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CommunityRepo) IncrFollowers(ctx context.Context, communityID uuid.UUID, delta int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE communities SET followers_count = GREATEST(0, followers_count + $2) WHERE id = $1`,
		communityID, delta,
	)
	return pgErr(err)
}

func (r *CommunityRepo) IncrStars(ctx context.Context, communityID uuid.UUID, delta int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE communities SET stars_count = GREATEST(0, stars_count + $2) WHERE id = $1`,
		communityID, delta,
	)
	return pgErr(err)
}

func (r *CommunityRepo) List(ctx context.Context, sort string, limit, offset int) ([]domain.Community, error) {
	orderBy := "followers_count DESC, created_at DESC"
	switch sort {
	case "new":
		orderBy = "created_at DESC"
	case "active":
		orderBy = "(SELECT COUNT(*) FROM posts WHERE community_id = communities.id AND created_at > NOW() - INTERVAL '7 days') DESC, followers_count DESC"
	case "popular":
		orderBy = "followers_count DESC, created_at DESC"
	}
	q := `SELECT id, owner_id, slug, name, description, avatar_url, banner_url,
	             followers_count,
	             (SELECT COUNT(*) FROM posts WHERE community_id = communities.id)::int AS posts_count,
	             stars_count, rules, created_at, updated_at
	      FROM communities
	      ORDER BY ` + orderBy + `
	      LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	out := []domain.Community{}
	for rows.Next() {
		c := domain.Community{}
		if err = rows.Scan(&c.ID, &c.OwnerID, &c.Slug, &c.Name, &c.Description,
			&c.AvatarURL, &c.BannerURL, &c.FollowersCount, &c.PostsCount,
			&c.StarsCount, &c.Rules, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if c.Rules == nil {
			c.Rules = []string{}
		}
		c.Tags = []string{}
		out = append(out, c)
	}
	return out, nil
}

func (r *CommunityRepo) scan(ctx context.Context, query string, args ...any) (*domain.Community, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	c := &domain.Community{}
	err := row.Scan(&c.ID, &c.OwnerID, &c.Slug, &c.Name, &c.Description,
		&c.AvatarURL, &c.BannerURL, &c.FollowersCount, &c.PostsCount,
		&c.StarsCount, &c.Rules, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if c.Rules == nil {
		c.Rules = []string{}
	}
	return c, pgErr(err)
}

func (r *CommunityRepo) withTags(ctx context.Context, c *domain.Community) (*domain.Community, error) {
	rows, err := r.pool.Query(ctx, `SELECT tag FROM community_tags WHERE community_id = $1`, c.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	c.Tags = []string{}
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, err
		}
		c.Tags = append(c.Tags, tag)
	}
	return c, nil
}
