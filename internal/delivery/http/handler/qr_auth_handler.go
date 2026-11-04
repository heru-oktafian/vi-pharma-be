package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	ws "pharma-be/internal/delivery/http/websocket"
	"pharma-be/internal/domain"
	"pharma-be/internal/usecase"
	"pharma-be/pkg/config"
	"pharma-be/pkg/utils"
)

type QRAuthHandler struct {
	qrUsecase    usecase.QRAuthUsecase
	sessionStore domain.SessionStore
	wsHub        *ws.Hub
	cfg          *config.Config
}

func NewQRAuthHandler(
	qrUsecase usecase.QRAuthUsecase,
	sessionStore domain.SessionStore,
	wsHub *ws.Hub,
	cfg *config.Config,
) *QRAuthHandler {
	return &QRAuthHandler{
		qrUsecase:    qrUsecase,
		sessionStore: sessionStore,
		wsHub:        wsHub,
		cfg:          cfg,
	}
}

// GenerateQR handles POST /api/v1/auth/qr/generate (Web client requests new QR)
func (h *QRAuthHandler) GenerateQR(c *fiber.Ctx) error {
	res, err := h.qrUsecase.GenerateSession(c.Context())
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate QR session", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "QR session generated successfully", res)
}

// AuthorizeQR handles POST /api/v1/auth/qr/authorize (Mobile vi-token scans QR and approves)
func (h *QRAuthHandler) AuthorizeQR(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized mobile user", nil)
	}

	var req domain.AuthorizeQRRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.SessionID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "session_id is required", nil)
	}

	session, err := h.qrUsecase.AuthorizeSession(c.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQRSessionNotFound):
			return ErrorResponse(c, fiber.StatusNotFound, "QR session not found or expired", nil)
		case errors.Is(err, domain.ErrQRSessionExpired):
			return ErrorResponse(c, fiber.StatusGone, "QR session has expired", nil)
		case errors.Is(err, domain.ErrQRSessionAlreadyAuthorized):
			return ErrorResponse(c, fiber.StatusConflict, "QR session already authorized", nil)
		case errors.Is(err, domain.ErrUserInactive):
			return ErrorResponse(c, fiber.StatusForbidden, "User account is inactive", nil)
		case errors.Is(err, domain.ErrUserSuspended):
			return ErrorResponse(c, fiber.StatusForbidden, "User account is suspended", nil)
		default:
			return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to authorize QR session", err)
		}
	}

	return SuccessResponse(c, fiber.StatusOK, "QR session authorized successfully", fiber.Map{
		"session_id": session.SessionID,
		"status":     session.Status,
		"branches":   session.Branches,
	})
}

// SelectBranch handles POST /api/v1/auth/qr/select-branch (Web user selects affiliated branch)
func (h *QRAuthHandler) SelectBranch(c *fiber.Ctx) error {
	var req domain.SelectBranchRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.SessionID == "" || req.BranchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "session_id and branch_id are required", nil)
	}

	res, err := h.qrUsecase.SelectBranch(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQRSessionNotFound):
			return ErrorResponse(c, fiber.StatusNotFound, "QR session not found or expired", nil)
		case errors.Is(err, domain.ErrQRSessionExpired):
			return ErrorResponse(c, fiber.StatusGone, "QR session has expired", nil)
		case errors.Is(err, domain.ErrQRSessionNotAuthorized):
			return ErrorResponse(c, fiber.StatusForbidden, "QR session has not been scanned/authorized by mobile yet", nil)
		case errors.Is(err, domain.ErrUnauthorizedBranch):
			return ErrorResponse(c, fiber.StatusForbidden, "User does not have access to this branch", nil)
		case errors.Is(err, domain.ErrBranchNotFound):
			return ErrorResponse(c, fiber.StatusNotFound, "Branch not found", nil)
		default:
			return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to select branch", err)
		}
	}

	return SuccessResponse(c, fiber.StatusOK, "Branch selected and web session created", res)
}

// GetSessionStatus handles GET /api/v1/auth/qr/status/:sessionId (Fallback polling endpoint)
func (h *QRAuthHandler) GetSessionStatus(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "sessionId is required", nil)
	}

	session, err := h.qrUsecase.GetSession(c.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrQRSessionExpired) {
			return ErrorResponse(c, fiber.StatusGone, "QR session has expired", nil)
		}
		return ErrorResponse(c, fiber.StatusNotFound, "QR session not found", nil)
	}

	return SuccessResponse(c, fiber.StatusOK, "QR session status retrieved", session)
}

// HandleWebSocket returns a Fiber WebSocket handler for GET /api/v1/auth/qr/ws/:sessionId
func (h *QRAuthHandler) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		sessionID := c.Params("sessionId")
		if sessionID == "" {
			_ = c.WriteJSON(domain.WSMessage{
				Event: domain.WSEventError,
				Data:  "Missing sessionId in websocket URL",
			})
			_ = c.Close()
			return
		}

		// Register connection with Hub
		h.wsHub.Register(sessionID, c)
		defer h.wsHub.Unregister(sessionID)

		// Send initial CONNECTED confirmation
		_ = c.WriteJSON(domain.WSMessage{
			Event: domain.WSEventConnected,
			Data: map[string]string{
				"session_id": sessionID,
				"status":     "LISTENING",
			},
		})

		// Keep connection open and listen for client messages / pings / disconnects
		for {
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				// Client disconnected
				break
			}
			// Ping-Pong response
			if messageType == websocket.PingMessage {
				_ = c.WriteMessage(websocket.PongMessage, []byte{})
			} else if string(msg) == "ping" {
				_ = c.WriteJSON(fiber.Map{"event": "PONG"})
			}
		}

		log.Printf("[WS] Connection closed for sessionId: %s", sessionID)
	})
}

// RemoteLogout handles POST /api/v1/auth/remote-logout (Called by vi-token to terminate web workstation session)
func (h *QRAuthHandler) RemoteLogout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized mobile user", nil)
	}

	// Revoke in session store
	_ = h.sessionStore.RevokeSessionByUserID(c.Context(), userID)

	// Broadcast REMOTE_LOGOUT event to all workstation WebSockets for this user
	_ = h.wsHub.BroadcastToWorkstation(userID, &domain.WSMessage{
		Event: "REMOTE_LOGOUT",
		Data: map[string]interface{}{
			"user_id": userID,
			"message": "Sesi workstation Anda telah diakhiri secara remote dari aplikasi vi-token. Silakan scan ulang untuk masuk kembali.",
			"reason":  "USER_TERMINATED",
		},
	})

	// Close workstation connections
	h.wsHub.CloseWorkstationConnections(userID)

	return SuccessResponse(c, fiber.StatusOK, "Sesi workstation web berhasil diputuskan dari vi-token", nil)
}

// GetActiveWorkstation handles GET /api/v1/auth/active-workstation (Called by vi-token to check web session)
func (h *QRAuthHandler) GetActiveWorkstation(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized mobile user", nil)
	}

	session, err := h.sessionStore.GetActiveSessionByUserID(c.Context(), userID)
	if err != nil || session == nil {
		return SuccessResponse(c, fiber.StatusOK, "Tidak ada workstation aktif", domain.WorkstationStatusResponse{
			HasActiveSession: false,
		})
	}

	return SuccessResponse(c, fiber.StatusOK, "Workstation aktif ditemukan", domain.WorkstationStatusResponse{
		HasActiveSession: true,
		Session:          session,
	})
}

// HandleWorkstationWebSocket returns a Fiber WebSocket handler for GET /api/v1/workstation/ws?token=<webToken>
func (h *QRAuthHandler) HandleWorkstationWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		if token == "" {
			_ = c.WriteJSON(domain.WSMessage{
				Event: domain.WSEventError,
				Data:  "Token required in query parameter",
			})
			_ = c.Close()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(token, h.cfg.JWTSecret)
		if err != nil {
			_ = c.WriteJSON(domain.WSMessage{
				Event: domain.WSEventError,
				Data:  "Invalid or expired web token",
			})
			_ = c.Close()
			return
		}

		// Check if already revoked
		if h.sessionStore.IsTokenRevoked(token) {
			_ = c.WriteJSON(domain.WSMessage{
				Event: "REMOTE_LOGOUT",
				Data:  "Session has already been revoked from vi-token",
			})
			_ = c.Close()
			return
		}

		userID := claims.UserID
		h.wsHub.RegisterWorkstation(userID, c)
		defer h.wsHub.UnregisterWorkstation(userID, c)

		// Send initial WORKSTATION_CONNECTED
		_ = c.WriteJSON(domain.WSMessage{
			Event: "WORKSTATION_CONNECTED",
			Data: map[string]string{
				"user_id":   userID,
				"username":  claims.Username,
				"branch_id": claims.BranchID,
				"status":    "ACTIVE",
			},
		})

		// Keep connection open and listen for pings / disconnects
		for {
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if messageType == websocket.PingMessage {
				_ = c.WriteMessage(websocket.PongMessage, []byte{})
			} else if string(msg) == "ping" {
				_ = c.WriteJSON(fiber.Map{"event": "PONG"})
			}
		}

		log.Printf("[Workstation WS] Closed connection for user: %s", userID)
	})
}
