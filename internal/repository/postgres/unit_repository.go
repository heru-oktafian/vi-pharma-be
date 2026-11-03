package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type unitRepository struct {
	db *gorm.DB
}

func NewUnitRepository(db *gorm.DB) domain.UnitRepository {
	return &unitRepository{db: db}
}

func (r *unitRepository) FindAll(ctx context.Context, search string, activeOnly bool) ([]domain.Unit, error) {
	var units []domain.Unit
	query := r.db.WithContext(ctx).Model(&domain.Unit{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR symbol ILIKE ?", searchPattern, searchPattern)
	}

	err := query.Order("name ASC").Find(&units).Error
	return units, err
}

func (r *unitRepository) FindByID(ctx context.Context, id string) (*domain.Unit, error) {
	var unit domain.Unit
	err := r.db.WithContext(ctx).First(&unit, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUnitNotFound
		}
		return nil, err
	}
	return &unit, nil
}

func (r *unitRepository) Create(ctx context.Context, unit *domain.Unit) error {
	return r.db.WithContext(ctx).Create(unit).Error
}

func (r *unitRepository) Update(ctx context.Context, unit *domain.Unit) error {
	return r.db.WithContext(ctx).Save(unit).Error
}

func (r *unitRepository) Delete(ctx context.Context, id string) error {
	// Soft delete by setting is_active = false
	result := r.db.WithContext(ctx).Model(&domain.Unit{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUnitNotFound
	}
	return nil
}
