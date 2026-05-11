package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BadgeStatsRepo struct {
	pool *pgxpool.Pool
}

func NewBadgeStatsRepo(pool *pgxpool.Pool) *BadgeStatsRepo {
	return &BadgeStatsRepo{pool: pool}
}

func (r *BadgeStatsRepo) UserPostCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM posts WHERE author_id = $1`, userID)
}

func (r *BadgeStatsRepo) UserCommentCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM comments WHERE author_id = $1`, userID)
}

func (r *BadgeStatsRepo) UserFollowedCommunitiesCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM community_follows WHERE user_id = $1`, userID)
}

func (r *BadgeStatsRepo) UserMaxPostVotes(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(votes_count), 0) FROM posts WHERE author_id = $1`, userID,
	).Scan(&n)
	return n, err
}

func (r *BadgeStatsRepo) UserReputation(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT reputation FROM users WHERE id = $1`, userID,
	).Scan(&n)
	return n, err
}

func (r *BadgeStatsRepo) UserCommunityFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(followers_count), 0) FROM communities WHERE owner_id = $1`, userID,
	).Scan(&n)
	return n, err
}

func (r *BadgeStatsRepo) UserCommunityStars(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(stars_count), 0) FROM communities WHERE owner_id = $1`, userID,
	).Scan(&n)
	return n, err
}

func (r *BadgeStatsRepo) HasCommunity(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM communities WHERE owner_id = $1)`, userID,
	).Scan(&exists)
	return exists, err
}

func (r *BadgeStatsRepo) count(ctx context.Context, query string, args ...any) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&n)
	return n, err
}
