package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

const shortIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func newShortID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		copy(b, uuid.New().String())
	}
	var sb strings.Builder
	sb.Grow(8)
	for i := 0; i < 8; i++ {
		sb.WriteByte(shortIDAlphabet[int(b[i])%len(shortIDAlphabet)])
	}
	return sb.String()
}

type PostRepo struct {
	pool *pgxpool.Pool
}

func NewPostRepo(pool *pgxpool.Pool) *PostRepo {
	return &PostRepo{pool: pool}
}

const postCols = `id, short_id, community_id, author_id, kind, title, content, views_count, votes_count, is_pinned,
       created_at, updated_at, source_id, external_url, cover_image_url, tags`

func scanPost(row pgx.Row, p *domain.Post) error {
	return row.Scan(
		&p.ID, &p.ShortID, &p.CommunityID, &p.AuthorID, &p.Kind, &p.Title, &p.Content,
		&p.ViewsCount, &p.VotesCount, &p.IsPinned, &p.CreatedAt, &p.UpdatedAt,
		&p.SourceID, &p.ExternalURL, &p.CoverImageURL, &p.Tags,
	)
}

func (r *PostRepo) Create(ctx context.Context, p *domain.Post) error {
	kind := p.Kind
	if kind == "" {
		kind = domain.PostKindDiscussion
	}
	for range 5 {
		short := newShortID()
		_, err := r.pool.Exec(ctx,
			`INSERT INTO posts (id, short_id, community_id, author_id, kind, title, content) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			p.ID, short, p.CommunityID, p.AuthorID, kind, p.Title, p.Content,
		)
		if err == nil {
			p.ShortID = short
			return nil
		}
		if isUniqueViolation(err, "idx_posts_short_id") {
			continue
		}
		return pgErr(err)
	}
	return domain.ErrInternal
}

// UpsertExternal вставляет или обновляет внешний пост (RSS-статья). Идемпотентен по external_url.
// Существующие votes/comments сохраняются (UPDATE не трогает счётчики).
func (r *PostRepo) UpsertExternal(ctx context.Context, e *domain.ExternalPost) error {
	for attempt := 0; attempt < 5; attempt++ {
		short := newShortID()
		_, err := r.pool.Exec(ctx,
			`INSERT INTO posts (id, short_id, source_id, external_url, title, content, cover_image_url, tags, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
			 ON CONFLICT (external_url) DO UPDATE
			    SET title = EXCLUDED.title,
			        content = EXCLUDED.content,
			        cover_image_url = EXCLUDED.cover_image_url,
			        tags = EXCLUDED.tags,
			        updated_at = EXCLUDED.updated_at`,
			e.ID, short, e.SourceID, e.URL, e.Title, e.Summary, e.CoverImageURL, e.Tags, e.PublishedAt,
		)
		if err == nil {
			return nil
		}
		if isUniqueViolation(err, "idx_posts_short_id") {
			continue
		}
		return pgErr(err)
	}
	return domain.ErrInternal
}

func (r *PostRepo) GetByShortID(ctx context.Context, shortID string) (*domain.Post, error) {
	p := &domain.Post{}
	err := scanPost(r.pool.QueryRow(ctx, `SELECT `+postCols+` FROM posts WHERE short_id = $1`, shortID), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, pgErr(err)
	}
	media, err := r.loadMedia(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Media = media
	return p, nil
}

func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	p := &domain.Post{}
	err := scanPost(r.pool.QueryRow(ctx, `SELECT `+postCols+` FROM posts WHERE id = $1`, id), p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, pgErr(err)
	}
	media, err := r.loadMedia(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Media = media
	return p, nil
}

func (r *PostRepo) listQuery(ctx context.Context, query string, args ...any) ([]domain.Post, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		if err = scanPost(rows, &p); err != nil {
			return nil, err
		}
		p.Media = []domain.PostMedia{}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *PostRepo) GetByCommunity(ctx context.Context, communityID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	return r.listQuery(ctx,
		`SELECT `+postCols+` FROM posts WHERE community_id = $1
		 ORDER BY is_pinned DESC, created_at DESC LIMIT $2 OFFSET $3`,
		communityID, limit, offset,
	)
}

func (r *PostRepo) GetByCommunityKind(ctx context.Context, communityID uuid.UUID, kind domain.PostKind, limit, offset int) ([]domain.Post, error) {
	return r.listQuery(ctx,
		`SELECT `+postCols+` FROM posts WHERE community_id = $1 AND kind = $2
		 ORDER BY is_pinned DESC, created_at DESC LIMIT $3 OFFSET $4`,
		communityID, kind, limit, offset,
	)
}

func (r *PostRepo) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	return r.listQuery(ctx,
		`SELECT `+postCols+` FROM posts WHERE author_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		authorID, limit, offset,
	)
}

func (r *PostRepo) Update(ctx context.Context, p *domain.Post) error {
	kind := p.Kind
	if kind == "" {
		kind = domain.PostKindDiscussion
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE posts SET kind=$2, title=$3, content=$4, updated_at=NOW() WHERE id = $1`,
		p.ID, kind, p.Title, p.Content,
	)
	return pgErr(err)
}

func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return pgErr(err)
}

func (r *PostRepo) IncrViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET views_count = views_count + 1 WHERE id = $1`, id)
	return pgErr(err)
}

func (r *PostRepo) AddVote(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx,
		`INSERT INTO post_votes (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE posts SET votes_count = votes_count + 1 WHERE id = $1`, postID); err != nil {
			return pgErr(err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) RemoveVote(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM post_votes WHERE user_id = $1 AND post_id = $2`, userID, postID)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() > 0 {
		if _, err = tx.Exec(ctx, `UPDATE posts SET votes_count = GREATEST(0, votes_count - 1) WHERE id = $1`, postID); err != nil {
			return pgErr(err)
		}
	}
	return tx.Commit(ctx)
}

func (r *PostRepo) BatchIsVoted(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT post_id FROM post_votes WHERE user_id = $1 AND post_id = ANY($2)`,
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

func (r *PostRepo) IsVoted(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM post_votes WHERE user_id = $1 AND post_id = $2)`,
		userID, postID,
	).Scan(&exists)
	return exists, err
}

func (r *PostRepo) GetTrending(ctx context.Context, since time.Time, limit int) ([]domain.Post, error) {
	return r.listQuery(ctx,
		`SELECT `+postCols+` FROM posts WHERE created_at >= $1 AND author_id IS NOT NULL
		 ORDER BY votes_count DESC, views_count DESC LIMIT $2`,
		since, limit,
	)
}

func (r *PostRepo) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE posts SET is_pinned = $2 WHERE id = $1`, id, pinned)
	return pgErr(err)
}

func (r *PostRepo) loadMedia(ctx context.Context, postID uuid.UUID) ([]domain.PostMedia, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, post_id, type, url, name, size, created_at FROM post_media WHERE post_id = $1 ORDER BY created_at`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	media := []domain.PostMedia{}
	for rows.Next() {
		var m domain.PostMedia
		if err = rows.Scan(&m.ID, &m.PostID, &m.Type, &m.URL, &m.Name, &m.Size, &m.CreatedAt); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}
