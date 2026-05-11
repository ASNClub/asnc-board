package service

import (
	"context"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type RoadmapService struct {
	repo port.RoadmapRepository
}

func NewRoadmapService(repo port.RoadmapRepository) *RoadmapService {
	return &RoadmapService{repo: repo}
}

func (s *RoadmapService) List(ctx context.Context) ([]domain.RoadmapItem, error) {
	return s.repo.List(ctx)
}

func (s *RoadmapService) Create(ctx context.Context, input domain.CreateRoadmapItemInput) (*domain.RoadmapItem, error) {
	if !input.Phase.Valid() {
		return nil, domain.ErrInvalidInput
	}
	if input.Title == "" {
		return nil, domain.ErrInvalidInput
	}
	if input.Tags == nil {
		input.Tags = []string{}
	}
	item := &domain.RoadmapItem{
		Phase:       input.Phase,
		Title:       input.Title,
		Description: input.Description,
		Tags:        input.Tags,
		ETA:         input.ETA,
		Featured:    input.Featured,
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RoadmapService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateRoadmapItemInput) (*domain.RoadmapItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Phase != nil {
		if !input.Phase.Valid() {
			return nil, domain.ErrInvalidInput
		}
		item.Phase = *input.Phase
	}
	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.Tags != nil {
		item.Tags = input.Tags
	}
	if input.ETA != nil {
		item.ETA = input.ETA
	}
	if input.Featured != nil {
		item.Featured = *input.Featured
	}
	if input.SortOrder != nil {
		item.SortOrder = *input.SortOrder
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RoadmapService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
