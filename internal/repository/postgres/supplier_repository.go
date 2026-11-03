package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) domain.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) FindAll(ctx context.Context, search string, activeOnly bool) ([]domain.Supplier, error) {
	var suppliers []domain.Supplier
	query := r.db.WithContext(ctx).Model(&domain.Supplier{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR pic_name ILIKE ? OR phone ILIKE ?", pattern, pattern, pattern, pattern)
	}

	err := query.Order("name ASC").Find(&suppliers).Error
	return suppliers, err
}

func (r *supplierRepository) FindByID(ctx context.Context, id string) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) FindByCode(ctx context.Context, code string) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return &supplier, nil
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *supplierRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Supplier{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrSupplierNotFound
	}
	return nil
}
