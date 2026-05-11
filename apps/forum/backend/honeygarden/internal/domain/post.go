package domain

import (
	"time"

	"github.com/google/uuid"
)

type PostKind string

const (
	PostKindDiscussion PostKind = "discussion"
	PostKindArticle    PostKind = "article"
	PostKindQuestion   PostKind = "question"
)

func (k PostKind) Valid() bool {
	switch k {
	case PostKindDiscussion, PostKindArticle, PostKindQuestion:
		return true
	}
	return false
}

type Post struct {
	ID            uuid.UUID   `json:"id"`
	ShortID       string      `json:"shortId,omitempty"`
	CommunityID   *uuid.UUID  `json:"communityId"`
	AuthorID      *uuid.UUID  `json:"authorId"`
	Kind          PostKind    `json:"kind"`
	Title         *string     `json:"title"`
	Content       string      `json:"content"`
	Media         []PostMedia `json:"media"`
	ViewsCount    int         `json:"viewsCount"`
	VotesCount    int         `json:"votesCount"`
	IsPinned      bool        `json:"isPinned"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`

	// Внешние посты (RSS): source_id и external_url выставлены, author_id и community_id — nil.
	SourceID      *uuid.UUID `json:"sourceId,omitempty"`
	ExternalURL   *string    `json:"externalUrl,omitempty"`
	CoverImageURL string     `json:"coverImageUrl,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
}

func (p *Post) IsExternal() bool {
	return p.SourceID != nil
}

type PostMedia struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"postId"`
	Type      MediaType `json:"type"`
	URL       string    `json:"url"`
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeFile  MediaType = "file"
)

type Comment struct {
	ID         uuid.UUID  `json:"id"`
	PostID     uuid.UUID  `json:"postId"`
	AuthorID   uuid.UUID  `json:"authorId"`
	ParentID   *uuid.UUID `json:"parentId"`
	Content    string     `json:"content"`
	VotesCount int        `json:"votesCount"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type CreatePostInput struct {
	Kind    PostKind     `json:"kind"`
	Title   *string      `json:"title"`
	Content string       `json:"content"`
	Media   []MediaInput `json:"media"`
}

type UpdatePostInput struct {
	Kind    *PostKind `json:"kind"`
	Title   *string   `json:"title"`
	Content *string   `json:"content"`
}

type MediaInput struct {
	Type MediaType `json:"type"`
	URL  string    `json:"url"`
	Name string    `json:"name"`
	Size int       `json:"size"`
}

type CreateCommentInput struct {
	Content  string     `json:"content"`
	ParentID *uuid.UUID `json:"parentId"`
}

// ExternalPost — данные для RSS worker
type ExternalPost struct {
	ID            uuid.UUID
	SourceID      uuid.UUID
	Title         string
	URL           string
	Summary       string
	CoverImageURL string
	Tags          []string
	PublishedAt   time.Time
}
