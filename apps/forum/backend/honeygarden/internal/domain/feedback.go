package domain

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackType string

const (
	FeedbackTypeIdea     FeedbackType = "idea"
	FeedbackTypeBug      FeedbackType = "bug"
	FeedbackTypeQuestion FeedbackType = "question"
	FeedbackTypeOther    FeedbackType = "other"
)

func (t FeedbackType) Valid() bool {
	switch t {
	case FeedbackTypeIdea, FeedbackTypeBug, FeedbackTypeQuestion, FeedbackTypeOther:
		return true
	}
	return false
}

type FeedbackStatus string

const (
	FeedbackStatusOpen       FeedbackStatus = "open"
	FeedbackStatusPlanned    FeedbackStatus = "planned"
	FeedbackStatusInProgress FeedbackStatus = "in_progress"
	FeedbackStatusDone       FeedbackStatus = "done"
	FeedbackStatusRejected   FeedbackStatus = "rejected"
)

func (s FeedbackStatus) Valid() bool {
	switch s {
	case FeedbackStatusOpen, FeedbackStatusPlanned, FeedbackStatusInProgress, FeedbackStatusDone, FeedbackStatusRejected:
		return true
	}
	return false
}

type Feedback struct {
	ID         uuid.UUID      `json:"id"`
	Type       FeedbackType   `json:"type"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	AuthorID   *uuid.UUID     `json:"-"`
	IsAnon     bool           `json:"isAnon"`
	Status     FeedbackStatus `json:"status"`
	VotesCount int            `json:"votesCount"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`

	// Enriched (omitempty when anon or no author).
	Author *UserBrief `json:"author,omitempty"`
	IsVoted bool      `json:"isVoted"`
}

type CreateFeedbackInput struct {
	Type   FeedbackType `json:"type"`
	Title  string       `json:"title"`
	Body   string       `json:"body"`
	IsAnon bool         `json:"isAnon"`
}

type UpdateFeedbackInput struct {
	Status *FeedbackStatus `json:"status"`
}
