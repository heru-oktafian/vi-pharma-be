package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type UnitHandler struct {
	unitUsecase domain.UnitUsecase
}

func NewUnitHandler(unitUsecase domain.UnitUsecase) *UnitHandler {
	return &UnitHandler{unitUsecase: unitUsecase}
}

// ListUnits handles GET /api/v1/units
func (h *UnitHandler) ListUnits(c *fiber.Ctx) error {
	search := c.Query("search", "")
	activeOnly := c.Query("active_only", "false") == "true"

	units, err := h.unitUsecase.GetAllUnits(c.Context(), search, activeOnly)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data satuan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data satuan berhasil diambil", units)
}

// GetUnitByID handles GET /api/v1/units/:id
func (h *UnitHandler) GetUnitByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID satuan wajib diisi", nil)
	}

	unit, err := h.unitUsecase.GetUnitByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Satuan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data satuan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data satuan ditemukan", unit)
}

// CreateUnit handles POST /api/v1/units
func (h *UnitHandler) CreateUnit(c *fiber.Ctx) error {
	var req domain.CreateUnitRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Name == "" || req.Symbol == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nama dan simbol satuan wajib diisi", nil)
	}

	unit, err := h.unitUsecase.CreateUnit(c.Context(), &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat satuan baru", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Satuan baru berhasil ditambahkan", unit)
}

// UpdateUnit handles PUT /api/v1/units/:id
func (h *UnitHandler) UpdateUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID satuan wajib diisi", nil)
	}

	var req domain.UpdateUnitRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	unit, err := h.unitUsecase.UpdateUnit(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Satuan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data satuan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data satuan berhasil diperbarui", unit)
}

// DeleteUnit handles DELETE /api/v1/units/:id
func (h *UnitHandler) DeleteUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID satuan wajib diisi", nil)
	}

	err := h.unitUsecase.DeleteUnit(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUnitNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Satuan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menonaktifkan satuan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Satuan berhasil dinonaktifkan", nil)
}
