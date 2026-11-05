package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"pharma-be/internal/domain"
)

type POSHandler struct {
	posUsecase domain.POSUsecase
}

func NewPOSHandler(posUsecase domain.POSUsecase) *POSHandler {
	return &POSHandler{posUsecase: posUsecase}
}

// CreateTransaction handles POST /api/v1/pos/transactions
func (h *POSHandler) CreateTransaction(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	cashierID := getUserID(c)

	var req domain.CreateSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	sale, err := h.posUsecase.Checkout(c.Context(), branchID, cashierID, req)
	if err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Transaksi penjualan berhasil diproses", sale)
}

// ListTransactions handles GET /api/v1/pos/transactions
func (h *POSHandler) ListTransactions(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	paymentMethod := c.Query("payment_method", "")
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.SaleFilter{
		Search:        search,
		PaymentMethod: paymentMethod,
		StartDate:     startDate,
		EndDate:       endDate,
		Page:          page,
		Limit:         limit,
	}

	result, err := h.posUsecase.GetSales(c.Context(), branchID, filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar transaksi penjualan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Daftar transaksi berhasil diambil", result)
}

// GetTransactionByID handles GET /api/v1/pos/transactions/:id
func (h *POSHandler) GetTransactionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	sale, err := h.posUsecase.GetSaleByID(c.Context(), id)
	if err != nil {
		return ErrorResponse(c, fiber.StatusNotFound, "Transaksi tidak ditemukan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail transaksi berhasil diambil", sale)
}

// GetTodaySummary handles GET /api/v1/pos/transactions/summary-today
func (h *POSHandler) GetTodaySummary(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	summary, err := h.posUsecase.GetTodaySummary(c.Context(), branchID)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil ringkasan kasir hari ini", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Ringkasan kasir hari ini berhasil diambil", summary)
}
