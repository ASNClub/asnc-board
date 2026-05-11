package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type BannedWordRepo struct {
	pool *pgxpool.Pool
}

func NewBannedWordRepo(pool *pgxpool.Pool) *BannedWordRepo {
	return &BannedWordRepo{pool: pool}
}

func (r *BannedWordRepo) List(ctx context.Context) ([]domain.BannedWord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, word, scope, created_at FROM banned_words ORDER BY word`)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	var items []domain.BannedWord
	for rows.Next() {
		var bw domain.BannedWord
		if err := rows.Scan(&bw.ID, &bw.Word, &bw.Scope, &bw.CreatedAt); err != nil {
			return nil, pgErr(err)
		}
		items = append(items, bw)
	}
	return items, pgErr(rows.Err())
}

func (r *BannedWordRepo) Create(ctx context.Context, bw *domain.BannedWord) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO banned_words (word, scope) VALUES ($1, $2)
		 RETURNING id, created_at`,
		strings.ToLower(bw.Word), bw.Scope,
	)
	return pgErr(row.Scan(&bw.ID, &bw.CreatedAt))
}

func (r *BannedWordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM banned_words WHERE id = $1`, id)
	if err != nil {
		return pgErr(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BannedWordRepo) IsWordBanned(ctx context.Context, word string, scope domain.BannedWordScope) (bool, error) {
	w := strings.ToLower(word)
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM banned_words
			WHERE $1 LIKE '%' || word || '%'
			  AND (scope = $2 OR scope = 'both')
		)`, w, scope).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, pgErr(err)
}
