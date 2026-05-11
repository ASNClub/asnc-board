package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type BookmarkRepo struct {
	pool *pgxpool.Pool
}

func NewBookmarkRepo(pool *pgxpool.Pool) *BookmarkRepo {
	return &BookmarkRepo{pool: pool}
}

func (r *BookmarkRepo) Add(ctx context.Context, userID, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO bookmarks (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	return pgErr(err)
}

func (r *BookmarkRepo) Remove(ctx context.Context, userID, postID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM bookmarks WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	return pgErr(err)
}

func (r *BookmarkRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.community_id, p.author_id, p.kind, p.title, p.content,
		        p.views_count, p.votes_count, p.is_pinned, p.created_at, p.updated_at,
		        p.source_id, p.external_url, p.cover_image_url, p.tags
		 FROM bookmarks b
		 JOIN posts p ON p.id = b.post_id
		 WHERE b.user_id = $1
		 ORDER BY b.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		if err = rows.Scan(&p.ID, &p.CommunityID, &p.AuthorID, &p.Kind, &p.Title, &p.Content,
			&p.ViewsCount, &p.VotesCount, &p.IsPinned, &p.CreatedAt, &p.UpdatedAt,
			&p.SourceID, &p.ExternalURL, &p.CoverImageURL, &p.Tags); err != nil {
			return nil, err
		}
		p.Media = []domain.PostMedia{}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *BookmarkRepo) IsBookmarked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	return exists, err
}

func (r *BookmarkRepo) BatchIsBookmarked(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT post_id FROM bookmarks WHERE user_id = $1 AND post_id = ANY($2)`,
		userID, postIDs,
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
