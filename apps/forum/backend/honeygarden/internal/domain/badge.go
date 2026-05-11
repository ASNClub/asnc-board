package domain

import (
	"time"

	"github.com/google/uuid"
)

type BadgeDefinition struct {
	ID          string `json:"id"`
	Glyph       string `json:"glyph"`
	Name        string `json:"name"`
	NameRu      string `json:"nameRu"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Rarity      string `json:"rarity"`
	SortOrder   int    `json:"sortOrder"`
}

type UserBadge struct {
	UserID   uuid.UUID        `json:"userId"`
	Badge    BadgeDefinition  `json:"badge"`
	EarnedAt time.Time        `json:"earnedAt"`
}
