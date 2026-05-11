package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type WakapiRepo struct {
	pool    *pgxpool.Pool
	secrets *SecretBox
}

func NewWakapiRepo(pool *pgxpool.Pool, secrets *SecretBox) *WakapiRepo {
	return &WakapiRepo{pool: pool, secrets: secrets}
}

func (r *WakapiRepo) Save(ctx context.Context, acc domain.WakapiAccount) error {
	apiKey := acc.APIKey
	var err error
	if r.secrets != nil {
		apiKey, err = r.secrets.Encrypt(apiKey)
		if err != nil {
			return err
		}
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO user_wakapi (user_id, instance_url, api_key, username)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO UPDATE
		 SET instance_url = EXCLUDED.instance_url,
		     api_key      = EXCLUDED.api_key,
		     username     = EXCLUDED.username`,
		acc.UserID, acc.InstanceURL, apiKey, acc.Username,
	)
	return pgErr(err)
}

func (r *WakapiRepo) Get(ctx context.Context, userID uuid.UUID) (domain.WakapiAccount, error) {
	var acc domain.WakapiAccount
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, instance_url, api_key, username, created_at
		 FROM user_wakapi WHERE user_id = $1`, userID,
	).Scan(&acc.UserID, &acc.InstanceURL, &acc.APIKey, &acc.Username, &acc.CreatedAt)
	if err != nil {
		return acc, pgErr(err)
	}
	if r.secrets != nil {
		acc.APIKey, err = r.secrets.Decrypt(acc.APIKey)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}

func (r *WakapiRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM user_wakapi WHERE user_id = $1`, userID)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
