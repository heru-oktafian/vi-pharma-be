package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type posRepository struct {
	db *gorm.DB
}

func NewPOSRepository(db *gorm.DB) domain.POSRepository {
	return &posRepository{db: db}
}

func (r *posRepository) GenerateInvoiceNumber(ctx context.Context, branchID string) (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	prefix := fmt.Sprintf("POS/%s/", dateStr)

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Sale{}).
		Where("invoice_number LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *posRepository) CreateSale(ctx context.Context, sale *domain.Sale) error {
	// Fallback to valid Superadmin user if cashierID is empty or SYSTEM to prevent FK failure
	if sale.CashierID == "" || sale.CashierID == "SYSTEM" {
		sale.CashierID = "USR000000000001"
	}
	if sale.BranchID == "" {
		sale.BranchID = "CBG000000000001"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Process each sale item & deduct from shelf stock (FEFO)
		for i := range sale.Items {
			item := &sale.Items[i]

			var batch domain.MedicineBatch
			if err := tx.Where("id = ? AND branch_id = ?", item.BatchID, sale.BranchID).First(&batch).Error; err != nil {
				return fmt.Errorf("batch %s: %w", item.BatchID, domain.ErrBatchNotFound)
			}

			// Verify shelf stock availability
			if batch.QtyShelf < item.QtyBase {
				return fmt.Errorf("obat %s pada batch %s: stok rak hanya %.2f, diminta %.2f: %w",
					item.MedicineID, batch.BatchNumber, batch.QtyShelf, item.QtyBase, domain.ErrInsufficientShelfStock)
			}

			// Deduct shelf stock
			batch.QtyShelf -= item.QtyBase
			batch.CalculateTotal()

			if err := tx.Model(&domain.MedicineBatch{}).
				Where("id = ?", batch.ID).
				Updates(map[string]interface{}{
					"qty_shelf":   batch.QtyShelf,
					"total_stock": batch.TotalStock,
					"updated_at":  time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to deduct shelf stock: %w", err)
			}

			// Record audit log in Digital Stock Card (Stock Mutation)
			mutation := domain.StockMutation{
				ID:           utils.GenerateStockMutationID(),
				BranchID:     sale.BranchID,
				MedicineID:   item.MedicineID,
				BatchID:      batch.ID,
				MutationType: domain.MutationSale,
				Location:     domain.LocationShelf,
				ReferenceNo:  sale.InvoiceNumber,
				QtyIn:        0,
				QtyOut:       item.QtyBase,
				BalanceAfter: batch.QtyShelf,
				TotalAfter:   batch.TotalStock,
				Notes:        fmt.Sprintf("Penjualan Kasir POS %s (%s)", sale.InvoiceNumber, sale.CustomerName),
				UserID:       sale.CashierID,
			}
			if err := tx.Create(&mutation).Error; err != nil {
				return fmt.Errorf("failed to record stock mutation: %w", err)
			}
		}

		// 2. Insert Sale header & items
		if err := tx.Create(sale).Error; err != nil {
			return fmt.Errorf("failed to save sale transaction: %w", err)
		}

		return nil
	})
}

func (r *posRepository) FindSales(ctx context.Context, branchID string, filter domain.SaleFilter) ([]domain.Sale, int64, error) {
	var sales []domain.Sale
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Sale{}).Where("branch_id = ?", branchID)

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where(
			"invoice_number ILIKE ? OR customer_name ILIKE ? OR id IN (SELECT sale_id FROM sale_items si JOIN medicines m ON m.id = si.medicine_id WHERE m.name ILIKE ? OR m.generic_name ILIKE ? OR m.indication ILIKE ?)",
			s, s, s, s, s,
		)
	}

	if filter.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filter.PaymentMethod)
	}

	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate+" 00:00:00")
	}

	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	err := query.
		Preload("Cashier").
		Preload("Items.Medicine").
		Preload("Items.Unit").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&sales).Error

	return sales, total, err
}

func (r *posRepository) FindSaleByID(ctx context.Context, id string) (*domain.Sale, error) {
	var sale domain.Sale
	err := r.db.WithContext(ctx).
		Preload("Cashier").
		Preload("Branch").
		Preload("Items.Medicine").
		Preload("Items.Unit").
		Preload("Items.Batch").
		Where("id = ?", id).
		First(&sale).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrSaleNotFound
		}
		return nil, err
	}
	return &sale, nil
}

func (r *posRepository) GetTodaySummary(ctx context.Context, branchID string) (*domain.DailyCashierSummaryResponse, error) {
	today := time.Now().Format("2006-01-02")
	startOfDay := today + " 00:00:00"
	endOfDay := today + " 23:59:59"

	var stats struct {
		TotalTransactions int64   `gorm:"column:total_transactions"`
		TotalRevenue      float64 `gorm:"column:total_revenue"`
		TotalCash         float64 `gorm:"column:total_cash"`
		TotalNonCash      float64 `gorm:"column:total_non_cash"`
	}

	err := r.db.WithContext(ctx).
		Model(&domain.Sale{}).
		Select(`
			COUNT(*) as total_transactions,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN payment_method = 'TUNAI' THEN total_amount ELSE 0 END), 0) as total_cash,
			COALESCE(SUM(CASE WHEN payment_method != 'TUNAI' THEN total_amount ELSE 0 END), 0) as total_non_cash
		`).
		Where("branch_id = ? AND payment_status = ? AND created_at BETWEEN ? AND ?",
			branchID, domain.PaymentStatusPaid, startOfDay, endOfDay).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	var totalItemsSold float64
	err = r.db.WithContext(ctx).
		Table("sale_items").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Where("sales.branch_id = ? AND sales.payment_status = ? AND sales.created_at BETWEEN ? AND ?",
			branchID, domain.PaymentStatusPaid, startOfDay, endOfDay).
		Select("COALESCE(SUM(sale_items.qty), 0)").
		Scan(&totalItemsSold).Error
	if err != nil {
		return nil, err
	}

	return &domain.DailyCashierSummaryResponse{
		Date:              today,
		TotalTransactions: stats.TotalTransactions,
		TotalRevenue:      stats.TotalRevenue,
		TotalCash:         stats.TotalCash,
		TotalNonCash:      stats.TotalNonCash,
		TotalItemsSold:    totalItemsSold,
	}, nil
}
