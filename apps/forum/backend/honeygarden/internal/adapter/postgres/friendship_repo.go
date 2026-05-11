package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type FriendshipRepo struct {
	pool *pgxpool.Pool
}

func NewFriendshipRepo(pool *pgxpool.Pool) *FriendshipRepo {
	return &FriendshipRepo{pool: pool}
}

func (r *FriendshipRepo) Create(ctx context.Context, f *domain.Friendship) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO friendships (id, requester_id, addressee_id, status) VALUES ($1, $2, $3, $4)`,
		f.ID, f.RequesterID, f.AddresseeID, f.Status,
	)
	return pgErr(err)
}

func (r *FriendshipRepo) UpdateStatus(ctx context.Context, requesterID, addresseeID uuid.UUID, status domain.FriendshipStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE friendships SET status = $3, updated_at = NOW() WHERE requester_id = $1 AND addressee_id = $2`,
		requesterID, addresseeID, status,
	)
	return pgErr(err)
}

func (r *FriendshipRepo) Get(ctx context.Context, userA, userB uuid.UUID) (*domain.Friendship, error) {
	f := &domain.Friendship{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, requester_id, addressee_id, status, created_at, updated_at
		 FROM friendships
		 WHERE (requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1)`,
		userA, userB,
	).Scan(&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return f, pgErr(err)
}

func (r *FriendshipRepo) GetFriends(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.auth_id, u.username, u.display_name, u.avatar_url, u.banner_url, u.bio,
		        u.reputation, u.privacy, u.created_at, u.updated_at
		 FROM users u
		 JOIN friendships f ON (f.requester_id = u.id OR f.addressee_id = u.id)
		 WHERE (f.requester_id = $1 OR f.addressee_id = $1) AND f.status = 'accepted' AND u.id != $1
		 ORDER BY f.updated_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(&u.ID, &u.AuthID, &u.Username, &u.DisplayName,
			&u.AvatarURL, &u.BannerURL, &u.Bio, &u.Reputation, &u.Privacy,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Tags = []string{}
		u.Platforms = []domain.UserPlatform{}
		users = append(users, u)
	}
	if users == nil {
		users = []domain.User{}
	}
	return users, nil
}

func (r *FriendshipRepo) GetPending(ctx context.Context, addresseeID uuid.UUID) ([]domain.Friendship, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, requester_id, addressee_id, status, created_at, updated_at
		 FROM friendships WHERE addressee_id = $1 AND status = 'pending' ORDER BY created_at DESC`,
		addresseeID,
	)
	if err != nil {
		return nil, pgErr(err)
	}
	defer rows.Close()
	var fs []domain.Friendship
	for rows.Next() {
		var f domain.Friendship
		if err = rows.Scan(&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	if fs == nil {
		fs = []domain.Friendship{}
	}
	return fs, nil
}
