package handler

import (
	"errors"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req domain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Username == "" || req.Password == "" || req.Name == "" || req.UserRole == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "username, password, name, and user_role are required", nil)
	}

	user, err := h.userUsecase.CreateUser(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsernameAlreadyExists):
			return ErrorResponse(c, fiber.StatusConflict, "Username is already taken", nil)
		case errors.Is(err, domain.ErrInvalidRole):
			return ErrorResponse(c, fiber.StatusBadRequest, "Invalid user role. Allowed: superadmin, admin, pharmacist, cashier, inventory", nil)
		case errors.Is(err, domain.ErrInvalidStatus):
			return ErrorResponse(c, fiber.StatusBadRequest, "Invalid user status. Allowed: active, inactive, suspended", nil)
		default:
			return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err)
		}
	}

	return SuccessResponse(c, fiber.StatusCreated, "User created successfully", user)
}

// GetUserByID handles GET /api/v1/users/:id
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "User ID is required", nil)
	}

	user, err := h.userUsecase.GetUserByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "User retrieved successfully", user)
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	users, total, err := h.userUsecase.ListUsers(c.Context(), page, pageSize)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve users", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return SuccessResponse(c, fiber.StatusOK, "Users retrieved successfully", PaginatedData{
		Items: users,
		Meta: PaginationMeta{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}
