package domain

import (
	"time"

	"github.com/google/uuid"
)

type BannedWordScope string

const (
	BannedWordScopeUsername BannedWordScope = "username"
	BannedWordScopeSlug    BannedWordScope = "slug"
	BannedWordScopeBoth    BannedWordScope = "both"
)

func (s BannedWordScope) Valid() bool {
	switch s {
	case BannedWordScopeUsername, BannedWordScopeSlug, BannedWordScopeBoth:
		return true
	}
	return false
}

func (s BannedWordScope) MatchesUsername() bool {
	return s == BannedWordScopeUsername || s == BannedWordScopeBoth
}

func (s BannedWordScope) MatchesSlug() bool {
	return s == BannedWordScopeSlug || s == BannedWordScopeBoth
}

type BannedWord struct {
	ID        uuid.UUID       `json:"id"`
	Word      string          `json:"word"`
	Scope     BannedWordScope `json:"scope"`
	CreatedAt time.Time       `json:"createdAt"`
}

type CreateBannedWordInput struct {
	Word  string          `json:"word"`
	Scope BannedWordScope `json:"scope"`
}
