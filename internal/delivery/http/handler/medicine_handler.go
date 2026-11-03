package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type MedicineHandler struct {
	medicineUsecase domain.MedicineUsecase
}

func NewMedicineHandler(medicineUsecase domain.MedicineUsecase) *MedicineHandler {
	return &MedicineHandler{medicineUsecase: medicineUsecase}
}

// ListMedicines handles GET /api/v1/medicines
func (h *MedicineHandler) ListMedicines(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	classification := domain.DrugClassification(c.Query("classification", ""))
	categoryID := c.Query("category_id", "")
	activeOnly := c.Query("active_only", "false") == "true"

	filter := domain.MedicineFilter{
		Search:         search,
		Classification: classification,
		CategoryID:     categoryID,
		ActiveOnly:     activeOnly,
		Page:           page,
		Limit:          limit,
	}

	result, err := h.medicineUsecase.GetAllMedicines(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data katalog obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Katalog obat berhasil diambil", result)
}

// GetMedicineByID handles GET /api/v1/medicines/:id
func (h *MedicineHandler) GetMedicineByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID obat wajib diisi", nil)
	}

	med, err := h.medicineUsecase.GetMedicineByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMedicineNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data obat tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data obat ditemukan", med)
}

// GetMedicineByBarcode handles GET /api/v1/medicines/barcode/:barcode
func (h *MedicineHandler) GetMedicineByBarcode(c *fiber.Ctx) error {
	barcode := c.Params("barcode")
	if barcode == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Barcode obat wajib diisi", nil)
	}

	med, err := h.medicineUsecase.GetMedicineByBarcode(c.Context(), barcode)
	if err != nil {
		if errors.Is(err, domain.ErrMedicineNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Obat dengan barcode tersebut tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mencari obat berdasarkan barcode", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Obat berhasil ditemukan via barcode", med)
}

// CreateMedicine handles POST /api/v1/medicines
func (h *MedicineHandler) CreateMedicine(c *fiber.Ctx) error {
	var req domain.CreateMedicineRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.ItemCode == "" || req.Name == "" || req.CategoryID == "" || req.BaseUnitID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Kode item, nama obat, kategori, dan satuan dasar wajib diisi", nil)
	}

	med, err := h.medicineUsecase.CreateMedicine(c.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrItemCodeAlreadyExists) {
			return ErrorResponse(c, fiber.StatusConflict, "Kode item obat sudah digunakan", nil)
		}
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Kategori obat tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrUnitNotFound) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Satuan dasar obat tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan obat baru ke katalog", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Obat baru berhasil ditambahkan ke katalog master", med)
}

// UpdateMedicine handles PUT /api/v1/medicines/:id
func (h *MedicineHandler) UpdateMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID obat wajib diisi", nil)
	}

	var req domain.UpdateMedicineRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	med, err := h.medicineUsecase.UpdateMedicine(c.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrMedicineNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data obat tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrItemCodeAlreadyExists) {
			return ErrorResponse(c, fiber.StatusConflict, "Kode item obat sudah digunakan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data obat berhasil diperbarui", med)
}

// DeleteMedicine handles DELETE /api/v1/medicines/:id
func (h *MedicineHandler) DeleteMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID obat wajib diisi", nil)
	}

	err := h.medicineUsecase.DeleteMedicine(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrMedicineNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data obat tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menonaktifkan obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Obat berhasil dinonaktifkan dari katalog", nil)
}
