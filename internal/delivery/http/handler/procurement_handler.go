package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type ProcurementHandler struct {
	procUsecase domain.ProcurementUsecase
}

func NewProcurementHandler(procUsecase domain.ProcurementUsecase) *ProcurementHandler {
	return &ProcurementHandler{procUsecase: procUsecase}
}

// CreatePurchaseOrder handles POST /api/v1/procurement/orders
func (h *ProcurementHandler) CreatePurchaseOrder(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.CreatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.SupplierID == "" || req.OrderDate == "" || req.SPType == "" || req.PharmacistName == "" || req.PharmacistSIPA == "" || len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Supplier, Tanggal Pesan, Jenis SP, Nama Apoteker, No SIPA, dan minimal 1 item obat wajib diisi", nil)
	}

	for _, item := range req.Items {
		if item.MedicineID == "" || item.UnitID == "" || item.QtyOrdered <= 0 {
			return ErrorResponse(c, fiber.StatusBadRequest, "Setiap item SP wajib memiliki ID obat, satuan, dan jumlah pesan > 0", nil)
		}
	}

	po, err := h.procUsecase.CreatePurchaseOrder(c.Context(), branchID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrNarcoticSingleItemOnly) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Surat Pesanan Narkotika resmi hanya boleh berisi 1 item obat per surat pesanan (Regulasi BPOM)", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat Surat Pesanan / Purchase Order", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Surat Pesanan (SP) berhasil dibuat", po)
}

// ListPurchaseOrders handles GET /api/v1/procurement/orders
func (h *ProcurementHandler) ListPurchaseOrders(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	supplierID := c.Query("supplier_id", "")
	spType := domain.SPType(c.Query("sp_type", ""))
	status := domain.POStatus(c.Query("status", ""))
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.POFilter{
		Search:     search,
		SupplierID: supplierID,
		SPType:     spType,
		Status:     status,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.procUsecase.ListPurchaseOrders(c.Context(), branchID, filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar Surat Pesanan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Daftar Surat Pesanan berhasil diambil", result)
}

// GetPurchaseOrderByID handles GET /api/v1/procurement/orders/:id
func (h *ProcurementHandler) GetPurchaseOrderByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID Purchase Order wajib diisi", nil)
	}

	po, err := h.procUsecase.GetPurchaseOrderByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseOrderNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Surat Pesanan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail Surat Pesanan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail Surat Pesanan berhasil diambil", po)
}

// UpdatePurchaseOrder handles PUT /api/v1/procurement/orders/:id
func (h *ProcurementHandler) UpdatePurchaseOrder(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID Purchase Order wajib diisi", nil)
	}

	var req domain.UpdatePurchaseOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	po, err := h.procUsecase.UpdatePurchaseOrder(c.Context(), branchID, id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrPOAlreadyProcessed) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Surat Pesanan yang sudah diproses atau selesai tidak dapat diubah", nil)
		}
		if errors.Is(err, domain.ErrNarcoticSingleItemOnly) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Surat Pesanan Narkotika resmi hanya boleh berisi 1 item obat per surat pesanan", nil)
		}
		if errors.Is(err, domain.ErrPurchaseOrderNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Surat Pesanan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui Surat Pesanan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Surat Pesanan berhasil diperbarui", po)
}

// UpdatePOStatus handles PATCH /api/v1/procurement/orders/:id/status
func (h *ProcurementHandler) UpdatePOStatus(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID Purchase Order wajib diisi", nil)
	}

	var req domain.UpdatePOStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Status == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Status baru wajib diisi", nil)
	}

	po, err := h.procUsecase.UpdatePOStatus(c.Context(), branchID, id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseOrderNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Surat Pesanan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengubah status Surat Pesanan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Status Surat Pesanan berhasil diubah", po)
}

// DeletePurchaseOrder handles DELETE /api/v1/procurement/orders/:id
func (h *ProcurementHandler) DeletePurchaseOrder(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID Purchase Order wajib diisi", nil)
	}

	if err := h.procUsecase.DeletePurchaseOrder(c.Context(), branchID, id); err != nil {
		if errors.Is(err, domain.ErrPOAlreadyProcessed) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Hanya Surat Pesanan berstatus DRAFT yang dapat dihapus", nil)
		}
		if errors.Is(err, domain.ErrPurchaseOrderNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Surat Pesanan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus Surat Pesanan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Surat Pesanan berhasil dihapus", nil)
}
