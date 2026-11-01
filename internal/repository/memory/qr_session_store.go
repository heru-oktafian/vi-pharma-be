package memory

import (
	"context"
	"sync"
	"time"

	"pharma-be/internal/domain"
)

type qrSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.QRSession
	stopChan chan struct{}
}

// NewQRSessionStore creates a thread-safe in-memory store with auto-cleanup background worker
func NewQRSessionStore() domain.QRSessionStore {
	store := &qrSessionStore{
		sessions: make(map[string]*domain.QRSession),
		stopChan: make(chan struct{}),
	}

	// Start auto-cleanup worker every 30 seconds
	go store.startCleanupWorker(30 * time.Second)

	return store
}

func (s *qrSessionStore) Save(ctx context.Context, session *domain.QRSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionID] = session
	return nil
}

func (s *qrSessionStore) Get(ctx context.Context, sessionID string) (*domain.QRSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, domain.ErrQRSessionNotFound
	}

	if session.IsExpired() {
		return nil, domain.ErrQRSessionExpired
	}

	return session, nil
}

func (s *qrSessionStore) Update(ctx context.Context, session *domain.QRSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[session.SessionID]; !exists {
		return domain.ErrQRSessionNotFound
	}

	s.sessions[session.SessionID] = session
	return nil
}

func (s *qrSessionStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

func (s *qrSessionStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cleaned := 0

	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) || session.Status == domain.QRSessionStatusConsumed {
			delete(s.sessions, id)
			cleaned++
		}
	}

	return cleaned
}

func (s *qrSessionStore) startCleanupWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CleanupExpired()
		case <-s.stopChan:
			return
		}
	}
}
