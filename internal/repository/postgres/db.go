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
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

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
}
