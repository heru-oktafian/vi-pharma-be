package usecase

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type inventoryUsecase struct {
	invRepo domain.InventoryRepository
	medRepo domain.MedicineRepository
}

func NewInventoryUsecase(
	invRepo domain.InventoryRepository,
	medRepo domain.MedicineRepository,
) domain.InventoryUsecase {
	return &inventoryUsecase{
		invRepo: invRepo,
		medRepo: medRepo,
	}
}

// -------------------------------------------------------------
// Purchase Invoices (Penerimaan Faktur PBF)
// -------------------------------------------------------------

func (u *inventoryUsecase) ReceivePurchaseInvoice(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreatePurchaseInvoiceRequest,
) (*domain.PurchaseInvoice, error) {
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		invoiceDate = time.Now()
	}

	var dueDate time.Time
	if req.DueDate != "" {
		if d, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			dueDate = d
		}
	}

	invoiceID := utils.GeneratePurchaseInvoiceID()
	var subtotal float64
	var items []domain.PurchaseInvoiceItem

	for _, itemInput := range req.Items {
		expDate, err := time.Parse("2006-01-02", itemInput.ExpiredDate)
		if err != nil {
			expDate = time.Now().AddDate(1, 0, 0) // default 1 year if invalid
		}

		conv := itemInput.ConversionFactor
		if conv <= 0 {
			conv = 1
		}
		qtyBase := itemInput.QtyPurchased * conv
		discountFactor := 1.0 - (itemInput.DiscountPercent / 100.0)
		if discountFactor < 0 {
			discountFactor = 0
		}
		itemSubtotal := itemInput.QtyPurchased * itemInput.BuyPrice * discountFactor
		subtotal += itemSubtotal

		// Unit buy price per base unit
		hppPerBase := (itemInput.BuyPrice / conv) * discountFactor

		// Batch Lookup or Creation
		batch, err := u.invRepo.FindBatchByNumberAndMedicine(ctx, branchID, itemInput.MedicineID, itemInput.BatchNumber)
		if err != nil {
			// Create new batch
			batch = &domain.MedicineBatch{
				ID:          utils.GenerateBatchID(),
				MedicineID:  itemInput.MedicineID,
				BranchID:    branchID,
				BatchNumber: itemInput.BatchNumber,
				ExpiredDate: expDate,
				BuyPrice:    hppPerBase,
				IsActive:    true,
			}
			if itemInput.PutToShelfDirectly {
				batch.QtyShelf = qtyBase
				batch.QtyWarehouse = 0
			} else {
				batch.QtyWarehouse = qtyBase
				batch.QtyShelf = 0
			}
			batch.CalculateTotal()
			if err := u.invRepo.CreateBatch(ctx, batch); err != nil {
				return nil, fmt.Errorf("failed to create batch: %w", err)
			}
		} else {
			// Update existing batch
			if itemInput.PutToShelfDirectly {
				batch.QtyShelf += qtyBase
			} else {
				batch.QtyWarehouse += qtyBase
			}
			batch.BuyPrice = hppPerBase
			batch.ExpiredDate = expDate
			batch.CalculateTotal()
			if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
				return nil, fmt.Errorf("failed to update batch: %w", err)
			}
		}

		// Log Stock Mutation
		targetLoc := domain.LocationWarehouse
		balAfter := batch.QtyWarehouse
		if itemInput.PutToShelfDirectly {
			targetLoc = domain.LocationShelf
			balAfter = batch.QtyShelf
		}

		mutation := &domain.StockMutation{
			ID:           utils.GenerateStockMutationID(),
			BranchID:     branchID,
			MedicineID:   itemInput.MedicineID,
			BatchID:      batch.ID,
			MutationType: domain.MutationPurchase,
			Location:     targetLoc,
			ReferenceNo:  req.InvoiceNumber,
			QtyIn:        qtyBase,
			QtyOut:       0,
			BalanceAfter: balAfter,
			TotalAfter:   batch.TotalStock,
			Notes:        fmt.Sprintf("Penerimaan Faktur PBF No. %s (Batch: %s)", req.InvoiceNumber, batch.BatchNumber),
			UserID:       userID,
		}
		if err := u.invRepo.CreateMutation(ctx, mutation); err != nil {
			return nil, fmt.Errorf("failed to log stock mutation: %w", err)
		}

		// Prepare invoice item
		items = append(items, domain.PurchaseInvoiceItem{
			ID:                 utils.GeneratePurchaseInvoiceItemID(),
			PurchaseInvoiceID:  invoiceID,
			MedicineID:         itemInput.MedicineID,
			UnitID:             itemInput.UnitID,
			ConversionFactor:   conv,
			QtyPurchased:       itemInput.QtyPurchased,
			QtyBaseUnit:        qtyBase,
			BatchNumber:        itemInput.BatchNumber,
			ExpiredDate:        expDate,
			BuyPrice:           itemInput.BuyPrice,
			DiscountPercent:    itemInput.DiscountPercent,
			Subtotal:           itemSubtotal,
			PutToShelfDirectly: itemInput.PutToShelfDirectly,
		})
	}

	totalAmount := subtotal - req.DiscountAmount + req.TaxAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	paidAmount := float64(0)
	paymentStatus := domain.PaymentStatusUnpaid
	if req.PaymentType == domain.PaymentCash {
		paymentStatus = domain.PaymentStatusPaid
		paidAmount = totalAmount
	}

	invoice := &domain.PurchaseInvoice{
		ID:             invoiceID,
		BranchID:       branchID,
		SupplierID:     req.SupplierID,
		InvoiceNumber:  req.InvoiceNumber,
		InvoiceDate:    invoiceDate,
		DueDate:        dueDate,
		PaymentType:    req.PaymentType,
		PaymentStatus:  paymentStatus,
		Subtotal:       subtotal,
		DiscountAmount: req.DiscountAmount,
		TaxAmount:      req.TaxAmount,
		TotalAmount:    totalAmount,
		PaidAmount:     paidAmount,
		Notes:          req.Notes,
		ReceivedByID:   userID,
		Items:          items,
	}

	if err := u.invRepo.CreatePurchaseInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	// If paid in cash immediately, record initial payment record
	if paidAmount > 0 {
		_ = u.invRepo.CreateInvoicePayment(ctx, &domain.PurchaseInvoicePayment{
			ID:                utils.GenerateInvoicePaymentID(),
			BranchID:          branchID,
			PurchaseInvoiceID: invoiceID,
			PaymentDate:       invoiceDate,
			PaymentMethod:     "CASH",
			Amount:            paidAmount,
			ReferenceNumber:   "TUNAI_SAAT_TERIMA",
			Notes:             "Pembayaran tunai saat penerimaan barang faktur",
			CreatedByID:       userID,
		})
	}

	return u.invRepo.GetPurchaseInvoiceByID(ctx, invoiceID)
}

func (u *inventoryUsecase) ListPurchaseInvoices(ctx context.Context, branchID string, page, limit int) ([]*domain.PurchaseInvoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.invRepo.ListPurchaseInvoices(ctx, branchID, page, limit)
}

func (u *inventoryUsecase) GetPurchaseInvoiceDetail(ctx context.Context, invoiceID string) (*domain.PurchaseInvoice, error) {
	return u.invRepo.GetPurchaseInvoiceByID(ctx, invoiceID)
}

// -------------------------------------------------------------
// Accounts Payable & Payments Usecase
// -------------------------------------------------------------

func (u *inventoryUsecase) GetAccountsPayableSummary(ctx context.Context, branchID string) (*domain.AccountsPayableSummary, error) {
	return u.invRepo.GetAccountsPayableSummary(ctx, branchID)
}

func (u *inventoryUsecase) ListAccountsPayableInvoices(ctx context.Context, branchID string, filter domain.AccountsPayableFilter) (*domain.PaginatedAccountsPayable, error) {
	data, total, err := u.invRepo.ListAccountsPayableInvoices(ctx, branchID, filter)
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

	return &domain.PaginatedAccountsPayable{
		Data:       data,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *inventoryUsecase) RecordInvoicePayment(
	ctx context.Context,
	branchID, userID, invoiceID string,
	req *domain.CreateInvoicePaymentRequest,
) (*domain.PurchaseInvoicePayment, error) {
	if req.Amount <= 0 {
		return nil, domain.ErrInvalidPaymentValue
	}

	inv, err := u.invRepo.GetPurchaseInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.BranchID != branchID {
		return nil, domain.ErrUnauthorizedBranch
	}

	if inv.PaymentStatus == domain.PaymentStatusPaid {
		return nil, domain.ErrInvoiceAlreadyPaid
	}

	remainingBalance := inv.TotalAmount - inv.PaidAmount
	if req.Amount > (remainingBalance + 0.01) { // Tolerate floating point micro-differences
		return nil, domain.ErrPaymentExceedsBalance
	}

	payDate := time.Now()
	if req.PaymentDate != "" {
		if t, err := time.Parse("2006-01-02", req.PaymentDate); err == nil {
			payDate = t
		}
	}

	method := req.PaymentMethod
	if method == "" {
		method = "BANK_TRANSFER"
	}

	payment := &domain.PurchaseInvoicePayment{
		ID:                utils.GenerateInvoicePaymentID(),
		BranchID:          branchID,
		PurchaseInvoiceID: invoiceID,
		PaymentDate:       payDate,
		PaymentMethod:     method,
		Amount:            req.Amount,
		ReferenceNumber:   req.ReferenceNumber,
		Notes:             req.Notes,
		CreatedByID:       userID,
	}

	if err := u.invRepo.CreateInvoicePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to record invoice payment: %w", err)
	}

	return payment, nil
}

func (u *inventoryUsecase) GetInvoicePayments(ctx context.Context, invoiceID string) ([]domain.PurchaseInvoicePayment, error) {
	return u.invRepo.GetInvoicePaymentsByInvoiceID(ctx, invoiceID)
}

// -------------------------------------------------------------
// Internal Stock Movements (Gudang <-> Rak)
// -------------------------------------------------------------

func (u *inventoryUsecase) RefillShelf(
	ctx context.Context,
	branchID, userID string,
	req *domain.RefillShelfRequest,
) (*domain.MedicineBatch, error) {
	batch, err := u.invRepo.FindBatchByID(ctx, req.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.BranchID != branchID {
		return nil, domain.ErrBatchNotFound
	}
	if batch.QtyWarehouse < req.Qty {
		return nil, domain.ErrInsufficientWarehouseStock
	}

	batch.QtyWarehouse -= req.Qty
	batch.QtyShelf += req.Qty
	batch.CalculateTotal()

	if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
		return nil, err
	}

	notes := req.Notes
	if notes == "" {
		notes = fmt.Sprintf("Refill Rak dari Gudang (Batch: %s)", batch.BatchNumber)
	}

	// Double-entry mutation: Gudang Out, Rak In
	_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
		ID:           utils.GenerateStockMutationID(),
		BranchID:     branchID,
		MedicineID:   batch.MedicineID,
		BatchID:      batch.ID,
		MutationType: domain.MutationRefillShelf,
		Location:     domain.LocationWarehouse,
		ReferenceNo:  "REFILL-" + batch.ID,
		QtyIn:        0,
		QtyOut:       req.Qty,
		BalanceAfter: batch.QtyWarehouse,
		TotalAfter:   batch.TotalStock,
		Notes:        notes,
		UserID:       userID,
	})

	_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
		ID:           utils.GenerateStockMutationID(),
		BranchID:     branchID,
		MedicineID:   batch.MedicineID,
		BatchID:      batch.ID,
		MutationType: domain.MutationRefillShelf,
		Location:     domain.LocationShelf,
		ReferenceNo:  "REFILL-" + batch.ID,
		QtyIn:        req.Qty,
		QtyOut:       0,
		BalanceAfter: batch.QtyShelf,
		TotalAfter:   batch.TotalStock,
		Notes:        notes,
		UserID:       userID,
	})

	return batch, nil
}

func (u *inventoryUsecase) ReturnToWarehouse(
	ctx context.Context,
	branchID, userID string,
	req *domain.ReturnToWarehouseRequest,
) (*domain.MedicineBatch, error) {
	batch, err := u.invRepo.FindBatchByID(ctx, req.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.BranchID != branchID {
		return nil, domain.ErrBatchNotFound
	}
	if batch.QtyShelf < req.Qty {
		return nil, domain.ErrInsufficientShelfStock
	}

	batch.QtyShelf -= req.Qty
	batch.QtyWarehouse += req.Qty
	batch.CalculateTotal()

	if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
		return nil, err
	}

	notes := req.Notes
	if notes == "" {
		notes = fmt.Sprintf("Kembali ke Gudang dari Rak (Batch: %s)", batch.BatchNumber)
	}

	// Double-entry mutation: Rak Out, Gudang In
	_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
		ID:           utils.GenerateStockMutationID(),
		BranchID:     branchID,
		MedicineID:   batch.MedicineID,
		BatchID:      batch.ID,
		MutationType: domain.MutationReturnToWarehouse,
		Location:     domain.LocationShelf,
		ReferenceNo:  "RETURN-" + batch.ID,
		QtyIn:        0,
		QtyOut:       req.Qty,
		BalanceAfter: batch.QtyShelf,
		TotalAfter:   batch.TotalStock,
		Notes:        notes,
		UserID:       userID,
	})

	_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
		ID:           utils.GenerateStockMutationID(),
		BranchID:     branchID,
		MedicineID:   batch.MedicineID,
		BatchID:      batch.ID,
		MutationType: domain.MutationReturnToWarehouse,
		Location:     domain.LocationWarehouse,
		ReferenceNo:  "RETURN-" + batch.ID,
		QtyIn:        req.Qty,
		QtyOut:       0,
		BalanceAfter: batch.QtyWarehouse,
		TotalAfter:   batch.TotalStock,
		Notes:        notes,
		UserID:       userID,
	})

	return batch, nil
}

func (u *inventoryUsecase) CreateInitialStockBatch(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreateInitialBatchRequest,
) (*domain.MedicineBatch, error) {
	if req.MedicineID == "" {
		return nil, fmt.Errorf("obat wajib dipilih")
	}
	if req.BatchNumber == "" {
		return nil, fmt.Errorf("nomor batch wajib diisi")
	}
	if req.QtyWarehouse < 0 || req.QtyShelf < 0 || (req.QtyWarehouse+req.QtyShelf) <= 0 {
		return nil, fmt.Errorf("total stok fisik awal (gudang + rak) harus lebih dari 0")
	}

	// Verify medicine exists
	med, err := u.medRepo.FindByID(ctx, req.MedicineID)
	if err != nil {
		return nil, fmt.Errorf("obat tidak ditemukan: %w", err)
	}

	expDate, err := time.Parse("2006-01-02", req.ExpiredDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal kadaluarsa (ED) tidak valid (gunakan format YYYY-MM-DD)")
	}

	// Safe fallbacks for userID and branchID
	if userID == "" || userID == "SYSTEM" {
		userID = "USR000000000001"
	}
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	// Check if batch with this number already exists for this medicine in this branch
	existingBatch, err := u.invRepo.FindBatchByNumberAndMedicine(ctx, branchID, req.MedicineID, req.BatchNumber)
	var batch *domain.MedicineBatch
	if err == nil && existingBatch != nil {
		// Existing batch: increment stock
		batch = existingBatch
		batch.QtyWarehouse += req.QtyWarehouse
		batch.QtyShelf += req.QtyShelf
		batch.CalculateTotal()
		if req.BuyPrice > 0 {
			batch.BuyPrice = req.BuyPrice
		}
		batch.ExpiredDate = expDate
		batch.IsActive = true
		if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
			return nil, fmt.Errorf("gagal memperbarui batch: %w", err)
		}
	} else {
		// New batch
		batch = &domain.MedicineBatch{
			ID:           utils.GenerateBatchID(),
			MedicineID:   req.MedicineID,
			BranchID:     branchID,
			BatchNumber:  req.BatchNumber,
			ExpiredDate:  expDate,
			QtyWarehouse: req.QtyWarehouse,
			QtyShelf:     req.QtyShelf,
			TotalStock:   req.QtyWarehouse + req.QtyShelf,
			BuyPrice:     req.BuyPrice,
			IsActive:     true,
		}
		if err := u.invRepo.CreateBatch(ctx, batch); err != nil {
			return nil, fmt.Errorf("gagal mencatat batch baru: %w", err)
		}
	}

	refNo := fmt.Sprintf("SALDO-AWAL-%s-%d", batch.BatchNumber, time.Now().Unix())
	userNotes := req.Notes
	if userNotes == "" {
		userNotes = fmt.Sprintf("Saldo awal fisik batch %s (%s)", batch.BatchNumber, med.Name)
	}

	// Record mutation for Gudang if QtyWarehouse > 0
	if req.QtyWarehouse > 0 {
		_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
			ID:           utils.GenerateStockMutationID(),
			BranchID:     branchID,
			MedicineID:   req.MedicineID,
			BatchID:      batch.ID,
			MutationType: domain.MutationInitialBalance,
			Location:     domain.LocationWarehouse,
			ReferenceNo:  refNo,
			QtyIn:        req.QtyWarehouse,
			QtyOut:       0,
			BalanceAfter: batch.QtyWarehouse,
			TotalAfter:   batch.QtyWarehouse,
			Notes:        fmt.Sprintf("%s (Gudang Belakang)", userNotes),
			UserID:       userID,
		})
	}

	// Record mutation for Rak if QtyShelf > 0
	if req.QtyShelf > 0 {
		_ = u.invRepo.CreateMutation(ctx, &domain.StockMutation{
			ID:           utils.GenerateStockMutationID(),
			BranchID:     branchID,
			MedicineID:   req.MedicineID,
			BatchID:      batch.ID,
			MutationType: domain.MutationInitialBalance,
			Location:     domain.LocationShelf,
			ReferenceNo:  refNo,
			QtyIn:        req.QtyShelf,
			QtyOut:       0,
			BalanceAfter: batch.QtyShelf,
			TotalAfter:   batch.TotalStock,
			Notes:        fmt.Sprintf("%s (Rak Penjualan Kasir)", userNotes),
			UserID:       userID,
		})
	}

	return batch, nil
}

// -------------------------------------------------------------
// Stock Summaries & Digital Stock Card
// -------------------------------------------------------------

func (u *inventoryUsecase) GetStockSummary(ctx context.Context, branchID, medicineID string) (*domain.StockSummaryResponse, error) {
	med, err := u.medRepo.FindByID(ctx, medicineID)
	if err != nil {
		return nil, err
	}

	batches, err := u.invRepo.FindBatchesByMedicine(ctx, branchID, medicineID)
	if err != nil {
		return nil, err
	}

	var totalWh, totalSh float64
	for _, b := range batches {
		totalWh += b.QtyWarehouse
		totalSh += b.QtyShelf
	}

	baseUnitName := "Pcs"
	if med.BaseUnit != nil {
		baseUnitName = med.BaseUnit.Name
	}

	var sellPrice float64
	for _, pr := range med.Prices {
		if pr.PriceType == domain.PriceTypeUmum && pr.Price > 0 {
			sellPrice = pr.Price
			break
		}
	}
	if sellPrice == 0 && len(med.Prices) > 0 {
		sellPrice = med.Prices[0].Price
	}

	return &domain.StockSummaryResponse{
		MedicineID:     med.ID,
		ItemCode:       med.ItemCode,
		Barcode:        med.Barcode,
		Name:           med.Name,
		BaseUnitID:     med.BaseUnitID,
		BaseUnitName:   baseUnitName,
		SellPrice:      sellPrice,
		Prices:         med.Prices,
		TotalWarehouse: totalWh,
		TotalShelf:     totalSh,
		TotalStock:     totalWh + totalSh,
		MinStock:       med.MinStock,
		Batches:        batches,
	}, nil
}

func (u *inventoryUsecase) ListStockSummaries(ctx context.Context, branchID string, search string) ([]domain.StockSummaryResponse, error) {
	return u.invRepo.GetStockSummaryByBranch(ctx, branchID, search)
}

func (u *inventoryUsecase) GetStockCard(ctx context.Context, branchID, medicineID string, startDate, endDate string) ([]*domain.StockMutation, error) {
	var sDate, eDate time.Time
	if startDate != "" {
		if d, err := time.Parse("2006-01-02", startDate); err == nil {
			sDate = d
		}
	}
	if endDate != "" {
		if d, err := time.Parse("2006-01-02", endDate); err == nil {
			eDate = d.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}
	return u.invRepo.GetMutationsByMedicine(ctx, branchID, medicineID, sDate, eDate)
}

// -------------------------------------------------------------
// FEFO & Expiry Watcher
// -------------------------------------------------------------

func (u *inventoryUsecase) GetFefoAlerts(ctx context.Context, branchID string) ([]domain.FefoAlertItem, error) {
	batches, err := u.invRepo.FindBatchesFEFO(ctx, branchID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var alerts []domain.FefoAlertItem

	for _, b := range batches {
		// Only display batches that still have stock or are expired
		if b.TotalStock <= 0 {
			continue
		}

		days := int(b.ExpiredDate.Sub(now).Hours() / 24)
		status := "SAFE"
		if days <= 0 {
			status = "EXPIRED"
		} else if days <= 90 {
			status = "CRITICAL" // < 3 bulan
		} else if days <= 180 {
			status = "WARNING"  // < 6 bulan
		}

		medName := "Unknown Medicine"
		itemCode := ""
		if b.Medicine != nil {
			medName = b.Medicine.Name
			itemCode = b.Medicine.ItemCode
		}

		alerts = append(alerts, domain.FefoAlertItem{
			BatchID:      b.ID,
			MedicineID:   b.MedicineID,
			MedicineName: medName,
			ItemCode:     itemCode,
			BatchNumber:  b.BatchNumber,
			ExpiredDate:  b.ExpiredDate,
			DaysToExpiry: days,
			QtyWarehouse: b.QtyWarehouse,
			QtyShelf:     b.QtyShelf,
			TotalStock:   b.TotalStock,
			AlertStatus:  status,
		})
	}

	return alerts, nil
}

// -------------------------------------------------------------
// Rolling Cycle Count (Stok Opname Bergulir Parsial)
// -------------------------------------------------------------

func (u *inventoryUsecase) CreateOpnameSession(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreateOpnameSessionRequest,
) (*domain.StockOpname, error) {
	opnameID := utils.GenerateStockOpnameID()
	opnameCode := fmt.Sprintf("SOP-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))

	opname := &domain.StockOpname{
		ID:          opnameID,
		BranchID:    branchID,
		OpnameCode:  opnameCode,
		OpnameDate:  time.Now(),
		Location:    req.Location,
		RackName:    req.RackName,
		Status:      domain.OpnameDraft,
		Notes:       req.Notes,
		CreatedByID: userID,
	}

	if err := u.invRepo.CreateOpname(ctx, opname); err != nil {
		return nil, err
	}

	// Add initial items if provided in request
	for _, it := range req.Items {
		var batch *domain.MedicineBatch
		if it.BatchID != "" {
			batch, _ = u.invRepo.FindBatchByID(ctx, it.BatchID)
		}
		if batch == nil {
			batchNo := it.BatchNumber
			if batchNo == "" {
				batchNo = "OPN-" + time.Now().Format("20060102")
			}
			expDate := time.Now().AddDate(1, 0, 0)
			if it.ExpiredDate != "" {
				if d, e := time.Parse("2006-01-02", it.ExpiredDate); e == nil {
					expDate = d
				}
			}
			existing, _ := u.invRepo.FindBatchByNumberAndMedicine(ctx, branchID, it.MedicineID, batchNo)
			if existing != nil {
				batch = existing
			} else {
				batch = &domain.MedicineBatch{
					ID:          utils.GenerateBatchID(),
					MedicineID:  it.MedicineID,
					BranchID:    branchID,
					BatchNumber: batchNo,
					ExpiredDate: expDate,
					IsActive:    true,
				}
				_ = u.invRepo.CreateBatch(ctx, batch)
			}
			it.BatchID = batch.ID
		}

		var sysQty float64
		if req.Location == domain.LocationWarehouse {
			sysQty = batch.QtyWarehouse
		} else {
			sysQty = batch.QtyShelf
		}

		diff := it.PhysicalQty - sysQty

		opItem := &domain.StockOpnameItem{
			ID:            utils.GenerateStockOpnameItemID(),
			StockOpnameID: opnameID,
			MedicineID:    it.MedicineID,
			BatchID:       it.BatchID,
			SystemQty:     sysQty,
			PhysicalQty:   it.PhysicalQty,
			DifferenceQty: diff,
			Notes:         it.Notes,
		}
		_ = u.invRepo.AddOpnameItem(ctx, opItem)
	}

	return u.invRepo.GetOpnameByID(ctx, opnameID)
}

func (u *inventoryUsecase) SubmitOpnameSession(
	ctx context.Context,
	branchID, userID, opnameID string,
	req *domain.SubmitOpnameRequest,
) (*domain.StockOpname, error) {
	opname, err := u.invRepo.GetOpnameByID(ctx, opnameID)
	if err != nil {
		return nil, err
	}
	if opname.BranchID != branchID {
		return nil, domain.ErrOpnameNotFound
	}
	if opname.Status == domain.OpnameCompleted {
		return nil, domain.ErrOpnameAlreadyCompleted
	}

	// Process each counted item and adjust stock immediately
	for _, it := range req.Items {
		var batch *domain.MedicineBatch
		if it.BatchID != "" {
			batch, _ = u.invRepo.FindBatchByID(ctx, it.BatchID)
		}
		if batch == nil {
			batchNo := it.BatchNumber
			if batchNo == "" {
				batchNo = "OPN-" + time.Now().Format("20060102")
			}
			expDate := time.Now().AddDate(1, 0, 0)
			if it.ExpiredDate != "" {
				if d, e := time.Parse("2006-01-02", it.ExpiredDate); e == nil {
					expDate = d
				}
			}
			existing, _ := u.invRepo.FindBatchByNumberAndMedicine(ctx, branchID, it.MedicineID, batchNo)
			if existing != nil {
				batch = existing
			} else {
				batch = &domain.MedicineBatch{
					ID:          utils.GenerateBatchID(),
					MedicineID:  it.MedicineID,
					BranchID:    branchID,
					BatchNumber: batchNo,
					ExpiredDate: expDate,
					IsActive:    true,
				}
				if err := u.invRepo.CreateBatch(ctx, batch); err != nil {
					continue
				}
			}
			it.BatchID = batch.ID
		}

		var sysQty float64
		if opname.Location == domain.LocationWarehouse {
			sysQty = batch.QtyWarehouse
			batch.QtyWarehouse = it.PhysicalQty
		} else {
			sysQty = batch.QtyShelf
			batch.QtyShelf = it.PhysicalQty
		}
		batch.CalculateTotal()

		diff := it.PhysicalQty - sysQty

		// Update batch balance
		if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
			return nil, err
		}

		// Log Stock Mutation for Opname Adjustment
		var qtyIn, qtyOut float64
		if diff > 0 {
			qtyIn = diff
		} else if diff < 0 {
			qtyOut = -diff
		}

		balAfter := batch.QtyWarehouse
		if opname.Location == domain.LocationShelf {
			balAfter = batch.QtyShelf
		}

		mutation := &domain.StockMutation{
			ID:           utils.GenerateStockMutationID(),
			BranchID:     branchID,
			MedicineID:   it.MedicineID,
			BatchID:      it.BatchID,
			MutationType: domain.MutationAdjustment,
			Location:     opname.Location,
			ReferenceNo:  opname.OpnameCode,
			QtyIn:        qtyIn,
			QtyOut:       qtyOut,
			BalanceAfter: balAfter,
			TotalAfter:   batch.TotalStock,
			Notes:        fmt.Sprintf("Penyesuaian Opname Fisik: %s (%s). %s", opname.RackName, opname.Location, it.Notes),
			UserID:       userID,
		}
		_ = u.invRepo.CreateMutation(ctx, mutation)

		// Record Opname Item
		opItem := &domain.StockOpnameItem{
			ID:            utils.GenerateStockOpnameItemID(),
			StockOpnameID: opnameID,
			MedicineID:    it.MedicineID,
			BatchID:       it.BatchID,
			SystemQty:     sysQty,
			PhysicalQty:   it.PhysicalQty,
			DifferenceQty: diff,
			Notes:         it.Notes,
		}
		_ = u.invRepo.AddOpnameItem(ctx, opItem)
	}

	opname.Status = domain.OpnameCompleted
	if err := u.invRepo.UpdateOpname(ctx, opname); err != nil {
		return nil, err
	}

	return u.invRepo.GetOpnameByID(ctx, opnameID)
}

func (u *inventoryUsecase) ListOpnames(ctx context.Context, branchID string, page, limit int) ([]*domain.StockOpname, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.invRepo.ListOpnames(ctx, branchID, page, limit)
}

func (u *inventoryUsecase) GetOpnameDetail(ctx context.Context, opnameID string) (*domain.StockOpname, error) {
	return u.invRepo.GetOpnameByID(ctx, opnameID)
}

// -------------------------------------------------------------
// Collaborative Multi-User Opname (Sistem Setor per Rak)
// -------------------------------------------------------------

func (u *inventoryUsecase) GetActiveParentOpname(ctx context.Context, branchID string) (*domain.ParentOpnameSummaryResponse, error) {
	opname, err := u.invRepo.GetActiveParentOpname(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if opname == nil {
		return nil, nil
	}

	var pendingCount, approvedCount, rejectedCount, totalItems, discrepancyCount int
	for _, rack := range opname.Racks {
		switch rack.Status {
		case domain.RackPending:
			pendingCount++
		case domain.RackApproved:
			approvedCount++
		case domain.RackRejected:
			rejectedCount++
		}
		totalItems += len(rack.Items)
		for _, item := range rack.Items {
			if item.DifferenceQty != 0 {
				discrepancyCount++
			}
		}
	}

	return &domain.ParentOpnameSummaryResponse{
		Opname:            opname,
		TotalRacks:        len(opname.Racks),
		PendingRacks:      pendingCount,
		ApprovedRacks:     approvedCount,
		RejectedRacks:     rejectedCount,
		TotalItemsCounted: totalItems,
		DiscrepancyCount:  discrepancyCount,
		Racks:             opname.Racks,
	}, nil
}

func (u *inventoryUsecase) CreateParentOpnameSession(
	ctx context.Context,
	branchID, userID string,
	req *domain.CreateParentOpnameRequest,
) (*domain.StockOpname, error) {
	// Verify no active parent session already exists for this branch
	active, err := u.invRepo.GetActiveParentOpname(ctx, branchID)
	if err == nil && active != nil {
		return nil, fmt.Errorf("masih terdapat Sesi Induk Opname yang aktif (%s). Silakan selesaikan sesi tersebut sebelum membuat sesi baru", active.OpnameCode)
	}

	opnameID := utils.GenerateStockOpnameID()
	opnameCode := fmt.Sprintf("SOP-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))

	loc := req.Location
	if loc == "" {
		loc = domain.LocationShelf
	}

	notes := req.Notes
	if notes == "" {
		notes = "Sesi Induk Stok Opname Kolaboratif Multi-User"
	}

	opname := &domain.StockOpname{
		ID:          opnameID,
		BranchID:    branchID,
		OpnameCode:  opnameCode,
		OpnameDate:  time.Now(),
		Location:    loc,
		RackName:    "Sesi Induk Opname",
		Status:      domain.OpnameOpen,
		Notes:       notes,
		CreatedByID: userID,
	}

	if err := u.invRepo.CreateOpname(ctx, opname); err != nil {
		return nil, fmt.Errorf("gagal membuat sesi induk opname: %w", err)
	}

	return u.invRepo.GetOpnameByID(ctx, opnameID)
}

func (u *inventoryUsecase) SubmitRackCount(
	ctx context.Context,
	branchID, userID, opnameID string,
	req *domain.SubmitRackCountRequest,
) (*domain.StockOpnameRack, error) {
	opname, err := u.invRepo.GetOpnameByID(ctx, opnameID)
	if err != nil {
		return nil, err
	}
	if opname.BranchID != branchID {
		return nil, domain.ErrOpnameNotFound
	}
	if opname.Status != domain.OpnameOpen {
		return nil, fmt.Errorf("sesi induk opname ini sudah tidak aktif (status: %s)", opname.Status)
	}
	if req.RackName == "" {
		return nil, fmt.Errorf("nama rak wajib diisi")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("minimal 1 item obat wajib dimasukkan dalam hitungan rak")
	}

	rackID := utils.GenerateStockOpnameRackID()
	loc := req.Location
	if loc == "" {
		loc = opname.Location
	}

	rack := &domain.StockOpnameRack{
		ID:            rackID,
		StockOpnameID: opnameID,
		RackName:      req.RackName,
		Location:      loc,
		Status:        domain.RackPending,
		CountedByID:   userID,
		Notes:         req.Notes,
	}

	if err := u.invRepo.CreateOpnameRack(ctx, rack); err != nil {
		return nil, fmt.Errorf("gagal membuat data rak opname: %w", err)
	}

	// Save all counted items in this rack
	for _, it := range req.Items {
		var batch *domain.MedicineBatch
		if it.BatchID != "" {
			batch, _ = u.invRepo.FindBatchByID(ctx, it.BatchID)
		}
		if batch == nil {
			batchNo := it.BatchNumber
			if batchNo == "" {
				batchNo = "OPN-" + time.Now().Format("20060102")
			}
			expDate := time.Now().AddDate(1, 0, 0)
			if it.ExpiredDate != "" {
				if d, e := time.Parse("2006-01-02", it.ExpiredDate); e == nil {
					expDate = d
				}
			}
			existing, _ := u.invRepo.FindBatchByNumberAndMedicine(ctx, branchID, it.MedicineID, batchNo)
			if existing != nil {
				batch = existing
			} else {
				batch = &domain.MedicineBatch{
					ID:          utils.GenerateBatchID(),
					MedicineID:  it.MedicineID,
					BranchID:    branchID,
					BatchNumber: batchNo,
					ExpiredDate: expDate,
					IsActive:    true,
				}
				_ = u.invRepo.CreateBatch(ctx, batch)
			}
			it.BatchID = batch.ID
		}

		var sysQty float64
		if loc == domain.LocationWarehouse {
			sysQty = batch.QtyWarehouse
		} else {
			sysQty = batch.QtyShelf
		}

		diff := it.PhysicalQty - sysQty

		opItem := &domain.StockOpnameItem{
			ID:                utils.GenerateStockOpnameItemID(),
			StockOpnameID:     opnameID,
			StockOpnameRackID: &rackID,
			RackName:          req.RackName,
			MedicineID:        it.MedicineID,
			BatchID:           it.BatchID,
			SystemQty:         sysQty,
			PhysicalQty:       it.PhysicalQty,
			DifferenceQty:     diff,
			Notes:             it.Notes,
			CountedByID:       &userID,
		}
		_ = u.invRepo.AddOpnameItem(ctx, opItem)
	}

	return u.invRepo.GetOpnameRackByID(ctx, rackID)
}

func (u *inventoryUsecase) GetOpnameRackDetail(ctx context.Context, rackID string) (*domain.StockOpnameRack, error) {
	return u.invRepo.GetOpnameRackByID(ctx, rackID)
}

func (u *inventoryUsecase) ApproveRackCount(
	ctx context.Context,
	branchID, userID, opnameID, rackID string,
) (*domain.StockOpnameRack, error) {
	opname, err := u.invRepo.GetOpnameByID(ctx, opnameID)
	if err != nil {
		return nil, err
	}
	if opname.BranchID != branchID {
		return nil, domain.ErrOpnameNotFound
	}

	rack, err := u.invRepo.GetOpnameRackByID(ctx, rackID)
	if err != nil {
		return nil, err
	}
	if rack.StockOpnameID != opnameID {
		return nil, fmt.Errorf("rak tidak terikat pada sesi opname ini")
	}
	if rack.Status == domain.RackApproved {
		return nil, domain.ErrOpnameRackAlreadyProcessed
	}

	// Adjust actual stock for all items counted in this approved rack
	for _, it := range rack.Items {
		batch, err := u.invRepo.FindBatchByID(ctx, it.BatchID)
		if err != nil || batch == nil {
			continue
		}

		var sysBefore float64
		if rack.Location == domain.LocationWarehouse {
			sysBefore = batch.QtyWarehouse
			batch.QtyWarehouse = it.PhysicalQty
		} else {
			sysBefore = batch.QtyShelf
			batch.QtyShelf = it.PhysicalQty
		}
		batch.CalculateTotal()

		if err := u.invRepo.UpdateBatch(ctx, batch); err != nil {
			return nil, fmt.Errorf("gagal memperbarui stok batch: %w", err)
		}

		// Calculate discrepancy
		diff := it.PhysicalQty - sysBefore
		if diff != 0 {
			var qtyIn, qtyOut float64
			if diff > 0 {
				qtyIn = diff
			} else {
				qtyOut = -diff
			}

			balAfter := batch.QtyWarehouse
			if rack.Location == domain.LocationShelf {
				balAfter = batch.QtyShelf
			}

			mutation := &domain.StockMutation{
				ID:           utils.GenerateStockMutationID(),
				BranchID:     branchID,
				MedicineID:   it.MedicineID,
				BatchID:      it.BatchID,
				MutationType: domain.MutationAdjustment,
				Location:     rack.Location,
				ReferenceNo:  fmt.Sprintf("%s / %s", opname.OpnameCode, rack.RackName),
				QtyIn:        qtyIn,
				QtyOut:       qtyOut,
				BalanceAfter: balAfter,
				TotalAfter:   batch.TotalStock,
				Notes:        fmt.Sprintf("ACC Opname Fisik Rak '%s' (%s). %s", rack.RackName, rack.Location, it.Notes),
				UserID:       userID,
			}
			_ = u.invRepo.CreateMutation(ctx, mutation)
		}
	}

	now := time.Now()
	rack.Status = domain.RackApproved
	rack.ApprovedByID = &userID
	rack.ApprovedAt = &now

	if err := u.invRepo.UpdateOpnameRack(ctx, rack); err != nil {
		return nil, fmt.Errorf("gagal memperbarui status rak: %w", err)
	}

	return u.invRepo.GetOpnameRackByID(ctx, rackID)
}

func (u *inventoryUsecase) RejectRackCount(
	ctx context.Context,
	branchID, userID, opnameID, rackID string,
	notes string,
) (*domain.StockOpnameRack, error) {
	opname, err := u.invRepo.GetOpnameByID(ctx, opnameID)
	if err != nil {
		return nil, err
	}
	if opname.BranchID != branchID {
		return nil, domain.ErrOpnameNotFound
	}

	rack, err := u.invRepo.GetOpnameRackByID(ctx, rackID)
	if err != nil {
		return nil, err
	}
	if rack.StockOpnameID != opnameID {
		return nil, fmt.Errorf("rak tidak terikat pada sesi opname ini")
	}
	if rack.Status == domain.RackApproved {
		return nil, fmt.Errorf("rak yang sudah disetujui tidak dapat ditolak")
	}

	now := time.Now()
	rack.Status = domain.RackRejected
	rack.ApprovedByID = &userID
	rack.ApprovedAt = &now
	if notes != "" {
		if rack.Notes != "" {
			rack.Notes += " | Catatan Penolakan: " + notes
		} else {
			rack.Notes = "Catatan Penolakan: " + notes
		}
	}

	if err := u.invRepo.UpdateOpnameRack(ctx, rack); err != nil {
		return nil, fmt.Errorf("gagal menolak rak: %w", err)
	}

	return u.invRepo.GetOpnameRackByID(ctx, rackID)
}

func (u *inventoryUsecase) CloseParentOpnameSession(
	ctx context.Context,
	branchID, userID, opnameID string,
) (*domain.StockOpname, error) {
	opname, err := u.invRepo.GetOpnameByID(ctx, opnameID)
	if err != nil {
		return nil, err
	}
	if opname.BranchID != branchID {
		return nil, domain.ErrOpnameNotFound
	}
	if opname.Status == domain.OpnameCompleted {
		return opname, nil
	}

	opname.Status = domain.OpnameCompleted
	if err := u.invRepo.UpdateOpname(ctx, opname); err != nil {
		return nil, fmt.Errorf("gagal menutup sesi induk opname: %w", err)
	}

	return u.invRepo.GetOpnameByID(ctx, opnameID)
}
