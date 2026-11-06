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
	
	dailyMarginTarget, _ := strconv.ParseFloat(c.Query("daily_margin_target", "0"), 64)
	if dailyMarginTarget == 0 {
		dailyMarginTarget, _ = strconv.ParseFloat(c.Query("target_margin", "0"), 64)
	}
	dailyBonusReward, _ := strconv.ParseFloat(c.Query("daily_bonus_reward", "0"), 64)
	if dailyBonusReward == 0 {
		dailyBonusReward, _ = strconv.ParseFloat(c.Query("bonus_reward", "0"), 64)
	}
	bonusPool, _ := strconv.ParseFloat(c.Query("bonus_pool", "0"), 64)

	filter := domain.ReportFilter{
		BranchID:          branchID,
		StartDate:         startDate,
		EndDate:           endDate,
		DailyMarginTarget: dailyMarginTarget,
		DailyBonusReward:  dailyBonusReward,
		BonusPoolBudget:   bonusPool,
	}

	report, err := h.reportUsecase.GetStaffPerformanceReport(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat laporan kinerja kontribusi petugas", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Laporan kinerja dan kontribusi margin petugas berhasil dimuat", report)
}

// UpdateBranchBonusSetting handles PUT /api/v1/reports/branch-bonus-setting
func (h *ReportHandler) UpdateBranchBonusSetting(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	var req domain.UpdateBranchBonusSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Format data tidak valid", err)
	}

	if req.DailyMarginTarget < 0 || req.DailyBonusReward < 0 {
		return ErrorResponse(c, fiber.StatusBadRequest, "Target margin dan nominal bonus tidak boleh bernilai negatif", nil)
	}

	if err := h.reportUsecase.UpdateBranchBonusSetting(c.Context(), branchID, req); err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui konfigurasi bonus cabang", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Konfigurasi target margin dan bonus harian cabang berhasil diperbarui", req)
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

// GetSIPNAPReport handles GET /api/v1/reports/sipnap
func (h *ReportHandler) GetSIPNAPReport(c *fiber.Ctx) error {
	branchID := getBranchID(c)
	month, _ := strconv.Atoi(c.Query("month", "0"))
	year, _ := strconv.Atoi(c.Query("year", "0"))
	classification := c.Query("classification", "ALL")

	filter := domain.SIPNAPReportFilter{
		BranchID:       branchID,
		Month:          month,
		Year:           year,
		Classification: classification,
	}

	report, err := h.reportUsecase.GetSIPNAPReport(c.Context(), filter)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memuat laporan SIPNAP Kemenkes", err)
	}

	return SuccessResponse(c, fiber.StatusOK, "Laporan resmi SIPNAP Kemenkes berhasil dimuat", report)
}

