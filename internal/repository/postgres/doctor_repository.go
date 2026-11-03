package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type doctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) domain.DoctorRepository {
	return &doctorRepository{db: db}
}

func (r *doctorRepository) FindAll(ctx context.Context, search string, activeOnly bool) ([]domain.Doctor, error) {
	var doctors []domain.Doctor
	query := r.db.WithContext(ctx).Model(&domain.Doctor{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR sip ILIKE ? OR specialization ILIKE ? OR clinic_hospital ILIKE ?", pattern, pattern, pattern, pattern)
	}

	err := query.Order("name ASC").Find(&doctors).Error
	return doctors, err
}

func (r *doctorRepository) FindByID(ctx context.Context, id string) (*domain.Doctor, error) {
	var doctor domain.Doctor
	err := r.db.WithContext(ctx).First(&doctor, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDoctorNotFound
		}
		return nil, err
	}
	return &doctor, nil
}

func (r *doctorRepository) Create(ctx context.Context, doctor *domain.Doctor) error {
	return r.db.WithContext(ctx).Create(doctor).Error
}

func (r *doctorRepository) Update(ctx context.Context, doctor *domain.Doctor) error {
	return r.db.WithContext(ctx).Save(doctor).Error
}

func (r *doctorRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Doctor{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrDoctorNotFound
	}
	return nil
}
