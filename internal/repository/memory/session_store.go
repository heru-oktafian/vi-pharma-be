package memory

import (
	"context"
	"sync"
	"time"

	"pharma-be/internal/domain"
)

type sessionStore struct {
	mu            sync.RWMutex
	userSessions  map[string]*domain.WorkstationSession // userID -> active session
	revokedTokens map[string]time.Time                  // tokenString -> revocation time
	stopChan      chan struct{}
}

// NewSessionStore creates a thread-safe in-memory store for workstation sessions and revoked tokens
func NewSessionStore() domain.SessionStore {
	store := &sessionStore{
		userSessions:  make(map[string]*domain.WorkstationSession),
		revokedTokens: make(map[string]time.Time),
		stopChan:      make(chan struct{}),
	}

	// Auto-cleanup worker for revoked tokens older than 24h
	go store.startCleanupWorker(10 * time.Minute)

	return store
}

func (s *sessionStore) CreateSession(ctx context.Context, session *domain.WorkstationSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If there's an existing active session for this user, revoke its old token
	if old, exists := s.userSessions[session.UserID]; exists && old != nil && old.Token != "" {
		s.revokedTokens[old.Token] = time.Now()
	}

	session.IsActive = true
	session.LastSeenAt = time.Now()
	s.userSessions[session.UserID] = session
	return nil
}

func (s *sessionStore) GetActiveSessionByUserID(ctx context.Context, userID string) (*domain.WorkstationSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.userSessions[userID]
	if !exists || session == nil || !session.IsActive {
		return nil, domain.ErrSessionNotFound
	}

	return session, nil
}

func (s *sessionStore) RevokeSessionByUserID(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.userSessions[userID]
	if exists && session != nil {
		session.IsActive = false
		if session.Token != "" {
			s.revokedTokens[session.Token] = time.Now()
		}
		delete(s.userSessions, userID)
	}

	return nil
}

func (s *sessionStore) IsTokenRevoked(token string) bool {
	if token == "" {
		return true
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, revoked := s.revokedTokens[token]
	return revoked
}

func (s *sessionStore) CleanupExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleaned := 0
	cutoff := time.Now().Add(-24 * time.Hour)

	for token, revokedAt := range s.revokedTokens {
		if revokedAt.Before(cutoff) {
			delete(s.revokedTokens, token)
			cleaned++
		}
	}

	return cleaned
}

func (s *sessionStore) startCleanupWorker(interval time.Duration) {
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
