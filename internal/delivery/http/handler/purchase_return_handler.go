package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type PurchaseReturnHandler struct {
	returnUsecase domain.PurchaseReturnUsecase
}

func NewPurchaseReturnHandler(returnUsecase domain.PurchaseReturnUsecase) *PurchaseReturnHandler {
	return &PurchaseReturnHandler{returnUsecase: returnUsecase}
}

// CreatePurchaseReturn handles POST /api/v1/procurement/returns
func (h *PurchaseReturnHandler) CreatePurchaseReturn(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.CreatePurchaseReturnRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.SupplierID == "" || req.ReturnDate == "" || req.Reason == "" || len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Supplier, Tanggal Retur, Alasan Utama, dan minimal 1 item obat wajib diisi", nil)
	}

	for _, item := range req.Items {
		if item.MedicineID == "" || item.BatchID == "" || item.Qty <= 0 {
			return ErrorResponse(c, fiber.StatusBadRequest, "Setiap item retur wajib memiliki obat, batch, dan jumlah > 0", nil)
		}
	}

	ret, err := h.returnUsecase.CreatePurchaseReturn(c.Context(), branchID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientReturnStock) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat berkas Retur Pembelian PBF", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Retur Pembelian PBF berhasil dibuat", ret)
}

// ListPurchaseReturns handles GET /api/v1/procurement/returns
func (h *PurchaseReturnHandler) ListPurchaseReturns(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	supplierID := c.Query("supplier_id", "")
	status := domain.PurchaseReturnStatus(c.Query("status", ""))
	reason := c.Query("reason", "")
	compType := domain.CompensationType(c.Query("compensation_type", ""))
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.PurchaseReturnFilter{
		Search:           search,
		SupplierID:       supplierID,
		Status:           status,
		Reason:           reason,
		CompensationType: compType,
		StartDate:        startDate,
		EndDate:          endDate,
		Page:             page,
		Limit:            limit,
	}

	result, err := h.returnUsecase.ListPurchaseReturns(c.Context(), branchID, filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar Retur Pembelian", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data retur pembelian berhasil diambil",
		"data":    result.Data,
		"meta": fiber.Map{
			"current_page": result.Page,
			"page_size":    result.Limit,
			"total_items":  result.TotalItems,
			"total_pages":  result.TotalPages,
		},
	})
}

// GetPurchaseReturnByID handles GET /api/v1/procurement/returns/:id
func (h *PurchaseReturnHandler) GetPurchaseReturnByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID retur tidak valid", nil)
	}

	ret, err := h.returnUsecase.GetPurchaseReturnByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data Retur Pembelian tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail Retur Pembelian", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail Retur Pembelian berhasil diambil", ret)
}

// UpdatePurchaseReturn handles PUT /api/v1/procurement/returns/:id
func (h *PurchaseReturnHandler) UpdatePurchaseReturn(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID retur tidak valid", nil)
	}

	var req domain.UpdatePurchaseReturnRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	ret, err := h.returnUsecase.UpdatePurchaseReturn(c.Context(), branchID, id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data Retur Pembelian tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrPurchaseReturnAlreadyProcessed) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrInsufficientReturnStock) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui Retur Pembelian", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Retur Pembelian berhasil diperbarui", ret)
}

// UpdateReturnStatus handles PATCH /api/v1/procurement/returns/:id/status
func (h *PurchaseReturnHandler) UpdateReturnStatus(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID retur tidak valid", nil)
	}

	var req domain.UpdateReturnStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Status == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Status retur tujuan wajib diisi", nil)
	}

	ret, err := h.returnUsecase.UpdateReturnStatus(c.Context(), branchID, userID, id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data Retur Pembelian tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrPurchaseReturnAlreadyProcessed) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		if errors.Is(err, domain.ErrInsufficientReturnStock) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui status Retur Pembelian", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Status Retur Pembelian berhasil diubah", ret)
}

// DeletePurchaseReturn handles DELETE /api/v1/procurement/returns/:id
func (h *PurchaseReturnHandler) DeletePurchaseReturn(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID retur tidak valid", nil)
	}

	if err := h.returnUsecase.DeletePurchaseReturn(c.Context(), branchID, id); err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data Retur Pembelian tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrPurchaseReturnAlreadyProcessed) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus Retur Pembelian", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Retur Pembelian berhasil dihapus", nil)
}
