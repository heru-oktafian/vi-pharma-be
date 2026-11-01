package memory

import (
	"context"
	"testing"
	"time"

	"pharma-be/internal/domain"
)

func TestQRSessionStore(t *testing.T) {
	store := NewQRSessionStore()
	ctx := context.Background()

	session := &domain.QRSession{
		SessionID: "test-session-123",
		Status:    domain.QRSessionStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}

	// 1. Test Save
	err := store.Save(ctx, session)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// 2. Test Get
	retrieved, err := store.Get(ctx, "test-session-123")
	if err != nil {
		t.Fatalf("Failed to retrieve session: %v", err)
	}
	if retrieved.SessionID != "test-session-123" {
		t.Errorf("Expected session_id 'test-session-123', got '%s'", retrieved.SessionID)
	}
	if retrieved.Status != domain.QRSessionStatusPending {
		t.Errorf("Expected status PENDING, got %s", retrieved.Status)
	}

	// 3. Test Update
	retrieved.Status = domain.QRSessionStatusAuthorized
	err = store.Update(ctx, retrieved)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	updated, _ := store.Get(ctx, "test-session-123")
	if updated.Status != domain.QRSessionStatusAuthorized {
		t.Errorf("Expected status AUTHORIZED, got %s", updated.Status)
	}

	// 4. Test Expired Session
	expiredSession := &domain.QRSession{
		SessionID: "expired-session-456",
		Status:    domain.QRSessionStatusPending,
		CreatedAt: time.Now().Add(-5 * time.Minute),
		ExpiresAt: time.Now().Add(-1 * time.Minute), // in the past
	}
	_ = store.Save(ctx, expiredSession)

	_, err = store.Get(ctx, "expired-session-456")
	if err != domain.ErrQRSessionExpired {
		t.Errorf("Expected ErrQRSessionExpired, got %v", err)
	}

	// 5. Test Cleanup
	cleaned := store.CleanupExpired()
	if cleaned != 1 {
		t.Errorf("Expected 1 session cleaned, got %d", cleaned)
	}
}
