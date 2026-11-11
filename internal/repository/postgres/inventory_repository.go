package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

// -------------------------------------------------------------
// Batches
// -------------------------------------------------------------

func (r *inventoryRepository) CreateBatch(ctx context.Context, batch *domain.MedicineBatch) error {
	batch.CalculateTotal()
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *inventoryRepository) UpdateBatch(ctx context.Context, batch *domain.MedicineBatch) error {
	batch.CalculateTotal()
	return r.db.WithContext(ctx).Save(batch).Error
}

func (r *inventoryRepository) FindBatchByID(ctx context.Context, batchID string) (*domain.MedicineBatch, error) {
	var batch domain.MedicineBatch
	err := r.db.WithContext(ctx).
		Preload("Medicine").
		Preload("Medicine.BaseUnit").
		Where("id = ?", batchID).
		First(&batch).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBatchNotFound
		}
		return nil, err
	}
	batch.CalculateTotal()
	return &batch, nil
}

func (r *inventoryRepository) FindBatchesByMedicine(ctx context.Context, branchID, medicineID string) ([]*domain.MedicineBatch, error) {
	var batches []*domain.MedicineBatch
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND medicine_id = ? AND is_active = true", branchID, medicineID).
		Order("expired_date ASC, created_at ASC"). // FEFO: First Expired, First Out
		Find(&batches).Error

	if err != nil {
		return nil, err
	}

	for _, b := range batches {
		b.CalculateTotal()
	}
	return batches, nil
}

func (r *inventoryRepository) FindBatchByNumberAndMedicine(ctx context.Context, branchID, medicineID, batchNumber string) (*domain.MedicineBatch, error) {
	var batch domain.MedicineBatch
	err := r.db.WithContext(ctx).
		Where("branch_id = ? AND medicine_id = ? AND batch_number = ?", branchID, medicineID, batchNumber).
		First(&batch).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBatchNotFound
		}
		return nil, err
	}
	batch.CalculateTotal()
	return &batch, nil
}

func (r *inventoryRepository) FindBatchesFEFO(ctx context.Context, branchID string) ([]*domain.MedicineBatch, error) {
	var batches []*domain.MedicineBatch
	err := r.db.WithContext(ctx).
		Preload("Medicine").
		Preload("Medicine.BaseUnit").
		Where("branch_id = ? AND is_active = true", branchID).
		Order("expired_date ASC").
		Find(&batches).Error

	if err != nil {
		return nil, err
	}

	for _, b := range batches {
		b.CalculateTotal()
	}
	return batches, nil
}

func (r *inventoryRepository) GetStockSummaryByBranch(ctx context.Context, branchID string, search string) ([]domain.StockSummaryResponse, error) {
	var medicines []domain.Medicine
	query := r.db.WithContext(ctx).
		Preload("BaseUnit").
		Preload("Prices")

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("name ILIKE ? OR item_code ILIKE ? OR barcode ILIKE ? OR generic_name ILIKE ? OR indication ILIKE ?", s, s, s, s, s)
	}

	if err := query.Order("name ASC").Find(&medicines).Error; err != nil {
		return nil, err
	}

	results := make([]domain.StockSummaryResponse, 0, len(medicines))

	for _, med := range medicines {
		var batches []*domain.MedicineBatch
		_ = r.db.WithContext(ctx).
			Where("branch_id = ? AND medicine_id = ? AND is_active = true", branchID, med.ID).
			Order("expired_date ASC").
			Find(&batches).Error

		var totalWh float64
		var totalSh float64

		for _, b := range batches {
			b.CalculateTotal()
			totalWh += b.QtyWarehouse
			totalSh += b.QtyShelf
		}

		baseUnitName := "Pcs"
		if med.BaseUnit != nil {
			baseUnitName = med.BaseUnit.Name
		}

		var sellPrice float64
		for _, pr := range med.Prices {
			if pr.PriceType == domain.PriceTypeUmum && pr.Price > 0 {
				sellPrice = pr.Price
				break
			}
		}
		if sellPrice == 0 && len(med.Prices) > 0 {
			sellPrice = med.Prices[0].Price
		}

		results = append(results, domain.StockSummaryResponse{
			MedicineID:     med.ID,
			ItemCode:       med.ItemCode,
			Barcode:        med.Barcode,
			Name:           med.Name,
			GenericName:    med.GenericName,
			Indication:     med.Indication,
			CategoryID:     med.CategoryID,
			BaseUnitID:     med.BaseUnitID,
			BaseUnitName:   baseUnitName,
			SellPrice:      sellPrice,
			Prices:         med.Prices,
			TotalWarehouse: totalWh,
			TotalShelf:     totalSh,
			TotalStock:     totalWh + totalSh,
			MinStock:       med.MinStock,
			Batches:        batches,
		})
	}

	return results, nil
}

// -------------------------------------------------------------
// Mutations / Digital Stock Card
// -------------------------------------------------------------

func (r *inventoryRepository) CreateMutation(ctx context.Context, mutation *domain.StockMutation) error {
	return r.db.WithContext(ctx).Create(mutation).Error
}

func (r *inventoryRepository) GetMutationsByMedicine(ctx context.Context, branchID, medicineID string, startDate, endDate time.Time) ([]*domain.StockMutation, error) {
	var mutations []*domain.StockMutation
	query := r.db.WithContext(ctx).
		Preload("Batch").
		Preload("User").
		Where("branch_id = ? AND medicine_id = ?", branchID, medicineID)

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Order("created_at DESC").Find(&mutations).Error
	return mutations, err
}

// -------------------------------------------------------------
// Purchase Invoices
// -------------------------------------------------------------

func (r *inventoryRepository) CreatePurchaseInvoice(ctx context.Context, invoice *domain.PurchaseInvoice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(invoice).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *inventoryRepository) GetPurchaseInvoiceByID(ctx context.Context, invoiceID string) (*domain.PurchaseInvoice, error) {
	var invoice domain.PurchaseInvoice
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("ReceivedBy").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Unit").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC, created_at DESC")
		}).
		Preload("Payments.CreatedBy").
		Where("id = ?", invoiceID).
		First(&invoice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *inventoryRepository) ListPurchaseInvoices(ctx context.Context, branchID string, page, limit int) ([]*domain.PurchaseInvoice, int64, error) {
	var invoices []*domain.PurchaseInvoice
	var total int64

	offset := (page - 1) * limit
	baseQuery := r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).Where("branch_id = ?", branchID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Preload("Supplier").
		Preload("ReceivedBy").
		Order("invoice_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invoices).Error

	return invoices, total, err
}

// -------------------------------------------------------------
// Accounts Payable & Invoice Payments
// -------------------------------------------------------------

func (r *inventoryRepository) GetAccountsPayableSummary(ctx context.Context, branchID string) (*domain.AccountsPayableSummary, error) {
	now := time.Now().Truncate(24 * time.Hour)
	next7Days := now.AddDate(0, 0, 7)
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	summary := &domain.AccountsPayableSummary{}

	// 1. Total Outstanding Debt (Total Tagihan - Total Terbayar pada faktur yang belum lunas)
	type DebtResult struct {
		TotalDebt float64
	}
	var debtRes DebtResult
	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0) as total_debt").
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED'", branchID).
		Scan(&debtRes)
	summary.TotalOutstandingDebt = debtRes.TotalDebt

	// 2. Total Invoices Count & Unpaid Invoices Count
	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("branch_id = ? AND payment_status != 'CANCELLED'", branchID).
		Count(&summary.TotalInvoices)

	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED'", branchID).
		Count(&summary.UnpaidInvoicesCount)

	// 3. Overdue Debt (Hutang macet / lewat jatuh tempo)
	var overdueRes DebtResult
	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0) as total_debt").
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED' AND due_date < ?", branchID, now).
		Scan(&overdueRes)
	summary.OverdueDebt = overdueRes.TotalDebt

	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED' AND due_date < ?", branchID, now).
		Count(&summary.OverdueCount)

	// 4. Due Soon (Jatuh tempo 7 hari ke depan)
	var dueSoonRes DebtResult
	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Select("COALESCE(SUM(total_amount - paid_amount), 0) as total_debt").
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED' AND due_date >= ? AND due_date <= ?", branchID, now, next7Days).
		Scan(&dueSoonRes)
	summary.DueSoonDebt = dueSoonRes.TotalDebt

	r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("branch_id = ? AND payment_status != 'PAID' AND payment_status != 'CANCELLED' AND due_date >= ? AND due_date <= ?", branchID, now, next7Days).
		Count(&summary.DueSoonCount)

	// 5. Total Paid This Month (Pelunasan hutang bulan berjalan)
	var paidThisMonthRes struct {
		TotalPaid float64
	}
	r.db.WithContext(ctx).Model(&domain.PurchaseInvoicePayment{}).
		Select("COALESCE(SUM(amount), 0) as total_paid").
		Where("branch_id = ? AND payment_date >= ?", branchID, firstOfMonth).
		Scan(&paidThisMonthRes)
	summary.TotalPaidThisMonth = paidThisMonthRes.TotalPaid

	return summary, nil
}

func (r *inventoryRepository) ListAccountsPayableInvoices(ctx context.Context, branchID string, filter domain.AccountsPayableFilter) ([]domain.PurchaseInvoice, int64, error) {
	var invoices []domain.PurchaseInvoice
	var total int64

	now := time.Now().Truncate(24 * time.Hour)
	next7Days := now.AddDate(0, 0, 7)

	query := r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("branch_id = ? AND payment_status != 'CANCELLED'", branchID)

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("(invoice_number ILIKE ? OR supplier_id IN (SELECT id FROM suppliers WHERE name ILIKE ?))", s, s)
	}

	if filter.SupplierID != "" {
		query = query.Where("supplier_id = ?", filter.SupplierID)
	}

	if filter.Status != "" {
		query = query.Where("payment_status = ?", filter.Status)
	}

	if filter.IsOverdue {
		query = query.Where("payment_status != 'PAID' AND due_date < ?", now)
	} else if filter.DueSoon {
		query = query.Where("payment_status != 'PAID' AND due_date >= ? AND due_date <= ?", now, next7Days)
	}

	if filter.StartDate != "" && filter.EndDate != "" {
		query = query.Where("invoice_date >= ? AND invoice_date <= ?", filter.StartDate, filter.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	switch filter.SortBy {
	case "due_date_desc":
		query = query.Order("due_date DESC NULLS LAST")
	case "debt_desc":
		query = query.Order("(total_amount - paid_amount) DESC")
	case "date_desc":
		query = query.Order("invoice_date DESC")
	default:
		// Default: due_date_asc (faktur yang paling mendesak jatuh tempo muncul paling atas)
		query = query.Order("due_date ASC NULLS LAST, invoice_date ASC")
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
		Preload("Supplier").
		Preload("ReceivedBy").
		Preload("Payments", func(db *gorm.DB) *gorm.DB {
			return db.Order("payment_date DESC")
		}).
		Offset(offset).
		Limit(limit).
		Find(&invoices).Error

	return invoices, total, err
}

func (r *inventoryRepository) CreateInvoicePayment(ctx context.Context, payment *domain.PurchaseInvoicePayment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// Recalculate total paid on this invoice
		var sumRes struct {
			TotalPaid float64
		}
		if err := tx.Model(&domain.PurchaseInvoicePayment{}).
			Select("COALESCE(SUM(amount), 0) as total_paid").
			Where("purchase_invoice_id = ?", payment.PurchaseInvoiceID).
			Scan(&sumRes).Error; err != nil {
			return err
		}

		// Fetch invoice to check total_amount
		var inv domain.PurchaseInvoice
		if err := tx.Select("id, total_amount").Where("id = ?", payment.PurchaseInvoiceID).First(&inv).Error; err != nil {
			return err
		}

		var newStatus domain.PaymentStatus
		if sumRes.TotalPaid >= inv.TotalAmount {
			newStatus = domain.PaymentStatusPaid
		} else if sumRes.TotalPaid > 0 {
			newStatus = domain.PaymentStatusPartial
		} else {
			newStatus = domain.PaymentStatusUnpaid
		}

		return tx.Model(&domain.PurchaseInvoice{}).
			Where("id = ?", payment.PurchaseInvoiceID).
			Updates(map[string]interface{}{
				"paid_amount":    sumRes.TotalPaid,
				"payment_status": newStatus,
			}).Error
	})
}

func (r *inventoryRepository) GetInvoicePaymentsByInvoiceID(ctx context.Context, invoiceID string) ([]domain.PurchaseInvoicePayment, error) {
	var payments []domain.PurchaseInvoicePayment
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		Where("purchase_invoice_id = ?", invoiceID).
		Order("payment_date DESC, created_at DESC").
		Find(&payments).Error
	return payments, err
}

func (r *inventoryRepository) UpdateInvoicePaymentStatus(ctx context.Context, invoiceID string, paidAmount float64, status domain.PaymentStatus) error {
	return r.db.WithContext(ctx).Model(&domain.PurchaseInvoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"paid_amount":    paidAmount,
			"payment_status": status,
		}).Error
}

// -------------------------------------------------------------
// Stock Opname
// -------------------------------------------------------------

func (r *inventoryRepository) CreateOpname(ctx context.Context, opname *domain.StockOpname) error {
	return r.db.WithContext(ctx).Create(opname).Error
}

func (r *inventoryRepository) GetOpnameByID(ctx context.Context, opnameID string) (*domain.StockOpname, error) {
	var opname domain.StockOpname
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Branch").
		Preload("Racks", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Racks.CountedBy").
		Preload("Racks.ApprovedBy").
		Preload("Racks.Items.Medicine").
		Preload("Racks.Items.Medicine.BaseUnit").
		Preload("Racks.Items.Batch").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Batch").
		Where("id = ?", opnameID).
		First(&opname).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOpnameNotFound
		}
		return nil, err
	}
	return &opname, nil
}

func (r *inventoryRepository) UpdateOpname(ctx context.Context, opname *domain.StockOpname) error {
	return r.db.WithContext(ctx).Save(opname).Error
}

func (r *inventoryRepository) ListOpnames(ctx context.Context, branchID string, page, limit int) ([]*domain.StockOpname, int64, error) {
	var opnames []*domain.StockOpname
	var total int64

	offset := (page - 1) * limit
	baseQuery := r.db.WithContext(ctx).Model(&domain.StockOpname{}).Where("branch_id = ?", branchID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Preload("CreatedBy").
		Order("opname_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&opnames).Error

	return opnames, total, err
}

func (r *inventoryRepository) AddOpnameItem(ctx context.Context, item *domain.StockOpnameItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *inventoryRepository) GetActiveParentOpname(ctx context.Context, branchID string) (*domain.StockOpname, error) {
	var opname domain.StockOpname
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Branch").
		Preload("Racks", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Racks.CountedBy").
		Preload("Racks.ApprovedBy").
		Preload("Racks.Items.Medicine").
		Preload("Racks.Items.Medicine.BaseUnit").
		Preload("Racks.Items.Batch").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Batch").
		Where("branch_id = ? AND status = ?", branchID, domain.OpnameOpen).
		Order("created_at DESC").
		First(&opname).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No active parent session currently
		}
		return nil, err
	}
	return &opname, nil
}

func (r *inventoryRepository) CreateOpnameRack(ctx context.Context, rack *domain.StockOpnameRack) error {
	return r.db.WithContext(ctx).Create(rack).Error
}

func (r *inventoryRepository) GetOpnameRackByID(ctx context.Context, rackID string) (*domain.StockOpnameRack, error) {
	var rack domain.StockOpnameRack
	err := r.db.WithContext(ctx).
		Preload("CountedBy").
		Preload("ApprovedBy").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Batch").
		Where("id = ?", rackID).
		First(&rack).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOpnameRackNotFound
		}
		return nil, err
	}
	return &rack, nil
}

func (r *inventoryRepository) UpdateOpnameRack(ctx context.Context, rack *domain.StockOpnameRack) error {
	return r.db.WithContext(ctx).Save(rack).Error
}

func (r *inventoryRepository) ListOpnameRacks(ctx context.Context, opnameID string) ([]domain.StockOpnameRack, error) {
	var racks []domain.StockOpnameRack
	err := r.db.WithContext(ctx).
		Preload("CountedBy").
		Preload("ApprovedBy").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Batch").
		Where("stock_opname_id = ?", opnameID).
		Order("created_at DESC").
		Find(&racks).Error
	return racks, err
}
