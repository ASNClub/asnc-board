package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type ModeratorRepo struct {
	pool *pgxpool.Pool
}

func NewModeratorRepo(pool *pgxpool.Pool) *ModeratorRepo {
	return &ModeratorRepo{pool: pool}
}

func (r *ModeratorRepo) Add(ctx context.Context, communityID, userID uuid.UUID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO community_moderators (community_id, user_id, role)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (community_id, user_id) DO UPDATE SET role = $3`,
		communityID, userID, role,
	)
	return pgErr(err)
}

func (r *ModeratorRepo) Remove(ctx context.Context, communityID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM community_moderators WHERE community_id = $1 AND user_id = $2`,
		communityID, userID,
	)
	return pgErr(err)
}

func (r *ModeratorRepo) GetByCommunity(ctx context.Context, communityID uuid.UUID) ([]domain.CommunityModerator, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT cm.community_id, cm.user_id, u.username, cm.role, cm.created_at
		 FROM community_moderators cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.community_id = $1
		 ORDER BY cm.created_at`,
		communityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mods := []domain.CommunityModerator{}
	for rows.Next() {
		var m domain.CommunityModerator
		if err = rows.Scan(&m.CommunityID, &m.UserID, &m.Username, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, rows.Err()
}

func (r *ModeratorRepo) IsModerator(ctx context.Context, communityID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM community_moderators WHERE community_id = $1 AND user_id = $2)`,
		communityID, userID,
	).Scan(&exists)
	return exists, err
}
