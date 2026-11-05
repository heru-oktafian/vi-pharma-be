package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"pharma-be/internal/domain"
)

type PrescriptionHandler struct {
	prescriptionUsecase domain.PrescriptionUsecase
}

func NewPrescriptionHandler(prescriptionUsecase domain.PrescriptionUsecase) *PrescriptionHandler {
	return &PrescriptionHandler{prescriptionUsecase: prescriptionUsecase}
}

// CreatePrescription handles POST /api/v1/prescriptions
func (h *PrescriptionHandler) CreatePrescription(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	var req domain.CreatePrescriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload resep tidak valid", err)
	}

	if req.PatientName == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nama pasien wajib diisi", nil)
	}
	if req.DoctorName == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nama dokter wajib diisi", nil)
	}
	if len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Minimal satu item obat/racikan wajib diisi", nil)
	}

	prescription, err := h.prescriptionUsecase.CreatePrescription(c.Context(), branchID, &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Resep dokter berhasil didaftarkan ke apotek", prescription)
}

// ListPrescriptions handles GET /api/v1/prescriptions
func (h *PrescriptionHandler) ListPrescriptions(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	status := c.Query("status", "")
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.PrescriptionFilter{
		BranchID:  branchID,
		Status:    status,
		Search:    search,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	}

	if isPaidStr := c.Query("is_paid"); isPaidStr != "" {
		isPaid := isPaidStr == "true" || isPaidStr == "1"
		filter.IsPaid = &isPaid
	}

	res, err := h.prescriptionUsecase.GetPrescriptions(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat daftar resep", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Daftar resep berhasil dimuat", res)
}

// GetPrescriptionDetail handles GET /api/v1/prescriptions/:id
func (h *PrescriptionHandler) GetPrescriptionDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID resep wajib disertakan", nil)
	}

	prescription, err := h.prescriptionUsecase.GetPrescriptionByID(c.Context(), id)
	if err != nil {
		return ErrorResponse(c, fiber.StatusNotFound, "Resep tidak ditemukan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail resep berhasil dimuat", prescription)
}

// UpdateStatus handles PUT /api/v1/prescriptions/:id/status
func (h *PrescriptionHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID resep wajib disertakan", nil)
	}

	var req domain.UpdatePrescriptionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	pharmacistID := getUserID(c)
	if req.PharmacistID == nil && pharmacistID != "" {
		req.PharmacistID = &pharmacistID
	}

	prescription, err := h.prescriptionUsecase.UpdateStatus(c.Context(), id, &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Status alur resep berhasil diperbarui", prescription)
}

// DispensePrescription handles POST /api/v1/prescriptions/:id/dispense
func (h *PrescriptionHandler) DispensePrescription(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID resep wajib disertakan", nil)
	}

	userID := getUserID(c)
	prescription, err := h.prescriptionUsecase.DispensePrescription(c.Context(), id, userID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Resep selesai diracik dan stok fisik rak berhasil dipotong", prescription)
}
