package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"pharma-be/internal/delivery/http/handler"
	"pharma-be/internal/delivery/http/middleware"
	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
)

// SetupRoutes registers all application endpoints on the Fiber app
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	qrAuthHandler *handler.QRAuthHandler,
	userHandler *handler.UserHandler,
	branchHandler *handler.BranchHandler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// 1. Health check endpoint
	v1.Get("/health", func(c *fiber.Ctx) error {
		return handler.SuccessResponse(c, fiber.StatusOK, "Vi-Pharma Backend Service is healthy", fiber.Map{
			"app":     cfg.AppName,
			"version": "1.0.0",
			"status":  "online",
		})
	})

	// 2. Authentication & QR Pairing Group
	auth := v1.Group("/auth")

	// 2.1 Standard Mobile Login (vi-token uses this to get mobile JWT)
	auth.Post("/login", authHandler.Login)

	// 2.2 QR Code Flow (vi-pharma web pairing with vi-token mobile)
	qr := auth.Group("/qr")
	qr.Post("/generate", qrAuthHandler.GenerateQR)                  // Web requests QR
	qr.Get("/status/:sessionId", qrAuthHandler.GetSessionStatus)    // Polling fallback
	qr.Post("/select-branch", qrAuthHandler.SelectBranch)           // Web chooses branch

	// 2.3 WebSocket upgrade middleware & endpoint for vi-pharma web
	qr.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	qr.Get("/ws/:sessionId", qrAuthHandler.HandleWebSocket())

	// 2.4 Mobile QR Authorization (Protected by Mobile JWT)
	qrProtected := qr.Group("")
	qrProtected.Use(middleware.JWTProtected(cfg))
	qrProtected.Post("/authorize", qrAuthHandler.AuthorizeQR)

	// 2.5 Profile Info (Protected)
	authProtected := auth.Group("")
	authProtected.Use(middleware.JWTProtected(cfg))
	authProtected.Get("/me", authHandler.GetMe)

	// 3. User Management (Protected: SuperAdmin / Admin)
	users := v1.Group("/users")
	users.Use(middleware.JWTProtected(cfg))
	users.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin))
	users.Post("", userHandler.CreateUser)
	users.Get("", userHandler.ListUsers)
	users.Get("/:id", userHandler.GetUserByID)

	// 4. Branch Management (Protected: SuperAdmin / Admin)
	branches := v1.Group("/branches")
	branches.Use(middleware.JWTProtected(cfg))
	branches.Use(middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleAdmin))
	branches.Post("", branchHandler.CreateBranch)
	branches.Get("", branchHandler.ListBranches)
	branches.Get("/:id", branchHandler.GetBranchByID)
	branches.Post("/assign", branchHandler.AssignUserBranch)

	// 5. Workstation / Active Branch Features (Protected by Selected Branch)
	workstation := v1.Group("/workstation")
	workstation.Use(middleware.JWTProtected(cfg))
	workstation.Use(middleware.RequireBranch())

	workstation.Get("/current", func(c *fiber.Ctx) error {
		userID := c.Locals("userID")
		role := c.Locals("role")
		branchID := c.Locals("branchID")

		return handler.SuccessResponse(c, fiber.StatusOK, "Active branch session verified", fiber.Map{
			"user_id":   userID,
			"role":      role,
			"branch_id": branchID,
		})
	})
}
