package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type CommentRepo struct {
	pool *pgxpool.Pool
}

func NewCommentRepo(pool *pgxpool.Pool) *CommentRepo {
	return &CommentRepo{pool: pool}
}

func (r *CommentRepo) Create(ctx context.Context, c *domain.Comment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO comments (id, post_id, author_id, parent_id, content) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.PostID, c.AuthorID, c.ParentID, c.Content,
	)
	return pgErr(err)
}

func (r *CommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	c := &domain.Comment{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, post_id, author_id, parent_id, content, votes_count, created_at, updated_at
		 FROM comments WHERE id = $1`, id,
	).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Content, &c.VotesCount, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return c, pgErr(err)
}

func (r *CommentRepo) GetByPost(ctx context.Context, postID uuid.UUID, limit, offset int) ([]domain.Comment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, post_id, author_id, parent_id, content, votes_count, created_at, updated_at
		 FROM comments WHERE post_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`,
		postID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := []domain.Comment{}
	for rows.Next() {
		var c domain.Comment
		if err = rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Content,
			&c.VotesCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *CommentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM comments WHERE id = $1`, id)
	return pgErr(err)
}

func (r *CommentRepo) AddVote(ctx context.Context, userID, commentID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`INSERT INTO comment_votes (user_id, comment_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, commentID,
	)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE comments SET votes_count = votes_count + 1 WHERE id = $1`, commentID); err != nil {
			return pgErr(err)
		}
	}
	return tx.Commit(ctx)
}

func (r *CommentRepo) RemoveVote(ctx context.Context, userID, commentID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM comment_votes WHERE user_id = $1 AND comment_id = $2`, userID, commentID)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE comments SET votes_count = GREATEST(0, votes_count - 1) WHERE id = $1`, commentID); err != nil {
			return pgErr(err)
		}
	}
	return tx.Commit(ctx)
}
