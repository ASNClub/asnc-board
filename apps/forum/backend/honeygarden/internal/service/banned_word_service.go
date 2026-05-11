package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type BannedWordService struct {
	repo port.BannedWordRepository
}

func NewBannedWordService(repo port.BannedWordRepository) *BannedWordService {
	return &BannedWordService{repo: repo}
}

func (s *BannedWordService) List(ctx context.Context) ([]domain.BannedWord, error) {
	return s.repo.List(ctx)
}

func (s *BannedWordService) Create(ctx context.Context, input domain.CreateBannedWordInput) (*domain.BannedWord, error) {
	w := strings.ToLower(strings.TrimSpace(input.Word))
	if w == "" {
		return nil, domain.ErrInvalidInput
	}
	if !input.Scope.Valid() {
		return nil, domain.ErrInvalidInput
	}
	bw := &domain.BannedWord{
		Word:  w,
		Scope: input.Scope,
	}
	if err := s.repo.Create(ctx, bw); err != nil {
		return nil, err
	}
	return bw, nil
}

func (s *BannedWordService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *BannedWordService) CheckUsername(ctx context.Context, username string) error {
	banned, err := s.repo.IsWordBanned(ctx, username, domain.BannedWordScopeUsername)
	if err != nil {
		return err
	}
	if banned {
		return domain.ErrForbidden
	}
	return nil
}

func (s *BannedWordService) CheckSlug(ctx context.Context, slug string) error {
	banned, err := s.repo.IsWordBanned(ctx, slug, domain.BannedWordScopeSlug)
	if err != nil {
		return err
	}
	if banned {
		return domain.ErrForbidden
	}
	return nil
}
