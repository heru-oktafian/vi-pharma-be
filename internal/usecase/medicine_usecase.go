package usecase

import (
	"context"
	"math"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type medicineUsecase struct {
	medRepo      domain.MedicineRepository
	categoryRepo domain.CategoryRepository
	unitRepo     domain.UnitRepository
}

func NewMedicineUsecase(
	medRepo domain.MedicineRepository,
	categoryRepo domain.CategoryRepository,
	unitRepo domain.UnitRepository,
) domain.MedicineUsecase {
	return &medicineUsecase{
		medRepo:      medRepo,
		categoryRepo: categoryRepo,
		unitRepo:     unitRepo,
	}
}

func (u *medicineUsecase) GetAllMedicines(ctx context.Context, filter domain.MedicineFilter) (*domain.PaginatedMedicines, error) {
	medicines, totalItems, err := u.medRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	data := make([]domain.MedicineResponse, len(medicines))
	for i, m := range medicines {
		data[i] = *m.ToResponse()
	}

	return &domain.PaginatedMedicines{
		Data:       data,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *medicineUsecase) GetMedicineByID(ctx context.Context, id string) (*domain.MedicineResponse, error) {
	medicine, err := u.medRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return medicine.ToResponse(), nil
}

func (u *medicineUsecase) GetMedicineByBarcode(ctx context.Context, barcode string) (*domain.MedicineResponse, error) {
	medicine, err := u.medRepo.FindByBarcode(ctx, barcode)
	if err != nil {
		return nil, err
	}
	return medicine.ToResponse(), nil
}

func (u *medicineUsecase) CreateMedicine(ctx context.Context, req *domain.CreateMedicineRequest) (*domain.MedicineResponse, error) {
	// 1. Verify Category exists
	if _, err := u.categoryRepo.FindByID(ctx, req.CategoryID); err != nil {
		return nil, domain.ErrCategoryNotFound
	}

	// 2. Verify Base Unit exists
	if _, err := u.unitRepo.FindByID(ctx, req.BaseUnitID); err != nil {
		return nil, domain.ErrUnitNotFound
	}

	// 3. Verify ItemCode uniqueness
	existing, _ := u.medRepo.FindByItemCode(ctx, req.ItemCode)
	if existing != nil {
		return nil, domain.ErrItemCodeAlreadyExists
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	labelType := req.LabelType
	if labelType == "" {
		labelType = "dalam"
	}

	doseMorning := req.DoseMorning
	if doseMorning == "" {
		doseMorning = "1"
	}
	doseNoon := req.DoseNoon
	if doseNoon == "" {
		doseNoon = "1"
	}
	doseNight := req.DoseNight
	if doseNight == "" {
		doseNight = "1"
	}

	medicineID := utils.GenerateMedicineID()
	medicine := &domain.Medicine{
		ID:                   medicineID,
		ItemCode:             req.ItemCode,
		Barcode:              req.Barcode,
		Name:                 req.Name,
		GenericName:          req.GenericName,
		NIE:                  req.NIE,
		Manufacturer:         req.Manufacturer,
		DosageForm:           req.DosageForm,
		DrugClassification:   req.DrugClassification,
		CategoryID:           req.CategoryID,
		BaseUnitID:           req.BaseUnitID,
		RequiresPrescription: req.RequiresPrescription,
		MinStock:             req.MinStock,
		MaxStock:             req.MaxStock,
		StorageLocation:      req.StorageLocation,
		DefaultSupplierID:    req.DefaultSupplierID,
		IsActive:             isActive,
		LabelType:            labelType,
		DefaultSigna:         req.DefaultSigna,
		ConsumptionTime:      req.ConsumptionTime,
		DoseMorning:          doseMorning,
		DoseNoon:             doseNoon,
		DoseNight:            doseNight,
		WarningTags:          req.WarningTags,
		SpecialInstructions:  req.SpecialInstructions,
		BUDDays:              req.BUDDays,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// 4. Prepare Units
	units := make([]domain.MedicineUnit, 0)
	hasBaseUnit := false
	for _, uInput := range req.Units {
		if uInput.IsBaseUnit || uInput.ConversionFactor == 1 {
			hasBaseUnit = true
		}
		units = append(units, domain.MedicineUnit{
			UnitID:           uInput.UnitID,
			ConversionFactor: uInput.ConversionFactor,
			IsBaseUnit:       uInput.IsBaseUnit,
			Barcode:          uInput.Barcode,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}

	// Auto ensure Base Unit is in units array if omitted
	if !hasBaseUnit {
		units = append(units, domain.MedicineUnit{
			UnitID:           req.BaseUnitID,
			ConversionFactor: 1,
			IsBaseUnit:       true,
			Barcode:          req.Barcode,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}

	// 5. Prepare Prices
	prices := make([]domain.MedicinePrice, 0)
	for _, pInput := range req.Prices {
		prices = append(prices, domain.MedicinePrice{
			UnitID:    pInput.UnitID,
			PriceType: pInput.PriceType,
			Price:     pInput.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	if err := u.medRepo.Create(ctx, medicine, units, prices); err != nil {
		return nil, err
	}

	// Fetch full populated record
	created, err := u.medRepo.FindByID(ctx, medicineID)
	if err != nil {
		return medicine.ToResponse(), nil
	}
	return created.ToResponse(), nil
}

func (u *medicineUsecase) UpdateMedicine(ctx context.Context, id string, req *domain.UpdateMedicineRequest) (*domain.MedicineResponse, error) {
	medicine, err := u.medRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if item_code changed and conflicts
	if medicine.ItemCode != req.ItemCode {
		existing, _ := u.medRepo.FindByItemCode(ctx, req.ItemCode)
		if existing != nil && existing.ID != id {
			return nil, domain.ErrItemCodeAlreadyExists
		}
	}

	medicine.ItemCode = req.ItemCode
	medicine.Barcode = req.Barcode
	medicine.Name = req.Name
	medicine.GenericName = req.GenericName
	medicine.NIE = req.NIE
	medicine.Manufacturer = req.Manufacturer
	medicine.DosageForm = req.DosageForm
	medicine.DrugClassification = req.DrugClassification
	medicine.CategoryID = req.CategoryID
	medicine.BaseUnitID = req.BaseUnitID
	medicine.RequiresPrescription = req.RequiresPrescription
	medicine.MinStock = req.MinStock
	medicine.MaxStock = req.MaxStock
	medicine.StorageLocation = req.StorageLocation
	medicine.DefaultSupplierID = req.DefaultSupplierID
	if req.IsActive != nil {
		medicine.IsActive = *req.IsActive
	}
	if req.LabelType != "" {
		medicine.LabelType = req.LabelType
	}
	medicine.DefaultSigna = req.DefaultSigna
	medicine.ConsumptionTime = req.ConsumptionTime
	medicine.DoseMorning = req.DoseMorning
	medicine.DoseNoon = req.DoseNoon
	medicine.DoseNight = req.DoseNight
	medicine.WarningTags = req.WarningTags
	medicine.SpecialInstructions = req.SpecialInstructions
	medicine.BUDDays = req.BUDDays
	medicine.UpdatedAt = time.Now()

	units := make([]domain.MedicineUnit, 0)
	for _, uInput := range req.Units {
		units = append(units, domain.MedicineUnit{
			UnitID:           uInput.UnitID,
			ConversionFactor: uInput.ConversionFactor,
			IsBaseUnit:       uInput.IsBaseUnit,
			Barcode:          uInput.Barcode,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}

	prices := make([]domain.MedicinePrice, 0)
	for _, pInput := range req.Prices {
		prices = append(prices, domain.MedicinePrice{
			UnitID:    pInput.UnitID,
			PriceType: pInput.PriceType,
			Price:     pInput.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	if err := u.medRepo.Update(ctx, medicine, units, prices); err != nil {
		return nil, err
	}

	updated, err := u.medRepo.FindByID(ctx, id)
	if err != nil {
		return medicine.ToResponse(), nil
	}
	return updated.ToResponse(), nil
}

func (u *medicineUsecase) DeleteMedicine(ctx context.Context, id string) error {
	return u.medRepo.Delete(ctx, id)
}
