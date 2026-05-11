package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type NotificationService struct {
	repo        port.NotificationRepository
	prefs       port.NotificationPreferenceRepository
	lookup      port.EntityLookup
	users       port.UserRepository
	posts       port.PostRepository
	comments    port.CommentRepository
	communities port.CommunityRepository
	log         zerolog.Logger
}

func NewNotificationService(
	repo port.NotificationRepository,
	prefs port.NotificationPreferenceRepository,
	lookup port.EntityLookup,
	users port.UserRepository,
	posts port.PostRepository,
	comments port.CommentRepository,
	communities port.CommunityRepository,
	log zerolog.Logger,
) *NotificationService {
	return &NotificationService{
		repo: repo, prefs: prefs, lookup: lookup,
		users: users, posts: posts, comments: comments, communities: communities,
		log: log,
	}
}

func (s *NotificationService) HandleEvent(ctx context.Context, subject string, data []byte) error {
	targetUserID, err := s.resolveTarget(ctx, subject, data)
	if err != nil {
		s.log.Warn().Err(err).Str("subject", subject).Msg("cannot resolve notification target")
		return nil
	}
	if targetUserID == uuid.Nil {
		return nil
	}
	enabled, _ := s.prefs.IsEnabled(ctx, targetUserID, subject)
	if !enabled {
		return nil
	}
	n := &domain.Notification{
		ID:      uuid.New(),
		UserID:  targetUserID,
		Type:    subject,
		Payload: json.RawMessage(data),
	}
	if err = s.repo.Create(ctx, n); err != nil {
		s.log.Error().Err(err).Str("subject", subject).Msg("failed to create notification")
		return err
	}
	return nil
}

func (s *NotificationService) resolveTarget(ctx context.Context, subject string, data []byte) (uuid.UUID, error) {
	var p map[string]any
	if err := json.Unmarshal(data, &p); err != nil {
		return uuid.Nil, err
	}
	parseID := func(key string) (uuid.UUID, error) {
		raw, ok := p[key].(string)
		if !ok {
			return uuid.Nil, nil
		}
		return uuid.Parse(raw)
	}

	switch subject {
	case "user.followed":
		return parseID("following_id")
	case "friendship.requested":
		return parseID("addressee_id")
	case "friendship.accepted":
		return parseID("requester_id")
	case "community.followed", "community.starred":
		communityID, err := parseID("community_id")
		if err != nil || communityID == uuid.Nil {
			return uuid.Nil, err
		}
		return s.lookup.GetCommunityOwnerID(ctx, communityID)
	case "comment.created":
		postID, err := parseID("post_id")
		if err != nil || postID == uuid.Nil {
			return uuid.Nil, err
		}
		return s.lookup.GetPostAuthorID(ctx, postID)
	case "post.voted":
		postID, err := parseID("post_id")
		if err != nil || postID == uuid.Nil {
			return uuid.Nil, err
		}
		authorID, err := s.lookup.GetPostAuthorID(ctx, postID)
		if err != nil {
			return uuid.Nil, err
		}
		voterID, _ := parseID("user_id")
		if authorID == voterID {
			return uuid.Nil, nil
		}
		return authorID, nil
	case "mention.created":
		return parseID("target_user_id")
	}
	return uuid.Nil, nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Notification, error) {
	notifs, err := s.repo.GetByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	s.enrich(ctx, notifs)
	return notifs, nil
}

func (s *NotificationService) enrich(ctx context.Context, notifs []domain.Notification) {
	userCache := map[uuid.UUID]*domain.User{}
	postCache := map[uuid.UUID]*domain.Post{}
	commCache := map[uuid.UUID]*domain.Community{}

	resolveUser := func(id uuid.UUID) *domain.NotifActor {
		if id == uuid.Nil {
			return nil
		}
		u, ok := userCache[id]
		if !ok {
			fetched, err := s.users.GetByID(ctx, id)
			if err == nil {
				userCache[id] = fetched
				u = fetched
			} else {
				userCache[id] = nil
			}
		}
		if u == nil {
			return nil
		}
		return &domain.NotifActor{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL}
	}
	resolvePost := func(id uuid.UUID) (*domain.NotifPostRef, *domain.Post) {
		if id == uuid.Nil {
			return nil, nil
		}
		p, ok := postCache[id]
		if !ok {
			fetched, err := s.posts.GetByID(ctx, id)
			if err == nil {
				postCache[id] = fetched
				p = fetched
			} else {
				postCache[id] = nil
			}
		}
		if p == nil {
			return nil, nil
		}
		title := ""
		if p.Title != nil {
			title = *p.Title
		}
		ref := &domain.NotifPostRef{ID: p.ID, Title: title}
		if p.CommunityID != nil {
			if c := resolveCommunityCached(ctx, s, commCache, *p.CommunityID); c != nil {
				slug := c.Slug
				ref.CommunitySlug = &slug
			}
		}
		return ref, p
	}
	resolveCommunity := func(id uuid.UUID) *domain.NotifCommRef {
		c := resolveCommunityCached(ctx, s, commCache, id)
		if c == nil {
			return nil
		}
		return &domain.NotifCommRef{ID: c.ID, Slug: c.Slug, Name: c.Name, AvatarURL: c.AvatarURL}
	}

	parseUUID := func(p map[string]any, key string) uuid.UUID {
		raw, _ := p[key].(string)
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil
		}
		return id
	}

	for i := range notifs {
		n := &notifs[i]
		var p map[string]any
		if err := json.Unmarshal(n.Payload, &p); err != nil {
			continue
		}
		var actorID uuid.UUID
		switch n.Type {
		case "user.followed":
			actorID = parseUUID(p, "follower_id")
		case "friendship.requested":
			actorID = parseUUID(p, "requester_id")
		case "friendship.accepted":
			actorID = parseUUID(p, "addressee_id")
		case "community.followed", "community.starred":
			actorID = parseUUID(p, "user_id")
		case "comment.created":
			actorID = parseUUID(p, "author_id")
		case "post.voted":
			actorID = parseUUID(p, "user_id")
		case "post.created":
			actorID = parseUUID(p, "author_id")
		case "mention.created", "mention":
			actorID = parseUUID(p, "actor_id")
		}
		n.Actor = resolveUser(actorID)

		if postID := parseUUID(p, "post_id"); postID != uuid.Nil {
			ref, _ := resolvePost(postID)
			n.Post = ref
		}
		if commID := parseUUID(p, "community_id"); commID != uuid.Nil {
			n.Community = resolveCommunity(commID)
		}

		if n.Type == "comment.created" {
			if cid := parseUUID(p, "comment_id"); cid != uuid.Nil {
				if c, err := s.comments.GetByID(ctx, cid); err == nil {
					n.Snippet = trimSnippet(c.Content, 140)
				}
			}
		}
	}
}

func resolveCommunityCached(ctx context.Context, s *NotificationService, cache map[uuid.UUID]*domain.Community, id uuid.UUID) *domain.Community {
	if id == uuid.Nil {
		return nil
	}
	if c, ok := cache[id]; ok {
		return c
	}
	c, err := s.communities.GetByID(ctx, id)
	if err != nil {
		cache[id] = nil
		return nil
	}
	cache[id] = c
	return c
}

func trimSnippet(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) ([]domain.NotificationPreference, error) {
	return s.prefs.Get(ctx, userID)
}

func (s *NotificationService) SetPreference(ctx context.Context, userID uuid.UUID, notifType string, enabled bool) error {
	return s.prefs.Set(ctx, userID, notifType, enabled)
}
