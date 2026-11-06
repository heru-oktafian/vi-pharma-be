package domain

import (
	"context"
)

// DailyStaffAchievement represents performance metrics for a staff member on a specific day
type DailyStaffAchievement struct {
	Date           string  `json:"date"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalCOGS      float64 `json:"total_cogs"`
	GrossMargin    float64 `json:"gross_margin"`
	TargetMargin   float64 `json:"target_margin"`
	TargetAchieved bool    `json:"target_achieved"`
	BonusEarned    float64 `json:"bonus_earned"`
	Transactions   int64   `json:"transactions"`
}

// StaffPerformanceItem represents individual contributor metrics
type StaffPerformanceItem struct {
	UserID                     string                  `json:"user_id"`
	UserName                   string                  `json:"user_name"`
	UserRole                   string                  `json:"user_role"`
	TotalTransactions          int64                   `json:"total_transactions"`
	TotalRevenue               float64                 `json:"total_revenue"`
	RevenueContributionPercent float64                 `json:"revenue_contribution_percent"`
	TotalCOGS                  float64                 `json:"total_cogs"`                  // Total Modal Beli (HPP)
	GrossMargin                float64                 `json:"gross_margin"`                // Total Keuntungan Kotor (Rp)
	MarginContributionPercent  float64                 `json:"margin_contribution_percent"` // % terhadap Total Laba Cabang
	MarkupPercent              float64                 `json:"markup_percent"`              // % Keuntungan terhadap Modal
	TuslahAmount               float64                 `json:"tuslah_amount"`               // Jasa Racik Apoteker
	ActiveDaysCount            int                     `json:"active_days_count"`           // Jumlah hari aktif bertransaksi
	DaysAchieved               int                     `json:"days_achieved"`               // Jumlah hari target margin harian tercapai
	DailyTargetMargin          float64                 `json:"daily_target_margin"`         // Nilai target margin harian (Rp)
	DailyBonusReward           float64                 `json:"daily_bonus_reward"`          // Bonus flat per hari jika target tercapai (Rp)
	TotalBonus                 float64                 `json:"total_bonus"`                 // DaysAchieved * DailyBonusReward (Rp)
	EstimatedBonus             float64                 `json:"estimated_bonus"`            // Backward compatibility alias
	DailyBreakdown             []DailyStaffAchievement `json:"daily_breakdown"`             // Rincian performa per hari
}

// StaffPerformanceReportResponse represents the full report structure
type StaffPerformanceReportResponse struct {
	BranchID            string                 `json:"branch_id"`
	BranchName          string                 `json:"branch_name"`
	PeriodStart         string                 `json:"period_start"`
	PeriodEnd           string                 `json:"period_end"`
	TotalBranchRevenue  float64                `json:"total_branch_revenue"`
	TotalBranchCOGS     float64                `json:"total_branch_cogs"`
	TotalBranchMargin   float64                `json:"total_branch_margin"`
	AverageBranchMarkup float64                `json:"average_branch_markup"` // (Total Margin / Total COGS) * 100
	DailyMarginTarget   float64                `json:"daily_margin_target"`   // Target margin harian per karyawan (Rp)
	DailyBonusReward    float64                `json:"daily_bonus_reward"`    // Bonus flat harian jika target tercapai (Rp)
	TotalBonusPaid      float64                `json:"total_bonus_paid"`      // Akumulasi seluruh bonus yang tercapai di cabang
	BonusPoolBudget     float64                `json:"bonus_pool_budget"`     // Backward compatibility
	TargetMarginPercent float64                `json:"target_margin_percent"` // Backward compatibility
	StaffContributions  []StaffPerformanceItem `json:"staff_contributions"`
}

// UpdateBranchBonusSettingRequest DTO to configure branch daily margin target and bonus
type UpdateBranchBonusSettingRequest struct {
	DailyMarginTarget float64 `json:"daily_margin_target"`
	DailyBonusReward  float64 `json:"daily_bonus_reward"`
}

// PaymentMethodSummary represents revenue breakdown by payment method
type PaymentMethodSummary struct {
	PaymentMethod string  `json:"payment_method"`
	Count         int64   `json:"count"`
	TotalAmount   float64 `json:"total_amount"`
	Percentage    float64 `json:"percentage"`
}

// CategoryMarginItem represents margin breakdown by drug category
type CategoryMarginItem struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	TotalQtySold  float64 `json:"total_qty_sold"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalCOGS     float64 `json:"total_cogs"`
	GrossMargin   float64 `json:"gross_margin"`
	MarginPercent float64 `json:"margin_percent"`
	MarkupPercent float64 `json:"markup_percent"`
}

// DailySalesTrend represents time series sales trend
type DailySalesTrend struct {
	Date         string  `json:"date"`
	Transactions int64   `json:"transactions"`
	Revenue      float64 `json:"revenue"`
	COGS         float64 `json:"cogs"`
	Margin       float64 `json:"margin"`
}

// SalesSummaryReportResponse represents comprehensive branch sales report
type SalesSummaryReportResponse struct {
	BranchID             string                 `json:"branch_id"`
	PeriodStart          string                 `json:"period_start"`
	PeriodEnd            string                 `json:"period_end"`
	TotalTransactions    int64                  `json:"total_transactions"`
	TotalRevenue         float64                `json:"total_revenue"`
	TotalCOGS            float64                `json:"total_cogs"`
	TotalMargin          float64                `json:"total_margin"`
	MarginPercent        float64                `json:"margin_percent"`
	PaymentMethodSummary []PaymentMethodSummary `json:"payment_methods"`
	DailyTrends          []DailySalesTrend      `json:"daily_trends"`
}

type ReportFilter struct {
	BranchID            string  `json:"branch_id"`
	StartDate           string  `json:"start_date"`
	EndDate             string  `json:"end_date"`
	DailyMarginTarget   float64 `json:"daily_margin_target"`
	DailyBonusReward    float64 `json:"daily_bonus_reward"`
	BonusPoolBudget     float64 `json:"bonus_pool_budget"`
	TargetMarginPercent float64 `json:"target_margin_percent"`
}

// SIPNAPInDetail represents incoming batch deliveries from PBF for SIPNAP Kemenkes
type SIPNAPInDetail struct {
	InvoiceNumber string  `json:"invoice_number"`
	InvoiceDate   string  `json:"invoice_date"`
	SupplierName  string  `json:"supplier_name"`
	BatchNumber   string  `json:"batch_number"`
	Qty           float64 `json:"qty"`
}

// SIPNAPOutDetail represents outgoing doctor prescription dispensations for SIPNAP Kemenkes
type SIPNAPOutDetail struct {
	TransactionNumber string  `json:"transaction_number"` // No. Resep / No. Nota Kasir
	TransactionDate   string  `json:"transaction_date"`
	DoctorName        string  `json:"doctor_name"`
	DoctorSIP         string  `json:"doctor_sip"`
	PatientName       string  `json:"patient_name"`
	PatientAddress    string  `json:"patient_address"`
	Qty               float64 `json:"qty"`
}

// SIPNAPReportItem represents each controlled drug item in the monthly SIPNAP report
type SIPNAPReportItem struct {
	MedicineID      string             `json:"medicine_id"`
	MedicineName    string             `json:"medicine_name"`
	Classification  DrugClassification `json:"classification"` // narkotika, psikotropika, prekursor, oot
	DosageForm      string             `json:"dosage_form"`    // Tablet, Sirup, Injeksi, Kapsul
	UnitName        string             `json:"unit_name"`      // Tablet, Botol, Ampul
	InitialStock    float64            `json:"initial_stock"`  // Saldo Awal per tgl 1 bulan tsb
	TotalIn         float64            `json:"total_in"`       // Total Pemasukan dari Faktur PBF
	TotalOut        float64            `json:"total_out"`      // Total Pengeluaran (Resep Dokter / Pasien)
	Adjustment      float64            `json:"adjustment"`     // Penyesuaian / Opname / Pemusnahan
	FinalStock      float64            `json:"final_stock"`    // Saldo Akhir (Stok Akhir Bulan)
	InTransactions  []SIPNAPInDetail   `json:"in_transactions"`
	OutTransactions []SIPNAPOutDetail  `json:"out_transactions"`
}

// SIPNAPReportResponse represents the full monthly official SIPNAP report
type SIPNAPReportResponse struct {
	BranchID             string             `json:"branch_id"`
	BranchName           string             `json:"branch_name"`
	BranchAddress        string             `json:"branch_address"`
	SIANumber            string             `json:"sia_number"`
	SIPANumber           string             `json:"sipa_number"`
	PharmacistName       string             `json:"pharmacist_name"`
	PeriodMonth          int                `json:"period_month"` // 1 - 12
	PeriodYear           int                `json:"period_year"`  // e.g. 2026
	ClassificationFilter string             `json:"classification_filter"` // ALL, narkotika, psikotropika, prekursor
	TotalMedicines       int                `json:"total_medicines"`
	Items                []SIPNAPReportItem `json:"items"`
}

type SIPNAPReportFilter struct {
	BranchID       string `json:"branch_id"`
	Month          int    `json:"month"` // 1 - 12
	Year           int    `json:"year"`  // e.g. 2026
	Classification string `json:"classification"` // ALL, narkotika, psikotropika, prekursor, oot
}

type ReportRepository interface {
	GetStaffPerformanceReport(ctx context.Context, filter ReportFilter) (*StaffPerformanceReportResponse, error)
	GetSalesSummaryReport(ctx context.Context, filter ReportFilter) (*SalesSummaryReportResponse, error)
	GetCategoryMarginsReport(ctx context.Context, filter ReportFilter) ([]CategoryMarginItem, error)
	GetSIPNAPReport(ctx context.Context, filter SIPNAPReportFilter) (*SIPNAPReportResponse, error)
	UpdateBranchBonusSetting(ctx context.Context, branchID string, dailyMarginTarget, dailyBonusReward float64) error
}

type ReportUsecase interface {
	GetStaffPerformanceReport(ctx context.Context, filter ReportFilter) (*StaffPerformanceReportResponse, error)
	GetSalesSummaryReport(ctx context.Context, filter ReportFilter) (*SalesSummaryReportResponse, error)
	GetCategoryMarginsReport(ctx context.Context, filter ReportFilter) ([]CategoryMarginItem, error)
	GetSIPNAPReport(ctx context.Context, filter SIPNAPReportFilter) (*SIPNAPReportResponse, error)
	UpdateBranchBonusSetting(ctx context.Context, branchID string, req UpdateBranchBonusSettingRequest) error
}
