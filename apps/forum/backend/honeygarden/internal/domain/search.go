package domain

type PostDocument struct {
	ID          string  `json:"id"`
	CommunityID string  `json:"community_id"`
	AuthorID    string  `json:"author_id"`
	Title       *string `json:"title"`
	Content     string  `json:"content"`
	CreatedAt   int64   `json:"created_at"`
}

type CommunityDocument struct {
	ID             string   `json:"id"`
	Slug           string   `json:"slug"`
	Name           string   `json:"name"`
	Description    *string  `json:"description"`
	Tags           []string `json:"tags"`
	FollowersCount int      `json:"followers_count"`
}

type SearchResult[T any] struct {
	Hits  []T `json:"hits"`
	Total int `json:"total"`
}
