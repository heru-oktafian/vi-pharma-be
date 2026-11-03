package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type SupplierHandler struct {
	supplierUsecase domain.SupplierUsecase
}

func NewSupplierHandler(supplierUsecase domain.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{supplierUsecase: supplierUsecase}
}

// ListSuppliers handles GET /api/v1/suppliers
func (h *SupplierHandler) ListSuppliers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	activeOnly := c.Query("active_only", "false") == "true"

	suppliers, err := h.supplierUsecase.GetAllSuppliers(c.Context(), search, activeOnly)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data supplier/PBF", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data supplier/PBF berhasil diambil", suppliers)
}

// GetSupplierByID handles GET /api/v1/suppliers/:id
func (h *SupplierHandler) GetSupplierByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID supplier wajib diisi", nil)
	}

	sup, err := h.supplierUsecase.GetSupplierByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Supplier/PBF tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data supplier/PBF", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data supplier/PBF ditemukan", sup)
}

// CreateSupplier handles POST /api/v1/suppliers
func (h *SupplierHandler) CreateSupplier(c *fiber.Ctx) error {
	var req domain.CreateSupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Code == "" || req.Name == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Kode dan nama supplier/PBF wajib diisi", nil)
	}

	sup, err := h.supplierUsecase.CreateSupplier(c.Context(), &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan supplier/PBF baru", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Supplier/PBF baru berhasil ditambahkan", sup)
}

// UpdateSupplier handles PUT /api/v1/suppliers/:id
func (h *SupplierHandler) UpdateSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID supplier wajib diisi", nil)
	}

	var req domain.UpdateSupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	sup, err := h.supplierUsecase.UpdateSupplier(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Supplier/PBF tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data supplier/PBF", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data supplier/PBF berhasil diperbarui", sup)
}

// DeleteSupplier handles DELETE /api/v1/suppliers/:id
func (h *SupplierHandler) DeleteSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID supplier wajib diisi", nil)
	}

	err := h.supplierUsecase.DeleteSupplier(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Supplier/PBF tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menonaktifkan supplier/PBF", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Supplier/PBF berhasil dinonaktifkan", nil)
}
