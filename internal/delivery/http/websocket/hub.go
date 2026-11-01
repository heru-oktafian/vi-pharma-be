package websocket

import (
	"sync"

	"github.com/gofiber/websocket/v2"

	"pharma-be/internal/domain"
)

// Hub manages active WebSocket connections for QR code sessions
type Hub struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
}

// NewHub creates a new WebSocket connection hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*websocket.Conn),
	}
}

// Register adds a WebSocket connection mapped to a sessionID
func (h *Hub) Register(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// If old connection exists, close it
	if oldConn, exists := h.connections[sessionID]; exists && oldConn != nil {
		_ = oldConn.Close()
	}

	h.connections[sessionID] = conn
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.connections, sessionID)
}

// BroadcastToSession sends a JSON payload to the specific session's WebSocket client
func (h *Hub) BroadcastToSession(sessionID string, msg *domain.WSMessage) error {
	h.mu.RLock()
	conn, exists := h.connections[sessionID]
	h.mu.RUnlock()

	if !exists || conn == nil {
		return nil
	}

	return conn.WriteJSON(msg)
}
