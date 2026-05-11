package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type BanRepo struct {
	pool *pgxpool.Pool
}

func NewBanRepo(pool *pgxpool.Pool) *BanRepo {
	return &BanRepo{pool: pool}
}

func (r *BanRepo) Ban(ctx context.Context, b *domain.CommunityBan) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO community_bans (community_id, user_id, type, reason, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (community_id, user_id) DO UPDATE
		 SET type = EXCLUDED.type, reason = EXCLUDED.reason, expires_at = EXCLUDED.expires_at`,
		b.CommunityID, b.UserID, b.Type, b.Reason, b.ExpiresAt,
	)
	return pgErr(err)
}

func (r *BanRepo) Unban(ctx context.Context, communityID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM community_bans WHERE community_id = $1 AND user_id = $2`,
		communityID, userID,
	)
	return pgErr(err)
}

func (r *BanRepo) IsBanned(ctx context.Context, communityID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
		     SELECT 1 FROM community_bans
		     WHERE community_id = $1 AND user_id = $2
		       AND (expires_at IS NULL OR expires_at > NOW())
		 )`,
		communityID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *BanRepo) GetBans(ctx context.Context, communityID uuid.UUID) ([]domain.CommunityBan, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT community_id, user_id, type, reason, expires_at, created_at
		 FROM community_bans WHERE community_id = $1 ORDER BY created_at DESC`,
		communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bans := []domain.CommunityBan{}
	for rows.Next() {
		var b domain.CommunityBan
		if err = rows.Scan(&b.CommunityID, &b.UserID, &b.Type, &b.Reason, &b.ExpiresAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		bans = append(bans, b)
	}
	return bans, nil
}
