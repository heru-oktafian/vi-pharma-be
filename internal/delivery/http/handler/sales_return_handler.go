package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"pharma-be/internal/domain"
)

type SalesReturnHandler struct {
	returnUsecase domain.SalesReturnUsecase
}

func NewSalesReturnHandler(returnUsecase domain.SalesReturnUsecase) *SalesReturnHandler {
	return &SalesReturnHandler{returnUsecase: returnUsecase}
}

// CreateReturn handles POST /api/v1/pos/returns
func (h *SalesReturnHandler) CreateReturn(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	cashierID := getUserID(c)

	var req domain.CreateSalesReturnRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format payload tidak valid", err)
	}

	if req.SaleID == "" || req.Reason == "" || len(req.Items) == 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Sale ID, alasan retur, dan minimal 1 item obat wajib diisi", nil)
	}

	salesReturn, err := h.returnUsecase.ProcessReturn(c.Context(), branchID, cashierID, req)
	if err != nil {
		if errors.Is(err, domain.ErrSaleNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Transaksi penjualan tidak ditemukan", nil)
		}
		if errors.Is(err, domain.ErrReturnQtyExceedsSale) {
			return ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), err)
	}

	return SuccessResponse(c, fiber.StatusCreated, "Retur penjualan berhasil diproses dan stok obat telah direstorasi", salesReturn)
}

// ListReturns handles GET /api/v1/pos/returns
func (h *SalesReturnHandler) ListReturns(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	refundType := domain.SalesReturnRefundType(c.Query("refund_type", ""))
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.SalesReturnFilter{
		Search:     search,
		RefundType: refundType,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       page,
		Limit:      limit,
	}

	result, err := h.returnUsecase.GetSalesReturns(c.Context(), branchID, filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar retur penjualan", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Daftar retur penjualan kasir berhasil diambil",
		"data":    result.Data,
		"meta": fiber.Map{
			"current_page": result.Page,
			"page_size":    result.Limit,
			"total_items":  result.TotalItems,
			"total_pages":  result.TotalPages,
		},
	})
}

// GetReturnByID handles GET /api/v1/pos/returns/:id
func (h *SalesReturnHandler) GetReturnByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID retur wajib diisi", nil)
	}

	salesReturn, err := h.returnUsecase.GetSalesReturnByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSalesReturnNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Data retur penjualan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail retur penjualan", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Detail retur penjualan berhasil diambil", salesReturn)
}

// GetSaleReturnable handles GET /api/v1/pos/sales/:id/returnable
func (h *SalesReturnHandler) GetSaleReturnable(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	if branchID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "Branch ID wajib ditentukan", nil)
	}
	saleID := c.Params("id")
	if saleID == "" {
		return ErrorResponse(c, fiber.StatusBadRequest, "ID transaksi penjualan wajib diisi", nil)
	}

	info, err := h.returnUsecase.GetSaleReturnableInfo(c.Context(), branchID, saleID)
	if err != nil {
		if errors.Is(err, domain.ErrSaleNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Transaksi penjualan tidak ditemukan", nil)
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengecek sisa retur transaksi", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Informasi sisa retur transaksi berhasil diambil", info)
}
