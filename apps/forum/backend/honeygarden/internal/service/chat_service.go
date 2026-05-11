package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
	"honeygarden/internal/port"
)

type ChatService struct {
	repo  port.ChatRepository
	users port.UserResolver

	mu          sync.RWMutex
	subscribers map[chan domain.ChatMessage]struct{}
}

func NewChatService(repo port.ChatRepository, users port.UserResolver) *ChatService {
	return &ChatService{
		repo:        repo,
		users:       users,
		subscribers: make(map[chan domain.ChatMessage]struct{}),
	}
}

func (s *ChatService) Send(ctx context.Context, authorID uuid.UUID, content string) (*domain.ChatMessage, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, domain.ErrInvalidInput
	}
	if len([]rune(content)) > 1000 {
		return nil, domain.ErrInvalidInput
	}
	m := &domain.ChatMessage{
		ID:       uuid.New(),
		AuthorID: authorID,
		Content:  content,
	}
	if err := s.repo.Insert(ctx, m); err != nil {
		return nil, err
	}
	if briefs, err := s.users.ResolveUsers(ctx, []uuid.UUID{authorID}); err == nil {
		if b, ok := briefs[authorID]; ok {
			m.Author = &b
		}
	}
	if m.Author == nil {
		return nil, errors.New("chat: failed to resolve author")
	}
	m.CreatedAt = time.Now().UTC()
	metrics.ChatSent.Inc()
	s.broadcast(*m)
	return m, nil
}

func (s *ChatService) List(ctx context.Context, limit int) ([]domain.ChatMessage, error) {
	msgs, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return msgs, nil
	}
	ids := make([]uuid.UUID, 0, len(msgs))
	seen := map[uuid.UUID]struct{}{}
	for _, m := range msgs {
		if _, ok := seen[m.AuthorID]; ok {
			continue
		}
		seen[m.AuthorID] = struct{}{}
		ids = append(ids, m.AuthorID)
	}
	briefs, err := s.users.ResolveUsers(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range msgs {
		if b, ok := briefs[msgs[i].AuthorID]; ok {
			msgs[i].Author = &b
		}
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *ChatService) Subscribe() chan domain.ChatMessage {
	ch := make(chan domain.ChatMessage, 32)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *ChatService) Unsubscribe(ch chan domain.ChatMessage) {
	s.mu.Lock()
	if _, ok := s.subscribers[ch]; ok {
		delete(s.subscribers, ch)
		close(ch)
	}
	s.mu.Unlock()
}

func (s *ChatService) broadcast(m domain.ChatMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers {
		select {
		case ch <- m:
		default:
		}
	}
}
