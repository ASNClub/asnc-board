package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type RoadmapRepo struct {
	pool *pgxpool.Pool
}

func NewRoadmapRepo(pool *pgxpool.Pool) *RoadmapRepo {
	return &RoadmapRepo{pool: pool}
}

func scanRoadmapItem(row pgx.Row, item *domain.RoadmapItem) error {
	return row.Scan(&item.ID, &item.Phase, &item.Title, &item.Description,
		&item.Tags, &item.ETA, &item.Featured, &item.SortOrder,
		&item.CreatedAt, &item.UpdatedAt)
}

const roadmapCols = `id, phase, title, description, tags, eta, featured, sort_order, created_at, updated_at`

func (r *RoadmapRepo) List(ctx context.Context) ([]domain.RoadmapItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+roadmapCols+` FROM roadmap_items ORDER BY phase, sort_order, created_at`)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	var items []domain.RoadmapItem
	for rows.Next() {
		var item domain.RoadmapItem
		if err := scanRoadmapItem(rows, &item); err != nil {
			return nil, pgErr(err)
		}
		items = append(items, item)
	}
	return items, pgErr(rows.Err())
}

func (r *RoadmapRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoadmapItem, error) {
	item := &domain.RoadmapItem{}
	err := scanRoadmapItem(r.pool.QueryRow(ctx,
		`SELECT `+roadmapCols+` FROM roadmap_items WHERE id = $1`, id), item)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return item, pgErr(err)
}

func (r *RoadmapRepo) Create(ctx context.Context, item *domain.RoadmapItem) error {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO roadmap_items (phase, title, description, tags, eta, featured, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		item.Phase, item.Title, item.Description, item.Tags, item.ETA,
		item.Featured, item.SortOrder,
	)
	return pgErr(row.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt))
}

func (r *RoadmapRepo) Update(ctx context.Context, item *domain.RoadmapItem) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE roadmap_items
		 SET phase = $2, title = $3, description = $4, tags = $5, eta = $6,
		     featured = $7, sort_order = $8, updated_at = now()
		 WHERE id = $1`,
		item.ID, item.Phase, item.Title, item.Description, item.Tags,
		item.ETA, item.Featured, item.SortOrder,
	)
	if err != nil {
		return pgErr(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RoadmapRepo) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM roadmap_items WHERE id = $1`, id)
	if err != nil {
		return pgErr(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
