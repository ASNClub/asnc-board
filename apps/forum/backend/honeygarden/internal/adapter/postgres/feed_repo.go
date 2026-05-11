package postgres

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"honeygarden/internal/domain"
)

type FeedRepo struct {
	pool *pgxpool.Pool
}

func NewFeedRepo(pool *pgxpool.Pool) *FeedRepo {
	return &FeedRepo{pool: pool}
}

func (r *FeedRepo) GetFeed(ctx context.Context, userID *uuid.UUID, cursor time.Time, limit int) ([]domain.FeedItem, error) {
	items := make([]domain.FeedItem, 0, limit*2)

	externalItems, err := r.getExternalItems(ctx, userID, cursor, limit)
	if err != nil {
		return nil, err
	}
	items = append(items, externalItems...)

	if userID != nil {
		postItems, err := r.getPostItems(ctx, *userID, cursor, limit)
		if err != nil {
			return nil, err
		}
		items = append(items, postItems...)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].PublishedAt.After(items[j].PublishedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (r *FeedRepo) getExternalItems(ctx context.Context, userID *uuid.UUID, cursor time.Time, limit int) ([]domain.FeedItem, error) {
	var rows pgx.Rows
	var err error

	if userID != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT p.id, p.short_id, p.title, p.external_url, p.content, p.cover_image_url, p.tags,
			       p.created_at, p.votes_count,
			       (SELECT COUNT(*) FROM comments WHERE post_id = p.id)::int AS comments_count,
			       rs.name, rs.favicon_url,
			       EXISTS(SELECT 1 FROM post_votes pv WHERE pv.post_id = p.id AND pv.user_id = $2) AS is_voted,
			       EXISTS(SELECT 1 FROM bookmarks bm WHERE bm.post_id = p.id AND bm.user_id = $2) AS is_bookmarked
			FROM posts p
			JOIN rss_sources rs ON rs.id = p.source_id
			WHERE p.source_id IS NOT NULL AND p.created_at < $1
			  AND (
			      p.tags && (SELECT array_agg(tag) FROM user_tags WHERE user_id = $2)
			      OR NOT EXISTS (SELECT 1 FROM user_tags WHERE user_id = $2)
			  )
			ORDER BY p.created_at DESC LIMIT $3`,
			cursor, *userID, limit,
		)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT p.id, p.short_id, p.title, p.external_url, p.content, p.cover_image_url, p.tags,
			       p.created_at, p.votes_count,
			       (SELECT COUNT(*) FROM comments WHERE post_id = p.id)::int AS comments_count,
			       rs.name, rs.favicon_url,
			       FALSE AS is_voted,
			       FALSE AS is_bookmarked
			FROM posts p
			JOIN rss_sources rs ON rs.id = p.source_id
			WHERE p.source_id IS NOT NULL AND p.created_at < $1
			ORDER BY p.created_at DESC LIMIT $2`,
			cursor, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.FeedItem{}
	for rows.Next() {
		var (
			id            uuid.UUID
			shortID       string
			title         string
			externalURL   string
			content       string
			coverImage    string
			tags          []string
			createdAt     time.Time
			votesCount    int
			commentsCount int
			sourceName    string
			sourceFavicon string
			isVoted       bool
			isBookmarked  bool
		)
		if err = rows.Scan(&id, &shortID, &title, &externalURL, &content, &coverImage, &tags,
			&createdAt, &votesCount, &commentsCount, &sourceName, &sourceFavicon, &isVoted, &isBookmarked); err != nil {
			return nil, err
		}
		postID := id
		items = append(items, domain.FeedItem{
			Type:          "rss",
			ID:            id.String(),
			ShortID:       shortID,
			PostID:        &postID,
			Title:         title,
			URL:           externalURL,
			Summary:       content,
			CoverImage:    coverImage,
			Tags:          tags,
			PublishedAt:   createdAt,
			Upvotes:       votesCount,
			CommentsCount: commentsCount,
			IsVoted:       isVoted,
			IsBookmarked:  isBookmarked,
			Source: &domain.FeedSource{
				Name:       sourceName,
				FaviconURL: sourceFavicon,
			},
		})
	}
	return items, rows.Err()
}

func (r *FeedRepo) getPostItems(ctx context.Context, userID uuid.UUID, cursor time.Time, limit int) ([]domain.FeedItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.short_id, COALESCE(p.title, '') AS title, p.kind, p.votes_count, p.created_at,
		       c.slug, u.username, COALESCE(u.avatar_url, '') AS avatar_url,
		       COALESCE(
		           (SELECT array_agg(ct.tag) FROM community_tags ct WHERE ct.community_id = p.community_id),
		           '{}'::text[]
		       ) AS tags,
		       (SELECT COUNT(*) FROM comments WHERE post_id = p.id)::int AS comments_count,
		       EXISTS(SELECT 1 FROM post_votes pv WHERE pv.post_id = p.id AND pv.user_id = $2) AS is_voted,
		       EXISTS(SELECT 1 FROM bookmarks bm WHERE bm.post_id = p.id AND bm.user_id = $2) AS is_bookmarked
		FROM posts p
		JOIN communities c ON c.id = p.community_id
		JOIN users u ON u.id = p.author_id
		WHERE p.created_at < $1
		  AND p.source_id IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM user_blocks ub
		       WHERE (ub.blocker_id = $2 AND ub.blocked_id = p.author_id)
		          OR (ub.blocker_id = p.author_id AND ub.blocked_id = $2)
		  )
		  AND (
		      EXISTS (
		          SELECT 1 FROM community_tags ct
		          JOIN user_tags ut ON ut.tag = ct.tag
		          WHERE ct.community_id = p.community_id AND ut.user_id = $2
		      )
		      OR EXISTS (
		          SELECT 1 FROM community_follows cf
		          WHERE cf.community_id = p.community_id AND cf.user_id = $2
		      )
		      OR EXISTS (
		          SELECT 1 FROM user_follows uf
		          WHERE uf.following_id = p.author_id AND uf.follower_id = $2
		      )
		  )
		ORDER BY p.created_at DESC LIMIT $3`,
		cursor, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.FeedItem{}
	for rows.Next() {
		var (
			postID        uuid.UUID
			shortID       string
			title         string
			kind          domain.PostKind
			votesCount    int
			createdAt     time.Time
			communitySlug string
			username      string
			avatarURL     string
			tags          []string
			commentsCount int
			isVoted       bool
			isBookmarked  bool
		)
		if err = rows.Scan(&postID, &shortID, &title, &kind, &votesCount, &createdAt,
			&communitySlug, &username, &avatarURL, &tags, &commentsCount, &isVoted, &isBookmarked); err != nil {
			return nil, err
		}
		items = append(items, domain.FeedItem{
			Type:          "post",
			ID:            postID.String(),
			ShortID:       shortID,
			Title:         title,
			Kind:          kind,
			Tags:          tags,
			PublishedAt:   createdAt,
			Upvotes:       votesCount,
			PostID:        &postID,
			CommunitySlug: communitySlug,
			CommentsCount: commentsCount,
			IsVoted:       isVoted,
			IsBookmarked:  isBookmarked,
			Author: &domain.FeedAuthor{
				Username:  username,
				AvatarURL: avatarURL,
			},
		})
	}
	return items, rows.Err()
}
