package usecase

import (
	"context"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type categoryUsecase struct {
	categoryRepo domain.CategoryRepository
}

func NewCategoryUsecase(categoryRepo domain.CategoryRepository) domain.CategoryUsecase {
	return &categoryUsecase{categoryRepo: categoryRepo}
}

func (u *categoryUsecase) GetAllCategories(ctx context.Context, search string, activeOnly bool) ([]domain.CategoryResponse, error) {
	categories, err := u.categoryRepo.FindAll(ctx, search, activeOnly)
	if err != nil {
		return nil, err
	}

	res := make([]domain.CategoryResponse, len(categories))
	for i, cat := range categories {
		res[i] = *cat.ToResponse()
	}
	return res, nil
}

func (u *categoryUsecase) GetCategoryByID(ctx context.Context, id string) (*domain.CategoryResponse, error) {
	cat, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return cat.ToResponse(), nil
}

func (u *categoryUsecase) CreateCategory(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.CategoryResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	cat := &domain.Category{
		ID:          utils.GenerateCategoryID(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.categoryRepo.Create(ctx, cat); err != nil {
		return nil, err
	}

	return cat.ToResponse(), nil
}

func (u *categoryUsecase) UpdateCategory(ctx context.Context, id string, req *domain.UpdateCategoryRequest) (*domain.CategoryResponse, error) {
	cat, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cat.Name = req.Name
	cat.Description = req.Description
	if req.IsActive != nil {
		cat.IsActive = *req.IsActive
	}
	cat.UpdatedAt = time.Now()

	if err := u.categoryRepo.Update(ctx, cat); err != nil {
		return nil, err
	}

	return cat.ToResponse(), nil
}

func (u *categoryUsecase) DeleteCategory(ctx context.Context, id string) error {
	return u.categoryRepo.Delete(ctx, id)
}
