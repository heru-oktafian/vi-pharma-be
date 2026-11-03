package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type DoctorHandler struct {
	doctorUsecase domain.DoctorUsecase
}

func NewDoctorHandler(doctorUsecase domain.DoctorUsecase) *DoctorHandler {
	return &DoctorHandler{doctorUsecase: doctorUsecase}
}

// ListDoctors handles GET /api/v1/doctors
func (h *DoctorHandler) ListDoctors(c *fiber.Ctx) error {
	search := c.Query("search", "")
	activeOnly := c.Query("active_only", "false") == "true"

	doctors, err := h.doctorUsecase.GetAllDoctors(c.Context(), search, activeOnly)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data dokter", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data dokter berhasil diambil", doctors)
}

// GetDoctorByID handles GET /api/v1/doctors/:id
func (h *DoctorHandler) GetDoctorByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID dokter wajib diisi", nil)
	}

	doc, err := h.doctorUsecase.GetDoctorByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDoctorNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Dokter tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data dokter", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data dokter ditemukan", doc)
}

// CreateDoctor handles POST /api/v1/doctors
func (h *DoctorHandler) CreateDoctor(c *fiber.Ctx) error {
	var req domain.CreateDoctorRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Name == "" || req.SIP == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nama dan SIP dokter wajib diisi", nil)
	}

	doc, err := h.doctorUsecase.CreateDoctor(c.Context(), &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan data dokter", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Data dokter baru berhasil ditambahkan", doc)
}

// UpdateDoctor handles PUT /api/v1/doctors/:id
func (h *DoctorHandler) UpdateDoctor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID dokter wajib diisi", nil)
	}

	var req domain.UpdateDoctorRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	doc, err := h.doctorUsecase.UpdateDoctor(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrDoctorNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Dokter tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data dokter", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data dokter berhasil diperbarui", doc)
}

// DeleteDoctor handles DELETE /api/v1/doctors/:id
func (h *DoctorHandler) DeleteDoctor(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID dokter wajib diisi", nil)
	}

	err := h.doctorUsecase.DeleteDoctor(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDoctorNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Dokter tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menonaktifkan dokter", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data dokter berhasil dinonaktifkan", nil)
}
