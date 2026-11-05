package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type posUsecase struct {
	posRepo domain.POSRepository
	db      *gorm.DB
}

func NewPOSUsecase(posRepo domain.POSRepository, db *gorm.DB) domain.POSUsecase {
	return &posUsecase{
		posRepo: posRepo,
		db:      db,
	}
}

func (u *posUsecase) Checkout(ctx context.Context, branchID, cashierID string, req domain.CreateSaleRequest) (*domain.Sale, error) {
	if len(req.Items) == 0 {
		return nil, domain.ErrCartEmpty
	}

	saleID := utils.GenerateSaleID()
	invoiceNum, err := u.posRepo.GenerateInvoiceNumber(ctx, branchID)
	if err != nil {
		invoiceNum = fmt.Sprintf("POS/%s/%s", time.Now().Format("20060102"), saleID[len(saleID)-4:])
	}

	var subtotal float64
	saleItems := make([]domain.SaleItem, len(req.Items))

	for i, it := range req.Items {
		if it.Qty <= 0 {
			return nil, fmt.Errorf("jumlah beli item #%d harus lebih dari 0", i+1)
		}

		conv := it.ConversionFactor
		if conv <= 0 {
			conv = 1
		}
		qtyBase := it.Qty * float64(conv)
		itemSubtotal := it.Qty * it.UnitPrice
		itemTotal := itemSubtotal - it.Discount
		if itemTotal < 0 {
			itemTotal = 0
		}
		subtotal += itemTotal

		// Auto FEFO batch selection if not specified by cashier
		batchID := it.BatchID
		if batchID == "" {
			var batch domain.MedicineBatch
			err := u.db.WithContext(ctx).
				Where("medicine_id = ? AND branch_id = ? AND is_active = ? AND qty_shelf >= ?",
					it.MedicineID, branchID, true, qtyBase).
				Order("expired_date ASC").
				First(&batch).Error
			if err != nil {
				// Fallback: pick any active batch with shelf stock > 0
				err = u.db.WithContext(ctx).
					Where("medicine_id = ? AND branch_id = ? AND is_active = ? AND qty_shelf > 0",
						it.MedicineID, branchID, true).
					Order("expired_date ASC").
					First(&batch).Error
				if err != nil {
					return nil, fmt.Errorf("obat %s tidak memiliki stok yang cukup di rak: %w", it.MedicineID, domain.ErrInsufficientShelfStock)
				}
			}
			batchID = batch.ID
		}

		saleItems[i] = domain.SaleItem{
			ID:               utils.GenerateSaleItemID(),
			SaleID:           saleID,
			MedicineID:       it.MedicineID,
			BatchID:          batchID,
			UnitID:           it.UnitID,
			ConversionFactor: conv,
			Qty:              it.Qty,
			QtyBase:          qtyBase,
			UnitPrice:        it.UnitPrice,
			Subtotal:         itemSubtotal,
			Discount:         it.Discount,
			Total:            itemTotal,
		}
	}

	totalAmount := subtotal - req.DiscountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	// Validate Cashier payment
	amountPaid := req.AmountPaid
	changeAmount := 0.0

	if req.PaymentMethod == domain.PaymentMethodTunai {
		if amountPaid < totalAmount {
			return nil, domain.ErrInvalidPaymentAmount
		}
		changeAmount = amountPaid - totalAmount
	} else {
		// Non-cash (QRIS, Debit, Transfer) is exact amount
		if amountPaid <= 0 {
			amountPaid = totalAmount
		}
		changeAmount = 0.0
	}

	customerName := req.CustomerName
	if customerName == "" {
		customerName = "Pelanggan Umum"
	}

	txType := req.TransactionType
	if txType == "" {
		txType = domain.SaleTypeOTC
	}

	sale := &domain.Sale{
		ID:              saleID,
		BranchID:        branchID,
		InvoiceNumber:   invoiceNum,
		CashierID:       cashierID,
		CustomerName:    customerName,
		CustomerPhone:   req.CustomerPhone,
		TransactionType: txType,
		Subtotal:        subtotal,
		DiscountAmount:  req.DiscountAmount,
		TaxAmount:       0,
		TotalAmount:     totalAmount,
		PaymentMethod:   req.PaymentMethod,
		AmountPaid:      amountPaid,
		ChangeAmount:    changeAmount,
		PaymentStatus:   domain.PaymentStatusPaid,
		Notes:           req.Notes,
		Items:           saleItems,
	}

	if err := u.posRepo.CreateSale(ctx, sale); err != nil {
		return nil, err
	}

	return u.posRepo.FindSaleByID(ctx, sale.ID)
}

func (u *posUsecase) GetSales(ctx context.Context, branchID string, filter domain.SaleFilter) (*domain.PaginatedSales, error) {
	sales, total, err := u.posRepo.FindSales(ctx, branchID, filter)
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
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedSales{
		Data:       sales,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *posUsecase) GetSaleByID(ctx context.Context, id string) (*domain.Sale, error) {
	return u.posRepo.FindSaleByID(ctx, id)
}

func (u *posUsecase) GetTodaySummary(ctx context.Context, branchID string) (*domain.DailyCashierSummaryResponse, error) {
	return u.posRepo.GetTodaySummary(ctx, branchID)
}
