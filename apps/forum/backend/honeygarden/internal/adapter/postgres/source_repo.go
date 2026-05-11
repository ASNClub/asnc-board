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

type SourceRepo struct {
	pool *pgxpool.Pool
}

func NewSourceRepo(pool *pgxpool.Pool) *SourceRepo {
	return &SourceRepo{pool: pool}
}

func (r *SourceRepo) Create(ctx context.Context, s *domain.RSSSource) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rss_sources (id, url, name, site_url, favicon_url, tags) VALUES ($1, $2, $3, $4, $5, $6)`,
		s.ID, s.URL, s.Name, s.SiteURL, s.FaviconURL, s.Tags,
	)
	return pgErr(err)
}

func (r *SourceRepo) List(ctx context.Context) ([]domain.RSSSource, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, url, name, site_url, favicon_url, tags, last_fetched, created_at FROM rss_sources ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sources := []domain.RSSSource{}
	for rows.Next() {
		var s domain.RSSSource
		if err = rows.Scan(&s.ID, &s.URL, &s.Name, &s.SiteURL, &s.FaviconURL,
			&s.Tags, &s.LastFetched, &s.CreatedAt); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *SourceRepo) UpdateLastFetched(ctx context.Context, id uuid.UUID, t time.Time) error {
	tag, err := r.pool.Exec(ctx, `UPDATE rss_sources SET last_fetched = $2 WHERE id = $1`, id, t)
	if err != nil {
		return pgErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SourceRepo) ResolveSources(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.FeedSource, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]domain.FeedSource{}, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT id, name, favicon_url FROM rss_sources WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[uuid.UUID]domain.FeedSource, len(ids))
	for rows.Next() {
		var id uuid.UUID
		var s domain.FeedSource
		if err = rows.Scan(&id, &s.Name, &s.FaviconURL); err != nil {
			return nil, err
		}
		result[id] = s
	}
	return result, rows.Err()
}

func (r *SourceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RSSSource, error) {
	var s domain.RSSSource
	err := r.pool.QueryRow(ctx,
		`SELECT id, url, name, site_url, favicon_url, tags, last_fetched, created_at FROM rss_sources WHERE id = $1`, id,
	).Scan(&s.ID, &s.URL, &s.Name, &s.SiteURL, &s.FaviconURL, &s.Tags, &s.LastFetched, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &s, pgErr(err)
}
