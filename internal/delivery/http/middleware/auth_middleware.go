package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/delivery/http/handler"
	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
	"pharma-be/pkg/utils"
)

// JWTProtected verifies the Bearer token in Authorization header
func JWTProtected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return handler.ErrorResponse(c, fiber.StatusUnauthorized, "Missing Authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return handler.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid token format, expected 'Bearer <token>'", nil)
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			return handler.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token", err)
		}

		// Store user and session info in Fiber context locals
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("branchID", claims.BranchID)
		c.Locals("division", claims.Division)

		return c.Next()
	}
}

// RequireRoles restricts route access to specific roles
func RequireRoles(allowedRoles ...domain.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: role not found", nil)
		}

		for _, allowed := range allowedRoles {
			if string(allowed) == role {
				return c.Next()
			}
		}

		return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: insufficient role permissions", nil)
	}
}

// RequireBranch ensures the active request has selected a valid branch context
func RequireBranch(allowedBranches ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		branchID, ok := c.Locals("branchID").(string)
		if !ok || branchID == "" {
			return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: please select an active branch first", nil)
		}

		if len(allowedBranches) == 0 {
			return c.Next()
		}

		for _, allowed := range allowedBranches {
			if allowed == branchID {
				return c.Next()
			}
		}

		return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: branch mismatch", nil)
	}
}

// RequireDivision ensures the user has selected an active division session
func RequireDivision(allowedDivisions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		division, ok := c.Locals("division").(string)
		if !ok || division == "" {
			return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: please select a division first", nil)
		}

		if len(allowedDivisions) == 0 {
			return c.Next()
		}

		for _, allowed := range allowedDivisions {
			if allowed == division {
				return c.Next()
			}
		}

		return handler.ErrorResponse(c, fiber.StatusForbidden, "Access forbidden: division mismatch", nil)
	}
}
