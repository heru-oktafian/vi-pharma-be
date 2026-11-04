package websocket

import (
	"sync"

	"github.com/gofiber/websocket/v2"

	"pharma-be/internal/domain"
)

// Hub manages active WebSocket connections for QR code sessions and active workstations
type Hub struct {
	mu           sync.RWMutex
	connections  map[string]*websocket.Conn          // sessionID -> QR listener conn
	workstations map[string]map[*websocket.Conn]bool // userID -> set of workstation conns
}

// NewHub creates a new WebSocket connection hub
func NewHub() *Hub {
	return &Hub{
		connections:  make(map[string]*websocket.Conn),
		workstations: make(map[string]map[*websocket.Conn]bool),
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

// RegisterWorkstation attaches an active web workstation connection for a user
func (h *Hub) RegisterWorkstation(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.workstations[userID]; !exists {
		h.workstations[userID] = make(map[*websocket.Conn]bool)
	}
	h.workstations[userID][conn] = true
}

// UnregisterWorkstation removes a workstation connection
func (h *Hub) UnregisterWorkstation(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, exists := h.workstations[userID]; exists {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.workstations, userID)
		}
	}
}

// BroadcastToWorkstation sends a JSON payload to all active workstations for a user
func (h *Hub) BroadcastToWorkstation(userID string, msg *domain.WSMessage) error {
	h.mu.RLock()
	conns, exists := h.workstations[userID]
	if !exists || len(conns) == 0 {
		h.mu.RUnlock()
		return nil
	}

	// Copy pointers under read lock to write safely
	activeList := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		if conn != nil {
			activeList = append(activeList, conn)
		}
	}
	h.mu.RUnlock()

	for _, conn := range activeList {
		_ = conn.WriteJSON(msg)
	}
	return nil
}

// CloseWorkstationConnections closes and cleans up all active WebSocket connections for a user
func (h *Hub) CloseWorkstationConnections(userID string) {
	h.mu.Lock()
	conns, exists := h.workstations[userID]
	if !exists {
		h.mu.Unlock()
		return
	}

	delete(h.workstations, userID)
	h.mu.Unlock()

	for conn := range conns {
		if conn != nil {
			_ = conn.Close()
		}
	}
}
