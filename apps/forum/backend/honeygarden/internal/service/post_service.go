package service

import (
	"context"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
	"honeygarden/internal/port"
)

var mentionRegexp = regexp.MustCompile(`(?:^|[^\w])@([a-zA-Z0-9_-]{2,32})`)

func extractMentions(text string) []string {
	matches := mentionRegexp.FindAllStringSubmatch(text, -1)
	seen := map[string]bool{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

type PostService struct {
	posts        port.PostRepository
	comments     port.CommentRepository
	media        port.MediaRepository
	community    port.CommunityAccess
	bookmarks    port.BookmarkRepository
	blocks       port.UserBlockRepository
	publisher    port.EventPublisher
	userResolver port.UserResolver
	slugResolver port.CommunitySlugResolver
	sources      port.SourceRepository
	log          zerolog.Logger
}

func NewPostService(
	posts port.PostRepository,
	comments port.CommentRepository,
	media port.MediaRepository,
	community port.CommunityAccess,
	bookmarks port.BookmarkRepository,
	blocks port.UserBlockRepository,
	publisher port.EventPublisher,
	userResolver port.UserResolver,
	slugResolver port.CommunitySlugResolver,
	sources port.SourceRepository,
	log zerolog.Logger,
) *PostService {
	return &PostService{
		posts: posts, comments: comments, media: media,
		community: community, bookmarks: bookmarks, blocks: blocks,
		publisher: publisher,
		userResolver: userResolver, slugResolver: slugResolver,
		sources: sources, log: log,
	}
}

func (s *PostService) Create(ctx context.Context, authorID uuid.UUID, communitySlug string, input domain.CreatePostInput) (*domain.Post, error) {
	communityID, err := s.community.GetCommunityIDBySlug(ctx, communitySlug)
	if err != nil {
		return nil, err
	}

	isOwner, _ := s.community.IsOwner(ctx, authorID, communityID)
	if !isOwner {
		return nil, domain.ErrForbidden
	}
	if input.Title != nil && len(*input.Title) > 300 {
		return nil, domain.ErrInvalidInput
	}
	if len(input.Content) > 100_000 {
		return nil, domain.ErrInvalidInput
	}
	kind := input.Kind
	if kind == "" {
		kind = domain.PostKindDiscussion
	}
	if !kind.Valid() {
		return nil, domain.ErrInvalidInput
	}
	p := &domain.Post{
		ID:          uuid.New(),
		CommunityID: &communityID,
		AuthorID:    &authorID,
		Kind:        kind,
		Title:       input.Title,
		Content:     input.Content,
		Media:       []domain.PostMedia{},
	}
	if err = s.posts.Create(ctx, p); err != nil {
		return nil, err
	}
	for _, m := range input.Media {
		pm := &domain.PostMedia{
			ID:     uuid.New(),
			PostID: p.ID,
			Type:   m.Type,
			URL:    m.URL,
			Name:   m.Name,
			Size:   m.Size,
		}
		if err = s.media.Create(ctx, pm); err != nil {
			return nil, err
		}
		p.Media = append(p.Media, *pm)
	}
	_ = s.publisher.Publish(ctx, "post.created", map[string]any{
		"post_id":      p.ID,
		"community_id": communityID,
		"author_id":    authorID,
		"title":        p.Title,
		"content":      p.Content,
		"created_at":   p.CreatedAt.Unix(),
	})
	metrics.PostsCreated.WithLabelValues(string(kind)).Inc()
	s.dispatchMentions(ctx, p.Content, authorID, &p.ID, nil)
	return p, nil
}

func (s *PostService) dispatchMentions(ctx context.Context, text string, actorID uuid.UUID, postID, commentID *uuid.UUID) {
	usernames := extractMentions(text)
	if len(usernames) == 0 {
		return
	}
	resolved, err := s.userResolver.ResolveUsernames(ctx, usernames)
	if err != nil || len(resolved) == 0 {
		return
	}
	for _, targetID := range resolved {
		if targetID == actorID {
			continue
		}
		payload := map[string]any{
			"target_user_id": targetID,
			"actor_id":       actorID,
		}
		if postID != nil {
			payload["post_id"] = *postID
		}
		if commentID != nil {
			payload["comment_id"] = *commentID
		}
		_ = s.publisher.Publish(ctx, "mention.created", payload)
	}
}

func (s *PostService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.posts.IncrViews(ctx, id)
	return p, nil
}

func (s *PostService) GetByIDOrShort(ctx context.Context, raw string) (*domain.Post, error) {
	if id, err := uuid.Parse(raw); err == nil {
		return s.GetByID(ctx, id)
	}
	p, err := s.posts.GetByShortID(ctx, raw)
	if err != nil {
		return nil, err
	}
	_ = s.posts.IncrViews(ctx, p.ID)
	return p, nil
}

func (s *PostService) GetByCommunity(ctx context.Context, communitySlug string, limit, offset int) ([]domain.Post, error) {
	communityID, err := s.community.GetCommunityIDBySlug(ctx, communitySlug)
	if err != nil {
		return nil, err
	}
	return s.posts.GetByCommunity(ctx, communityID, limit, offset)
}

func (s *PostService) GetByCommunityKind(ctx context.Context, communitySlug string, kind domain.PostKind, limit, offset int) ([]domain.Post, error) {
	communityID, err := s.community.GetCommunityIDBySlug(ctx, communitySlug)
	if err != nil {
		return nil, err
	}
	return s.posts.GetByCommunityKind(ctx, communityID, kind, limit, offset)
}

func (s *PostService) GetTrending(ctx context.Context, limit int) ([]domain.Post, error) {
	since := time.Now().UTC().Add(-24 * time.Hour)
	return s.posts.GetTrending(ctx, since, limit)
}

func (s *PostService) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	return s.posts.GetByAuthor(ctx, authorID, limit, offset)
}

func (s *PostService) Update(ctx context.Context, authorID, postID uuid.UUID, input domain.UpdatePostInput) (*domain.Post, error) {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.IsExternal() || p.AuthorID == nil || *p.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}
	if input.Kind != nil {
		if !input.Kind.Valid() {
			return nil, domain.ErrInvalidInput
		}
		p.Kind = *input.Kind
	}
	if input.Title != nil {
		if len(*input.Title) > 300 {
			return nil, domain.ErrInvalidInput
		}
		p.Title = input.Title
	}
	if input.Content != nil {
		if len(*input.Content) > 100_000 {
			return nil, domain.ErrInvalidInput
		}
		p.Content = *input.Content
	}
	if err = s.posts.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) Delete(ctx context.Context, requesterID, postID uuid.UUID) error {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if p.IsExternal() {
		return domain.ErrForbidden
	}
	if p.AuthorID != nil && *p.AuthorID == requesterID {
		return s.posts.Delete(ctx, postID)
	}
	if p.CommunityID == nil {
		return domain.ErrForbidden
	}
	isOwner, _ := s.community.IsOwner(ctx, requesterID, *p.CommunityID)
	if !isOwner {
		return domain.ErrForbidden
	}
	return s.posts.Delete(ctx, postID)
}

func (s *PostService) Vote(ctx context.Context, userID, postID uuid.UUID) error {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	// Внешние посты (RSS) — голосовать может любой залогиненный пользователь;
	// проверка членства в комьюнити пропускается, т.к. комьюнити нет.
	if !p.IsExternal() && p.CommunityID != nil {
		if err = s.requireMember(ctx, userID, *p.CommunityID); err != nil {
			return err
		}
	}
	if err = s.posts.AddVote(ctx, userID, postID); err != nil {
		return err
	}
	metrics.PostVotes.Inc()
	_ = s.publisher.Publish(ctx, "post.voted", map[string]any{"user_id": userID, "post_id": postID})
	return nil
}

func (s *PostService) Unvote(ctx context.Context, userID, postID uuid.UUID) error {
	return s.posts.RemoveVote(ctx, userID, postID)
}

func (s *PostService) Pin(ctx context.Context, requesterID, postID uuid.UUID) error {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if p.CommunityID == nil {
		return domain.ErrForbidden
	}
	isOwner, _ := s.community.IsOwner(ctx, requesterID, *p.CommunityID)
	if !isOwner {
		return domain.ErrForbidden
	}
	return s.posts.SetPinned(ctx, postID, true)
}

func (s *PostService) Unpin(ctx context.Context, requesterID, postID uuid.UUID) error {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if p.CommunityID == nil {
		return domain.ErrForbidden
	}
	isOwner, _ := s.community.IsOwner(ctx, requesterID, *p.CommunityID)
	if !isOwner {
		return domain.ErrForbidden
	}
	return s.posts.SetPinned(ctx, postID, false)
}

func (s *PostService) CreateComment(ctx context.Context, authorID, postID uuid.UUID, input domain.CreateCommentInput) (*domain.Comment, error) {
	p, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	// Внешние посты — комментировать может любой залогиненный.
	if !p.IsExternal() && p.CommunityID != nil {
		if err = s.requireMember(ctx, authorID, *p.CommunityID); err != nil {
			return nil, err
		}
	}
	if input.Content == "" || len(input.Content) > 10_000 {
		return nil, domain.ErrInvalidInput
	}
	c := &domain.Comment{
		ID:       uuid.New(),
		PostID:   postID,
		AuthorID: authorID,
		ParentID: input.ParentID,
		Content:  input.Content,
	}
	if err = s.comments.Create(ctx, c); err != nil {
		return nil, err
	}
	payload := map[string]any{
		"comment_id": c.ID,
		"post_id":    postID,
		"author_id":  authorID,
	}
	if p.CommunityID != nil {
		payload["community_id"] = *p.CommunityID
	}
	_ = s.publisher.Publish(ctx, "comment.created", payload)
	metrics.CommentsCreated.Inc()
	s.dispatchMentions(ctx, c.Content, authorID, &postID, &c.ID)
	return c, nil
}

func (s *PostService) GetComments(ctx context.Context, postID uuid.UUID, limit, offset int) ([]domain.Comment, error) {
	return s.comments.GetByPost(ctx, postID, limit, offset)
}

func (s *PostService) DeleteComment(ctx context.Context, requesterID, commentID uuid.UUID) error {
	c, err := s.comments.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if c.AuthorID == requesterID {
		return s.comments.Delete(ctx, commentID)
	}
	p, err := s.posts.GetByID(ctx, c.PostID)
	if err != nil {
		return err
	}
	if p.CommunityID == nil {
		return domain.ErrForbidden
	}
	isOwner, _ := s.community.IsOwner(ctx, requesterID, *p.CommunityID)
	if !isOwner {
		return domain.ErrForbidden
	}
	return s.comments.Delete(ctx, commentID)
}

func (s *PostService) VoteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	if err := s.comments.AddVote(ctx, userID, commentID); err != nil {
		return err
	}
	metrics.CommentVotes.Inc()
	return nil
}

func (s *PostService) UnvoteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	return s.comments.RemoveVote(ctx, userID, commentID)
}

func (s *PostService) Bookmark(ctx context.Context, userID, postID uuid.UUID) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return err
	}
	if err := s.bookmarks.Add(ctx, userID, postID); err != nil {
		return err
	}
	metrics.BookmarksAdded.Inc()
	return nil
}

func (s *PostService) Unbookmark(ctx context.Context, userID, postID uuid.UUID) error {
	return s.bookmarks.Remove(ctx, userID, postID)
}

func (s *PostService) GetBookmarks(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Post, error) {
	return s.bookmarks.GetByUser(ctx, userID, limit, offset)
}

func (s *PostService) EnrichPost(ctx context.Context, p *domain.Post, viewerID *uuid.UUID) (*domain.PostView, error) {
	views, err := s.EnrichPosts(ctx, []domain.Post{*p}, viewerID)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, domain.ErrNotFound
	}
	return &views[0], nil
}

func (s *PostService) EnrichPosts(ctx context.Context, posts []domain.Post, viewerID *uuid.UUID) ([]domain.PostView, error) {
	if len(posts) == 0 {
		return []domain.PostView{}, nil
	}

	if viewerID != nil {
		blockSet, err := s.blocks.ListBlockSet(ctx, *viewerID)
		if err == nil && len(blockSet) > 0 {
			filtered := posts[:0]
			for _, p := range posts {
				if p.AuthorID != nil && blockSet[*p.AuthorID] {
					continue
				}
				filtered = append(filtered, p)
			}
			posts = filtered
			if len(posts) == 0 {
				return []domain.PostView{}, nil
			}
		}
	}
	authorIDs := make([]uuid.UUID, 0, len(posts))
	communityIDs := make([]uuid.UUID, 0, len(posts))
	sourceIDs := make([]uuid.UUID, 0, len(posts))
	seenAuthors := map[uuid.UUID]bool{}
	seenCommunities := map[uuid.UUID]bool{}
	seenSources := map[uuid.UUID]bool{}
	for _, p := range posts {
		if p.AuthorID != nil && !seenAuthors[*p.AuthorID] {
			authorIDs = append(authorIDs, *p.AuthorID)
			seenAuthors[*p.AuthorID] = true
		}
		if p.CommunityID != nil && !seenCommunities[*p.CommunityID] {
			communityIDs = append(communityIDs, *p.CommunityID)
			seenCommunities[*p.CommunityID] = true
		}
		if p.SourceID != nil && !seenSources[*p.SourceID] {
			sourceIDs = append(sourceIDs, *p.SourceID)
			seenSources[*p.SourceID] = true
		}
	}
	users, err := s.userResolver.ResolveUsers(ctx, authorIDs)
	if err != nil {
		return nil, err
	}
	slugs, err := s.slugResolver.ResolveCommunitySlugs(ctx, communityIDs)
	if err != nil {
		return nil, err
	}
	sources, err := s.sources.ResolveSources(ctx, sourceIDs)
	if err != nil {
		return nil, err
	}
	var voted map[uuid.UUID]bool
	var bookmarked map[uuid.UUID]bool
	if viewerID != nil {
		postIDs := make([]uuid.UUID, len(posts))
		for i, p := range posts {
			postIDs[i] = p.ID
		}
		voted, err = s.posts.BatchIsVoted(ctx, *viewerID, postIDs)
		if err != nil {
			return nil, err
		}
		bookmarked, err = s.bookmarks.BatchIsBookmarked(ctx, *viewerID, postIDs)
		if err != nil {
			return nil, err
		}
	}
	views := make([]domain.PostView, 0, len(posts))
	for _, p := range posts {
		if p.AuthorID != nil {
			if _, ok := users[*p.AuthorID]; !ok && p.SourceID == nil {
				continue
			}
		}
		view := domain.PostView{Post: p}
		if p.AuthorID != nil {
			if u, ok := users[*p.AuthorID]; ok {
				view.Author = &u
			}
		}
		if p.CommunityID != nil {
			view.CommunitySlug = slugs[*p.CommunityID]
		}
		if p.SourceID != nil {
			if src, ok := sources[*p.SourceID]; ok {
				view.Source = &src
			}
		}
		view.IsVoted = voted[p.ID]
		view.IsBookmarked = bookmarked[p.ID]
		views = append(views, view)
	}
	return views, nil
}

func (s *PostService) EnrichComments(ctx context.Context, comments []domain.Comment, viewerID *uuid.UUID) ([]domain.CommentView, error) {
	if len(comments) == 0 {
		return []domain.CommentView{}, nil
	}
	if viewerID != nil {
		blockSet, err := s.blocks.ListBlockSet(ctx, *viewerID)
		if err == nil && len(blockSet) > 0 {
			filtered := comments[:0]
			for _, c := range comments {
				if blockSet[c.AuthorID] {
					continue
				}
				filtered = append(filtered, c)
			}
			comments = filtered
			if len(comments) == 0 {
				return []domain.CommentView{}, nil
			}
		}
	}
	authorIDs := make([]uuid.UUID, 0, len(comments))
	seen := map[uuid.UUID]bool{}
	for _, c := range comments {
		if !seen[c.AuthorID] {
			authorIDs = append(authorIDs, c.AuthorID)
			seen[c.AuthorID] = true
		}
	}
	users, err := s.userResolver.ResolveUsers(ctx, authorIDs)
	if err != nil {
		return nil, err
	}
	views := make([]domain.CommentView, len(comments))
	for i, c := range comments {
		views[i] = domain.CommentView{
			Comment: c,
			Author:  users[c.AuthorID],
		}
	}
	return views, nil
}

func (s *PostService) EnrichComment(ctx context.Context, c *domain.Comment) (*domain.CommentView, error) {
	users, err := s.userResolver.ResolveUsers(ctx, []uuid.UUID{c.AuthorID})
	if err != nil {
		return nil, err
	}
	return &domain.CommentView{
		Comment: *c,
		Author:  users[c.AuthorID],
	}, nil
}

func (s *PostService) requireMember(ctx context.Context, userID, communityID uuid.UUID) error {
	isOwner, _ := s.community.IsOwner(ctx, userID, communityID)
	if isOwner {
		return nil
	}
	isFollower, _ := s.community.IsFollower(ctx, userID, communityID)
	if !isFollower {
		return domain.ErrForbidden
	}
	return nil
}

func (s *PostService) AdminDelete(ctx context.Context, postID uuid.UUID) error {
	return s.posts.Delete(ctx, postID)
}

func (s *PostService) AdminDeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return s.comments.Delete(ctx, commentID)
}
