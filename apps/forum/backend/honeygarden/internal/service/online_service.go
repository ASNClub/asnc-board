package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const onlineTTL = 5 * time.Minute

type OnlineService struct {
	mu    sync.RWMutex
	users map[uuid.UUID]time.Time
}

func NewOnlineService() *OnlineService {
	s := &OnlineService{users: make(map[uuid.UUID]time.Time)}
	go s.cleanup()
	return s
}

func (s *OnlineService) Heartbeat(userID uuid.UUID) {
	s.mu.Lock()
	s.users[userID] = time.Now()
	s.mu.Unlock()
}

func (s *OnlineService) IsOnline(userID uuid.UUID) bool {
	s.mu.RLock()
	t, ok := s.users[userID]
	s.mu.RUnlock()
	return ok && time.Since(t) < onlineTTL
}

func (s *OnlineService) OnlineUsers(userIDs []uuid.UUID) map[uuid.UUID]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[uuid.UUID]bool, len(userIDs))
	now := time.Now()
	for _, id := range userIDs {
		if t, ok := s.users[id]; ok && now.Sub(t) < onlineTTL {
			result[id] = true
		}
	}
	return result
}

func (s *OnlineService) OnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	now := time.Now()
	for _, t := range s.users {
		if now.Sub(t) < onlineTTL {
			count++
		}
	}
	return count
}

func (s *OnlineService) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, t := range s.users {
			if now.Sub(t) >= onlineTTL {
				delete(s.users, id)
			}
		}
		s.mu.Unlock()
	}
}
