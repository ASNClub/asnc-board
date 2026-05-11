package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoadmapPhase string

const (
	RoadmapPhaseWIP   RoadmapPhase = "wip"
	RoadmapPhaseNext  RoadmapPhase = "next"
	RoadmapPhaseLater RoadmapPhase = "later"
	RoadmapPhaseDone  RoadmapPhase = "done"
)

func (p RoadmapPhase) Valid() bool {
	switch p {
	case RoadmapPhaseWIP, RoadmapPhaseNext, RoadmapPhaseLater, RoadmapPhaseDone:
		return true
	}
	return false
}

type RoadmapItem struct {
	ID          uuid.UUID    `json:"id"`
	Phase       RoadmapPhase `json:"phase"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	ETA         *string      `json:"eta"`
	Featured    bool         `json:"featured"`
	SortOrder   int          `json:"sortOrder"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type CreateRoadmapItemInput struct {
	Phase       RoadmapPhase `json:"phase"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	ETA         *string      `json:"eta"`
	Featured    bool         `json:"featured"`
	SortOrder   int          `json:"sortOrder"`
}

type UpdateRoadmapItemInput struct {
	Phase       *RoadmapPhase `json:"phase"`
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Tags        []string      `json:"tags"`
	ETA         *string       `json:"eta"`
	Featured    *bool         `json:"featured"`
	SortOrder   *int          `json:"sortOrder"`
}
