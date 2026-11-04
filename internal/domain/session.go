package domain

import (
	"context"
	"time"
)

// WorkstationSession represents an active workstation web login
type WorkstationSession struct {
	SessionID  string    `json:"session_id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	BranchID   string    `json:"branch_id"`
	BranchName string    `json:"branch_name"`
	Token      string    `json:"token,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	IsActive   bool      `json:"is_active"`
}

// WorkstationStatusResponse represents DTO sent to mobile app
type WorkstationStatusResponse struct {
	HasActiveSession bool                `json:"has_active_session"`
	Session          *WorkstationSession `json:"session,omitempty"`
}

// SessionStore manages active workstation sessions and token revocation
type SessionStore interface {
	CreateSession(ctx context.Context, session *WorkstationSession) error
	GetActiveSessionByUserID(ctx context.Context, userID string) (*WorkstationSession, error)
	RevokeSessionByUserID(ctx context.Context, userID string) error
	IsTokenRevoked(token string) bool
	CleanupExpired() int
}
