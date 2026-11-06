package postgres

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
	"pharma-be/pkg/utils"
)

// InitDB initializes PostgreSQL connection using GORM, prepares custom ENUM types, and runs auto-migration
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	dbLogLevel := logger.Warn
	if cfg.AppEnv == "development" {
		dbLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(dbLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	// Connection Pool Settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic database object: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Step 1: Ensure all custom ENUM types exist in PostgreSQL
	initEnumSQL := `
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
			CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'pharmacist', 'cashier', 'inventory');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
			CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'journal_method') THEN
			CREATE TYPE journal_method AS ENUM ('perpetual', 'periodic');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'branch_status') THEN
			CREATE TYPE branch_status AS ENUM ('active', 'inactive', 'suspended');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_type') THEN
			CREATE TYPE subscription_type AS ENUM ('trial', 'basic', 'pro', 'enterprise');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'real_asset') THEN
			CREATE TYPE real_asset AS ENUM ('owned', 'rented', 'partnership');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'drug_classification') THEN
			CREATE TYPE drug_classification AS ENUM (
				'bebas', 'bebas_terbatas', 'keras', 'narkotika', 'psikotropika',
				'prekursor', 'oot', 'alkes', 'herbal', 'suplemen', 'lainnya'
			);
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'price_type') THEN
			CREATE TYPE price_type AS ENUM ('umum', 'resep', 'member');
		END IF;
	END $$;
	`
	if err := db.Exec(initEnumSQL).Error; err != nil {
		return nil, fmt.Errorf("failed to create custom postgres enum types: %w", err)
	}

	// Step 2: AutoMigrate tables in order
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Branch{},
		&domain.UserBranch{},
		&domain.Unit{},
		&domain.Category{},
		&domain.Supplier{},
		&domain.Doctor{},
		&domain.Medicine{},
		&domain.MedicineUnit{},
		&domain.MedicinePrice{},
		&domain.MedicineBatch{},
		&domain.StockMutation{},
		&domain.PurchaseInvoice{},
		&domain.PurchaseInvoiceItem{},
		&domain.PurchaseInvoicePayment{},
		&domain.StockOpname{},
		&domain.StockOpnameItem{},
		&domain.PurchaseOrder{},
		&domain.PurchaseOrderItem{},
		&domain.Sale{},
		&domain.SaleItem{},
		&domain.Prescription{},
		&domain.PrescriptionItem{},
		&domain.PrescriptionCompoundItem{},
		&domain.PurchaseReturn{},
		&domain.PurchaseReturnItem{},
		&domain.SalesReturn{},
		&domain.SalesReturnItem{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	// Step 2.5: Ensure daily bonus target columns exist in branches table
	_ = db.Exec("ALTER TABLE branches ADD COLUMN IF NOT EXISTS daily_margin_target numeric(15,2) DEFAULT 250000").Error
	_ = db.Exec("ALTER TABLE branches ADD COLUMN IF NOT EXISTS daily_bonus_reward numeric(15,2) DEFAULT 50000").Error
	_ = db.Exec("ALTER TABLE purchase_invoices ADD COLUMN IF NOT EXISTS paid_amount numeric(14,2) DEFAULT 0").Error

	// Step 3: Seed initial data (Admin & Default Branches)
	seedInitialData(db)

	return db, nil
}

func seedInitialData(db *gorm.DB) {
	// 1. Seed Admin
	var userCount int64
	db.Model(&domain.User{}).Count(&userCount)
	adminID := "USR000000000001"
	if userCount == 0 {
		hashedPassword, err := utils.HashPassword("admin123")
		if err == nil {
			defaultAdmin := domain.User{
				ID:         adminID,
				Username:   "admin",
				Password:   hashedPassword,
				Name:       "Administrator Vi-Pharma",
				UserRole:   domain.RoleSuperAdmin,
				UserStatus: domain.StatusActive,
			}
			if err := db.Create(&defaultAdmin).Error; err == nil {
				log.Printf("[SEEDER] Default admin created: username='admin', password='admin123'")
			}
		}
	}

	// 2. Seed Default Branches
	var branchCount int64
	db.Model(&domain.Branch{}).Count(&branchCount)
	branch1ID := "CBG000000000001"
	branch2ID := "CBG000000000002"

	if branchCount == 0 {
		branch1 := domain.Branch{
			ID:               branch1ID,
			BranchCode:       "CBG-PST-01",
			BranchName:       "Apotek Vi-Pharma Pusat",
			Address:          "Jl. Pemuda No. 10, Jakarta",
			Phone:            "021-5551234",
			Email:            "pusat@vipharma.com",
			SiaID:            "SIA-101/2026",
			SiaName:          "Apotek Vi-Pharma Pusat",
			PsaID:            "PSA-101",
			PsaName:          "PT Vi-Pharma Medika",
			Sipa:             "SIPA-2026/001",
			SipaName:         "apt. Hendra Wijaya, S.Farm",
			ApingID:          "APING-2026/001",
			ApingName:        "apt. Siti Nurhaliza, S.Farm",
			BankName:         "BCA",
			AccountName:      "PT Vi-Pharma Medika",
			AccountNumber:    "8880192837",
			TaxPercentage:    11,
			JournalMethod:    domain.JournalMethodPerpetual,
			BranchStatus:     domain.BranchStatusActive,
			SubscriptionType: domain.SubscriptionEnterprise,
			Quota:            1000,
			RealAsset:        domain.RealAssetOwned,
		}

		branch2 := domain.Branch{
			ID:               branch2ID,
			BranchCode:       "CBG-SDR-02",
			BranchName:       "Apotek Vi-Pharma Sudirman",
			Address:          "Jl. Jend. Sudirman No. 45, Jakarta",
			Phone:            "021-5555678",
			Email:            "sudirman@vipharma.com",
			SiaID:            "SIA-102/2026",
			SiaName:          "Apotek Vi-Pharma Sudirman",
			PsaID:            "PSA-101",
			PsaName:          "PT Vi-Pharma Medika",
			Sipa:             "SIPA-2026/002",
			SipaName:         "apt. Dewi Lestari, S.Farm",
			ApingID:          "APING-2026/002",
			ApingName:        "apt. Rian Pratama, S.Farm",
			BankName:         "Mandiri",
			AccountName:      "PT Vi-Pharma Medika",
			AccountNumber:    "140001928374",
			TaxPercentage:    11,
			JournalMethod:    domain.JournalMethodPerpetual,
			BranchStatus:     domain.BranchStatusActive,
			SubscriptionType: domain.SubscriptionPro,
			Quota:            500,
			RealAsset:        domain.RealAssetRented,
		}

		_ = db.Create(&branch1).Error
		_ = db.Create(&branch2).Error
		log.Printf("[SEEDER] Default branches created: 'CBG-PST-01' and 'CBG-SDR-02'")
	}

	// 3. Ensure Admin has access to both branches
	var userBranchCount int64
	db.Model(&domain.UserBranch{}).Where("user_id = ?", adminID).Count(&userBranchCount)
	if userBranchCount == 0 {
		ub1 := domain.UserBranch{
			ID:        "UBR000000000001",
			UserID:    adminID,
			BranchID:  branch1ID,
			IsDefault: true,
		}
		ub2 := domain.UserBranch{
			ID:        "UBR000000000002",
			UserID:    adminID,
			BranchID:  branch2ID,
			IsDefault: false,
		}
		_ = db.Create(&ub1).Error
		_ = db.Create(&ub2).Error
		log.Printf("[SEEDER] Admin linked to default branches (Pusat as default).")
	}

	// 4. Seed Standard Units
	var unitCount int64
	db.Model(&domain.Unit{}).Count(&unitCount)
	if unitCount == 0 {
		units := []domain.Unit{
			{ID: "UNT000000000001", Name: "Tablet", Symbol: "tab", Description: "Satuan sediaan padat tablet", IsActive: true},
			{ID: "UNT000000000002", Name: "Strip", Symbol: "str", Description: "Kemasan blister/strip", IsActive: true},
			{ID: "UNT000000000003", Name: "Box", Symbol: "box", Description: "Kemasan kotak/dus besar", IsActive: true},
			{ID: "UNT000000000004", Name: "Kapsul", Symbol: "kap", Description: "Satuan sediaan kapsul", IsActive: true},
			{ID: "UNT000000000005", Name: "Botol", Symbol: "btl", Description: "Kemasan botol sirup/cair", IsActive: true},
			{ID: "UNT000000000006", Name: "Tube", Symbol: "tb", Description: "Kemasan salep/krim", IsActive: true},
			{ID: "UNT000000000007", Name: "Sachet", Symbol: "sct", Description: "Kemasan sachet serbuk", IsActive: true},
			{ID: "UNT000000000008", Name: "Ampul", Symbol: "amp", Description: "Sediaan injeksi ampul", IsActive: true},
			{ID: "UNT000000000009", Name: "Vial", Symbol: "vial", Description: "Sediaan injeksi vial", IsActive: true},
			{ID: "UNT000000000010", Name: "Pcs", Symbol: "pcs", Description: "Pieces alat kesehatan", IsActive: true},
		}
		for _, u := range units {
			_ = db.Create(&u).Error
		}
		log.Printf("[SEEDER] Standard pharmaceutical units seeded.")
	}

	// 5. Seed Standard Categories
	var catCount int64
	db.Model(&domain.Category{}).Count(&catCount)
	if catCount == 0 {
		categories := []domain.Category{
			{ID: "CAT000000000001", Name: "Antibiotik", Description: "Obat pembunuh atau penghambat bakteri", IsActive: true},
			{ID: "CAT000000000002", Name: "Analgesik & Antipiretik", Description: "Pereda nyeri dan penurun demam", IsActive: true},
			{ID: "CAT000000000003", Name: "Vitamin & Suplemen", Description: "Multivitamin dan suplemen daya tahan tubuh", IsActive: true},
			{ID: "CAT000000000004", Name: "Saluran Cerna", Description: "Antasida, anti-maag, dan pencernaan", IsActive: true},
			{ID: "CAT000000000005", Name: "Antihipertensi & Jantung", Description: "Pengontrol tekanan darah dan kardiovaskular", IsActive: true},
			{ID: "CAT000000000006", Name: "Antidiabetes", Description: "Pengontrol gula darah", IsActive: true},
			{ID: "CAT000000000007", Name: "Antihistamin & Alergi", Description: "Pereda alergi dan gatal", IsActive: true},
			{ID: "CAT000000000008", Name: "Obat Luar & Kulit", Description: "Salep, krim antijamur, dan antiseptik", IsActive: true},
		}
		for _, c := range categories {
			_ = db.Create(&c).Error
		}
		log.Printf("[SEEDER] Standard pharmaceutical categories seeded.")
	}

	// 6. Seed Standard Suppliers / PBF
	var supCount int64
	db.Model(&domain.Supplier{}).Count(&supCount)
	pbf1ID := "SUP000000000001"
	if supCount == 0 {
		suppliers := []domain.Supplier{
			{
				ID:              pbf1ID,
				Code:            "PBF-KF-01",
				Name:            "PT Kimia Farma Trading & Distribution",
				Address:         "Jl. Budi Utomo No. 1, Jakarta Pusat",
				Phone:           "021-3847709",
				Email:           "order@kftd.co.id",
				PICName:         "Bambang Irawan",
				PICPhone:        "081298765432",
				NPWP:            "01.001.608.7-051.000",
				PaymentTermDays: 30,
				IsActive:        true,
			},
			{
				ID:              "SUP000000000002",
				Code:            "PBF-AAM-02",
				Name:            "PT Anugrah Argon Medica",
				Address:         "Titan Center Lantai 3, Bintaro Jaya Sektor 7, Tangerang Selatan",
				Phone:           "021-7459000",
				Email:           "cs@argon-group.com",
				PICName:         "Dewi Sartika",
				PICPhone:        "081387654321",
				NPWP:            "01.328.776.8-058.000",
				PaymentTermDays: 21,
				IsActive:        true,
			},
			{
				ID:              "SUP000000000003",
				Code:            "PBF-MBS-03",
				Name:            "PT Mensa Binasukses",
				Address:         "Jl. Pulogadung No. 25, Kawasan Industri Pulogadung, Jakarta Timur",
				Phone:           "021-4609080",
				Email:           "info@mbs.co.id",
				PICName:         "Rahmat Hidayat",
				PICPhone:        "081198761234",
				NPWP:            "01.123.456.7-003.000",
				PaymentTermDays: 30,
				IsActive:        true,
			},
		}
		for _, s := range suppliers {
			_ = db.Create(&s).Error
		}
		log.Printf("[SEEDER] Default PBF / Suppliers seeded.")
	}

	// 7. Seed Standard Doctors
	var docCount int64
	db.Model(&domain.Doctor{}).Count(&docCount)
	if docCount == 0 {
		doctors := []domain.Doctor{
			{
				ID:             "DOC000000000001",
				Name:           "dr. Hendra Pratama, Sp.A",
				SIP:            "SIP.503/442/DU/2024",
				Specialization: "Spesialis Anak (Pediatri)",
				ClinicHospital: "RSIA Bunda & Klinik Sehat Mandiri",
				Phone:          "081234567890",
				Address:        "Jl. Kenanga No. 15, Jakarta",
				IsActive:       true,
			},
			{
				ID:             "DOC000000000002",
				Name:           "dr. Siti Rahmawati, Sp.PD",
				SIP:            "SIP.503/118/SP/2023",
				Specialization: "Spesialis Penyakit Dalam (Internis)",
				ClinicHospital: "RSUD Pratama Jakarta",
				Phone:          "081345678901",
				Address:        "Jl. Melati No. 8, Jakarta",
				IsActive:       true,
			},
		}
		for _, d := range doctors {
			_ = db.Create(&d).Error
		}
		log.Printf("[SEEDER] Default Doctors seeded.")
	}

	// 8. Seed Initial Medicines with Multi-packaging & Pricing
	var medCount int64
	db.Model(&domain.Medicine{}).Count(&medCount)
	if medCount == 0 {
		unitTabID := "UNT000000000001"
		unitStrID := "UNT000000000002"
		unitBoxID := "UNT000000000003"
		unitBtlID := "UNT000000000005"
		catAntiID := "CAT000000000001"
		catAnalID := "CAT000000000002"

		// Item 1: Amoxicillin 500 mg
		med1ID := "MED000000000001"
		med1 := domain.Medicine{
			ID:                   med1ID,
			ItemCode:             "OBT-00001",
			Barcode:              "8991234560011",
			Name:                 "Amoxicillin 500 mg",
			GenericName:          "Amoxicillin Trihydrate",
			NIE:                  "GKL9805024104A1",
			Manufacturer:         "PT Kimia Farma",
			DosageForm:           "Kaplet",
			DrugClassification:   domain.DrugClassKeras,
			CategoryID:           catAntiID,
			BaseUnitID:           unitTabID,
			RequiresPrescription: true,
			MinStock:             50,
			MaxStock:             1000,
			StorageLocation:      "Rak A-01",
			DefaultSupplierID:    &pbf1ID,
			IsActive:             true,
			LabelType:            "dalam",
			DefaultSigna:         "3 x Sehari 1 Kaplet (Tiap 8 Jam)",
			ConsumptionTime:      "Sesudah Makan",
			DoseMorning:          "1",
			DoseNoon:             "1",
			DoseNight:            "1",
			WarningTags:          "Harus Dihabiskan (Antibiotik)",
			SpecialInstructions:  "Wajib dihabiskan untuk mencegah resistensi antibiotik",
			BUDDays:              0,
		}
		_ = db.Create(&med1).Error

		// Packaging conversions for Amoxicillin:
		// 1 Box = 100 Tablet, 1 Strip = 10 Tablet, 1 Tablet = 1 Tablet (Base)
		unitsMed1 := []domain.MedicineUnit{
			{ID: "MNU000000000001", MedicineID: med1ID, UnitID: unitTabID, ConversionFactor: 1, IsBaseUnit: true, Barcode: "8991234560011"},
			{ID: "MNU000000000002", MedicineID: med1ID, UnitID: unitStrID, ConversionFactor: 10, IsBaseUnit: false, Barcode: "8991234560012"},
			{ID: "MNU000000000003", MedicineID: med1ID, UnitID: unitBoxID, ConversionFactor: 100, IsBaseUnit: false, Barcode: "8991234560013"},
		}
		for _, u := range unitsMed1 {
			_ = db.Create(&u).Error
		}

		// Pricing for Amoxicillin:
		pricesMed1 := []domain.MedicinePrice{
			{ID: "MPR000000000001", MedicineID: med1ID, UnitID: unitTabID, PriceType: domain.PriceTypeUmum, Price: 800},
			{ID: "MPR000000000002", MedicineID: med1ID, UnitID: unitTabID, PriceType: domain.PriceTypeResep, Price: 1000},
			{ID: "MPR000000000003", MedicineID: med1ID, UnitID: unitStrID, PriceType: domain.PriceTypeUmum, Price: 7500},
			{ID: "MPR000000000004", MedicineID: med1ID, UnitID: unitBoxID, PriceType: domain.PriceTypeUmum, Price: 70000},
		}
		for _, p := range pricesMed1 {
			_ = db.Create(&p).Error
		}

		// Item 2: Panadol Extra 500mg/65mg
		med2ID := "MED000000000002"
		med2 := domain.Medicine{
			ID:                   med2ID,
			ItemCode:             "OBT-00002",
			Barcode:              "8992741123456",
			Name:                 "Panadol Extra 500mg/65mg",
			GenericName:          "Paracetamol + Caffeine",
			NIE:                  "DTL9424502004A1",
			Manufacturer:         "PT Sterling Products Indonesia",
			DosageForm:           "Kaplet",
			DrugClassification:   domain.DrugClassBebas,
			CategoryID:           catAnalID,
			BaseUnitID:           unitTabID,
			RequiresPrescription: false,
			MinStock:             30,
			MaxStock:             500,
			StorageLocation:      "Rak B-02 (OTC)",
			DefaultSupplierID:    &pbf1ID,
			IsActive:             true,
			LabelType:            "dalam",
			DefaultSigna:         "3 x Sehari 1 Kaplet",
			ConsumptionTime:      "Sesudah Makan (Bila Demam / Sakit)",
			DoseMorning:          "1",
			DoseNoon:             "1",
			DoseNight:            "1",
			WarningTags:          "Hentikan jika demam reda",
			SpecialInstructions:  "Hati-hati penggunaan jangka panjang tanpa resep dokter",
			BUDDays:              0,
		}
		_ = db.Create(&med2).Error

		unitsMed2 := []domain.MedicineUnit{
			{ID: "MNU000000000004", MedicineID: med2ID, UnitID: unitTabID, ConversionFactor: 1, IsBaseUnit: true, Barcode: "8992741123456"},
			{ID: "MNU000000000005", MedicineID: med2ID, UnitID: unitStrID, ConversionFactor: 10, IsBaseUnit: false, Barcode: "8992741123457"},
			{ID: "MNU000000000006", MedicineID: med2ID, UnitID: unitBoxID, ConversionFactor: 100, IsBaseUnit: false, Barcode: "8992741123458"},
		}
		for _, u := range unitsMed2 {
			_ = db.Create(&u).Error
		}

		pricesMed2 := []domain.MedicinePrice{
			{ID: "MPR000000000005", MedicineID: med2ID, UnitID: unitTabID, PriceType: domain.PriceTypeUmum, Price: 1500},
			{ID: "MPR000000000006", MedicineID: med2ID, UnitID: unitStrID, PriceType: domain.PriceTypeUmum, Price: 14000},
			{ID: "MPR000000000007", MedicineID: med2ID, UnitID: unitBoxID, PriceType: domain.PriceTypeUmum, Price: 135000},
		}
		for _, p := range pricesMed2 {
			_ = db.Create(&p).Error
		}

		// Item 3: Sanmol Sirup 60 ml
		med3ID := "MED000000000003"
		med3 := domain.Medicine{
			ID:                   med3ID,
			ItemCode:             "OBT-00003",
			Barcode:              "8998838120019",
			Name:                 "Sanmol Sirup 60 ml",
			GenericName:          "Paracetamol 120mg/5ml",
			NIE:                  "DBL7622235637A1",
			Manufacturer:         "PT Sanbe Farma",
			DosageForm:           "Sirup",
			DrugClassification:   domain.DrugClassBebas,
			CategoryID:           catAnalID,
			BaseUnitID:           unitBtlID,
			RequiresPrescription: false,
			MinStock:             15,
			MaxStock:             200,
			StorageLocation:      "Rak Sirup Anak C-01",
			DefaultSupplierID:    &pbf1ID,
			IsActive:             true,
			LabelType:            "dalam",
			DefaultSigna:         "3 x Sehari 1 Sendok Takar (5 ml)",
			ConsumptionTime:      "Sesudah Makan",
			DoseMorning:          "5 ml",
			DoseNoon:             "5 ml",
			DoseNight:            "5 ml",
			WarningTags:          "Kocok Dahulu",
			SpecialInstructions:  "Gunakan sendok takar obat resmi yang tersedia di kemasan",
			BUDDays:              30,
		}
		_ = db.Create(&med3).Error

		unitsMed3 := []domain.MedicineUnit{
			{ID: "MNU000000000007", MedicineID: med3ID, UnitID: unitBtlID, ConversionFactor: 1, IsBaseUnit: true, Barcode: "8998838120019"},
		}
		for _, u := range unitsMed3 {
			_ = db.Create(&u).Error
		}

		pricesMed3 := []domain.MedicinePrice{
			{ID: "MPR000000000008", MedicineID: med3ID, UnitID: unitBtlID, PriceType: domain.PriceTypeUmum, Price: 25000},
			{ID: "MPR000000000009", MedicineID: med3ID, UnitID: unitBtlID, PriceType: domain.PriceTypeResep, Price: 26000},
		}
		for _, p := range pricesMed3 {
			_ = db.Create(&p).Error
		}

		log.Printf("[SEEDER] Default sample medicines with packaging & prices seeded.")
	}

	// 8. Seed MIPI Minyak Gosok 60 ml (with physical batch ED: 2029)
	var mipiCount int64
	db.Model(&domain.Medicine{}).Where("barcode = ?", "8993430174989").Count(&mipiCount)
	if mipiCount == 0 {
		unitBtlID := "UNT000000000005"
		catLuarID := "CAT000000000008"
		med4ID := "MED000000000004"

		med4 := domain.Medicine{
			ID:                   med4ID,
			ItemCode:             "OBT-00004",
			Barcode:              "8993430174989",
			Name:                 "MIPI Minyak Gosok Pereda Nyeri Otot 60 ml",
			GenericName:          "Methyl Salicylate, Menthol, Camphor, Capsicum Oleoresin, Clove Oil",
			NIE:                  "POM QD 061 606 731",
			Manufacturer:         "PT Etercon Pharma (Novell)",
			DosageForm:           "Cairan Obat Luar",
			DrugClassification:   domain.DrugClassHerbal,
			CategoryID:           catLuarID,
			BaseUnitID:           unitBtlID,
			RequiresPrescription: false,
			MinStock:             5,
			MaxStock:             100,
			StorageLocation:      "Rak Minyak Gosok A-01",
			DefaultSupplierID:    &pbf1ID,
			IsActive:             true,
			LabelType:            "luar",
			DefaultSigna:         "Dioleskan & dipijat 10-15 menit pada area nyeri/pegal",
			ConsumptionTime:      "Bila Perlu / Saat Nyeri",
			DoseMorning:          "Secukupnya",
			DoseNoon:             "Secukupnya",
			DoseNight:            "Secukupnya",
			WarningTags:          "Hanya Untuk Obat Luar,Hindari Area Mata dan Luka Terbuka,Tidak Untuk Anak < 2 Tahun",
			SpecialInstructions:  "Simpan pada suhu di bawah 30°C. Jangan digunakan langsung di bawah lubang hidung.",
			BUDDays:              90,
		}
		_ = db.Create(&med4).Error

		unitsMed4 := []domain.MedicineUnit{
			{ID: "MNU000000000010", MedicineID: med4ID, UnitID: unitBtlID, ConversionFactor: 1, IsBaseUnit: true, Barcode: "8993430174989"},
		}
		for _, u := range unitsMed4 {
			_ = db.Create(&u).Error
		}

		pricesMed4 := []domain.MedicinePrice{
			{ID: "MPR000000000010", MedicineID: med4ID, UnitID: unitBtlID, PriceType: domain.PriceTypeUmum, Price: 24500},
			{ID: "MPR000000000011", MedicineID: med4ID, UnitID: unitBtlID, PriceType: domain.PriceTypeResep, Price: 25000},
			{ID: "MPR000000000012", MedicineID: med4ID, UnitID: unitBtlID, PriceType: domain.PriceTypeMember, Price: 23000},
		}
		for _, p := range pricesMed4 {
			_ = db.Create(&p).Error
		}

		// Seed initial physical batch: No. Batch EFG146, Expired Date 2029-07-31
		expDate, _ := time.Parse("2006-01-02", "2029-07-31")
		batchMipi := domain.MedicineBatch{
			ID:           "BTC000000000004",
			MedicineID:   med4ID,
			BranchID:     "CBG000000000001",
			BatchNumber:  "EFG146",
			ExpiredDate:  expDate,
			QtyWarehouse: 5,
			QtyShelf:     10,
			TotalStock:   15,
			BuyPrice:     19500,
			IsActive:     true,
		}
		_ = db.Create(&batchMipi).Error

		log.Printf("[SEEDER] MIPI Minyak Gosok 60ml with batch EFG146 (ED: 2029) seeded.")
	}

	// 9. Seed initial stock mutations for MIPI (for Digital Stock Card history)
	var mipiMutCount int64
	db.Model(&domain.StockMutation{}).Where("medicine_id = ?", "MED000000000004").Count(&mipiMutCount)
	if mipiMutCount == 0 {
		mutGudang := domain.StockMutation{
			ID:           "MUT000000000001",
			BranchID:     "CBG000000000001",
			MedicineID:   "MED000000000004",
			BatchID:      "BTC000000000004",
			MutationType: domain.MutationPurchase,
			Location:     domain.LocationWarehouse,
			ReferenceNo:  "SALDO-AWAL-001",
			QtyIn:        5,
			QtyOut:       0,
			BalanceAfter: 5,
			TotalAfter:   5,
			Notes:        "Stok Awal Fisik Gudang Belakang (Batch EFG146)",
			UserID:       "USR000000000001",
		}
		_ = db.Create(&mutGudang).Error

		mutRak := domain.StockMutation{
			ID:           "MUT000000000002",
			BranchID:     "CBG000000000001",
			MedicineID:   "MED000000000004",
			BatchID:      "BTC000000000004",
			MutationType: domain.MutationPurchase,
			Location:     domain.LocationShelf,
			ReferenceNo:  "SALDO-AWAL-001",
			QtyIn:        10,
			QtyOut:       0,
			BalanceAfter: 10,
			TotalAfter:   15,
			Notes:        "Stok Awal Fisik Rak Penjualan (Batch EFG146)",
			UserID:       "USR000000000001",
		}
		_ = db.Create(&mutRak).Error

		log.Printf("[SEEDER] Initial stock card mutations for MIPI seeded.")
	}
}
