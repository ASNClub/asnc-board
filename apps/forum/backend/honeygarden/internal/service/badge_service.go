package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type badgeRule struct {
	ID    string
	Check func(ctx context.Context, stats port.BadgeStatsProvider, userID uuid.UUID) (bool, error)
}

var rules = []badgeRule{
	{"first-honey", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserMaxPostVotes(ctx, u)
		return n >= 1, err
	}},
	{"bee-keeper", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		return s.HasCommunity(ctx, u)
	}},
	{"pollinator", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserFollowedCommunitiesCount(ctx, u)
		return n >= 5, err
	}},
	{"hive-mind", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserCommentCount(ctx, u)
		return n >= 10, err
	}},
	{"honey-flow", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserPostCount(ctx, u)
		return n >= 10, err
	}},
	{"queen-bee", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserCommunityFollowers(ctx, u)
		return n >= 100, err
	}},
	{"golden-comb", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserMaxPostVotes(ctx, u)
		return n >= 50, err
	}},
	{"nectar-hunter", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserCommunityStars(ctx, u)
		return n >= 10, err
	}},
	{"royal-jelly", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserReputation(ctx, u)
		return n >= 500, err
	}},
	{"swarm-leader", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserCommunityFollowers(ctx, u)
		return n >= 1000, err
	}},
	{"honey-badger", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserReputation(ctx, u)
		return n >= 1000, err
	}},
	{"mythical-bloom", func(ctx context.Context, s port.BadgeStatsProvider, u uuid.UUID) (bool, error) {
		n, err := s.UserMaxPostVotes(ctx, u)
		return n >= 500, err
	}},
}

type BadgeService struct {
	badges port.BadgeRepository
	stats  port.BadgeStatsProvider
	log    zerolog.Logger
}

func NewBadgeService(badges port.BadgeRepository, stats port.BadgeStatsProvider, log zerolog.Logger) *BadgeService {
	return &BadgeService{badges: badges, stats: stats, log: log}
}

func (s *BadgeService) ListDefinitions(ctx context.Context) ([]domain.BadgeDefinition, error) {
	return s.badges.ListDefinitions(ctx)
}

func (s *BadgeService) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.UserBadge, error) {
	return s.badges.GetUserBadges(ctx, userID)
}

// CheckAndAward проверяет все правила и начисляет новые бейджи.
// Вызывается после действий пользователя
func (s *BadgeService) CheckAndAward(ctx context.Context, userID uuid.UUID) {
	for _, rule := range rules {
		has, _ := s.badges.HasBadge(ctx, userID, rule.ID)
		if has {
			continue
		}
		ok, err := rule.Check(ctx, s.stats, userID)
		if err != nil {
			s.log.Warn().Err(err).Str("badge", rule.ID).Msg("badge check failed")
			continue
		}
		if ok {
			if err = s.badges.Award(ctx, userID, rule.ID); err != nil {
				s.log.Warn().Err(err).Str("badge", rule.ID).Msg("badge award failed")
			}
		}
	}
}
