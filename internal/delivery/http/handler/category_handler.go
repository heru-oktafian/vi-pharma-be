package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type CategoryHandler struct {
	categoryUsecase domain.CategoryUsecase
}

func NewCategoryHandler(categoryUsecase domain.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{categoryUsecase: categoryUsecase}
}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	search := c.Query("search", "")
	activeOnly := c.Query("active_only", "false") == "true"

	categories, err := h.categoryUsecase.GetAllCategories(c.Context(), search, activeOnly)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data kategori", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data kategori berhasil diambil", categories)
}

// GetCategoryByID handles GET /api/v1/categories/:id
func (h *CategoryHandler) GetCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID kategori wajib diisi", nil)
	}

	cat, err := h.categoryUsecase.GetCategoryByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Kategori tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data kategori", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data kategori ditemukan", cat)
}

// CreateCategory handles POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req domain.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Name == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nama kategori wajib diisi", nil)
	}

	cat, err := h.categoryUsecase.CreateCategory(c.Context(), &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan kategori baru", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Kategori baru berhasil ditambahkan", cat)
}

// UpdateCategory handles PUT /api/v1/categories/:id
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID kategori wajib diisi", nil)
	}

	var req domain.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	cat, err := h.categoryUsecase.UpdateCategory(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Kategori tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui kategori", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data kategori berhasil diperbarui", cat)
}

// DeleteCategory handles DELETE /api/v1/categories/:id
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID kategori wajib diisi", nil)
	}

	err := h.categoryUsecase.DeleteCategory(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Kategori tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menonaktifkan kategori", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Kategori berhasil dinonaktifkan", nil)
}
