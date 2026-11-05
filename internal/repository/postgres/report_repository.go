package postgres

import (
	"context"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
	"pharma-be/internal/domain"
)

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) domain.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetStaffPerformanceReport(ctx context.Context, filter domain.ReportFilter) (*domain.StaffPerformanceReportResponse, error) {
	branchID := filter.BranchID
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	startDate := filter.StartDate
	if startDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}

	endDate := filter.EndDate
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	startDateTime := startDate + " 00:00:00"
	endDateTime := endDate + " 23:59:59"

	// 1. Get Branch Info & Threshold Configurations
	var branch domain.Branch
	_ = r.db.WithContext(ctx).Where("id = ?", branchID).First(&branch).Error
	branchName := branch.BranchName
	if branchName == "" {
		branchName = "Apotek Vi-Pharma Pusat"
	}

	dailyMarginTarget := branch.DailyMarginTarget
	if filter.DailyMarginTarget > 0 {
		dailyMarginTarget = filter.DailyMarginTarget
	} else if dailyMarginTarget <= 0 {
		dailyMarginTarget = 250000 // Default: Rp 250.000 margin per hari
	}

	dailyBonusReward := branch.DailyBonusReward
	if filter.DailyBonusReward > 0 {
		dailyBonusReward = filter.DailyBonusReward
	} else if dailyBonusReward <= 0 {
		dailyBonusReward = 50000 // Default: Rp 50.000 flat bonus per hari tercapai
	}

	// 2. Query Cashier sales segmented by Day
	type CashierDailyStat struct {
		UserID            string  `gorm:"column:user_id"`
		UserName          string  `gorm:"column:user_name"`
		UserRole          string  `gorm:"column:user_role"`
		TxDate            string  `gorm:"column:tx_date"`
		TotalTransactions int64   `gorm:"column:total_transactions"`
		TotalRevenue      float64 `gorm:"column:total_revenue"`
		TotalCOGS         float64 `gorm:"column:total_cogs"`
	}

	var cashierStats []CashierDailyStat
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(s.cashier_id, 'KASIR_CABANG') as user_id,
			COALESCE(u.name, 'Petugas Kasir') as user_name,
			COALESCE(u.user_role::text, 'cashier') as user_role,
			TO_CHAR(s.created_at, 'YYYY-MM-DD') as tx_date,
			COUNT(DISTINCT s.id) as total_transactions,
			COALESCE(SUM(si.total), 0) as total_revenue,
			COALESCE(SUM(si.qty_base * COALESCE(b.buy_price, 0)), 0) as total_cogs
		FROM sales s
		LEFT JOIN users u ON u.id = s.cashier_id
		JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN medicine_batches b ON b.id = si.batch_id
		WHERE s.branch_id = ? 
		  AND s.payment_status = 'PAID'
		  AND s.created_at BETWEEN ? AND ?
		GROUP BY s.cashier_id, u.name, u.user_role, TO_CHAR(s.created_at, 'YYYY-MM-DD')
		ORDER BY tx_date ASC
	`, branchID, startDateTime, endDateTime).Scan(&cashierStats).Error
	if err != nil {
		return nil, err
	}

	// 3. Query Pharmacist compounded prescriptions segmented by Day
	type PharmacistDailyStat struct {
		UserID       string  `gorm:"column:user_id"`
		UserName     string  `gorm:"column:user_name"`
		UserRole     string  `gorm:"column:user_role"`
		TxDate       string  `gorm:"column:tx_date"`
		RxCount      int64   `gorm:"column:rx_count"`
		RxRevenue    float64 `gorm:"column:rx_revenue"`
		TotalTuslah  float64 `gorm:"column:total_tuslah"`
	}

	var pharmacistStats []PharmacistDailyStat
	_ = r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(p.pharmacist_id, 'APOTEKER_CABANG') as user_id,
			COALESCE(u.name, 'Apoteker Pelayan') as user_name,
			COALESCE(u.user_role::text, 'pharmacist') as user_role,
			TO_CHAR(p.created_at, 'YYYY-MM-DD') as tx_date,
			COUNT(DISTINCT p.id) as rx_count,
			COALESCE(SUM(p.grand_total), 0) as rx_revenue,
			COALESCE(SUM(p.total_tuslah), 0) as total_tuslah
		FROM prescriptions p
		LEFT JOIN users u ON u.id = p.pharmacist_id
		WHERE p.branch_id = ? 
		  AND p.status IN ('SIAP', 'SELESAI')
		  AND p.created_at BETWEEN ? AND ?
		  AND p.pharmacist_id IS NOT NULL
		GROUP BY p.pharmacist_id, u.name, u.user_role, TO_CHAR(p.created_at, 'YYYY-MM-DD')
		ORDER BY tx_date ASC
	`, branchID, startDateTime, endDateTime).Scan(&pharmacistStats)

	// 4. Merge stats into unified map per user and per day
	userMap := make(map[string]*domain.StaffPerformanceItem)
	userDailyMap := make(map[string]map[string]*domain.DailyStaffAchievement)

	for _, cs := range cashierStats {
		uID := cs.UserID
		if uID == "" {
			uID = "USR000000000001"
		}
		if _, exists := userMap[uID]; !exists {
			userMap[uID] = &domain.StaffPerformanceItem{
				UserID:   uID,
				UserName: cs.UserName,
				UserRole: cs.UserRole,
			}
		}
		item := userMap[uID]
		item.TotalTransactions += cs.TotalTransactions
		item.TotalRevenue += cs.TotalRevenue
		item.TotalCOGS += cs.TotalCOGS
		item.GrossMargin += (cs.TotalRevenue - cs.TotalCOGS)

		if _, exists := userDailyMap[uID]; !exists {
			userDailyMap[uID] = make(map[string]*domain.DailyStaffAchievement)
		}
		dMap := userDailyMap[uID]
		dateKey := cs.TxDate
		if _, exists := dMap[dateKey]; !exists {
			dMap[dateKey] = &domain.DailyStaffAchievement{
				Date:         dateKey,
				TargetMargin: dailyMarginTarget,
			}
		}
		dItem := dMap[dateKey]
		dItem.TotalRevenue += cs.TotalRevenue
		dItem.TotalCOGS += cs.TotalCOGS
		dItem.GrossMargin += (cs.TotalRevenue - cs.TotalCOGS)
		dItem.Transactions += cs.TotalTransactions
	}

	for _, ps := range pharmacistStats {
		uID := ps.UserID
		if uID == "" {
			continue
		}
		if _, exists := userMap[uID]; !exists {
			userMap[uID] = &domain.StaffPerformanceItem{
				UserID:   uID,
				UserName: ps.UserName,
				UserRole: ps.UserRole,
			}
		}
		item := userMap[uID]
		item.TotalTransactions += ps.RxCount
		item.TotalRevenue += ps.RxRevenue
		item.GrossMargin += ps.RxRevenue
		item.TuslahAmount += ps.TotalTuslah

		if _, exists := userDailyMap[uID]; !exists {
			userDailyMap[uID] = make(map[string]*domain.DailyStaffAchievement)
		}
		dMap := userDailyMap[uID]
		dateKey := ps.TxDate
		if _, exists := dMap[dateKey]; !exists {
			dMap[dateKey] = &domain.DailyStaffAchievement{
				Date:         dateKey,
				TargetMargin: dailyMarginTarget,
			}
		}
		dItem := dMap[dateKey]
		dItem.TotalRevenue += ps.RxRevenue
		dItem.GrossMargin += ps.RxRevenue
		dItem.Transactions += ps.RxCount
	}

	// 5. Evaluate daily threshold bonus per user
	for uID, item := range userMap {
		dMap := userDailyMap[uID]
		item.ActiveDaysCount = len(dMap)
		item.DailyTargetMargin = dailyMarginTarget
		item.DailyBonusReward = dailyBonusReward

		var daysAchieved int
		var dailyList []domain.DailyStaffAchievement

		for _, dItem := range dMap {
			if dItem.GrossMargin >= dailyMarginTarget {
				dItem.TargetAchieved = true
				dItem.BonusEarned = dailyBonusReward // Flat bonus: tidak bertambah meski melebihi target
				daysAchieved++
			} else {
				dItem.TargetAchieved = false
				dItem.BonusEarned = 0
			}
			dailyList = append(dailyList, *dItem)
		}

		sort.Slice(dailyList, func(i, j int) bool {
			return dailyList[i].Date > dailyList[j].Date
		})

		item.DaysAchieved = daysAchieved
		item.TotalBonus = float64(daysAchieved) * dailyBonusReward
		item.EstimatedBonus = item.TotalBonus
		item.DailyBreakdown = dailyList
	}

	// Calculate branch totals
	var totalBranchRevenue float64
	var totalBranchCOGS float64
	var totalBranchMargin float64
	var totalBonusPaid float64

	for _, it := range userMap {
		totalBranchRevenue += it.TotalRevenue
		totalBranchCOGS += it.TotalCOGS
		totalBranchMargin += it.GrossMargin
		totalBonusPaid += it.TotalBonus
	}

	var staffList []domain.StaffPerformanceItem
	for _, it := range userMap {
		// Calculate % of Revenue
		if totalBranchRevenue > 0 {
			it.RevenueContributionPercent = math.Round((it.TotalRevenue/totalBranchRevenue)*10000) / 100
		}

		// Calculate % of Margin
		if totalBranchMargin > 0 {
			it.MarginContributionPercent = math.Round((it.GrossMargin/totalBranchMargin)*10000) / 100
		}

		// Calculate Markup % (Laba terhadap Modal)
		if it.TotalCOGS > 0 {
			it.MarkupPercent = math.Round((it.GrossMargin/it.TotalCOGS)*10000) / 100
		} else if it.GrossMargin > 0 {
			it.MarkupPercent = 100
		}

		staffList = append(staffList, *it)
	}

	// Sort staffList descending by GrossMargin
	sort.Slice(staffList, func(i, j int) bool {
		return staffList[i].GrossMargin > staffList[j].GrossMargin
	})

	var averageBranchMarkup float64
	if totalBranchCOGS > 0 {
		averageBranchMarkup = math.Round((totalBranchMargin/totalBranchCOGS)*10000) / 100
	}

	return &domain.StaffPerformanceReportResponse{
		BranchID:            branchID,
		BranchName:          branchName,
		PeriodStart:         startDate,
		PeriodEnd:           endDate,
		TotalBranchRevenue:  totalBranchRevenue,
		TotalBranchCOGS:     totalBranchCOGS,
		TotalBranchMargin:   totalBranchMargin,
		AverageBranchMarkup: averageBranchMarkup,
		DailyMarginTarget:   dailyMarginTarget,
		DailyBonusReward:    dailyBonusReward,
		TotalBonusPaid:      totalBonusPaid,
		BonusPoolBudget:     filter.BonusPoolBudget,
		TargetMarginPercent: filter.TargetMarginPercent,
		StaffContributions:  staffList,
	}, nil
}

func (r *reportRepository) UpdateBranchBonusSetting(ctx context.Context, branchID string, dailyMarginTarget, dailyBonusReward float64) error {
	return r.db.WithContext(ctx).Model(&domain.Branch{}).
		Where("id = ?", branchID).
		Updates(map[string]interface{}{
			"daily_margin_target": dailyMarginTarget,
			"daily_bonus_reward":  dailyBonusReward,
		}).Error
}

func (r *reportRepository) GetSalesSummaryReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesSummaryReportResponse, error) {
	branchID := filter.BranchID
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	startDate := filter.StartDate
	if startDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}

	endDate := filter.EndDate
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	startDateTime := startDate + " 00:00:00"
	endDateTime := endDate + " 23:59:59"

	// 1. Total metrics
	var overall struct {
		TotalTransactions int64   `gorm:"column:total_transactions"`
		TotalRevenue      float64 `gorm:"column:total_revenue"`
		TotalCOGS         float64 `gorm:"column:total_cogs"`
	}

	_ = r.db.WithContext(ctx).Raw(`
		SELECT 
			COUNT(DISTINCT s.id) as total_transactions,
			COALESCE(SUM(si.total), 0) as total_revenue,
			COALESCE(SUM(si.qty_base * COALESCE(b.buy_price, 0)), 0) as total_cogs
		FROM sales s
		JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN medicine_batches b ON b.id = si.batch_id
		WHERE s.branch_id = ? 
		  AND s.payment_status = 'PAID'
		  AND s.created_at BETWEEN ? AND ?
	`, branchID, startDateTime, endDateTime).Scan(&overall)

	totalMargin := overall.TotalRevenue - overall.TotalCOGS
	var marginPercent float64
	if overall.TotalRevenue > 0 {
		marginPercent = math.Round((totalMargin/overall.TotalRevenue)*10000) / 100
	}

	// 2. Payment methods breakdown
	var paymentStats []struct {
		PaymentMethod string  `gorm:"column:payment_method"`
		Count         int64   `gorm:"column:count"`
		TotalAmount   float64 `gorm:"column:total_amount"`
	}

	_ = r.db.WithContext(ctx).Raw(`
		SELECT 
			payment_method,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as total_amount
		FROM sales
		WHERE branch_id = ?
		  AND payment_status = 'PAID'
		  AND created_at BETWEEN ? AND ?
		GROUP BY payment_method
		ORDER BY total_amount DESC
	`, branchID, startDateTime, endDateTime).Scan(&paymentStats)

	var pMethods []domain.PaymentMethodSummary
	for _, p := range paymentStats {
		pct := 0.0
		if overall.TotalRevenue > 0 {
			pct = math.Round((p.TotalAmount/overall.TotalRevenue)*10000) / 100
		}
		pMethods = append(pMethods, domain.PaymentMethodSummary{
			PaymentMethod: p.PaymentMethod,
			Count:         p.Count,
			TotalAmount:   p.TotalAmount,
			Percentage:    pct,
		})
	}

	// 3. Daily trends
	var trendStats []struct {
		Date         string  `gorm:"column:date"`
		Transactions int64   `gorm:"column:transactions"`
		Revenue      float64 `gorm:"column:revenue"`
		COGS         float64 `gorm:"column:cogs"`
	}

	_ = r.db.WithContext(ctx).Raw(`
		SELECT 
			TO_CHAR(s.created_at, 'YYYY-MM-DD') as date,
			COUNT(DISTINCT s.id) as transactions,
			COALESCE(SUM(si.total), 0) as revenue,
			COALESCE(SUM(si.qty_base * COALESCE(b.buy_price, 0)), 0) as cogs
		FROM sales s
		JOIN sale_items si ON si.sale_id = s.id
		LEFT JOIN medicine_batches b ON b.id = si.batch_id
		WHERE s.branch_id = ? 
		  AND s.payment_status = 'PAID'
		  AND s.created_at BETWEEN ? AND ?
		GROUP BY TO_CHAR(s.created_at, 'YYYY-MM-DD')
		ORDER BY date ASC
	`, branchID, startDateTime, endDateTime).Scan(&trendStats)

	var dailyTrends []domain.DailySalesTrend
	for _, t := range trendStats {
		margin := t.Revenue - t.COGS
		dailyTrends = append(dailyTrends, domain.DailySalesTrend{
			Date:         t.Date,
			Transactions: t.Transactions,
			Revenue:      t.Revenue,
			COGS:         t.COGS,
			Margin:       margin,
		})
	}

	return &domain.SalesSummaryReportResponse{
		BranchID:             branchID,
		PeriodStart:          startDate,
		PeriodEnd:            endDate,
		TotalTransactions:    overall.TotalTransactions,
		TotalRevenue:         overall.TotalRevenue,
		TotalCOGS:            overall.TotalCOGS,
		TotalMargin:          totalMargin,
		MarginPercent:        marginPercent,
		PaymentMethodSummary: pMethods,
		DailyTrends:          dailyTrends,
	}, nil
}

func (r *reportRepository) GetCategoryMarginsReport(ctx context.Context, filter domain.ReportFilter) ([]domain.CategoryMarginItem, error) {
	branchID := filter.BranchID
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	startDate := filter.StartDate
	if startDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}

	endDate := filter.EndDate
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	startDateTime := startDate + " 00:00:00"
	endDateTime := endDate + " 23:59:59"

	var raw []struct {
		CategoryID   string  `gorm:"column:category_id"`
		CategoryName string  `gorm:"column:category_name"`
		TotalQtySold float64 `gorm:"column:total_qty_sold"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalCOGS    float64 `gorm:"column:total_cogs"`
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			c.id as category_id,
			COALESCE(c.name, 'Tanpa Kategori') as category_name,
			COALESCE(SUM(si.qty), 0) as total_qty_sold,
			COALESCE(SUM(si.total), 0) as total_revenue,
			COALESCE(SUM(si.qty_base * COALESCE(b.buy_price, 0)), 0) as total_cogs
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		JOIN medicines m ON m.id = si.medicine_id
		LEFT JOIN categories c ON c.id = m.category_id
		LEFT JOIN medicine_batches b ON b.id = si.batch_id
		WHERE s.branch_id = ? 
		  AND s.payment_status = 'PAID'
		  AND s.created_at BETWEEN ? AND ?
		GROUP BY c.id, c.name
		ORDER BY total_revenue DESC
	`, branchID, startDateTime, endDateTime).Scan(&raw).Error
	if err != nil {
		return nil, err
	}

	var results []domain.CategoryMarginItem
	for _, item := range raw {
		margin := item.TotalRevenue - item.TotalCOGS
		marginPct := 0.0
		if item.TotalRevenue > 0 {
			marginPct = math.Round((margin/item.TotalRevenue)*10000) / 100
		}
		markupPct := 0.0
		if item.TotalCOGS > 0 {
			markupPct = math.Round((margin/item.TotalCOGS)*10000) / 100
		}

		results = append(results, domain.CategoryMarginItem{
			CategoryID:    item.CategoryID,
			CategoryName:  item.CategoryName,
			TotalQtySold:  item.TotalQtySold,
			TotalRevenue:  item.TotalRevenue,
			TotalCOGS:     item.TotalCOGS,
			GrossMargin:   margin,
			MarginPercent: marginPct,
			MarkupPercent: markupPct,
		})
	}

	return results, nil
}
