package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type NotifPrefRepo struct {
	pool *pgxpool.Pool
}

func NewNotifPrefRepo(pool *pgxpool.Pool) *NotifPrefRepo {
	return &NotifPrefRepo{pool: pool}
}

func (r *NotifPrefRepo) Get(ctx context.Context, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, type, enabled FROM notification_settings WHERE user_id = $1 ORDER BY type`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := []domain.NotificationPreference{}
	for rows.Next() {
		var p domain.NotificationPreference
		if err = rows.Scan(&p.UserID, &p.Type, &p.Enabled); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, rows.Err()
}

func (r *NotifPrefRepo) Set(ctx context.Context, userID uuid.UUID, notifType string, enabled bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notification_settings (user_id, type, enabled)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, type) DO UPDATE SET enabled = $3`,
		userID, notifType, enabled,
	)
	return pgErr(err)
}

func (r *NotifPrefRepo) IsEnabled(ctx context.Context, userID uuid.UUID, notifType string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(
			(SELECT enabled FROM notification_settings WHERE user_id = $1 AND type = $2),
			true
		)`,
		userID, notifType,
	).Scan(&enabled)
	return enabled, err
}
