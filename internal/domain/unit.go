package domain

import (
	"context"
	"time"
)

// Unit represents measurement units for pharmaceutical items (e.g. Box, Strip, Tablet)
type Unit struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Symbol      string    `gorm:"type:varchar(20);not null" json:"symbol"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CreateUnitRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Symbol      string `json:"symbol" validate:"required,min=1,max=20"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateUnitRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Symbol      string `json:"symbol" validate:"required,min=1,max=20"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UnitResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u *Unit) ToResponse() *UnitResponse {
	return &UnitResponse{
		ID:          u.ID,
		Name:        u.Name,
		Symbol:      u.Symbol,
		Description: u.Description,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

type UnitRepository interface {
	FindAll(ctx context.Context, search string, activeOnly bool) ([]Unit, error)
	FindByID(ctx context.Context, id string) (*Unit, error)
	Create(ctx context.Context, unit *Unit) error
	Update(ctx context.Context, unit *Unit) error
	Delete(ctx context.Context, id string) error
}

type UnitUsecase interface {
	GetAllUnits(ctx context.Context, search string, activeOnly bool) ([]UnitResponse, error)
	GetUnitByID(ctx context.Context, id string) (*UnitResponse, error)
	CreateUnit(ctx context.Context, req *CreateUnitRequest) (*UnitResponse, error)
	UpdateUnit(ctx context.Context, id string, req *UpdateUnitRequest) (*UnitResponse, error)
	DeleteUnit(ctx context.Context, id string) error
}
