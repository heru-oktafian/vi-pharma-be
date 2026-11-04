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
		Preload("BaseUnit")

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("name ILIKE ? OR item_code ILIKE ? OR barcode ILIKE ?", s, s, s)
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

		results = append(results, domain.StockSummaryResponse{
			MedicineID:     med.ID,
			ItemCode:       med.ItemCode,
			Name:           med.Name,
			BaseUnitName:   baseUnitName,
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
// Stock Opname
// -------------------------------------------------------------

func (r *inventoryRepository) CreateOpname(ctx context.Context, opname *domain.StockOpname) error {
	return r.db.WithContext(ctx).Create(opname).Error
}

func (r *inventoryRepository) GetOpnameByID(ctx context.Context, opnameID string) (*domain.StockOpname, error) {
	var opname domain.StockOpname
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
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
