package domain

import (
	"context"
	"time"
)

// Supplier represents registered pharmaceutical distributors (Pedagang Besar Farmasi / PBF)
type Supplier struct {
	ID              string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	Code            string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	Address         string    `gorm:"type:text" json:"address"`
	Phone           string    `gorm:"type:varchar(50)" json:"phone"`
	Email           string    `gorm:"type:varchar(100)" json:"email"`
	PICName         string    `gorm:"type:varchar(150)" json:"pic_name"`
	PICPhone        string    `gorm:"type:varchar(50)" json:"pic_phone"`
	NPWP            string    `gorm:"type:varchar(100)" json:"npwp"`
	PaymentTermDays int       `gorm:"not null;default:30" json:"payment_term_days"` // e.g. 30 days, 21 days
	IsActive        bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CreateSupplierRequest struct {
	Code            string `json:"code" validate:"required,min=2,max=50"`
	Name            string `json:"name" validate:"required,min=2,max=255"`
	Address         string `json:"address"`
	Phone           string `json:"phone"`
	Email           string `json:"email" validate:"omitempty,email"`
	PICName         string `json:"pic_name"`
	PICPhone        string `json:"pic_phone"`
	NPWP            string `json:"npwp"`
	PaymentTermDays int    `json:"payment_term_days" validate:"gte=0"`
	IsActive        *bool  `json:"is_active"`
}

type UpdateSupplierRequest struct {
	Code            string `json:"code" validate:"required,min=2,max=50"`
	Name            string `json:"name" validate:"required,min=2,max=255"`
	Address         string `json:"address"`
	Phone           string `json:"phone"`
	Email           string `json:"email" validate:"omitempty,email"`
	PICName         string `json:"pic_name"`
	PICPhone        string `json:"pic_phone"`
	NPWP            string `json:"npwp"`
	PaymentTermDays int    `json:"payment_term_days" validate:"gte=0"`
	IsActive        *bool  `json:"is_active"`
}

type SupplierResponse struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	Phone           string    `json:"phone"`
	Email           string    `json:"email"`
	PICName         string    `json:"pic_name"`
	PICPhone        string    `json:"pic_phone"`
	NPWP            string    `json:"npwp"`
	PaymentTermDays int       `json:"payment_term_days"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *Supplier) ToResponse() *SupplierResponse {
	return &SupplierResponse{
		ID:              s.ID,
		Code:            s.Code,
		Name:            s.Name,
		Address:         s.Address,
		Phone:           s.Phone,
		Email:           s.Email,
		PICName:         s.PICName,
		PICPhone:        s.PICPhone,
		NPWP:            s.NPWP,
		PaymentTermDays: s.PaymentTermDays,
		IsActive:        s.IsActive,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

type SupplierRepository interface {
	FindAll(ctx context.Context, search string, activeOnly bool) ([]Supplier, error)
	FindByID(ctx context.Context, id string) (*Supplier, error)
	FindByCode(ctx context.Context, code string) (*Supplier, error)
	Create(ctx context.Context, supplier *Supplier) error
	Update(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id string) error
}

type SupplierUsecase interface {
	GetAllSuppliers(ctx context.Context, search string, activeOnly bool) ([]SupplierResponse, error)
	GetSupplierByID(ctx context.Context, id string) (*SupplierResponse, error)
	CreateSupplier(ctx context.Context, req *CreateSupplierRequest) (*SupplierResponse, error)
	UpdateSupplier(ctx context.Context, id string, req *UpdateSupplierRequest) (*SupplierResponse, error)
	DeleteSupplier(ctx context.Context, id string) error
}
