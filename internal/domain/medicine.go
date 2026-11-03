package domain

import (
	"context"
	"time"
)

// DrugClassification represents official Indonesian pharmaceutical regulatory categories
type DrugClassification string

const (
	DrugClassBebas         DrugClassification = "bebas"          // Lingkaran Hijau
	DrugClassBebasTerbatas DrugClassification = "bebas_terbatas" // Lingkaran Biru
	DrugClassKeras         DrugClassification = "keras"          // Lingkaran Merah K (Daftar G)
	DrugClassNarkotika     DrugClassification = "narkotika"      // Palang Merah
	DrugClassPsikotropika  DrugClassification = "psikotropika"
	DrugClassPrekursor     DrugClassification = "prekursor"
	DrugClassOOT           DrugClassification = "oot"            // Obat-Obat Tertentu
	DrugClassAlkes         DrugClassification = "alkes"          // Alat Kesehatan & BMHP
	DrugClassHerbal        DrugClassification = "herbal"         // Jamu / Fitofarmaka
	DrugClassSuplemen      DrugClassification = "suplemen"
	DrugClassLainnya       DrugClassification = "lainnya"
)

// PriceType represents pricing tiers for pharmacy items
type PriceType string

const (
	PriceTypeUmum  PriceType = "umum"  // Penjualan bebas OTC
	PriceTypeResep PriceType = "resep" // Penjualan dengan resep dokter
	PriceTypeMember PriceType = "member" // Pelanggan / member apotek
)

// LabelType represents pharmacy dispensing label categories (Etiket)
type LabelType string

const (
	LabelTypeDalam LabelType = "dalam" // Etiket Putih (Obat Dalam / Oral)
	LabelTypeLuar  LabelType = "luar"  // Etiket Biru (Obat Luar / Topikal / Tetes / Inhalasi)
)

// Medicine represents a pharmaceutical drug or medical product
type Medicine struct {
	ID                   string             `gorm:"type:varchar(15);primaryKey" json:"id"`
	ItemCode             string             `gorm:"type:varchar(50);not null;uniqueIndex" json:"item_code"`
	Barcode              string             `gorm:"type:varchar(50);index" json:"barcode"`
	Name                 string             `gorm:"type:varchar(255);not null;index" json:"name"`
	GenericName          string             `gorm:"type:varchar(255);index" json:"generic_name"`
	NIE                  string             `gorm:"type:varchar(100)" json:"nie"` // Nomor Izin Edar BPOM
	Manufacturer         string             `gorm:"type:varchar(150)" json:"manufacturer"`
	DosageForm           string             `gorm:"type:varchar(100)" json:"dosage_form"` // Tablet, Sirup, Salep, dll.
	DrugClassification   DrugClassification `gorm:"type:drug_classification;not null;default:'bebas'" json:"drug_classification"`
	CategoryID           string             `gorm:"type:varchar(15);not null;index" json:"category_id"`
	Category             *Category          `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	BaseUnitID           string             `gorm:"type:varchar(15);not null" json:"base_unit_id"`
	BaseUnit             *Unit              `gorm:"foreignKey:BaseUnitID" json:"base_unit,omitempty"`
	RequiresPrescription bool               `gorm:"not null;default:false" json:"requires_prescription"`
	MinStock             int                `gorm:"not null;default:10" json:"min_stock"`
	MaxStock             int                `gorm:"not null;default:1000" json:"max_stock"`
	StorageLocation      string             `gorm:"type:varchar(100)" json:"storage_location"`
	DefaultSupplierID    *string            `gorm:"type:varchar(15)" json:"default_supplier_id"`
	DefaultSupplier      *Supplier          `gorm:"foreignKey:DefaultSupplierID" json:"default_supplier,omitempty"`
	IsActive             bool               `gorm:"not null;default:true;index" json:"is_active"`

	// Etiket & Aturan Pakai (Dispensing Regulation & Etiket Print Support)
	LabelType           string             `gorm:"type:varchar(20);not null;default:'dalam'" json:"label_type"` // dalam / luar
	DefaultSigna        string             `gorm:"type:varchar(150)" json:"default_signa"`                      // e.g. "3 x Sehari 1 Tablet"
	ConsumptionTime     string             `gorm:"type:varchar(100)" json:"consumption_time"`                   // e.g. "Sesudah Makan"
	DoseMorning         string             `gorm:"type:varchar(50);default:'1'" json:"dose_morning"`            // e.g. "1"
	DoseNoon            string             `gorm:"type:varchar(50);default:'1'" json:"dose_noon"`               // e.g. "1"
	DoseNight           string             `gorm:"type:varchar(50);default:'1'" json:"dose_night"`              // e.g. "1"
	WarningTags         string             `gorm:"type:text" json:"warning_tags"`                               // e.g. "Harus Dihabiskan,Kocok Dahulu"
	SpecialInstructions string             `gorm:"type:text" json:"special_instructions"`                       // e.g. "Larutkan dalam 1 gelas air"
	BUDDays             int                `gorm:"not null;default:0" json:"bud_days"`                          // Beyond Use Date in days

	Units                []MedicineUnit     `gorm:"foreignKey:MedicineID" json:"units,omitempty"`
	Prices               []MedicinePrice    `gorm:"foreignKey:MedicineID" json:"prices,omitempty"`
	CreatedAt            time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// MedicineUnit represents multi-tier packaging conversions (e.g. 1 Box = 10 Strip = 100 Tablet)
type MedicineUnit struct {
	ID               string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	MedicineID       string    `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	UnitID           string    `gorm:"type:varchar(15);not null;index" json:"unit_id"`
	Unit             *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	ConversionFactor int       `gorm:"not null;default:1" json:"conversion_factor"` // Multiplier to base unit
	IsBaseUnit       bool      `gorm:"not null;default:false" json:"is_base_unit"`
	Barcode          string    `gorm:"type:varchar(50)" json:"barcode"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// MedicinePrice represents multi-tier selling prices per unit & tier
type MedicinePrice struct {
	ID         string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	MedicineID string    `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	UnitID     string    `gorm:"type:varchar(15);not null;index" json:"unit_id"`
	Unit       *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	PriceType  PriceType `gorm:"type:price_type;not null;default:'umum'" json:"price_type"`
	Price      float64   `gorm:"type:numeric(15,2);not null;default:0" json:"price"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// DTOs
type MedicineUnitInput struct {
	UnitID           string `json:"unit_id" validate:"required"`
	ConversionFactor int    `json:"conversion_factor" validate:"required,gt=0"`
	IsBaseUnit       bool   `json:"is_base_unit"`
	Barcode          string `json:"barcode"`
}

type MedicinePriceInput struct {
	UnitID    string    `json:"unit_id" validate:"required"`
	PriceType PriceType `json:"price_type" validate:"required"`
	Price     float64   `json:"price" validate:"gte=0"`
}

type CreateMedicineRequest struct {
	ItemCode             string               `json:"item_code" validate:"required,min=2,max=50"`
	Barcode              string               `json:"barcode"`
	Name                 string               `json:"name" validate:"required,min=2,max=255"`
	GenericName          string               `json:"generic_name"`
	NIE                  string               `json:"nie"`
	Manufacturer         string               `json:"manufacturer"`
	DosageForm           string               `json:"dosage_form"`
	DrugClassification   DrugClassification   `json:"drug_classification" validate:"required"`
	CategoryID           string               `json:"category_id" validate:"required"`
	BaseUnitID           string               `json:"base_unit_id" validate:"required"`
	RequiresPrescription bool                 `json:"requires_prescription"`
	MinStock             int                  `json:"min_stock" validate:"gte=0"`
	MaxStock             int                  `json:"max_stock" validate:"gte=0"`
	StorageLocation      string               `json:"storage_location"`
	DefaultSupplierID    *string              `json:"default_supplier_id"`
	IsActive             *bool                `json:"is_active"`
	LabelType            string               `json:"label_type"`
	DefaultSigna         string               `json:"default_signa"`
	ConsumptionTime      string               `json:"consumption_time"`
	DoseMorning          string               `json:"dose_morning"`
	DoseNoon             string               `json:"dose_noon"`
	DoseNight            string               `json:"dose_night"`
	WarningTags          string               `json:"warning_tags"`
	SpecialInstructions  string               `json:"special_instructions"`
	BUDDays              int                  `json:"bud_days"`
	Units                []MedicineUnitInput  `json:"units"`
	Prices               []MedicinePriceInput `json:"prices"`
}

type UpdateMedicineRequest struct {
	ItemCode             string               `json:"item_code" validate:"required,min=2,max=50"`
	Barcode              string               `json:"barcode"`
	Name                 string               `json:"name" validate:"required,min=2,max=255"`
	GenericName          string               `json:"generic_name"`
	NIE                  string               `json:"nie"`
	Manufacturer         string               `json:"manufacturer"`
	DosageForm           string               `json:"dosage_form"`
	DrugClassification   DrugClassification   `json:"drug_classification" validate:"required"`
	CategoryID           string               `json:"category_id" validate:"required"`
	BaseUnitID           string               `json:"base_unit_id" validate:"required"`
	RequiresPrescription bool                 `json:"requires_prescription"`
	MinStock             int                  `json:"min_stock" validate:"gte=0"`
	MaxStock             int                  `json:"max_stock" validate:"gte=0"`
	StorageLocation      string               `json:"storage_location"`
	DefaultSupplierID    *string              `json:"default_supplier_id"`
	IsActive             *bool                `json:"is_active"`
	LabelType            string               `json:"label_type"`
	DefaultSigna         string               `json:"default_signa"`
	ConsumptionTime      string               `json:"consumption_time"`
	DoseMorning          string               `json:"dose_morning"`
	DoseNoon             string               `json:"dose_noon"`
	DoseNight            string               `json:"dose_night"`
	WarningTags          string               `json:"warning_tags"`
	SpecialInstructions  string               `json:"special_instructions"`
	BUDDays              int                  `json:"bud_days"`
	Units                []MedicineUnitInput  `json:"units"`
	Prices               []MedicinePriceInput `json:"prices"`
}

type MedicineFilter struct {
	Search         string             `json:"search"`
	Classification DrugClassification `json:"classification"`
	CategoryID     string             `json:"category_id"`
	ActiveOnly     bool               `json:"active_only"`
	Page           int                `json:"page"`
	Limit          int                `json:"limit"`
}

type MedicineResponse struct {
	ID                   string             `json:"id"`
	ItemCode             string             `json:"item_code"`
	Barcode              string             `json:"barcode"`
	Name                 string             `json:"name"`
	GenericName          string             `json:"generic_name"`
	NIE                  string             `json:"nie"`
	Manufacturer         string             `json:"manufacturer"`
	DosageForm           string             `json:"dosage_form"`
	DrugClassification   DrugClassification `json:"drug_classification"`
	CategoryID           string             `json:"category_id"`
	CategoryName         string             `json:"category_name,omitempty"`
	BaseUnitID           string             `json:"base_unit_id"`
	BaseUnitName         string             `json:"base_unit_name,omitempty"`
	RequiresPrescription bool               `json:"requires_prescription"`
	MinStock             int                `json:"min_stock"`
	MaxStock             int                `json:"max_stock"`
	StorageLocation      string             `json:"storage_location"`
	DefaultSupplierID    *string            `json:"default_supplier_id,omitempty"`
	DefaultSupplierName  string             `json:"default_supplier_name,omitempty"`
	IsActive             bool               `json:"is_active"`

	// Etiket & Aturan Pakai
	LabelType           string             `json:"label_type"`
	DefaultSigna        string             `json:"default_signa"`
	ConsumptionTime     string             `json:"consumption_time"`
	DoseMorning         string             `json:"dose_morning"`
	DoseNoon            string             `json:"dose_noon"`
	DoseNight           string             `json:"dose_night"`
	WarningTags         string             `json:"warning_tags"`
	SpecialInstructions string             `json:"special_instructions"`
	BUDDays             int                `json:"bud_days"`

	Units                []MedicineUnit     `json:"units,omitempty"`
	Prices               []MedicinePrice    `json:"prices,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type PaginatedMedicines struct {
	Data       []MedicineResponse `json:"data"`
	TotalItems int64              `json:"total_items"`
	TotalPages int                `json:"total_pages"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
}

func (m *Medicine) ToResponse() *MedicineResponse {
	resp := &MedicineResponse{
		ID:                   m.ID,
		ItemCode:             m.ItemCode,
		Barcode:              m.Barcode,
		Name:                 m.Name,
		GenericName:          m.GenericName,
		NIE:                  m.NIE,
		Manufacturer:         m.Manufacturer,
		DosageForm:           m.DosageForm,
		DrugClassification:   m.DrugClassification,
		CategoryID:           m.CategoryID,
		BaseUnitID:           m.BaseUnitID,
		RequiresPrescription: m.RequiresPrescription,
		MinStock:             m.MinStock,
		MaxStock:             m.MaxStock,
		StorageLocation:      m.StorageLocation,
		DefaultSupplierID:    m.DefaultSupplierID,
		IsActive:             m.IsActive,
		LabelType:            m.LabelType,
		DefaultSigna:         m.DefaultSigna,
		ConsumptionTime:      m.ConsumptionTime,
		DoseMorning:          m.DoseMorning,
		DoseNoon:             m.DoseNoon,
		DoseNight:            m.DoseNight,
		WarningTags:          m.WarningTags,
		SpecialInstructions:  m.SpecialInstructions,
		BUDDays:              m.BUDDays,
		Units:                m.Units,
		Prices:               m.Prices,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}

	if m.Category != nil {
		resp.CategoryName = m.Category.Name
	}
	if m.BaseUnit != nil {
		resp.BaseUnitName = m.BaseUnit.Name
	}
	if m.DefaultSupplier != nil {
		resp.DefaultSupplierName = m.DefaultSupplier.Name
	}

	return resp
}

type MedicineRepository interface {
	FindAll(ctx context.Context, filter MedicineFilter) ([]Medicine, int64, error)
	FindByID(ctx context.Context, id string) (*Medicine, error)
	FindByItemCode(ctx context.Context, itemCode string) (*Medicine, error)
	FindByBarcode(ctx context.Context, barcode string) (*Medicine, error)
	Create(ctx context.Context, medicine *Medicine, units []MedicineUnit, prices []MedicinePrice) error
	Update(ctx context.Context, medicine *Medicine, units []MedicineUnit, prices []MedicinePrice) error
	Delete(ctx context.Context, id string) error
}

type MedicineUsecase interface {
	GetAllMedicines(ctx context.Context, filter MedicineFilter) (*PaginatedMedicines, error)
	GetMedicineByID(ctx context.Context, id string) (*MedicineResponse, error)
	GetMedicineByBarcode(ctx context.Context, barcode string) (*MedicineResponse, error)
	CreateMedicine(ctx context.Context, req *CreateMedicineRequest) (*MedicineResponse, error)
	UpdateMedicine(ctx context.Context, id string, req *UpdateMedicineRequest) (*MedicineResponse, error)
	DeleteMedicine(ctx context.Context, id string) error
}
