package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type prescriptionUsecase struct {
	prescriptionRepo domain.PrescriptionRepository
	medRepo          domain.MedicineRepository
	db               *gorm.DB
}

func NewPrescriptionUsecase(
	prescriptionRepo domain.PrescriptionRepository,
	medRepo domain.MedicineRepository,
	db *gorm.DB,
) domain.PrescriptionUsecase {
	return &prescriptionUsecase{
		prescriptionRepo: prescriptionRepo,
		medRepo:          medRepo,
		db:               db,
	}
}

func (u *prescriptionUsecase) CreatePrescription(ctx context.Context, branchID string, req *domain.CreatePrescriptionRequest) (*domain.Prescription, error) {
	if branchID == "" {
		branchID = "CBG000000000001"
	}

	rxDate := time.Now()
	if req.PrescriptionDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.PrescriptionDate)
		if err == nil {
			rxDate = parsedDate
		}
	}

	rxNumber, err := u.prescriptionRepo.GetNextPrescriptionNumber(ctx, branchID, rxDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prescription number: %w", err)
	}

	prescriptionID := utils.GeneratePrescriptionID()

	prescription := &domain.Prescription{
		ID:                 prescriptionID,
		BranchID:           branchID,
		PrescriptionNumber: rxNumber,
		DoctorID:           req.DoctorID,
		DoctorName:         req.DoctorName,
		DoctorSIP:          req.DoctorSIP,
		PatientName:        req.PatientName,
		PatientAge:         req.PatientAge,
		PatientWeight:      req.PatientWeight,
		PatientGender:      req.PatientGender,
		PatientPhone:       req.PatientPhone,
		PatientAddress:     req.PatientAddress,
		PrescriptionDate:   rxDate,
		Status:             domain.PrescriptionStatusScreening,
		ScreeningAdmin:     req.ScreeningAdmin,
		ScreeningPharma:    req.ScreeningPharma,
		ScreeningClinical:  req.ScreeningClinical,
		ScreeningPassed:    req.ScreeningAdmin && req.ScreeningPharma && req.ScreeningClinical,
		ScreeningNotes:     req.ScreeningNotes,
		Notes:              req.Notes,
		IsPaid:             false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	var totalMedPrice float64
	var totalEmbalase float64
	var totalTuslah float64
	var items []domain.PrescriptionItem

	for _, itemReq := range req.Items {
		itemID := utils.GeneratePrescriptionItemID()
		item := domain.PrescriptionItem{
			ID:                  itemID,
			PrescriptionID:      prescriptionID,
			ItemType:            itemReq.ItemType,
			LabelType:           itemReq.LabelType,
			Signa:               itemReq.Signa,
			UsageTime:           itemReq.UsageTime,
			DoseMorning:         itemReq.DoseMorning,
			DoseNoon:            itemReq.DoseNoon,
			DoseNight:           itemReq.DoseNight,
			WarningTags:         itemReq.WarningTags,
			SpecialInstructions: itemReq.SpecialInstructions,
			BUDDays:             itemReq.BUDDays,
			DeturStatus:         itemReq.DeturStatus,
			CreatedAt:           time.Now(),
		}

		if item.LabelType == "" {
			item.LabelType = "dalam"
		}
		if item.DeturStatus == "" {
			item.DeturStatus = "det"
		}

		if itemReq.ItemType == domain.PrescriptionItemNonRacik {
			// Non-racik item
			item.MedicineID = itemReq.MedicineID
			item.BatchID = itemReq.BatchID
			item.UnitID = itemReq.UnitID
			item.Qty = itemReq.Qty
			if item.Qty <= 0 {
				item.Qty = 1
			}

			unitPrice := itemReq.UnitPrice
			if unitPrice <= 0 && itemReq.MedicineID != nil {
				// Fetch price from DB
				var medPrice domain.MedicinePrice
				if err := u.db.Where("medicine_id = ? AND price_type = 'resep'", *itemReq.MedicineID).First(&medPrice).Error; err == nil {
					unitPrice = medPrice.Price
				} else if err := u.db.Where("medicine_id = ? AND price_type = 'umum'", *itemReq.MedicineID).First(&medPrice).Error; err == nil {
					unitPrice = medPrice.Price
				}
			}
			item.UnitPrice = unitPrice
			item.Subtotal = unitPrice * item.Qty
			item.Total = item.Subtotal

			totalMedPrice += item.Subtotal
		} else {
			// Racikan item
			item.RacikanName = itemReq.RacikanName
			if item.RacikanName == "" {
				item.RacikanName = "Racikan Puyer/Kapsul"
			}
			item.RacikanDoseForm = itemReq.RacikanDoseForm
			item.RacikanQty = itemReq.RacikanQty
			if item.RacikanQty <= 0 {
				item.RacikanQty = 1
			}
			item.RacikanUnit = itemReq.RacikanUnit
			if item.RacikanUnit == "" {
				item.RacikanUnit = "Bungkus"
			}
			item.RacikanMethod = itemReq.RacikanMethod
			if item.RacikanMethod == "" {
				item.RacikanMethod = domain.RacikanMethodDTD
			}
			item.EmbalasePrice = itemReq.EmbalasePrice
			item.TuslahPrice = itemReq.TuslahPrice

			totalEmbalase += item.EmbalasePrice
			totalTuslah += item.TuslahPrice

			var compoundSubtotal float64
			var compoundItems []domain.PrescriptionCompoundItem

			for _, cReq := range itemReq.CompoundItems {
				compID := utils.GeneratePrescriptionCompoundItemID()
				comp := domain.PrescriptionCompoundItem{
					ID:                 compID,
					PrescriptionItemID: itemID,
					MedicineID:         cReq.MedicineID,
					BatchID:            cReq.BatchID,
					UnitID:             cReq.UnitID,
					DosePerPack:        cReq.DosePerPack,
					TotalQty:           cReq.TotalQty,
					UnitPrice:          cReq.UnitPrice,
					CreatedAt:          time.Now(),
				}

				if comp.UnitPrice <= 0 {
					var medPrice domain.MedicinePrice
					if err := u.db.Where("medicine_id = ? AND price_type = 'resep'", cReq.MedicineID).First(&medPrice).Error; err == nil {
						comp.UnitPrice = medPrice.Price
					} else if err := u.db.Where("medicine_id = ? AND price_type = 'umum'", cReq.MedicineID).First(&medPrice).Error; err == nil {
						comp.UnitPrice = medPrice.Price
					}
				}

				comp.Subtotal = comp.UnitPrice * comp.TotalQty
				compoundSubtotal += comp.Subtotal
				compoundItems = append(compoundItems, comp)
			}

			item.CompoundItems = compoundItems
			item.Subtotal = compoundSubtotal
			item.Total = compoundSubtotal + item.EmbalasePrice + item.TuslahPrice
			totalMedPrice += compoundSubtotal
		}

		items = append(items, item)
	}

	prescription.TotalMedPrice = totalMedPrice
	prescription.TotalEmbalase = totalEmbalase
	prescription.TotalTuslah = totalTuslah
	prescription.GrandTotal = totalMedPrice + totalEmbalase + totalTuslah
	prescription.Items = items

	if err := u.prescriptionRepo.Create(ctx, prescription); err != nil {
		return nil, fmt.Errorf("failed to save prescription: %w", err)
	}

	return u.prescriptionRepo.FindByID(ctx, prescriptionID)
}

func (u *prescriptionUsecase) GetPrescriptions(ctx context.Context, filter domain.PrescriptionFilter) (*domain.PaginatedPrescriptions, error) {
	prescriptions, totalItems, err := u.prescriptionRepo.FindAll(ctx, filter)
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

	return &domain.PaginatedPrescriptions{
		Data:       prescriptions,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (u *prescriptionUsecase) GetPrescriptionByID(ctx context.Context, id string) (*domain.Prescription, error) {
	return u.prescriptionRepo.FindByID(ctx, id)
}

func (u *prescriptionUsecase) UpdateStatus(ctx context.Context, id string, req *domain.UpdatePrescriptionStatusRequest) (*domain.Prescription, error) {
	prescription, err := u.prescriptionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status != "" {
		prescription.Status = req.Status
	}
	if req.IsPaid != nil {
		prescription.IsPaid = *req.IsPaid
	}
	if req.SaleID != nil {
		prescription.SaleID = req.SaleID
	}
	if req.PharmacistID != nil {
		prescription.PharmacistID = req.PharmacistID
	}
	if req.Notes != nil {
		prescription.Notes = *req.Notes
	}
	if req.ScreeningAdmin != nil {
		prescription.ScreeningAdmin = *req.ScreeningAdmin
	}
	if req.ScreeningPharma != nil {
		prescription.ScreeningPharma = *req.ScreeningPharma
	}
	if req.ScreeningClinical != nil {
		prescription.ScreeningClinical = *req.ScreeningClinical
	}
	if req.ScreeningPassed != nil {
		prescription.ScreeningPassed = *req.ScreeningPassed
	} else {
		prescription.ScreeningPassed = prescription.ScreeningAdmin && prescription.ScreeningPharma && prescription.ScreeningClinical
	}
	if req.ScreeningNotes != nil {
		prescription.ScreeningNotes = *req.ScreeningNotes
	}

	prescription.UpdatedAt = time.Now()
	if err := u.prescriptionRepo.Update(ctx, prescription); err != nil {
		return nil, err
	}

	return u.prescriptionRepo.FindByID(ctx, id)
}

func (u *prescriptionUsecase) DispensePrescription(ctx context.Context, id string, userID string) (*domain.Prescription, error) {
	prescription, err := u.prescriptionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if prescription.Status == domain.PrescriptionStatusSelesai {
		return prescription, nil
	}

	if userID == "" {
		userID = "USR000000000001"
	}

	// Transaction to deduct stock from shelf & record stock mutations
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range prescription.Items {
			if item.ItemType == domain.PrescriptionItemNonRacik {
				if item.MedicineID == nil || item.BatchID == nil {
					continue
				}

				var batch domain.MedicineBatch
				if err := tx.Where("id = ? AND branch_id = ?", *item.BatchID, prescription.BranchID).First(&batch).Error; err != nil {
					// Fallback to active batch if specified batch not found
					if err := tx.Where("medicine_id = ? AND branch_id = ? AND is_active = true AND qty_shelf >= ?", *item.MedicineID, prescription.BranchID, item.Qty).Order("expired_date ASC").First(&batch).Error; err != nil {
						return fmt.Errorf("obat %s: stok tidak mencukupi untuk resep: %w", *item.MedicineID, domain.ErrInsufficientShelfStock)
					}
				}

				if batch.QtyShelf < item.Qty {
					return fmt.Errorf("batch %s: stok rak hanya %.2f, diminta %.2f: %w", batch.BatchNumber, batch.QtyShelf, item.Qty, domain.ErrInsufficientShelfStock)
				}

				batch.QtyShelf -= item.Qty
				batch.TotalStock -= item.Qty
				batch.UpdatedAt = time.Now()
				if err := tx.Save(&batch).Error; err != nil {
					return err
				}

				mutID := utils.GenerateStockMutationID()
				mutation := domain.StockMutation{
					ID:           mutID,
					BranchID:     prescription.BranchID,
					MedicineID:   *item.MedicineID,
					BatchID:      batch.ID,
					MutationType: domain.MutationPrescription,
					Location:     domain.LocationShelf,
					ReferenceNo:  prescription.PrescriptionNumber,
					QtyIn:        0,
					QtyOut:       item.Qty,
					BalanceAfter: batch.QtyShelf,
					TotalAfter:   batch.TotalStock,
					Notes:        fmt.Sprintf("Pelayanan Resep Pasien: %s (Signa: %s)", prescription.PatientName, item.Signa),
					UserID:       userID,
					CreatedAt:    time.Now(),
				}
				if err := tx.Create(&mutation).Error; err != nil {
					return err
				}
			} else {
				// Racikan: Deduct each compound ingredient
				for _, comp := range item.CompoundItems {
					var batch domain.MedicineBatch
					if comp.BatchID != nil && *comp.BatchID != "" {
						_ = tx.Where("id = ? AND branch_id = ?", *comp.BatchID, prescription.BranchID).First(&batch).Error
					}
					if batch.ID == "" {
						if err := tx.Where("medicine_id = ? AND branch_id = ? AND is_active = true AND qty_shelf >= ?", comp.MedicineID, prescription.BranchID, comp.TotalQty).Order("expired_date ASC").First(&batch).Error; err != nil {
							// If not enough shelf, try first active batch
							if err := tx.Where("medicine_id = ? AND branch_id = ? AND is_active = true", comp.MedicineID, prescription.BranchID).Order("expired_date ASC").First(&batch).Error; err != nil {
								return fmt.Errorf("bahan racikan obat %s tidak ditemukan di inventori: %w", comp.MedicineID, errors.New("batch not found"))
							}
						}
					}

					if batch.QtyShelf < comp.TotalQty {
						return fmt.Errorf("bahan racikan %s: stok rak hanya %.2f, diminta %.2f: %w", batch.BatchNumber, batch.QtyShelf, comp.TotalQty, domain.ErrInsufficientShelfStock)
					}

					batch.QtyShelf -= comp.TotalQty
					batch.TotalStock -= comp.TotalQty
					batch.UpdatedAt = time.Now()
					if err := tx.Save(&batch).Error; err != nil {
						return err
					}

					mutID := utils.GenerateStockMutationID()
					mutation := domain.StockMutation{
						ID:           mutID,
						BranchID:     prescription.BranchID,
						MedicineID:   comp.MedicineID,
						BatchID:      batch.ID,
						MutationType: domain.MutationPrescription,
						Location:     domain.LocationShelf,
						ReferenceNo:  prescription.PrescriptionNumber,
						QtyIn:        0,
						QtyOut:       comp.TotalQty,
						BalanceAfter: batch.QtyShelf,
						TotalAfter:   batch.TotalStock,
						Notes:        fmt.Sprintf("Bahan Racikan '%s' untuk Resep: %s (Pasien: %s)", item.RacikanName, prescription.PrescriptionNumber, prescription.PatientName),
						UserID:       userID,
						CreatedAt:    time.Now(),
					}
					if err := tx.Create(&mutation).Error; err != nil {
						return err
					}
				}
			}
		}

		prescription.Status = domain.PrescriptionStatusSiap
		prescription.UpdatedAt = time.Now()
		return tx.Save(prescription).Error
	})

	if err != nil {
		return nil, err
	}

	return u.prescriptionRepo.FindByID(ctx, id)
}
