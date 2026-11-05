package usecase

import (
	"context"
	"math"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type procurementUsecase struct {
	procRepo domain.ProcurementRepository
	supRepo  domain.SupplierRepository
	medRepo  domain.MedicineRepository
	unitRepo domain.UnitRepository
}

func NewProcurementUsecase(
	procRepo domain.ProcurementRepository,
	supRepo domain.SupplierRepository,
	medRepo domain.MedicineRepository,
	unitRepo domain.UnitRepository,
) domain.ProcurementUsecase {
	return &procurementUsecase{
		procRepo: procRepo,
		supRepo:  supRepo,
		medRepo:  medRepo,
		unitRepo: unitRepo,
	}
}

// resolveUnitID safely determines a valid unit ID from the input, the medicine, or the database
func (u *procurementUsecase) resolveUnitID(ctx context.Context, medicineID, rawUnitID string) string {
	// 1. Direct match by Unit ID if provided
	if rawUnitID != "" {
		if uObj, err := u.unitRepo.FindByID(ctx, rawUnitID); err == nil && uObj != nil {
			return uObj.ID
		}
		// 2. Search by name or symbol (e.g. "Box", "Pcs", "Botol", "Tablet")
		if units, err := u.unitRepo.FindAll(ctx, rawUnitID, true); err == nil && len(units) > 0 {
			return units[0].ID
		}
	}

	// 3. Fallback to medicine's base_unit_id
	if medicineID != "" {
		if med, err := u.medRepo.FindByID(ctx, medicineID); err == nil && med != nil && med.BaseUnitID != "" {
			return med.BaseUnitID
		}
	}

	// 4. Default to first active unit in database
	if allUnits, err := u.unitRepo.FindAll(ctx, "", true); err == nil && len(allUnits) > 0 {
		return allUnits[0].ID
	}

	return "UNT000000000001"
}

func (u *procurementUsecase) CreatePurchaseOrder(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreatePurchaseOrderRequest,
) (*domain.PurchaseOrder, error) {
	// 1. Regulatory rule: Narkotika Surat Pesanan CAN ONLY contain 1 item
	if req.SPType == domain.SPTypeNarcotic && len(req.Items) > 1 {
		return nil, domain.ErrNarcoticSingleItemOnly
	}

	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		orderDate = time.Now()
	}

	var estDelivery *time.Time
	if req.EstimatedDeliveryDate != "" {
		if d, err := time.Parse("2006-01-02", req.EstimatedDeliveryDate); err == nil {
			estDelivery = &d
		}
	}

	poNumber, err := u.procRepo.GetNextPONumber(ctx, branchID, req.SPType)
	if err != nil {
		return nil, err
	}

	poID := utils.GeneratePurchaseOrderID()
	var items []domain.PurchaseOrderItem
	var totalEstimated float64

	for _, it := range req.Items {
		validUnitID := u.resolveUnitID(ctx, it.MedicineID, it.UnitID)
		itemSubtotal := it.QtyOrdered * it.EstimatedPrice
		totalEstimated += itemSubtotal

		items = append(items, domain.PurchaseOrderItem{
			ID:              utils.GeneratePurchaseOrderItemID(),
			PurchaseOrderID: poID,
			MedicineID:      it.MedicineID,
			UnitID:          validUnitID,
			QtyOrdered:      it.QtyOrdered,
			QtyReceived:     0,
			EstimatedPrice:  it.EstimatedPrice,
			Subtotal:        itemSubtotal,
			Notes:           it.Notes,
		})
	}

	po := &domain.PurchaseOrder{
		ID:                    poID,
		BranchID:              branchID,
		PONumber:              poNumber,
		OrderDate:             orderDate,
		EstimatedDeliveryDate: estDelivery,
		SupplierID:            req.SupplierID,
		SPType:                req.SPType,
		Status:                domain.POStatusDraft, // Start as DRAFT
		PharmacistName:        req.PharmacistName,
		PharmacistSIPA:        req.PharmacistSIPA,
		TotalItems:            len(items),
		TotalEstimatedAmount:  totalEstimated,
		Notes:                 req.Notes,
		CreatedByID:           userID,
		Items:                 items,
	}

	if err := u.procRepo.CreatePurchaseOrder(ctx, po); err != nil {
		return nil, err
	}

	return u.procRepo.GetPurchaseOrderByID(ctx, poID)
}

func (u *procurementUsecase) GetPurchaseOrderByID(ctx context.Context, id string) (*domain.PurchaseOrder, error) {
	return u.procRepo.GetPurchaseOrderByID(ctx, id)
}

func (u *procurementUsecase) ListPurchaseOrders(
	ctx context.Context,
	branchID string,
	filter domain.POFilter,
) (*domain.PaginatedPurchaseOrders, error) {
	orders, total, err := u.procRepo.FindAllPurchaseOrders(ctx, branchID, filter)
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

	return &domain.PaginatedPurchaseOrders{
		Data:       orders,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *procurementUsecase) UpdatePurchaseOrder(
	ctx context.Context,
	branchID string,
	id string,
	req *domain.UpdatePurchaseOrderRequest,
) (*domain.PurchaseOrder, error) {
	po, err := u.procRepo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.BranchID != branchID {
		return nil, domain.ErrPurchaseOrderNotFound
	}
	if po.Status != domain.POStatusDraft {
		return nil, domain.ErrPOAlreadyProcessed
	}

	spType := po.SPType
	if req.SPType != "" {
		spType = req.SPType
	}

	if spType == domain.SPTypeNarcotic && len(req.Items) > 1 {
		return nil, domain.ErrNarcoticSingleItemOnly
	}

	if req.SupplierID != "" {
		po.SupplierID = req.SupplierID
	}
	if req.OrderDate != "" {
		if d, err := time.Parse("2006-01-02", req.OrderDate); err == nil {
			po.OrderDate = d
		}
	}
	if req.EstimatedDeliveryDate != "" {
		if d, err := time.Parse("2006-01-02", req.EstimatedDeliveryDate); err == nil {
			po.EstimatedDeliveryDate = &d
		}
	}
	if req.SPType != "" {
		po.SPType = req.SPType
	}
	if req.PharmacistName != "" {
		po.PharmacistName = req.PharmacistName
	}
	if req.PharmacistSIPA != "" {
		po.PharmacistSIPA = req.PharmacistSIPA
	}
	if req.Notes != "" {
		po.Notes = req.Notes
	}

	var items []domain.PurchaseOrderItem
	if len(req.Items) > 0 {
		var totalEstimated float64
		for _, it := range req.Items {
			validUnitID := u.resolveUnitID(ctx, it.MedicineID, it.UnitID)
			subtotal := it.QtyOrdered * it.EstimatedPrice
			totalEstimated += subtotal
			items = append(items, domain.PurchaseOrderItem{
				ID:              utils.GeneratePurchaseOrderItemID(),
				PurchaseOrderID: po.ID,
				MedicineID:      it.MedicineID,
				UnitID:          validUnitID,
				QtyOrdered:      it.QtyOrdered,
				QtyReceived:     0,
				EstimatedPrice:  it.EstimatedPrice,
				Subtotal:        subtotal,
				Notes:           it.Notes,
			})
		}
		po.TotalItems = len(items)
		po.TotalEstimatedAmount = totalEstimated
	}

	if err := u.procRepo.UpdatePurchaseOrder(ctx, po, items); err != nil {
		return nil, err
	}

	return u.procRepo.GetPurchaseOrderByID(ctx, id)
}

func (u *procurementUsecase) UpdatePOStatus(
	ctx context.Context,
	branchID string,
	id string,
	req *domain.UpdatePOStatusRequest,
) (*domain.PurchaseOrder, error) {
	po, err := u.procRepo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.BranchID != branchID {
		return nil, domain.ErrPurchaseOrderNotFound
	}

	if err := u.procRepo.UpdatePOStatus(ctx, id, req.Status); err != nil {
		return nil, err
	}

	return u.procRepo.GetPurchaseOrderByID(ctx, id)
}

func (u *procurementUsecase) DeletePurchaseOrder(ctx context.Context, branchID string, id string) error {
	po, err := u.procRepo.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return err
	}
	if po.BranchID != branchID {
		return domain.ErrPurchaseOrderNotFound
	}
	if po.Status != domain.POStatusDraft {
		return domain.ErrPOAlreadyProcessed
	}

	return u.procRepo.DeletePurchaseOrder(ctx, id)
}
