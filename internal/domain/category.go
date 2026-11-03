package domain

import (
	"context"
	"time"
)

// Category represents pharmacological or therapeutic categories (e.g. Antibiotik, Analgesik)
type Category struct {
	ID          string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(150);not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=150"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=150"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (c *Category) ToResponse() *CategoryResponse {
	return &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

type CategoryRepository interface {
	FindAll(ctx context.Context, search string, activeOnly bool) ([]Category, error)
	FindByID(ctx context.Context, id string) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

type CategoryUsecase interface {
	GetAllCategories(ctx context.Context, search string, activeOnly bool) ([]CategoryResponse, error)
	GetCategoryByID(ctx context.Context, id string) (*CategoryResponse, error)
	CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error)
	UpdateCategory(ctx context.Context, id string, req *UpdateCategoryRequest) (*CategoryResponse, error)
	DeleteCategory(ctx context.Context, id string) error
}
