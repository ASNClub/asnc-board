package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID  `json:"id"`
	AuthorID  uuid.UUID  `json:"authorId"`
	Author    *UserBrief `json:"author,omitempty"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"createdAt"`
}
