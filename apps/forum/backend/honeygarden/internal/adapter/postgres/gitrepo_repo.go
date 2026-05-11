package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type GitAccountRepo struct {
	pool    *pgxpool.Pool
	secrets *SecretBox
}

func NewGitAccountRepo(pool *pgxpool.Pool, secrets *SecretBox) *GitAccountRepo {
	return &GitAccountRepo{pool: pool, secrets: secrets}
}

const gitAccountCols = `id, user_id, provider, access_token, refresh_token, expires_at,
	username, instance_url, created_at, updated_at`

func (r *GitAccountRepo) scanGitAccount(row pgx.Row, a *domain.GitAccount) error {
	if err := row.Scan(
		&a.ID, &a.UserID, &a.Provider, &a.AccessToken, &a.RefreshToken, &a.ExpiresAt,
		&a.Username, &a.InstanceURL, &a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return err
	}
	if r.secrets == nil {
		return nil
	}
	var err error
	a.AccessToken, err = r.secrets.Decrypt(a.AccessToken)
	if err != nil {
		return err
	}
	if a.RefreshToken != nil {
		refresh, err := r.secrets.Decrypt(*a.RefreshToken)
		if err != nil {
			return err
		}
		a.RefreshToken = &refresh
	}
	return nil
}

func (r *GitAccountRepo) Upsert(ctx context.Context, a *domain.GitAccount) error {
	accessToken, refreshToken, err := r.encryptTokens(a.AccessToken, a.RefreshToken)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO user_git_accounts
		   (id, user_id, provider, access_token, refresh_token, expires_at,
		    username, instance_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		 ON CONFLICT (user_id, provider, instance_url)
		 DO UPDATE SET access_token  = EXCLUDED.access_token,
		               refresh_token = EXCLUDED.refresh_token,
		               expires_at    = EXCLUDED.expires_at,
		               username      = EXCLUDED.username,
		               updated_at    = NOW()`,
		a.ID, a.UserID, a.Provider, accessToken, refreshToken, a.ExpiresAt,
		a.Username, a.InstanceURL,
	)
	return pgErr(err)
}

func (r *GitAccountRepo) encryptTokens(accessToken string, refreshToken *string) (string, *string, error) {
	if r.secrets == nil {
		return accessToken, refreshToken, nil
	}
	access, err := r.secrets.Encrypt(accessToken)
	if err != nil {
		return "", nil, err
	}
	if refreshToken == nil {
		return access, nil, nil
	}
	refresh, err := r.secrets.Encrypt(*refreshToken)
	if err != nil {
		return "", nil, err
	}
	return access, &refresh, nil
}

func (r *GitAccountRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.GitAccount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+gitAccountCols+`
		 FROM user_git_accounts WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	accounts := []domain.GitAccount{}
	for rows.Next() {
		var a domain.GitAccount
		if err = r.scanGitAccount(rows, &a); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *GitAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GitAccount, error) {
	a := &domain.GitAccount{}
	err := r.scanGitAccount(
		r.pool.QueryRow(ctx, `SELECT `+gitAccountCols+` FROM user_git_accounts WHERE id = $1`, id),
		a,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return a, pgErr(err)
}

func (r *GitAccountRepo) GetByUserProvider(ctx context.Context, userID uuid.UUID, provider domain.GitProvider, instanceURL *string) (*domain.GitAccount, error) {
	a := &domain.GitAccount{}
	err := r.scanGitAccount(
		r.pool.QueryRow(ctx,
			`SELECT `+gitAccountCols+` FROM user_git_accounts
			 WHERE user_id = $1 AND provider = $2
			   AND instance_url IS NOT DISTINCT FROM $3`,
			userID, provider, instanceURL,
		),
		a,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return a, pgErr(err)
}

func (r *GitAccountRepo) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken string, refreshToken *string, expiresAt *time.Time) error {
	accessToken, refreshToken, err := r.encryptTokens(accessToken, refreshToken)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE user_git_accounts
		 SET access_token = $2, refresh_token = $3, expires_at = $4, updated_at = NOW()
		 WHERE id = $1`,
		id, accessToken, refreshToken, expiresAt,
	)
	return pgErr(err)
}

func (r *GitAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_git_accounts WHERE id = $1`, id)
	return pgErr(err)
}

type PinnedRepoRepo struct {
	pool *pgxpool.Pool
}

func NewPinnedRepoRepo(pool *pgxpool.Pool) *PinnedRepoRepo {
	return &PinnedRepoRepo{pool: pool}
}

func (r *PinnedRepoRepo) Upsert(ctx context.Context, repo *domain.PinnedRepo) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO pinned_repos
		   (id, user_id, git_account_id, external_id, name, description, url,
		    language, stars_count, forks_count, is_fork, topics, sort_order, synced_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
		 ON CONFLICT (user_id, git_account_id, external_id)
		 DO UPDATE SET name=$5, description=$6, url=$7, language=$8,
		               stars_count=$9, forks_count=$10, is_fork=$11,
		               topics=$12, sort_order=$13, synced_at=NOW()`,
		repo.ID, repo.UserID, repo.GitAccountID, repo.ExternalID,
		repo.Name, repo.Description, repo.URL, repo.Language,
		repo.StarsCount, repo.ForksCount, repo.IsFork, repo.Topics, repo.SortOrder,
	)
	return pgErr(err)
}

func (r *PinnedRepoRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.PinnedRepo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, git_account_id, external_id, name, description, url,
		        language, stars_count, forks_count, is_fork, topics, sort_order, synced_at
		 FROM pinned_repos WHERE user_id = $1 ORDER BY sort_order`,
		userID,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	repos := []domain.PinnedRepo{}
	for rows.Next() {
		var repo domain.PinnedRepo
		if err = rows.Scan(
			&repo.ID, &repo.UserID, &repo.GitAccountID, &repo.ExternalID,
			&repo.Name, &repo.Description, &repo.URL, &repo.Language,
			&repo.StarsCount, &repo.ForksCount, &repo.IsFork, &repo.Topics,
			&repo.SortOrder, &repo.SyncedAt,
		); err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

func (r *PinnedRepoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM pinned_repos WHERE id = $1`, id)
	return pgErr(err)
}

func (r *PinnedRepoRepo) UpdateOrder(ctx context.Context, userID uuid.UUID, repoIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for i, id := range repoIDs {
		if _, err = tx.Exec(ctx,
			`UPDATE pinned_repos SET sort_order = $1 WHERE id = $2 AND user_id = $3`,
			i, id, userID,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// replaceAll
func (r *PinnedRepoRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM pinned_repos WHERE user_id = $1`, userID)
	return pgErr(err)
}
