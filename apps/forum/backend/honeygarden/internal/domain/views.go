package domain

import "github.com/google/uuid"

type UserBrief struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl"`
	Reputation  int       `json:"reputation"`
}

type PostView struct {
	Post
	Author        *UserBrief  `json:"author,omitempty"`
	CommunitySlug string      `json:"communitySlug,omitempty"`
	Source        *FeedSource `json:"source,omitempty"`
	IsVoted       bool        `json:"isVoted"`
	IsBookmarked  bool        `json:"isBookmarked"`
}

type CommentView struct {
	Comment
	Author UserBrief `json:"author"`
}
