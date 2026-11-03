package usecase

import (
	"context"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type unitUsecase struct {
	unitRepo domain.UnitRepository
}

func NewUnitUsecase(unitRepo domain.UnitRepository) domain.UnitUsecase {
	return &unitUsecase{unitRepo: unitRepo}
}

func (u *unitUsecase) GetAllUnits(ctx context.Context, search string, activeOnly bool) ([]domain.UnitResponse, error) {
	units, err := u.unitRepo.FindAll(ctx, search, activeOnly)
	if err != nil {
		return nil, err
	}

	res := make([]domain.UnitResponse, len(units))
	for i, unit := range units {
		res[i] = *unit.ToResponse()
	}
	return res, nil
}

func (u *unitUsecase) GetUnitByID(ctx context.Context, id string) (*domain.UnitResponse, error) {
	unit, err := u.unitRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return unit.ToResponse(), nil
}

func (u *unitUsecase) CreateUnit(ctx context.Context, req *domain.CreateUnitRequest) (*domain.UnitResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	unit := &domain.Unit{
		ID:          utils.GenerateUnitID(),
		Name:        req.Name,
		Symbol:      req.Symbol,
		Description: req.Description,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := u.unitRepo.Create(ctx, unit); err != nil {
		return nil, err
	}

	return unit.ToResponse(), nil
}

func (u *unitUsecase) UpdateUnit(ctx context.Context, id string, req *domain.UpdateUnitRequest) (*domain.UnitResponse, error) {
	unit, err := u.unitRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	unit.Name = req.Name
	unit.Symbol = req.Symbol
	unit.Description = req.Description
	if req.IsActive != nil {
		unit.IsActive = *req.IsActive
	}
	unit.UpdatedAt = time.Now()

	if err := u.unitRepo.Update(ctx, unit); err != nil {
		return nil, err
	}

	return unit.ToResponse(), nil
}

func (u *unitUsecase) DeleteUnit(ctx context.Context, id string) error {
	return u.unitRepo.Delete(ctx, id)
}
