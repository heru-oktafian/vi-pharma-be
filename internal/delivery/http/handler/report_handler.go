package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"pharma-be/internal/domain"
)

type ReportHandler struct {
	reportUsecase domain.ReportUsecase
}

func NewReportHandler(reportUsecase domain.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: reportUsecase}
}

// GetStaffPerformance handles GET /api/v1/reports/staff-performance
func (h *ReportHandler) GetStaffPerformance(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")
	bonusPool, _ := strconv.ParseFloat(c.Query("bonus_pool", "0"), 64)
	targetMargin, _ := strconv.ParseFloat(c.Query("target_margin", "0"), 64)

	filter := domain.ReportFilter{
		BranchID:            branchID,
		StartDate:           startDate,
		EndDate:             endDate,
		BonusPoolBudget:     bonusPool,
		TargetMarginPercent: targetMargin,
	}

	report, err := h.reportUsecase.GetStaffPerformanceReport(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat laporan kinerja kontribusi petugas", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Laporan kinerja dan kontribusi margin petugas berhasil dimuat", report)
}

// GetSalesSummary handles GET /api/v1/reports/sales-summary
func (h *ReportHandler) GetSalesSummary(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.ReportFilter{
		BranchID:  branchID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	report, err := h.reportUsecase.GetSalesSummaryReport(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat rekapitulasi penjualan apotek", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Rekapitulasi penjualan cabang berhasil dimuat", report)
}

// GetCategoryMargins handles GET /api/v1/reports/category-margins
func (h *ReportHandler) GetCategoryMargins(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	filter := domain.ReportFilter{
		BranchID:  branchID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	items, err := h.reportUsecase.GetCategoryMarginsReport(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat analisis margin kategori obat", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Analisis profitabilitas kategori berhasil dimuat", items)
}
