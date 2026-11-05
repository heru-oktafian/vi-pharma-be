package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"pharma-be/internal/domain"
)

type prescriptionRepository struct {
	db *gorm.DB
}

func NewPrescriptionRepository(db *gorm.DB) domain.PrescriptionRepository {
	return &prescriptionRepository{db: db}
}

func (r *prescriptionRepository) GetNextPrescriptionNumber(ctx context.Context, branchID string, date time.Time) (string, error) {
	dateStr := date.Format("20060102")
	prefix := fmt.Sprintf("RXP/%s/", dateStr)

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Prescription{}).
		Where("prescription_number LIKE ?", prefix+"%").
		Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *prescriptionRepository) Create(ctx context.Context, prescription *domain.Prescription) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Insert Prescription
		if err := tx.Create(prescription).Error; err != nil {
			return err
		}

		// 2. Insert Items & Compound Items
		for i := range prescription.Items {
			item := &prescription.Items[i]
			item.PrescriptionID = prescription.ID
			if err := tx.Create(item).Error; err != nil {
				return err
			}

			// If it has compound items
			for j := range item.CompoundItems {
				comp := &item.CompoundItems[j]
				comp.PrescriptionItemID = item.ID
				if err := tx.Create(comp).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *prescriptionRepository) FindByID(ctx context.Context, id string) (*domain.Prescription, error) {
	var prescription domain.Prescription
	err := r.db.WithContext(ctx).
		Preload("Doctor").
		Preload("Branch").
		Preload("Pharmacist").
		Preload("Items.Medicine").
		Preload("Items.Medicine.Category").
		Preload("Items.Unit").
		Preload("Items.Batch").
		Preload("Items.CompoundItems.Medicine").
		Preload("Items.CompoundItems.Unit").
		Preload("Items.CompoundItems.Batch").
		Where("id = ? OR prescription_number = ?", id, id).
		First(&prescription).Error
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

func (r *prescriptionRepository) FindByNumber(ctx context.Context, number string) (*domain.Prescription, error) {
	var prescription domain.Prescription
	err := r.db.WithContext(ctx).
		Preload("Doctor").
		Preload("Branch").
		Preload("Pharmacist").
		Preload("Items.Medicine").
		Preload("Items.Unit").
		Preload("Items.Batch").
		Preload("Items.CompoundItems.Medicine").
		Preload("Items.CompoundItems.Unit").
		Preload("Items.CompoundItems.Batch").
		Where("prescription_number = ?", number).
		First(&prescription).Error
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

func (r *prescriptionRepository) FindAll(ctx context.Context, filter domain.PrescriptionFilter) ([]domain.Prescription, int64, error) {
	var prescriptions []domain.Prescription
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Prescription{})

	if filter.BranchID != "" {
		query = query.Where("branch_id = ?", filter.BranchID)
	}

	if filter.Status != "" && filter.Status != "SEMUA" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.IsPaid != nil {
		query = query.Where("is_paid = ?", *filter.IsPaid)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("prescription_number ILIKE ? OR patient_name ILIKE ? OR doctor_name ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if filter.StartDate != "" {
		query = query.Where("prescription_date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("prescription_date <= ?", filter.EndDate+" 23:59:59")
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
		Preload("Doctor").
		Preload("Pharmacist").
		Preload("Items").
		Preload("Items.Medicine").
		Preload("Items.CompoundItems").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&prescriptions).Error
	if err != nil {
		return nil, 0, err
	}

	return prescriptions, total, nil
}

func (r *prescriptionRepository) Update(ctx context.Context, prescription *domain.Prescription) error {
	return r.db.WithContext(ctx).Save(prescription).Error
}

func (r *prescriptionRepository) UpdateStatus(ctx context.Context, id string, status domain.PrescriptionStatus, pharmacistID *string, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if pharmacistID != nil {
		updates["pharmacist_id"] = *pharmacistID
	}
	if notes != "" {
		updates["notes"] = notes
	}

	return r.db.WithContext(ctx).
		Model(&domain.Prescription{}).
		Where("id = ?", id).
		Updates(updates).Error
}
