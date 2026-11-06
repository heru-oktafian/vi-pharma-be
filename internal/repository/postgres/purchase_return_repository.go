package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type purchaseReturnRepository struct {
	db *gorm.DB
}

func NewPurchaseReturnRepository(db *gorm.DB) domain.PurchaseReturnRepository {
	return &purchaseReturnRepository{db: db}
}

func (r *purchaseReturnRepository) CreatePurchaseReturn(ctx context.Context, ret *domain.PurchaseReturn) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(ret).Error
	})
}

func (r *purchaseReturnRepository) GetPurchaseReturnByID(ctx context.Context, id string) (*domain.PurchaseReturn, error) {
	var ret domain.PurchaseReturn
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Branch").
		Preload("CreatedBy").
		Preload("Items.Medicine").
		Preload("Items.Medicine.BaseUnit").
		Preload("Items.Batch").
		Preload("Items.Unit").
		Where("id = ?", id).
		First(&ret).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (r *purchaseReturnRepository) FindAllPurchaseReturns(ctx context.Context, branchID string, filter domain.PurchaseReturnFilter) ([]domain.PurchaseReturn, int64, error) {
	var returns []domain.PurchaseReturn
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.PurchaseReturn{}).Where("branch_id = ?", branchID)

	if filter.Search != "" {
		s := "%" + filter.Search + "%"
		query = query.Where("return_number ILIKE ? OR invoice_number ILIKE ? OR notes ILIKE ?", s, s, s)
	}

	if filter.SupplierID != "" {
		query = query.Where("supplier_id = ?", filter.SupplierID)
	}

	if filter.Status != "" && filter.Status != "ALL" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Reason != "" && filter.Reason != "ALL" {
		query = query.Where("reason = ?", filter.Reason)
	}

	if filter.CompensationType != "" && filter.CompensationType != "ALL" {
		query = query.Where("compensation_type = ?", filter.CompensationType)
	}

	if filter.StartDate != "" {
		if sDate, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			query = query.Where("return_date >= ?", sDate)
		}
	}

	if filter.EndDate != "" {
		if eDate, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			query = query.Where("return_date <= ?", eDate)
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
		Preload("Items.Medicine").
		Preload("Items.Batch").
		Preload("Items.Unit").
		Order("return_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&returns).Error

	return returns, total, err
}

func (r *purchaseReturnRepository) UpdatePurchaseReturn(ctx context.Context, ret *domain.PurchaseReturn, items []domain.PurchaseReturnItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(ret).Error; err != nil {
			return err
		}

		if len(items) > 0 {
			if err := tx.Where("purchase_return_id = ?", ret.ID).Delete(&domain.PurchaseReturnItem{}).Error; err != nil {
				return err
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *purchaseReturnRepository) UpdateReturnStatus(ctx context.Context, id string, status domain.PurchaseReturnStatus, notes string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if notes != "" {
		updates["notes"] = notes
	}

	result := r.db.WithContext(ctx).Model(&domain.PurchaseReturn{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPurchaseReturnNotFound
	}
	return nil
}

func (r *purchaseReturnRepository) ProcessStockDeduction(ctx context.Context, ret *domain.PurchaseReturn, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sup domain.Supplier
		_ = tx.Where("id = ?", ret.SupplierID).First(&sup).Error
		supName := sup.Name
		if supName == "" {
			supName = "Distributor PBF"
		}

		for _, item := range ret.Items {
			var batch domain.MedicineBatch
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND branch_id = ?", item.BatchID, ret.BranchID).
				First(&batch).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrBatchNotFound
				}
				return err
			}

			var balanceLoc float64
			if item.Location == domain.LocationShelf {
				if batch.QtyShelf < item.QtyBaseUnit {
					return domain.ErrInsufficientShelfStock
				}
				batch.QtyShelf -= item.QtyBaseUnit
				balanceLoc = batch.QtyShelf
			} else {
				// Default to GUDANG
				if batch.QtyWarehouse < item.QtyBaseUnit {
					return domain.ErrInsufficientWarehouseStock
				}
				batch.QtyWarehouse -= item.QtyBaseUnit
				balanceLoc = batch.QtyWarehouse
			}

			batch.CalculateTotal()
			if err := tx.Save(&batch).Error; err != nil {
				return err
			}

			mutationNotes := fmt.Sprintf("Retur ke PBF %s (No. Retur: %s)", supName, ret.ReturnNumber)
			if item.ItemReason != "" {
				mutationNotes += " - " + item.ItemReason
			} else if ret.Reason != "" {
				mutationNotes += " - " + ret.Reason
			}

			mutation := domain.StockMutation{
				ID:           utils.GenerateStockMutationID(),
				BranchID:     ret.BranchID,
				MedicineID:   item.MedicineID,
				BatchID:      item.BatchID,
				MutationType: domain.MutationReturnSupplier, // RETUR_PBF
				Location:     item.Location,
				ReferenceNo:  ret.ReturnNumber,
				QtyIn:        0,
				QtyOut:       item.QtyBaseUnit,
				BalanceAfter: balanceLoc,
				TotalAfter:   batch.TotalStock,
				Notes:        mutationNotes,
				UserID:       userID,
			}

			if err := tx.Create(&mutation).Error; err != nil {
				return err
			}
		}

		// Update return status to COMPLETED
		return tx.Model(&domain.PurchaseReturn{}).
			Where("id = ?", ret.ID).
			Update("status", domain.ReturnStatusCompleted).Error
	})
}

func (r *purchaseReturnRepository) DeletePurchaseReturn(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret domain.PurchaseReturn
		if err := tx.Where("id = ?", id).First(&ret).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrPurchaseReturnNotFound
			}
			return err
		}

		if ret.Status != domain.ReturnStatusDraft {
			return domain.ErrPurchaseReturnAlreadyProcessed
		}

		if err := tx.Where("purchase_return_id = ?", id).Delete(&domain.PurchaseReturnItem{}).Error; err != nil {
			return err
		}

		return tx.Delete(&ret).Error
	})
}

func (r *purchaseReturnRepository) GetNextReturnNumber(ctx context.Context, branchID string) (string, error) {
	now := time.Now()
	monthStr := now.Format("200601")
	var count int64
	likePattern := fmt.Sprintf("RET/PBF/%s/%%", monthStr)

	err := r.db.WithContext(ctx).Model(&domain.PurchaseReturn{}).
		Where("return_number LIKE ?", likePattern).
		Count(&count).Error
	if err != nil {
		return "", err
	}

	nextSeq := count + 1
	return fmt.Sprintf("RET/PBF/%s/%04d", monthStr, nextSeq), nil
}
