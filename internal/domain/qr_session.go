package domain

import (
	"context"
	"time"
)

type QRSessionStatus string

const (
	QRSessionStatusPending    QRSessionStatus = "PENDING"
	QRSessionStatusAuthorized QRSessionStatus = "AUTHORIZED"
	QRSessionStatusConsumed   QRSessionStatus = "CONSUMED"
	QRSessionStatusExpired    QRSessionStatus = "EXPIRED"
)

// QRSession represents temporary state for QR pairing
type QRSession struct {
	SessionID        string                  `json:"session_id"`
	Status           QRSessionStatus         `json:"status"`
	UserID           string                  `json:"user_id,omitempty"`
	User             *UserResponse           `json:"user,omitempty"`
	Branches         []BranchSummaryResponse `json:"branches,omitempty"`
	SelectedBranchID string                  `json:"selected_branch_id,omitempty"`
	DeviceID         string                  `json:"device_id,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	ExpiresAt        time.Time               `json:"expires_at"`
}

func (s *QRSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Request and Response DTOs for QR Flow
type GenerateQRResponse struct {
	SessionID        string    `json:"session_id"`
	QRString         string    `json:"qr_string"`
	ExpiresAt        time.Time `json:"expires_at"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type AuthorizeQRRequest struct {
	SessionID  string `json:"session_id" validate:"required"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	ActionType string `json:"action_type,omitempty"` // For future vi-payroll extension: WEB_LOGIN, ATTENDANCE, etc.
}

type SelectBranchRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	BranchID  string `json:"branch_id" validate:"required"`
}

type SelectBranchResponse struct {
	Token  string                `json:"token"`
	Branch BranchSummaryResponse `json:"branch"`
	User   UserResponse          `json:"user"`
}

// WebSocket Event Types
const (
	WSEventConnected  = "CONNECTED"
	WSEventAuthorized = "QR_AUTHORIZED"
	WSEventConsumed   = "QR_CONSUMED"
	WSEventExpired    = "QR_EXPIRED"
	WSEventError      = "ERROR"
)

type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data,omitempty"`
}

// QRSessionStore interface
type QRSessionStore interface {
	Save(ctx context.Context, session *QRSession) error
	Get(ctx context.Context, sessionID string) (*QRSession, error)
	Update(ctx context.Context, session *QRSession) error
	Delete(ctx context.Context, sessionID string) error
	CleanupExpired() int
}
