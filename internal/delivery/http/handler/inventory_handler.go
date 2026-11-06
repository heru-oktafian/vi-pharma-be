package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type InventoryHandler struct {
	invUsecase domain.InventoryUsecase
}

func NewInventoryHandler(invUsecase domain.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{
		invUsecase: invUsecase,
	}
}

// Helper to get branch ID from header, query, or locals
func getBranchID(c *fiber.Ctx) string {
	if headerBranch := c.Get("X-Branch-Id"); headerBranch != "" {
		return headerBranch
	}
	if queryBranch := c.Query("branch_id", ""); queryBranch != "" {
		return queryBranch
	}
	if branchID, ok := c.Locals("branchID").(string); ok && branchID != "" {
		return branchID
	}
	return "CBG000000000001"
}

// Helper to get user ID
func getUserID(c *fiber.Ctx) string {
	if userID, ok := c.Locals("userID").(string); ok && userID != "" {
		return userID
	}
	return "USR000000000001" // Safe fallback to Superadmin / Default User
}

// -------------------------------------------------------------
// Purchase Invoices (PBF)
// -------------------------------------------------------------

// ReceivePurchaseInvoice handles POST /api/v1/inventory/invoices
func (h *InventoryHandler) ReceivePurchaseInvoice(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.CreatePurchaseInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.SupplierID == "" || req.InvoiceNumber == "" || req.InvoiceDate == "" || len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Supplier, nomor faktur, tanggal faktur, dan minimal 1 item obat wajib diisi", nil)
	}

	for _, item := range req.Items {
		if item.MedicineID == "" || item.UnitID == "" || item.BatchNumber == "" || item.ExpiredDate == "" || item.QtyPurchased <= 0 {
			return ErrorResponse(c, fiber.StatusBadRequest, "Setiap item faktur wajib memiliki ID obat, satuan, batch, expired date, dan jumlah beli > 0", nil)
		}
	}

	invoice, err := h.invUsecase.ReceivePurchaseInvoice(c.Context(), branchID, userID, &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mencatat faktur pembelian PBF", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Faktur pembelian PBF berhasil diterima dan stok batch dicatat", invoice)
}

// ListPurchaseInvoices handles GET /api/v1/inventory/invoices
func (h *InventoryHandler) ListPurchaseInvoices(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	invoices, total, err := h.invUsecase.ListPurchaseInvoices(c.Context(), branchID, page, limit)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar faktur pembelian", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	return SuccessResponse(c, fiber.StatusOK, "Daftar faktur pembelian berhasil diambil", PaginatedData{
		Items: invoices,
		Meta: PaginationMeta{
			CurrentPage: page,
			PageSize:    limit,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}

// GetPurchaseInvoiceDetail handles GET /api/v1/inventory/invoices/:id
func (h *InventoryHandler) GetPurchaseInvoiceDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID faktur wajib diisi", nil)
	}

	invoice, err := h.invUsecase.GetPurchaseInvoiceDetail(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Faktur pembelian tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail faktur", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail faktur pembelian berhasil diambil", invoice)
}

// -------------------------------------------------------------
// Accounts Payable & Payments (Hutang Pembelian & Jatuh Tempo)
// -------------------------------------------------------------

// GetAccountsPayableSummary handles GET /api/v1/inventory/accounts-payable/summary
func (h *InventoryHandler) GetAccountsPayableSummary(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	summary, err := h.invUsecase.GetAccountsPayableSummary(c.Context(), branchID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil ringkasan hutang pembelian", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Ringkasan hutang pembelian PBF berhasil diambil", summary)
}

// ListAccountsPayableInvoices handles GET /api/v1/inventory/accounts-payable/invoices
func (h *InventoryHandler) ListAccountsPayableInvoices(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	supplierID := c.Query("supplier_id", "")
	status := domain.PaymentStatus(c.Query("status", ""))
	isOverdue := c.Query("is_overdue") == "true"
	dueSoon := c.Query("due_soon") == "true"
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	sortBy := c.Query("sort_by", "due_date_asc")

	filter := domain.AccountsPayableFilter{
		Search:     search,
		SupplierID: supplierID,
		Status:     status,
		IsOverdue:  isOverdue,
		DueSoon:    dueSoon,
		StartDate:  startDate,
		EndDate:    endDate,
		SortBy:     sortBy,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.invUsecase.ListAccountsPayableInvoices(c.Context(), branchID, filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar hutang faktur PBF", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Daftar hutang faktur PBF berhasil diambil",
		"data":    result.Data,
		"meta": fiber.Map{
			"current_page": result.Page,
			"page_size":    result.Limit,
			"total_items":  result.TotalItems,
			"total_pages":  result.TotalPages,
		},
	})
}

// RecordInvoicePayment handles POST /api/v1/inventory/invoices/:id/payments
func (h *InventoryHandler) RecordInvoicePayment(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)
	invoiceID := c.Params("id")
	if invoiceID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID faktur wajib diisi", nil)
	}

	var req domain.CreateInvoicePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Amount <= 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Nominal pembayaran harus lebih besar dari 0", nil)
	}

	payment, err := h.invUsecase.RecordInvoicePayment(c.Context(), branchID, userID, invoiceID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Faktur pembelian tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrInvoiceAlreadyPaid) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Faktur pembelian ini sudah berstatus lunas", nil)
		}
		if errors.Is(err, domain.ErrPaymentExceedsBalance) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Nominal pembayaran melebihi sisa hutang faktur", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mencatat pelunasan hutang faktur", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Pembayaran hutang faktur berhasil dicatat", payment)
}

// GetInvoicePayments handles GET /api/v1/inventory/invoices/:id/payments
func (h *InventoryHandler) GetInvoicePayments(c *fiber.Ctx) error {
	invoiceID := c.Params("id")
	if invoiceID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID faktur wajib diisi", nil)
	}

	payments, err := h.invUsecase.GetInvoicePayments(c.Context(), invoiceID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil riwayat pembayaran faktur", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Riwayat pembayaran faktur berhasil diambil", payments)
}

// -------------------------------------------------------------
// Stock Transfers (Gudang <-> Rak)
// -------------------------------------------------------------

// RefillShelf handles POST /api/v1/inventory/refill-shelf
func (h *InventoryHandler) RefillShelf(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.RefillShelfRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.BatchID == "" || req.Qty <= 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Batch ID dan jumlah transfer (qty > 0) wajib diisi", nil)
	}

	batch, err := h.invUsecase.RefillShelf(c.Context(), branchID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientWarehouseStock) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Stok gudang tidak mencukupi untuk dipindahkan ke rak", err)
		}
		if errors.Is(err, domain.ErrBatchNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Batch obat tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memindahkan stok ke rak", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Stok berhasil dipindahkan dari Gudang ke Rak Penjualan", batch)
}

// ReturnToWarehouse handles POST /api/v1/inventory/return-warehouse
func (h *InventoryHandler) ReturnToWarehouse(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.ReturnToWarehouseRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.BatchID == "" || req.Qty <= 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Batch ID dan jumlah pengembalian (qty > 0) wajib diisi", nil)
	}

	batch, err := h.invUsecase.ReturnToWarehouse(c.Context(), branchID, userID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientShelfStock) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Stok di rak tidak mencukupi untuk dikembalikan ke gudang", err)
		}
		if errors.Is(err, domain.ErrBatchNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Batch obat tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengembalikan stok ke gudang", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Stok berhasil dikembalikan dari Rak ke Gudang", batch)
}

// CreateInitialStockBatch handles POST /api/v1/inventory/batches/initial-stock
func (h *InventoryHandler) CreateInitialStockBatch(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.CreateInitialBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	batch, err := h.invUsecase.CreateInitialStockBatch(c.Context(), branchID, userID, &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Saldo awal batch fisik berhasil dicatat", batch)
}

// -------------------------------------------------------------
// Stock Summaries & Digital Stock Card
// -------------------------------------------------------------

// ListStockSummaries handles GET /api/v1/inventory/stock-summary
func (h *InventoryHandler) ListStockSummaries(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	search := c.Query("search", "")
	summaries, err := h.invUsecase.ListStockSummaries(c.Context(), branchID, search)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil rekapitulasi stok 3-level", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Rekapitulasi stok 3-level berhasil diambil", summaries)
}

// GetStockSummary handles GET /api/v1/inventory/stock-summary/:medicineId
func (h *InventoryHandler) GetStockSummary(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	medID := c.Params("medicineId")
	if medID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID obat wajib diisi", nil)
	}

	summary, err := h.invUsecase.GetStockSummary(c.Context(), branchID, medID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil ringkasan stok obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Ringkasan stok obat berhasil diambil", summary)
}

// GetStockCard handles GET /api/v1/inventory/stock-card/:medicineId
func (h *InventoryHandler) GetStockCard(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	medID := c.Params("medicineId")
	if medID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID obat wajib diisi", nil)
	}

	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	mutations, err := h.invUsecase.GetStockCard(c.Context(), branchID, medID, startDate, endDate)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data kartu stok digital", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Kartu stok digital berhasil diambil", mutations)
}

// -------------------------------------------------------------
// FEFO & Expiry Watcher
// -------------------------------------------------------------

// GetFefoAlerts handles GET /api/v1/inventory/fefo-alerts
func (h *InventoryHandler) GetFefoAlerts(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	alerts, err := h.invUsecase.GetFefoAlerts(c.Context(), branchID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data monitoring kadaluarsa FEFO", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Data monitoring kadaluarsa FEFO berhasil diambil", alerts)
}

// -------------------------------------------------------------
// Rolling Cycle Count (Stok Opname Bergulir)
// -------------------------------------------------------------

// CreateOpnameSession handles POST /api/v1/inventory/opnames
func (h *InventoryHandler) CreateOpnameSession(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)

	var req domain.CreateOpnameSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.Location == "" || req.RackName == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Lokasi (GUDANG/RAK) dan Nama Rak/Zona wajib diisi", nil)
	}

	opname, err := h.invUsecase.CreateOpnameSession(c.Context(), branchID, userID, &req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat sesi opname fisik", err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Sesi stok opname bergulir berhasil dibuat (Status Draft)", opname)
}

// SubmitOpnameSession handles POST /api/v1/inventory/opnames/:id/submit
func (h *InventoryHandler) SubmitOpnameSession(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	userID := getUserID(c)
	opnameID := c.Params("id")
	if opnameID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID sesi opname wajib diisi", nil)
	}

	var req domain.SubmitOpnameRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Minimal satu item hasil hitung opname wajib disertakan", nil)
	}

	opname, err := h.invUsecase.SubmitOpnameSession(c.Context(), branchID, userID, opnameID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrOpnameAlreadyCompleted) {
			return ErrorResponse(c, fiber.StatusBadRequest, "Sesi opname ini sudah selesai dieksekusi sebelumnya", nil)
		}
		if errors.Is(err, domain.ErrOpnameNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Sesi opname tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memproses penyesuaian stok opname", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Hasil stok opname berhasil diterapkan dan saldo stok disesuaikan", opname)
}

// ListOpnames handles GET /api/v1/inventory/opnames
func (h *InventoryHandler) ListOpnames(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	opnames, total, err := h.invUsecase.ListOpnames(c.Context(), branchID, page, limit)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar sesi opname", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	return SuccessResponse(c, fiber.StatusOK, "Daftar sesi opname berhasil diambil", PaginatedData{
		Items: opnames,
		Meta: PaginationMeta{
			CurrentPage: page,
			PageSize:    limit,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}

// GetOpnameDetail handles GET /api/v1/inventory/opnames/:id
func (h *InventoryHandler) GetOpnameDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID sesi opname wajib diisi", nil)
	}

	opname, err := h.invUsecase.GetOpnameDetail(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOpnameNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Sesi opname tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail sesi opname", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail sesi opname berhasil diambil", opname)
}
