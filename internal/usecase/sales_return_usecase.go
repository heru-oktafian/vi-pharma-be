package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type salesReturnUsecase struct {
	retRepo  domain.SalesReturnRepository
	posRepo  domain.POSRepository
	medRepo  domain.MedicineRepository
	unitRepo domain.UnitRepository
	invRepo  domain.InventoryRepository
}

func NewSalesReturnUsecase(
	retRepo domain.SalesReturnRepository,
	posRepo domain.POSRepository,
	medRepo domain.MedicineRepository,
	unitRepo domain.UnitRepository,
	invRepo domain.InventoryRepository,
) domain.SalesReturnUsecase {
	return &salesReturnUsecase{
		retRepo:  retRepo,
		posRepo:  posRepo,
		medRepo:  medRepo,
		unitRepo: unitRepo,
		invRepo:  invRepo,
	}
}

func (u *salesReturnUsecase) ProcessReturn(
	ctx context.Context,
	branchID, cashierID string,
	req domain.CreateSalesReturnRequest,
) (*domain.SalesReturn, error) {
	if req.SaleID == "" {
		return nil, fmt.Errorf("sale_id wajib ditentukan")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("minimal 1 item obat wajib diretur")
	}

	// 1. Fetch sale
	sale, err := u.posRepo.FindSaleByID(ctx, req.SaleID)
	if err != nil {
		return nil, domain.ErrSaleNotFound
	}
	if sale.BranchID != branchID {
		return nil, domain.ErrUnauthorizedBranch
	}

	// 2. Map existing sale items
	saleItemMap := make(map[string]domain.SaleItem)
	for _, it := range sale.Items {
		saleItemMap[it.ID] = it
	}

	// 3. Get already returned quantities for this sale
	alreadyReturned, err := u.retRepo.GetReturnedQtysBySaleID(ctx, req.SaleID)
	if err != nil {
		return nil, fmt.Errorf("failed to check previous returns: %w", err)
	}

	// 4. Validate and construct return items
	returnID := utils.GenerateSalesReturnID()
	var returnItems []domain.SalesReturnItem
	var totalRefundAmount float64

	for idx, input := range req.Items {
		if input.Qty <= 0 {
			return nil, fmt.Errorf("item ke-%d: jumlah retur harus lebih besar dari 0", idx+1)
		}

		saleItem, exists := saleItemMap[input.SaleItemID]
		if !exists {
			return nil, fmt.Errorf("item ke-%d: item penjualan tidak ditemukan pada struk ini", idx+1)
		}

		prevQty := alreadyReturned[input.SaleItemID]
		maxReturnable := saleItem.Qty - prevQty
		if input.Qty > (maxReturnable + 0.0001) {
			return nil, fmt.Errorf("item %s: jumlah retur (%.2f) melebihi sisa yang dapat diretur (%.2f)",
				saleItem.Medicine.Name, input.Qty, maxReturnable)
		}

		conv := saleItem.ConversionFactor
		if conv <= 0 {
			conv = 1
		}
		qtyBase := input.Qty * float64(conv)
		itemSubtotal := input.Qty * saleItem.UnitPrice
		totalRefundAmount += itemSubtotal

		condition := input.Condition
		if condition == "" {
			condition = domain.ReturnConditionGood
		}

		returnItems = append(returnItems, domain.SalesReturnItem{
			ID:               utils.GenerateSalesReturnItemID(),
			SalesReturnID:    returnID,
			SaleItemID:       saleItem.ID,
			MedicineID:       saleItem.MedicineID,
			BatchID:          saleItem.BatchID,
			UnitID:           saleItem.UnitID,
			ConversionFactor: conv,
			Qty:              input.Qty,
			QtyBase:          qtyBase,
			UnitPrice:        saleItem.UnitPrice,
			Subtotal:         itemSubtotal,
			Condition:        condition,
			ItemNotes:        input.ItemNotes,
		})
	}

	// 5. Generate Return Number
	returnNumber, err := u.retRepo.GenerateReturnNumber(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate return number: %w", err)
	}

	// 6. Parse return date
	retDate := time.Now()
	if req.ReturnDate != "" {
		if t, err := time.Parse("2006-01-02", req.ReturnDate); err == nil {
			retDate = t
		}
	}

	refundType := req.RefundType
	if refundType == "" {
		refundType = domain.RefundCash
	}

	salesReturn := &domain.SalesReturn{
		ID:            returnID,
		BranchID:      branchID,
		ReturnNumber:  returnNumber,
		SaleID:        sale.ID,
		InvoiceNumber: sale.InvoiceNumber,
		ReturnDate:    retDate,
		CustomerName:  sale.CustomerName,
		Reason:        req.Reason,
		RefundType:    refundType,
		TotalItems:    len(returnItems),
		TotalAmount:   totalRefundAmount,
		Notes:         req.Notes,
		CashierID:     cashierID,
		Status:        domain.SalesReturnStatusCompleted,
		Items:         returnItems,
	}

	// 7. Save to DB with stock restoration
	if err := u.retRepo.CreateSalesReturn(ctx, salesReturn); err != nil {
		return nil, fmt.Errorf("failed to process customer sales return: %w", err)
	}

	return u.retRepo.FindSalesReturnByID(ctx, returnID)
}

func (u *salesReturnUsecase) GetSalesReturns(ctx context.Context, branchID string, filter domain.SalesReturnFilter) (*domain.PaginatedSalesReturns, error) {
	data, total, err := u.retRepo.FindSalesReturns(ctx, branchID, filter)
	if err != nil {
		return nil, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &domain.PaginatedSalesReturns{
		Data:       data,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *salesReturnUsecase) GetSalesReturnByID(ctx context.Context, id string) (*domain.SalesReturn, error) {
	return u.retRepo.FindSalesReturnByID(ctx, id)
}

func (u *salesReturnUsecase) GetSaleReturnableInfo(ctx context.Context, branchID, saleID string) (*domain.SaleReturnableResponse, error) {
	sale, err := u.posRepo.FindSaleByID(ctx, saleID)
	if err != nil {
		return nil, domain.ErrSaleNotFound
	}
	if sale.BranchID != branchID {
		return nil, domain.ErrUnauthorizedBranch
	}

	returnedQtys, err := u.retRepo.GetReturnedQtysBySaleID(ctx, saleID)
	if err != nil {
		return nil, err
	}

	remainingQtys := make(map[string]float64)
	isFullyReturned := true

	for _, it := range sale.Items {
		retQty := returnedQtys[it.ID]
		rem := math.Max(0, it.Qty-retQty)
		remainingQtys[it.ID] = rem
		if rem > 0 {
			isFullyReturned = false
		}
	}

	return &domain.SaleReturnableResponse{
		Sale:            sale,
		ReturnedQtys:    returnedQtys,
		RemainingQtys:   remainingQtys,
		IsFullyReturned: isFullyReturned,
	}, nil
}
