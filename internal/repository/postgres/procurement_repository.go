package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type procurementRepository struct {
	db *gorm.DB
}

func NewProcurementRepository(db *gorm.DB) domain.ProcurementRepository {
	return &procurementRepository{db: db}
}

func (r *procurementRepository) CreatePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(po).Error
	})
}

func (r *procurementRepository) GetPurchaseOrderByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Branch").
		Preload("CreatedBy").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Unit").
		Where("id = ?", id).
		First(&po).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	return &po, nil
}

func (r *procurementRepository) FindAllPurchaseOrders(ctx context.Context, branchID string, filter domain.POFilter) ([]domain.PurchaseOrder, int64, error) {
	var orders []domain.PurchaseOrder
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PurchaseOrder{}).Where("branch_id = ?", branchID)

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("po_number ILIKE ? OR notes ILIKE ? OR pharmacist_name ILIKE ?", s, s, s)
	}

	if filter.SupplierID != "" {
		query = query.Where("supplier_id = ?", filter.SupplierID)
	}

	if filter.SPType != "" {
		query = query.Where("sp_type = ?", filter.SPType)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.StartDate != "" {
		if sDate, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			query = query.Where("order_date >= ?", sDate)
		}
	}

	if filter.EndDate != "" {
		if eDate, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			query = query.Where("order_date <= ?", eDate)
		}
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
		Preload("Supplier").
		Preload("CreatedBy").
		Order("order_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *procurementRepository) UpdatePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder, items []domain.PurchaseOrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update PO header
		if err := tx.Save(po).Error; err != nil {
			return err
		}

		// 2. Delete existing items if new items provided
		if len(items) > 0 {
			if err := tx.Where("purchase_order_id = ?", po.ID).Delete(&domain.PurchaseOrderItem{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *procurementRepository) UpdatePOStatus(ctx context.Context, id string, status domain.POStatus) error {
	result := r.db.WithContext(ctx).Model(&domain.PurchaseOrder{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPurchaseOrderNotFound
	}
	return nil
}

func (r *procurementRepository) DeletePurchaseOrder(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Only delete if DRAFT
		var po domain.PurchaseOrder
		if err := tx.Where("id = ?", id).First(&po).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrPurchaseOrderNotFound
			}
			return err
		}

		if po.Status != domain.POStatusDraft {
			return domain.ErrPOAlreadyProcessed
		}

		// Delete items
		if err := tx.Where("purchase_order_id = ?", id).Delete(&domain.PurchaseOrderItem{}).Error; err != nil {
			return err
		}

		return tx.Delete(&po).Error
	})
}

func (r *procurementRepository) GetNextPONumber(ctx context.Context, branchID string, spType domain.SPType) (string, error) {
	now := time.Now()
	prefixType := "REG"
	switch spType {
	case domain.SPTypePrecursor:
		prefixType = "PREK"
	case domain.SPTypeOOT:
		prefixType = "OOT"
	case domain.SPTypePsychotropic:
		prefixType = "PSI"
	case domain.SPTypeNarcotic:
		prefixType = "NARK"
	}

	monthStr := now.Format("200601")
	var count int64
	likePattern := fmt.Sprintf("SP/%s/%s/%%", prefixType, monthStr)

	err := r.db.WithContext(ctx).Model(&domain.PurchaseOrder{}).
		Where("po_number LIKE ?", likePattern).
		Count(&count).Error
	if err != nil {
		return "", err
	}

	nextSeq := count + 1
	return fmt.Sprintf("SP/%s/%s/%04d", prefixType, monthStr, nextSeq), nil
}
