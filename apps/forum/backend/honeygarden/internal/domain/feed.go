package domain

import (
	"time"

	"github.com/google/uuid"
)

type RSSSource struct {
	ID          uuid.UUID  `json:"id"`
	URL         string     `json:"url"`
	Name        string     `json:"name"`
	SiteURL     string     `json:"siteUrl"`
	FaviconURL  string     `json:"faviconUrl"`
	Tags        []string   `json:"tags"`
	LastFetched *time.Time `json:"lastFetched"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type FeedItem struct {
	Type    string `json:"type"` // "rss" | "post"
	ID      string `json:"id"`
	ShortID string `json:"shortId,omitempty"`

	URL     string      `json:"url,omitempty"`
	Summary string      `json:"summary,omitempty"`
	Source  *FeedSource `json:"source,omitempty"`

	PostID        *uuid.UUID  `json:"postId,omitempty"`
	CommunitySlug string      `json:"communitySlug,omitempty"`
	Author        *FeedAuthor `json:"author,omitempty"`
	CommentsCount int         `json:"commentsCount,omitempty"`
	Kind          PostKind    `json:"kind,omitempty"`
	IsBookmarked  bool        `json:"isBookmarked"`

	Title       string    `json:"title"`
	CoverImage  string    `json:"coverImage,omitempty"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"publishedAt"`
	Upvotes     int       `json:"upvotes"`
	IsVoted     bool      `json:"isVoted"`
}

type FeedSource struct {
	Name       string `json:"name"`
	FaviconURL string `json:"faviconUrl,omitempty"`
}

type FeedAuthor struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type CreateSourceInput struct {
	URL        string   `json:"url"`
	Name       string   `json:"name"`
	SiteURL    string   `json:"siteUrl"`
	FaviconURL string   `json:"faviconUrl"`
	Tags       []string `json:"tags"`
}
