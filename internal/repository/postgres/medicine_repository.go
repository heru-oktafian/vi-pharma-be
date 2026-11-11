package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type medicineRepository struct {
	db *gorm.DB
}

func NewMedicineRepository(db *gorm.DB) domain.MedicineRepository {
	return &medicineRepository{db: db}
}

func (r *medicineRepository) FindAll(ctx context.Context, filter domain.MedicineFilter) ([]domain.Medicine, int64, error) {
	var medicines []domain.Medicine
	var totalItems int64

	query := r.db.WithContext(ctx).Model(&domain.Medicine{})

	if filter.ActiveOnly {
		query = query.Where("medicines.is_active = ?", true)
	}

	if filter.Classification != "" {
		query = query.Where("medicines.drug_classification = ?", filter.Classification)
	}

	if filter.CategoryID != "" {
		query = query.Where("medicines.category_id = ?", filter.CategoryID)
	}

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		query = query.Where("medicines.name ILIKE ? OR medicines.generic_name ILIKE ? OR medicines.item_code ILIKE ? OR medicines.barcode ILIKE ? OR medicines.nie ILIKE ? OR medicines.indication ILIKE ?",
			pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if err := query.Count(&totalItems).Error; err != nil {
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
		Preload("Category").
		Preload("BaseUnit").
		Preload("DefaultSupplier").
		Preload("Units.Unit").
		Preload("Prices.Unit").
		Order("medicines.name ASC").
		Offset(offset).
		Limit(limit).
		Find(&medicines).Error

	return medicines, totalItems, err
}

func (r *medicineRepository) FindByID(ctx context.Context, id string) (*domain.Medicine, error) {
	var medicine domain.Medicine
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("BaseUnit").
		Preload("DefaultSupplier").
		Preload("Units.Unit").
		Preload("Prices.Unit").
		First(&medicine, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMedicineNotFound
		}
		return nil, err
	}
	return &medicine, nil
}

func (r *medicineRepository) FindByItemCode(ctx context.Context, itemCode string) (*domain.Medicine, error) {
	var medicine domain.Medicine
	err := r.db.WithContext(ctx).First(&medicine, "item_code = ?", itemCode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMedicineNotFound
		}
		return nil, err
	}
	return &medicine, nil
}

func (r *medicineRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.Medicine, error) {
	var medicine domain.Medicine
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("BaseUnit").
		Preload("DefaultSupplier").
		Preload("Units.Unit").
		Preload("Prices.Unit").
		First(&medicine, "barcode = ?", barcode).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMedicineNotFound
		}
		return nil, err
	}
	return &medicine, nil
}

func (r *medicineRepository) Create(ctx context.Context, medicine *domain.Medicine, units []domain.MedicineUnit, prices []domain.MedicinePrice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create Medicine record
		if err := tx.Create(medicine).Error; err != nil {
			return err
		}

		// 2. Create Medicine Units
		for i := range units {
			units[i].ID = utils.GenerateMedicineUnitID()
			units[i].MedicineID = medicine.ID
			if err := tx.Create(&units[i]).Error; err != nil {
				return err
			}
		}

		// 3. Create Medicine Prices
		for i := range prices {
			prices[i].ID = utils.GenerateMedicinePriceID()
			prices[i].MedicineID = medicine.ID
			if err := tx.Create(&prices[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *medicineRepository) Update(ctx context.Context, medicine *domain.Medicine, units []domain.MedicineUnit, prices []domain.MedicinePrice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		medicine.UpdatedAt = time.Now()
		if err := tx.Save(medicine).Error; err != nil {
			return err
		}

		// Replace Units if provided
		if len(units) > 0 {
			if err := tx.Where("medicine_id = ?", medicine.ID).Delete(&domain.MedicineUnit{}).Error; err != nil {
				return err
			}
			for i := range units {
				units[i].ID = utils.GenerateMedicineUnitID()
				units[i].MedicineID = medicine.ID
				if err := tx.Create(&units[i]).Error; err != nil {
					return err
				}
			}
		}

		// Replace Prices if provided
		if len(prices) > 0 {
			if err := tx.Where("medicine_id = ?", medicine.ID).Delete(&domain.MedicinePrice{}).Error; err != nil {
				return err
			}
			for i := range prices {
				prices[i].ID = utils.GenerateMedicinePriceID()
				prices[i].MedicineID = medicine.ID
				if err := tx.Create(&prices[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *medicineRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Medicine{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrMedicineNotFound
	}
	return nil
}
