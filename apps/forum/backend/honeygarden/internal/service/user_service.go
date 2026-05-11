package service

import (
	"context"
	"fmt"

	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
	"honeygarden/internal/port"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type UserService struct {
	users       port.UserRepository
	follows     port.UserFollowRepository
	friendships port.FriendshipRepository
	blocks      port.UserBlockRepository
	pinnedRepos port.PinnedRepoRepository
	bannedWords port.BannedWordRepository
	publisher   port.EventPublisher
	log         zerolog.Logger
}

func NewUserService(
	users port.UserRepository,
	follows port.UserFollowRepository,
	friendships port.FriendshipRepository,
	blocks port.UserBlockRepository,
	pinnedRepos port.PinnedRepoRepository,
	bannedWords port.BannedWordRepository,
	publisher port.EventPublisher,
	log zerolog.Logger,
) *UserService {
	return &UserService{
		users:       users,
		follows:     follows,
		friendships: friendships,
		blocks:      blocks,
		pinnedRepos: pinnedRepos,
		bannedWords: bannedWords,
		publisher:   publisher,
		log:         log,
	}
}

func (s *UserService) GetOrProvision(ctx context.Context, authID, username, displayName string) (*domain.User, error) {
	user, err := s.users.GetByAuthID(ctx, authID)
	if err == nil {
		changed := false
		if username != authID && user.Username == authID {
			user.Username = username
			changed = true
		}
		if displayName != authID && user.DisplayName == authID {
			user.DisplayName = displayName
			changed = true
		}
		if changed {
			_ = s.users.Update(ctx, user)
		}
		return user, nil
	}
	if err != domain.ErrNotFound {
		return nil, err
	}
	user = &domain.User{
		ID:          uuid.New(),
		AuthID:      authID,
		Username:    username,
		DisplayName: displayName,
		Privacy:     domain.PrivacyPublic,
		Tags:        []string{},
		Platforms:   []domain.UserPlatform{},
	}
	if err = s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	metrics.UsersRegistered.Inc()
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.users.GetByUsername(ctx, username)
}

func (s *UserService) Search(ctx context.Context, query string, limit int) ([]domain.User, error) {
	return s.users.Search(ctx, query, limit)
}

func (s *UserService) UpdateMe(ctx context.Context, userID uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if input.Username != nil {
		if banned, _ := s.bannedWords.IsWordBanned(ctx, *input.Username, domain.BannedWordScopeUsername); banned {
			return nil, fmt.Errorf("имя содержит запрещённое слово: %w", domain.ErrForbidden)
		}
		user.Username = *input.Username
	}
	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.BannerURL != nil {
		user.BannerURL = input.BannerURL
	}
	if input.Privacy != nil {
		user.Privacy = *input.Privacy
	}
	if input.OnboardingDone != nil {
		user.OnboardingDone = *input.OnboardingDone
	}
	if input.ShowActivity != nil {
		user.ShowActivity = *input.ShowActivity
	}
	if err = s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) SetTags(ctx context.Context, userID uuid.UUID, tags []string) error {
	return s.users.SetTags(ctx, userID, tags)
}

func (s *UserService) SetPlatforms(ctx context.Context, userID uuid.UUID, platforms []domain.UserPlatform) error {
	for i := range platforms {
		platforms[i].ID = uuid.New()
		platforms[i].UserID = userID
	}
	return s.users.SetPlatforms(ctx, userID, platforms)
}

func (s *UserService) Follow(ctx context.Context, followerID uuid.UUID, targetUsername string) error {
	target, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	if followerID == target.ID {
		return domain.ErrInvalidInput
	}
	if err = s.follows.Follow(ctx, followerID, target.ID); err != nil {
		return err
	}
	_ = s.publisher.Publish(ctx, "user.followed", map[string]any{
		"follower_id":  followerID,
		"following_id": target.ID,
	})
	return nil
}

func (s *UserService) Unfollow(ctx context.Context, followerID uuid.UUID, targetUsername string) error {
	target, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	return s.follows.Unfollow(ctx, followerID, target.ID)
}

func (s *UserService) Block(ctx context.Context, blockerID uuid.UUID, targetUsername string) error {
	target, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	if blockerID == target.ID {
		return domain.ErrInvalidInput
	}
	if err = s.blocks.Block(ctx, blockerID, target.ID); err != nil {
		return err
	}
	// на ошибки тут в целом похуй, не срем ими
	_ = s.follows.Unfollow(ctx, blockerID, target.ID)
	_ = s.follows.Unfollow(ctx, target.ID, blockerID)
	return nil
}

func (s *UserService) Unblock(ctx context.Context, blockerID uuid.UUID, targetUsername string) error {
	target, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	return s.blocks.Unblock(ctx, blockerID, target.ID)
}

func (s *UserService) ListBlocks(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	return s.blocks.ListBlocks(ctx, userID)
}

func (s *UserService) GetFollowers(ctx context.Context, username string, limit, offset int) ([]domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return s.follows.GetFollowers(ctx, user.ID, limit, offset)
}

func (s *UserService) GetFollowing(ctx context.Context, username string, limit, offset int) ([]domain.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return s.follows.GetFollowing(ctx, user.ID, limit, offset)
}

func (s *UserService) RequestFriendship(ctx context.Context, requesterID uuid.UUID, targetUsername string) error {
	target, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	if requesterID == target.ID {
		return domain.ErrInvalidInput
	}
	existing, err := s.friendships.Get(ctx, requesterID, target.ID)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil {
		if existing.Status == domain.FriendshipAccepted {
			return domain.ErrAlreadyExists
		}
		if existing.RequesterID == target.ID && existing.Status == domain.FriendshipPending {
			if err = s.friendships.UpdateStatus(ctx, target.ID, requesterID, domain.FriendshipAccepted); err != nil {
				return err
			}
			_ = s.publisher.Publish(ctx, "friendship.accepted", map[string]any{
				"requester_id": target.ID,
				"addressee_id": requesterID,
			})
			return nil
		}
		return domain.ErrAlreadyExists
	}
	f := &domain.Friendship{
		ID:          uuid.New(),
		RequesterID: requesterID,
		AddresseeID: target.ID,
		Status:      domain.FriendshipPending,
	}
	if err = s.friendships.Create(ctx, f); err != nil {
		return err
	}
	_ = s.publisher.Publish(ctx, "friendship.requested", map[string]any{
		"requester_id": requesterID,
		"addressee_id": target.ID,
	})
	return nil
}

func (s *UserService) AcceptFriendship(ctx context.Context, addresseeID uuid.UUID, requesterUsername string) error {
	requester, err := s.users.GetByUsername(ctx, requesterUsername)
	if err != nil {
		return err
	}
	f, err := s.friendships.Get(ctx, requester.ID, addresseeID)
	if err != nil {
		return err
	}
	if f.AddresseeID != addresseeID {
		return domain.ErrForbidden
	}
	if err = s.friendships.UpdateStatus(ctx, requester.ID, addresseeID, domain.FriendshipAccepted); err != nil {
		return err
	}
	_ = s.publisher.Publish(ctx, "friendship.accepted", map[string]any{
		"requester_id": requester.ID,
		"addressee_id": addresseeID,
	})
	return nil
}

func (s *UserService) RejectFriendship(ctx context.Context, addresseeID uuid.UUID, requesterUsername string) error {
	requester, err := s.users.GetByUsername(ctx, requesterUsername)
	if err != nil {
		return err
	}
	f, err := s.friendships.Get(ctx, requester.ID, addresseeID)
	if err != nil {
		return err
	}
	if f.AddresseeID != addresseeID {
		return domain.ErrForbidden
	}
	return s.friendships.UpdateStatus(ctx, requester.ID, addresseeID, domain.FriendshipRejected)
}

func (s *UserService) GetFriends(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.User, error) {
	return s.friendships.GetFriends(ctx, userID, limit, offset)
}

func (s *UserService) GetPendingRequesters(ctx context.Context, addresseeID uuid.UUID) ([]domain.User, error) {
	friendships, err := s.friendships.GetPending(ctx, addresseeID)
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(friendships))
	for _, f := range friendships {
		u, err := s.users.GetByID(ctx, f.RequesterID)
		if err != nil {
			s.log.Warn().Err(err).Str("requester_id", f.RequesterID.String()).Msg("pending: user not found")
			continue
		}
		users = append(users, *u)
	}
	return users, nil
}

func (s *UserService) GetPinnedRepos(ctx context.Context, userID uuid.UUID) ([]domain.PinnedRepo, error) {
	return s.pinnedRepos.GetByUser(ctx, userID)
}

func (s *UserService) BanUser(ctx context.Context, targetUsername string, ban bool) error {
	u, err := s.users.GetByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}
	return s.users.SetBanned(ctx, u.ID, ban)
}
