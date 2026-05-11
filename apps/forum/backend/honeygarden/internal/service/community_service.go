package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
	"honeygarden/internal/port"
)

type CommunityService struct {
	communities port.CommunityRepository
	follows     port.CommunityFollowRepository
	stars       port.StarRepository
	bans        port.BanRepository
	mods        port.ModeratorRepository
	users       port.UserRepository
	bannedWords port.BannedWordRepository
	publisher   port.EventPublisher
	log         zerolog.Logger
}

func NewCommunityService(
	communities port.CommunityRepository,
	follows port.CommunityFollowRepository,
	stars port.StarRepository,
	bans port.BanRepository,
	mods port.ModeratorRepository,
	users port.UserRepository,
	bannedWords port.BannedWordRepository,
	publisher port.EventPublisher,
	log zerolog.Logger,
) *CommunityService {
	return &CommunityService{communities: communities, follows: follows, stars: stars, bans: bans, mods: mods, users: users, bannedWords: bannedWords, publisher: publisher, log: log}
}

func (s *CommunityService) Create(ctx context.Context, ownerID uuid.UUID, input domain.CreateCommunityInput) (*domain.Community, error) {
	if banned, _ := s.bannedWords.IsWordBanned(ctx, input.Slug, domain.BannedWordScopeSlug); banned {
		return nil, fmt.Errorf("название содержит запрещённое слово: %w", domain.ErrForbidden)
	}
	c := &domain.Community{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Slug:        input.Slug,
		Name:        input.Name,
		Description: input.Description,
		AvatarURL:   input.AvatarURL,
		BannerURL:   input.BannerURL,
		Tags:        []string{},
		Rules:       input.Rules,
	}
	if c.Rules == nil {
		c.Rules = []string{}
	}
	if err := s.communities.Create(ctx, c); err != nil {
		return nil, err
	}
	_ = s.publisher.Publish(ctx, "community.created", map[string]any{
		"community_id": c.ID,
		"owner_id":     ownerID,
		"slug":         c.Slug,
		"name":         c.Name,
		"description":  c.Description,
	})
	metrics.CommunitiesCreated.Inc()
	return c, nil
}

func (s *CommunityService) GetBySlug(ctx context.Context, slug string) (*domain.Community, error) {
	return s.communities.GetBySlug(ctx, slug)
}

func (s *CommunityService) GetBySlugForViewer(ctx context.Context, slug string, viewerID *uuid.UUID) (*domain.Community, error) {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if viewerID != nil {
		if followed, err := s.follows.IsFollowing(ctx, *viewerID, c.ID); err == nil {
			c.Followed = followed
		}
		if starred, err := s.stars.IsStarred(ctx, *viewerID, c.ID); err == nil {
			c.Starred = starred
		}
	}
	return c, nil
}

func (s *CommunityService) List(ctx context.Context, sort string, limit, offset int) ([]domain.Community, error) {
	return s.communities.List(ctx, sort, limit, offset)
}

func (s *CommunityService) GetByOwner(ctx context.Context, ownerID uuid.UUID) (*domain.Community, error) {
	return s.communities.GetByOwner(ctx, ownerID)
}

func (s *CommunityService) Update(ctx context.Context, ownerID uuid.UUID, slug string, input domain.UpdateCommunityInput) (*domain.Community, error) {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if c.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	if input.Name != nil {
		c.Name = *input.Name
	}
	if input.Description != nil {
		c.Description = input.Description
	}
	if input.AvatarURL != nil {
		c.AvatarURL = input.AvatarURL
	}
	if input.BannerURL != nil {
		c.BannerURL = input.BannerURL
	}
	if input.Rules != nil {
		c.Rules = input.Rules
	}
	if err = s.communities.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CommunityService) Delete(ctx context.Context, ownerID uuid.UUID, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	return s.communities.Delete(ctx, c.ID)
}

func (s *CommunityService) SetTags(ctx context.Context, ownerID uuid.UUID, slug string, tags []string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	return s.communities.SetTags(ctx, c.ID, tags)
}

func (s *CommunityService) Follow(ctx context.Context, userID uuid.UUID, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	banned, _ := s.bans.IsBanned(ctx, c.ID, userID)
	if banned {
		return domain.ErrForbidden
	}
	inserted, err := s.follows.Follow(ctx, userID, c.ID)
	if err != nil {
		return err
	}
	if inserted {
		_ = s.communities.IncrFollowers(ctx, c.ID, 1)
		_ = s.publisher.Publish(ctx, "community.followed", map[string]any{
			"user_id":      userID,
			"community_id": c.ID,
		})
	}
	return nil
}

func (s *CommunityService) Unfollow(ctx context.Context, userID uuid.UUID, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	deleted, err := s.follows.Unfollow(ctx, userID, c.ID)
	if err != nil {
		return err
	}
	if deleted {
		_ = s.communities.IncrFollowers(ctx, c.ID, -1)
	}
	return nil
}

func (s *CommunityService) GetFollowed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Community, error) {
	return s.follows.GetFollowed(ctx, userID, limit, offset)
}

func (s *CommunityService) GetMembers(ctx context.Context, slug string, limit, offset int) ([]domain.User, error) {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return s.follows.GetFollowers(ctx, c.ID, limit, offset)
}

func (s *CommunityService) Star(ctx context.Context, userID uuid.UUID, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	following, err := s.follows.IsFollowing(ctx, userID, c.ID)
	if err != nil || !following {
		return domain.ErrForbidden
	}
	inserted, err := s.stars.Star(ctx, userID, c.ID)
	if err != nil {
		return err
	}
	if inserted {
		_ = s.communities.IncrStars(ctx, c.ID, 1)
		_ = s.publisher.Publish(ctx, "community.starred", map[string]any{
			"user_id":      userID,
			"community_id": c.ID,
		})
	}
	return nil
}

func (s *CommunityService) Unstar(ctx context.Context, userID uuid.UUID, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	deleted, err := s.stars.Unstar(ctx, userID, c.ID)
	if err != nil {
		return err
	}
	if deleted {
		_ = s.communities.IncrStars(ctx, c.ID, -1)
	}
	return nil
}

func (s *CommunityService) Ban(ctx context.Context, ownerID uuid.UUID, slug string, targetUserID uuid.UUID, banType domain.BanType, reason *string, expiresAt *time.Time) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	b := &domain.CommunityBan{
		CommunityID: c.ID,
		UserID:      targetUserID,
		Type:        banType,
		Reason:      reason,
		ExpiresAt:   expiresAt,
	}
	return s.bans.Ban(ctx, b)
}

func (s *CommunityService) Unban(ctx context.Context, ownerID uuid.UUID, slug string, targetUserID uuid.UUID) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	return s.bans.Unban(ctx, c.ID, targetUserID)
}

func (s *CommunityService) GetBans(ctx context.Context, ownerID uuid.UUID, slug string) ([]domain.CommunityBan, error) {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if c.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	return s.bans.GetBans(ctx, c.ID)
}

func (s *CommunityService) GetModerators(ctx context.Context, slug string) ([]domain.CommunityModerator, error) {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return s.mods.GetByCommunity(ctx, c.ID)
}

func (s *CommunityService) AddModerator(ctx context.Context, ownerID uuid.UUID, slug, username, role string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	target, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return s.mods.Add(ctx, c.ID, target.ID, role)
}

func (s *CommunityService) RemoveModerator(ctx context.Context, ownerID uuid.UUID, slug, username string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if c.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	target, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	return s.mods.Remove(ctx, c.ID, target.ID)
}

func (s *CommunityService) AdminDelete(ctx context.Context, slug string) error {
	c, err := s.communities.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	return s.communities.Delete(ctx, c.ID)
}
