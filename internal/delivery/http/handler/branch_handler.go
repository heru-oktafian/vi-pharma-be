package handler

import (
	"errors"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
	"pharma-be/internal/usecase"
)

type BranchHandler struct {
	branchUsecase usecase.BranchUsecase
}

func NewBranchHandler(branchUsecase usecase.BranchUsecase) *BranchHandler {
	return &BranchHandler{branchUsecase: branchUsecase}
}

// ListBranches handles GET /api/v1/branches
func (h *BranchHandler) ListBranches(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	branches, total, err := h.branchUsecase.ListBranches(c.Context(), page, pageSize)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to list branches", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return SuccessResponse(c, fiber.StatusOK, "Branches retrieved successfully", PaginatedData{
		Items: branches,
		Meta: PaginationMeta{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}

// GetBranchByID handles GET /api/v1/branches/:id
func (h *BranchHandler) GetBranchByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID is required", nil)
	}

	branch, err := h.branchUsecase.GetBranchByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBranchNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Branch not found", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve branch", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Branch retrieved successfully", branch)
}

// CreateBranch handles POST /api/v1/branches
func (h *BranchHandler) CreateBranch(c *fiber.Ctx) error {
	var branch domain.Branch
	if err := c.BodyParser(&branch); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if branch.BranchCode == "" || branch.BranchName == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "branch_code and branch_name are required", nil)
	}

	created, err := h.branchUsecase.CreateBranch(c.Context(), &branch)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create branch", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Branch created successfully", created)
}

type AssignBranchRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	BranchID  string `json:"branch_id" validate:"required"`
	IsDefault bool   `json:"is_default"`
}

// AssignUserBranch handles POST /api/v1/branches/assign
func (h *BranchHandler) AssignUserBranch(c *fiber.Ctx) error {
	var req AssignBranchRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.UserID == "" || req.BranchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "user_id and branch_id are required", nil)
	}

	if err := h.branchUsecase.AssignUserBranch(c.Context(), req.UserID, req.BranchID, req.IsDefault); err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Failed to assign user to branch", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "User successfully assigned to branch", nil)
}
