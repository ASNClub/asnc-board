package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type FeedbackService struct {
	repo  port.FeedbackRepository
	users port.UserResolver
}

func NewFeedbackService(repo port.FeedbackRepository, users port.UserResolver) *FeedbackService {
	return &FeedbackService{repo: repo, users: users}
}

const (
	feedbackTitleMin = 8
	feedbackTitleMax = 200
	feedbackBodyMin  = 20
	feedbackBodyMax  = 4000
	feedbackRateMax  = 5  // per window
	feedbackRateWin  = 60 // minutes
)

func (s *FeedbackService) Create(ctx context.Context, authorID uuid.UUID, input domain.CreateFeedbackInput) (*domain.Feedback, error) {
	if !input.Type.Valid() {
		return nil, fmt.Errorf("type: %w", domain.ErrInvalidInput)
	}
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	if len(title) < feedbackTitleMin || len(title) > feedbackTitleMax {
		return nil, fmt.Errorf("title length: %w", domain.ErrInvalidInput)
	}
	if len(body) < feedbackBodyMin || len(body) > feedbackBodyMax {
		return nil, fmt.Errorf("body length: %w", domain.ErrInvalidInput)
	}

	count, err := s.repo.CountByAuthorSince(ctx, authorID, feedbackRateWin)
	if err == nil && count >= feedbackRateMax {
		return nil, fmt.Errorf("too many feedback submissions: %w", domain.ErrForbidden)
	}

	f := &domain.Feedback{
		Type:     input.Type,
		Title:    title,
		Body:     body,
		AuthorID: &authorID,
		IsAnon:   input.IsAnon,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FeedbackService) List(ctx context.Context, sort, status string, viewerID *uuid.UUID, limit, offset int) ([]domain.Feedback, error) {
	items, err := s.repo.List(ctx, sort, status, limit, offset)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}

	authorIDs := []uuid.UUID{}
	for _, f := range items {
		if !f.IsAnon && f.AuthorID != nil {
			authorIDs = append(authorIDs, *f.AuthorID)
		}
	}
	authorMap := map[uuid.UUID]domain.UserBrief{}
	if len(authorIDs) > 0 {
		if m, err := s.users.ResolveUsers(ctx, authorIDs); err == nil {
			authorMap = m
		}
	}

	votedSet := map[uuid.UUID]bool{}
	if viewerID != nil {
		ids := make([]uuid.UUID, len(items))
		for i, f := range items {
			ids[i] = f.ID
		}
		if v, err := s.repo.BatchIsVoted(ctx, *viewerID, ids); err == nil {
			votedSet = v
		}
	}

	for i := range items {
		f := &items[i]
		if !f.IsAnon && f.AuthorID != nil {
			if a, ok := authorMap[*f.AuthorID]; ok {
				cp := a
				f.Author = &cp
			}
		}
		f.IsVoted = votedSet[f.ID]
	}
	return items, nil
}

func (s *FeedbackService) Vote(ctx context.Context, userID, feedbackID uuid.UUID) error {
	_, err := s.repo.Vote(ctx, userID, feedbackID)
	return err
}

func (s *FeedbackService) Unvote(ctx context.Context, userID, feedbackID uuid.UUID) error {
	_, err := s.repo.Unvote(ctx, userID, feedbackID)
	return err
}

func (s *FeedbackService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *FeedbackService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.FeedbackStatus) error {
	if !status.Valid() {
		return domain.ErrInvalidInput
	}
	return s.repo.UpdateStatus(ctx, id, status)
}
