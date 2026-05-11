package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type BadgeRepo struct {
	pool *pgxpool.Pool
}

func NewBadgeRepo(pool *pgxpool.Pool) *BadgeRepo {
	return &BadgeRepo{pool: pool}
}

func (r *BadgeRepo) ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, glyph, name, name_ru, description, color, rarity, sort_order
		 FROM badge_definitions ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	defs := []domain.BadgeDefinition{}
	for rows.Next() {
		var d domain.BadgeDefinition
		if err = rows.Scan(&d.ID, &d.Glyph, &d.Name, &d.NameRu, &d.Description, &d.Color, &d.Rarity, &d.SortOrder); err != nil {
			return nil, err
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (r *BadgeRepo) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ub.user_id, ub.earned_at,
		        bd.id, bd.glyph, bd.name, bd.name_ru, bd.description, bd.color, bd.rarity, bd.sort_order
		 FROM user_badges ub
		 JOIN badge_definitions bd ON bd.id = ub.badge_id
		 WHERE ub.user_id = $1
		 ORDER BY bd.sort_order`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	badges := []domain.UserBadge{}
	for rows.Next() {
		var b domain.UserBadge
		if err = rows.Scan(&b.UserID, &b.EarnedAt,
			&b.Badge.ID, &b.Badge.Glyph, &b.Badge.Name, &b.Badge.NameRu,
			&b.Badge.Description, &b.Badge.Color, &b.Badge.Rarity, &b.Badge.SortOrder); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}

func (r *BadgeRepo) Award(ctx context.Context, userID uuid.UUID, badgeID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_badges (user_id, badge_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, badgeID,
	)
	return pgErr(err)
}

func (r *BadgeRepo) HasBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_badges WHERE user_id = $1 AND badge_id = $2)`,
		userID, badgeID,
	).Scan(&exists)
	return exists, err
}
