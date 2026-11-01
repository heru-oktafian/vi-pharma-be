package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Username == "" || req.Password == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Username and password are required", nil)
	}

	res, err := h.authUsecase.Login(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return ErrorResponse(c, fiber.StatusUnauthorized, "Invalid username or password", nil)
		case errors.Is(err, domain.ErrUserInactive):
			return ErrorResponse(c, fiber.StatusForbidden, "User account is inactive", nil)
		case errors.Is(err, domain.ErrUserSuspended):
			return ErrorResponse(c, fiber.StatusForbidden, "User account is suspended", nil)
		default:
			return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to login", err)
		}
	}

	return SuccessResponse(c, fiber.StatusOK, "Login successful", res)
}

// SelectDivision handles POST /api/v1/auth/select-division
func (h *AuthHandler) SelectDivision(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized access", nil)
	}

	var req domain.SelectDivisionRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Division == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Division name is required", nil)
	}

	res, err := h.authUsecase.SelectDivision(c.Context(), userID, req.Division)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthorizedDivision):
			return ErrorResponse(c, fiber.StatusForbidden, "You do not have access to this division", nil)
		case errors.Is(err, domain.ErrUserNotFound):
			return ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		case errors.Is(err, domain.ErrUserInactive):
			return ErrorResponse(c, fiber.StatusForbidden, "User account is inactive", nil)
		default:
			return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to select division", err)
		}
	}

	return SuccessResponse(c, fiber.StatusOK, "Division selected successfully", res)
}

// GetMe handles GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized access", nil)
	}

	profile, err := h.authUsecase.GetProfile(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get user profile", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "User profile retrieved successfully", profile)
}
