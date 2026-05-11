package postgres

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type FeedbackRepo struct {
	pool *pgxpool.Pool
}

func NewFeedbackRepo(pool *pgxpool.Pool) *FeedbackRepo {
	return &FeedbackRepo{pool: pool}
}

const feedbackCols = `id, type, title, body, author_id, is_anon, status, votes_count, created_at, updated_at`

func scanFeedback(row pgx.Row, f *domain.Feedback) error {
	return row.Scan(&f.ID, &f.Type, &f.Title, &f.Body, &f.AuthorID,
		&f.IsAnon, &f.Status, &f.VotesCount, &f.CreatedAt, &f.UpdatedAt)
}

func (r *FeedbackRepo) Create(ctx context.Context, f *domain.Feedback) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO feedback (type, title, body, author_id, is_anon)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, status, votes_count, created_at, updated_at`,
		f.Type, f.Title, f.Body, f.AuthorID, f.IsAnon,
	)
	return pgErr(row.Scan(&f.ID, &f.Status, &f.VotesCount, &f.CreatedAt, &f.UpdatedAt))
}

func (r *FeedbackRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Feedback, error) {
	f := &domain.Feedback{}
	err := scanFeedback(r.pool.QueryRow(ctx, `SELECT `+feedbackCols+` FROM feedback WHERE id = $1`, id), f)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return f, pgErr(err)
}

func (r *FeedbackRepo) List(ctx context.Context, sort, status string, limit, offset int) ([]domain.Feedback, error) {
	orderBy := "votes_count DESC, created_at DESC"
	if sort == "new" {
		orderBy = "created_at DESC"
	}
	q := `SELECT ` + feedbackCols + ` FROM feedback`
	args := []any{}
	if status != "" {
		args = append(args, status)
		q += ` WHERE status = $1`
	}
	q += ` ORDER BY ` + orderBy + ` LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	out := []domain.Feedback{}
	for rows.Next() {
		f := domain.Feedback{}
		if err = scanFeedback(rows, &f); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FeedbackRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FeedbackStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE feedback SET status = $2, updated_at = NOW() WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FeedbackRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM feedback WHERE id = $1`, id)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *FeedbackRepo) Vote(ctx context.Context, userID, feedbackID uuid.UUID) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`INSERT INTO feedback_votes (user_id, feedback_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, feedbackID,
	)
	if err != nil {
		return false, pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE feedback SET votes_count = votes_count + 1 WHERE id = $1`, feedbackID); err != nil {
			return false, pgErr(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *FeedbackRepo) Unvote(ctx context.Context, userID, feedbackID uuid.UUID) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM feedback_votes WHERE user_id = $1 AND feedback_id = $2`, userID, feedbackID)
	if err != nil {
		return false, pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE feedback SET votes_count = GREATEST(0, votes_count - 1) WHERE id = $1`, feedbackID); err != nil {
			return false, pgErr(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *FeedbackRepo) BatchIsVoted(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT feedback_id FROM feedback_votes WHERE user_id = $1 AND feedback_id = ANY($2)`,
		userID, ids,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func (r *FeedbackRepo) CountByAuthorSince(ctx context.Context, authorID uuid.UUID, sinceMinutes int) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM feedback WHERE author_id = $1 AND created_at > NOW() - ($2 || ' minutes')::interval`,
		authorID, sinceMinutes,
	).Scan(&n)
	return n, pgErr(err)
}

