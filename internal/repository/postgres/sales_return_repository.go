package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type salesReturnRepository struct {
	db *gorm.DB
}

func NewSalesReturnRepository(db *gorm.DB) domain.SalesReturnRepository {
	return &salesReturnRepository{db: db}
}

func (r *salesReturnRepository) CreateSalesReturn(ctx context.Context, ret *domain.SalesReturn) error {
	// Fallback to valid Superadmin user if cashierID is empty or SYSTEM to prevent FK failure
	if ret.CashierID == "" || ret.CashierID == "SYSTEM" {
		ret.CashierID = "USR000000000001"
	}
	if ret.BranchID == "" {
		ret.BranchID = "CBG000000000001"
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Process each returned item: restore stock to batch & log stock mutation
		for i := range ret.Items {
			item := &ret.Items[i]

			// Fetch batch to restore stock
			var batch domain.MedicineBatch
			if err := tx.Where("id = ? AND branch_id = ?", item.BatchID, ret.BranchID).First(&batch).Error; err != nil {
				return fmt.Errorf("batch not found (%s): %w", item.BatchID, err)
			}

			var location domain.StockLocation
			var balAfter float64

			if item.Condition == domain.ReturnConditionGood {
				// Restore to RAK Kasir (Bisa dijual kembali)
				location = domain.LocationShelf
				batch.QtyShelf += item.QtyBase
				balAfter = batch.QtyShelf
			} else {
				// Put in GUDANG Karantina / Rusak
				location = domain.LocationWarehouse
				batch.QtyWarehouse += item.QtyBase
				balAfter = batch.QtyWarehouse
			}

			batch.CalculateTotal()

			if err := tx.Model(&domain.MedicineBatch{}).
				Where("id = ?", batch.ID).
				Updates(map[string]interface{}{
					"qty_shelf":     batch.QtyShelf,
					"qty_warehouse": batch.QtyWarehouse,
					"total_stock":   batch.TotalStock,
					"updated_at":    time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("failed to update batch stock: %w", err)
			}

			// Log Stock Mutation (Kartu Stok Digital)
			mutation := &domain.StockMutation{
				ID:           utils.GenerateStockMutationID(),
				BatchID:      batch.ID,
				BranchID:     ret.BranchID,
				MedicineID:   item.MedicineID,
				MutationType: domain.MutationReturnCustomer, // RETUR_PASIEN
				Location:     location,
				ReferenceNo:  ret.ReturnNumber,
				QtyIn:        item.QtyBase,
				QtyOut:       0,
				BalanceAfter: balAfter,
				TotalAfter:   batch.TotalStock,
				Notes:        fmt.Sprintf("Retur Pasien No. %s (Struk: %s, Alasan: %s)", ret.ReturnNumber, ret.InvoiceNumber, ret.Reason),
				UserID:       ret.CashierID,
			}

			if err := tx.Create(mutation).Error; err != nil {
				return fmt.Errorf("failed to log stock mutation: %w", err)
			}
		}

		// 2. Insert SalesReturn header & all return items (GORM creates items automatically through association)
		if err := tx.Create(ret).Error; err != nil {
			return fmt.Errorf("failed to create sales return: %w", err)
		}

		return nil
	})
}

func (r *salesReturnRepository) FindSalesReturnByID(ctx context.Context, id string) (*domain.SalesReturn, error) {
	var ret domain.SalesReturn
	err := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("Cashier").
		Preload("Sale").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Unit").
		Preload("Items.Batch").
		Preload("Items.SaleItem").
		Where("id = ?", id).
		First(&ret).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSalesReturnNotFound
		}
		return nil, err
	}

	return &ret, nil
}

func (r *salesReturnRepository) FindSalesReturns(ctx context.Context, branchID string, filter domain.SalesReturnFilter) ([]domain.SalesReturn, int64, error) {
	var returns []domain.SalesReturn
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.SalesReturn{}).Where("branch_id = ?", branchID)

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("(return_number ILIKE ? OR invoice_number ILIKE ? OR customer_name ILIKE ?)", s, s, s)
	}

	if filter.RefundType != "" {
		query = query.Where("refund_type = ?", filter.RefundType)
	}

	if filter.StartDate != "" {
		query = query.Where("return_date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("return_date <= ?", filter.EndDate)
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
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&returns).Error

	return returns, total, err
}

func (r *salesReturnRepository) GetReturnedQtyBySaleItem(ctx context.Context, saleItemID string) (float64, error) {
	var res struct {
		TotalQty float64
	}
	err := r.db.WithContext(ctx).Table("sales_return_items").
		Joins("JOIN sales_returns ON sales_returns.id = sales_return_items.sales_return_id").
		Select("COALESCE(SUM(sales_return_items.qty), 0) as total_qty").
		Where("sales_return_items.sale_item_id = ? AND sales_returns.status = 'COMPLETED'", saleItemID).
		Scan(&res).Error
	return res.TotalQty, err
}

func (r *salesReturnRepository) GetReturnedQtysBySaleID(ctx context.Context, saleID string) (map[string]float64, error) {
	type Row struct {
		SaleItemID string  `gorm:"column:sale_item_id"`
		TotalQty   float64 `gorm:"column:total_qty"`
	}
	var rows []Row
	err := r.db.WithContext(ctx).Table("sales_return_items").
		Joins("JOIN sales_returns ON sales_returns.id = sales_return_items.sales_return_id").
		Select("sales_return_items.sale_item_id, COALESCE(SUM(sales_return_items.qty), 0) as total_qty").
		Where("sales_returns.sale_id = ? AND sales_returns.status = 'COMPLETED'", saleID).
		Group("sales_return_items.sale_item_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	for _, r := range rows {
		result[r.SaleItemID] = r.TotalQty
	}
	return result, nil
}

func (r *salesReturnRepository) GenerateReturnNumber(ctx context.Context, branchID string) (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	var count int64
	likePattern := fmt.Sprintf("RET/POS/%s/%%", dateStr)

	err := r.db.WithContext(ctx).Model(&domain.SalesReturn{}).
		Where("return_number LIKE ?", likePattern).
		Count(&count).Error
	if err != nil {
		return "", err
	}

	nextSeq := count + 1
	return fmt.Sprintf("RET/POS/%s/%04d", dateStr, nextSeq), nil
}
