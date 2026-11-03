package usecase

import (
	"context"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type supplierUsecase struct {
	supplierRepo domain.SupplierRepository
}

func NewSupplierUsecase(supplierRepo domain.SupplierRepository) domain.SupplierUsecase {
	return &supplierUsecase{supplierRepo: supplierRepo}
}

func (u *supplierUsecase) GetAllSuppliers(ctx context.Context, search string, activeOnly bool) ([]domain.SupplierResponse, error) {
	suppliers, err := u.supplierRepo.FindAll(ctx, search, activeOnly)
	if err != nil {
		return nil, err
	}

	res := make([]domain.SupplierResponse, len(suppliers))
	for i, sup := range suppliers {
		res[i] = *sup.ToResponse()
	}
	return res, nil
}

func (u *supplierUsecase) GetSupplierByID(ctx context.Context, id string) (*domain.SupplierResponse, error) {
	sup, err := u.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return sup.ToResponse(), nil
}

func (u *supplierUsecase) CreateSupplier(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.SupplierResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	term := req.PaymentTermDays
	if term <= 0 {
		term = 30
	}

	sup := &domain.Supplier{
		ID:              utils.GenerateSupplierID(),
		Code:            req.Code,
		Name:            req.Name,
		Address:         req.Address,
		Phone:           req.Phone,
		Email:           req.Email,
		PICName:         req.PICName,
		PICPhone:        req.PICPhone,
		NPWP:            req.NPWP,
		PaymentTermDays: term,
		IsActive:        isActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := u.supplierRepo.Create(ctx, sup); err != nil {
		return nil, err
	}

	return sup.ToResponse(), nil
}

func (u *supplierUsecase) UpdateSupplier(ctx context.Context, id string, req *domain.UpdateSupplierRequest) (*domain.SupplierResponse, error) {
	sup, err := u.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sup.Code = req.Code
	sup.Name = req.Name
	sup.Address = req.Address
	sup.Phone = req.Phone
	sup.Email = req.Email
	sup.PICName = req.PICName
	sup.PICPhone = req.PICPhone
	sup.NPWP = req.NPWP
	sup.PaymentTermDays = req.PaymentTermDays
	if req.IsActive != nil {
		sup.IsActive = *req.IsActive
	}
	sup.UpdatedAt = time.Now()

	if err := u.supplierRepo.Update(ctx, sup); err != nil {
		return nil, err
	}

	return sup.ToResponse(), nil
}

func (u *supplierUsecase) DeleteSupplier(ctx context.Context, id string) error {
	return u.supplierRepo.Delete(ctx, id)
}
