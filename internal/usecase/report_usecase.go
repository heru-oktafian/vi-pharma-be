package usecase

import (
	"context"

	"pharma-be/internal/domain"
)

type reportUsecase struct {
	reportRepo domain.ReportRepository
}

func NewReportUsecase(reportRepo domain.ReportRepository) domain.ReportUsecase {
	return &reportUsecase{
		reportRepo: reportRepo,
	}
}

func (u *reportUsecase) GetStaffPerformanceReport(ctx context.Context, filter domain.ReportFilter) (*domain.StaffPerformanceReportResponse, error) {
	return u.reportRepo.GetStaffPerformanceReport(ctx, filter)
}

func (u *reportUsecase) GetSalesSummaryReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesSummaryReportResponse, error) {
	return u.reportRepo.GetSalesSummaryReport(ctx, filter)
}

func (u *reportUsecase) GetCategoryMarginsReport(ctx context.Context, filter domain.ReportFilter) ([]domain.CategoryMarginItem, error) {
	return u.reportRepo.GetCategoryMarginsReport(ctx, filter)
}
