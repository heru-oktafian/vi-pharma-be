package usecase

import (
	"context"
	"math"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type purchaseReturnUsecase struct {
	retRepo  domain.PurchaseReturnRepository
	supRepo  domain.SupplierRepository
	medRepo  domain.MedicineRepository
	unitRepo domain.UnitRepository
	invRepo  domain.InventoryRepository
}

func NewPurchaseReturnUsecase(
	retRepo domain.PurchaseReturnRepository,
	supRepo domain.SupplierRepository,
	medRepo domain.MedicineRepository,
	unitRepo domain.UnitRepository,
	invRepo domain.InventoryRepository,
) domain.PurchaseReturnUsecase {
	return &purchaseReturnUsecase{
		retRepo:  retRepo,
		supRepo:  supRepo,
		medRepo:  medRepo,
		unitRepo: unitRepo,
		invRepo:  invRepo,
	}
}

func (u *purchaseReturnUsecase) resolveUnitID(ctx context.Context, medicineID, rawUnitID string) string {
	if rawUnitID != "" {
		if uObj, err := u.unitRepo.FindByID(ctx, rawUnitID); err == nil && uObj != nil {
			return uObj.ID
		}
		if units, err := u.unitRepo.FindAll(ctx, rawUnitID, true); err == nil && len(units) > 0 {
			return units[0].ID
		}
	}

	if medicineID != "" {
		if med, err := u.medRepo.FindByID(ctx, medicineID); err == nil && med != nil && med.BaseUnitID != "" {
			return med.BaseUnitID
		}
	}

	if allUnits, err := u.unitRepo.FindAll(ctx, "", true); err == nil && len(allUnits) > 0 {
		return allUnits[0].ID
	}

	return "UNT000000000001"
}

func (u *purchaseReturnUsecase) CreatePurchaseReturn(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreatePurchaseReturnRequest,
) (*domain.PurchaseReturn, error) {
	if req.SupplierID != "" {
		if _, err := u.supRepo.FindByID(ctx, req.SupplierID); err != nil {
			return nil, domain.ErrSupplierNotFound
		}
	}

	returnDate, err := time.Parse("2006-01-02", req.ReturnDate)
	if err != nil {
		returnDate = time.Now()
	}

	returnNumber, err := u.retRepo.GetNextReturnNumber(ctx, branchID)
	if err != nil {
		return nil, err
	}

	retID := utils.GeneratePurchaseReturnID()
	var items []domain.PurchaseReturnItem
	var totalAmount float64

	for _, it := range req.Items {
		validUnitID := u.resolveUnitID(ctx, it.MedicineID, it.UnitID)

		conv := it.ConversionFactor
		if conv <= 0 {
			conv = 1
		}
		qtyBase := it.Qty * conv

		// Verify batch stock availability
		batch, err := u.invRepo.FindBatchByID(ctx, it.BatchID)
		if err != nil {
			return nil, domain.ErrBatchNotFound
		}

		loc := it.Location
		if loc == "" {
			loc = domain.LocationWarehouse
		}

		if loc == domain.LocationShelf {
			if batch.QtyShelf < qtyBase {
				return nil, domain.ErrInsufficientShelfStock
			}
		} else {
			if batch.QtyWarehouse < qtyBase {
				return nil, domain.ErrInsufficientWarehouseStock
			}
		}

		subtotal := it.Qty * it.UnitPrice
		totalAmount += subtotal

		items = append(items, domain.PurchaseReturnItem{
			ID:               utils.GeneratePurchaseReturnItemID(),
			PurchaseReturnID: retID,
			MedicineID:       it.MedicineID,
			BatchID:          it.BatchID,
			UnitID:           validUnitID,
			Location:         loc,
			ConversionFactor: conv,
			Qty:              it.Qty,
			QtyBaseUnit:      qtyBase,
			UnitPrice:        it.UnitPrice,
			Subtotal:         subtotal,
			ItemReason:       it.ItemReason,
		})
	}

	ret := &domain.PurchaseReturn{
		ID:               retID,
		BranchID:         branchID,
		ReturnNumber:     returnNumber,
		ReturnDate:       returnDate,
		SupplierID:       req.SupplierID,
		InvoiceNumber:    req.InvoiceNumber,
		Status:           domain.ReturnStatusDraft, // starts as DRAFT
		Reason:           req.Reason,
		CompensationType: req.CompensationType,
		TotalItems:       len(items),
		TotalAmount:      totalAmount,
		Notes:            req.Notes,
		CreatedByID:      userID,
		Items:            items,
	}

	if err := u.retRepo.CreatePurchaseReturn(ctx, ret); err != nil {
		return nil, err
	}

	return u.retRepo.GetPurchaseReturnByID(ctx, retID)
}

func (u *purchaseReturnUsecase) GetPurchaseReturnByID(ctx context.Context, id string) (*domain.PurchaseReturn, error) {
	return u.retRepo.GetPurchaseReturnByID(ctx, id)
}

func (u *purchaseReturnUsecase) ListPurchaseReturns(
	ctx context.Context,
	branchID string,
	filter domain.PurchaseReturnFilter,
) (*domain.PaginatedPurchaseReturns, error) {
	returns, total, err := u.retRepo.FindAllPurchaseReturns(ctx, branchID, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedPurchaseReturns{
		Data:       returns,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *purchaseReturnUsecase) UpdatePurchaseReturn(
	ctx context.Context,
	branchID string,
	id string,
	req *domain.UpdatePurchaseReturnRequest,
) (*domain.PurchaseReturn, error) {
	ret, err := u.retRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ret.BranchID != branchID {
		return nil, domain.ErrPurchaseReturnNotFound
	}
	if ret.Status != domain.ReturnStatusDraft {
		return nil, domain.ErrPurchaseReturnAlreadyProcessed
	}

	if req.SupplierID != "" {
		ret.SupplierID = req.SupplierID
	}
	if req.ReturnDate != "" {
		if d, err := time.Parse("2006-01-02", req.ReturnDate); err == nil {
			ret.ReturnDate = d
		}
	}
	if req.InvoiceNumber != "" {
		ret.InvoiceNumber = req.InvoiceNumber
	}
	if req.Reason != "" {
		ret.Reason = req.Reason
	}
	if req.CompensationType != "" {
		ret.CompensationType = req.CompensationType
	}
	if req.Notes != "" {
		ret.Notes = req.Notes
	}

	var items []domain.PurchaseReturnItem
	if len(req.Items) > 0 {
		var totalAmount float64
		for _, it := range req.Items {
			validUnitID := u.resolveUnitID(ctx, it.MedicineID, it.UnitID)
			conv := it.ConversionFactor
			if conv <= 0 {
				conv = 1
			}
			qtyBase := it.Qty * conv

			loc := it.Location
			if loc == "" {
				loc = domain.LocationWarehouse
			}

			subtotal := it.Qty * it.UnitPrice
			totalAmount += subtotal

			items = append(items, domain.PurchaseReturnItem{
				ID:               utils.GeneratePurchaseReturnItemID(),
				PurchaseReturnID: ret.ID,
				MedicineID:       it.MedicineID,
				BatchID:          it.BatchID,
				UnitID:           validUnitID,
				Location:         loc,
				ConversionFactor: conv,
				Qty:              it.Qty,
				QtyBaseUnit:      qtyBase,
				UnitPrice:        it.UnitPrice,
				Subtotal:         subtotal,
				ItemReason:       it.ItemReason,
			})
		}
		ret.TotalItems = len(items)
		ret.TotalAmount = totalAmount
	}

	if err := u.retRepo.UpdatePurchaseReturn(ctx, ret, items); err != nil {
		return nil, err
	}

	return u.retRepo.GetPurchaseReturnByID(ctx, id)
}

func (u *purchaseReturnUsecase) UpdateReturnStatus(
	ctx context.Context,
	branchID, userID string,
	id string,
	req *domain.UpdateReturnStatusRequest,
) (*domain.PurchaseReturn, error) {
	ret, err := u.retRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ret.BranchID != branchID {
		return nil, domain.ErrPurchaseReturnNotFound
	}

	if ret.Status == domain.ReturnStatusCompleted {
		return nil, domain.ErrPurchaseReturnAlreadyProcessed
	}

	if req.Status == domain.ReturnStatusCompleted {
		// Finalize return & deduct physical stock with audit mutations
		if err := u.retRepo.ProcessStockDeduction(ctx, ret, userID); err != nil {
			return nil, err
		}
		return u.retRepo.GetPurchaseReturnByID(ctx, id)
	}

	// For REQUESTED, DRAFT, or CANCELLED
	if err := u.retRepo.UpdateReturnStatus(ctx, id, req.Status, req.Notes); err != nil {
		return nil, err
	}

	return u.retRepo.GetPurchaseReturnByID(ctx, id)
}

func (u *purchaseReturnUsecase) DeletePurchaseReturn(ctx context.Context, branchID, id string) error {
	ret, err := u.retRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return err
	}
	if ret.BranchID != branchID {
		return domain.ErrPurchaseReturnNotFound
	}
	if ret.Status != domain.ReturnStatusDraft {
		return domain.ErrPurchaseReturnAlreadyProcessed
	}
	return u.retRepo.DeletePurchaseReturn(ctx, id)
}
