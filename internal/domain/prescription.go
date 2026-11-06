package domain

import (
	"context"
	"time"
)

type PrescriptionStatus string

const (
	PrescriptionStatusDraft      PrescriptionStatus = "DRAFT"
	PrescriptionStatusScreening  PrescriptionStatus = "SCREENING"
	PrescriptionStatusRacik      PrescriptionStatus = "RACIK"
	PrescriptionStatusSiap       PrescriptionStatus = "SIAP"
	PrescriptionStatusSelesai    PrescriptionStatus = "SELESAI"
	PrescriptionStatusBatal      PrescriptionStatus = "BATAL"
)

type PrescriptionItemType string

const (
	PrescriptionItemNonRacik PrescriptionItemType = "NON_RACIK"
	PrescriptionItemRacikan  PrescriptionItemType = "RACIKAN"
)

type RacikanMethod string

const (
	RacikanMethodDTD    RacikanMethod = "DTD"     // Da Tales Doses (dosis per bungkus)
	RacikanMethodNonDTD RacikanMethod = "NON_DTD" // Divisio (total dosis dibagi n)
)

// Prescription represents a medical doctor prescription handled in the pharmacy
type Prescription struct {
	ID                 string             `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID           string             `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch             *Branch            `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	PrescriptionNumber string             `gorm:"type:varchar(50);not null;uniqueIndex" json:"prescription_number"`
	DoctorID           *string            `gorm:"type:varchar(15);index" json:"doctor_id"`
	Doctor             *Doctor            `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	DoctorName         string             `gorm:"type:varchar(150);not null" json:"doctor_name"`
	DoctorSIP          string             `gorm:"type:varchar(100)" json:"doctor_sip"`
	PatientName        string             `gorm:"type:varchar(150);not null;index" json:"patient_name"`
	PatientAge         int                `gorm:"not null;default:0" json:"patient_age"`
	PatientWeight      float64            `gorm:"type:numeric(6,2);default:0" json:"patient_weight"` // in kg (pediatric check)
	PatientGender      string             `gorm:"type:varchar(10);default:'L'" json:"patient_gender"` // L / P
	PatientPhone       string             `gorm:"type:varchar(30)" json:"patient_phone"`
	PatientAddress     string             `gorm:"type:text" json:"patient_address"`
	PrescriptionDate   time.Time          `gorm:"not null" json:"prescription_date"`
	Status             PrescriptionStatus `gorm:"type:varchar(20);not null;default:'SCREENING';index" json:"status"`
	
	// Clinical & Administrative Screening Checklist
	ScreeningAdmin     bool               `gorm:"not null;default:true" json:"screening_admin"`      // Keabsahan resep & SIP
	ScreeningPharma    bool               `gorm:"not null;default:true" json:"screening_pharma"`     // Bentuk sediaan & kompatibilitas
	ScreeningClinical  bool               `gorm:"not null;default:true" json:"screening_clinical"`   // Dosis, alergi, duplikasi
	ScreeningPassed    bool               `gorm:"not null;default:true" json:"screening_passed"`
	ScreeningNotes     string             `gorm:"type:text" json:"screening_notes"`
	PharmacistID       *string            `gorm:"type:varchar(15);index" json:"pharmacist_id"`
	Pharmacist         *User              `gorm:"foreignKey:PharmacistID" json:"pharmacist,omitempty"`

	// Financials
	TotalMedPrice      float64            `gorm:"type:numeric(15,2);not null;default:0" json:"total_med_price"`
	TotalEmbalase      float64            `gorm:"type:numeric(15,2);not null;default:0" json:"total_embalase"`
	TotalTuslah        float64            `gorm:"type:numeric(15,2);not null;default:0" json:"total_tuslah"`
	GrandTotal         float64            `gorm:"type:numeric(15,2);not null;default:0" json:"grand_total"`
	
	// POS Integration
	IsPaid             bool               `gorm:"not null;default:false;index" json:"is_paid"`
	SaleID             *string            `gorm:"type:varchar(15);index" json:"sale_id"`
	Sale               *Sale              `gorm:"foreignKey:SaleID" json:"sale,omitempty"`

	Notes              string             `gorm:"type:text" json:"notes"`
	Items              []PrescriptionItem `gorm:"foreignKey:PrescriptionID" json:"items,omitempty"`
	CreatedAt          time.Time          `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
}

// PrescriptionItem represents an item in a prescription (Non-racik or Racikan)
type PrescriptionItem struct {
	ID                  string                     `gorm:"type:varchar(15);primaryKey" json:"id"`
	PrescriptionID      string                     `gorm:"type:varchar(15);not null;index" json:"prescription_id"`
	ItemType            PrescriptionItemType       `gorm:"type:varchar(20);not null;default:'NON_RACIK'" json:"item_type"` // NON_RACIK / RACIKAN
	
	// If NON_RACIK:
	MedicineID          *string                    `gorm:"type:varchar(15);index" json:"medicine_id"`
	Medicine            *Medicine                  `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID             *string                    `gorm:"type:varchar(15);index" json:"batch_id"`
	Batch               *MedicineBatch             `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	UnitID              *string                    `gorm:"type:varchar(15)" json:"unit_id"`
	Unit                *Unit                      `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Qty                 float64                    `gorm:"type:numeric(12,2);not null;default:1" json:"qty"`
	UnitPrice           float64                    `gorm:"type:numeric(15,2);not null;default:0" json:"unit_price"`

	// If RACIKAN:
	RacikanName         string                     `gorm:"type:varchar(150)" json:"racikan_name"` // e.g. "Puyer Batuk Pilek Anak"
	RacikanDoseForm     string                     `gorm:"type:varchar(50)" json:"racikan_dose_form"` // Puyer, Kapsul, Salep, Sirup
	RacikanQty          int                        `gorm:"not null;default:1" json:"racikan_qty"` // e.g. 10 bungkus
	RacikanUnit         string                     `gorm:"type:varchar(30);default:'Bungkus'" json:"racikan_unit"`
	RacikanMethod       RacikanMethod              `gorm:"type:varchar(20);default:'DTD'" json:"racikan_method"` // DTD / NON_DTD
	EmbalasePrice       float64                    `gorm:"type:numeric(15,2);not null;default:0" json:"embalase_price"`
	TuslahPrice         float64                    `gorm:"type:numeric(15,2);not null;default:0" json:"tuslah_price"`
	CompoundItems       []PrescriptionCompoundItem `gorm:"foreignKey:PrescriptionItemID" json:"compound_items,omitempty"`

	// Signa & Etiket Information
	LabelType           string                     `gorm:"type:varchar(20);not null;default:'dalam'" json:"label_type"` // dalam / luar
	Signa               string                     `gorm:"type:varchar(150);not null" json:"signa"` // e.g. "3 x Sehari 1 Bungkus"
	UsageTime           string                     `gorm:"type:varchar(100)" json:"usage_time"` // e.g. "Sesudah Makan"
	DoseMorning         string                     `gorm:"type:varchar(50);default:'1'" json:"dose_morning"`
	DoseNoon            string                     `gorm:"type:varchar(50);default:'1'" json:"dose_noon"`
	DoseNight           string                     `gorm:"type:varchar(50);default:'1'" json:"dose_night"`
	WarningTags         string                     `gorm:"type:text" json:"warning_tags"`
	SpecialInstructions string                     `gorm:"type:text" json:"special_instructions"`
	BUDDays             int                        `gorm:"not null;default:0" json:"bud_days"` // Beyond Use Date (days)
	
	// Copy Receipt Marker
	DeturStatus         string                     `gorm:"type:varchar(20);not null;default:'det'" json:"detur_status"` // det / ne_det / did (da in dimidio)

	Subtotal            float64                    `gorm:"type:numeric(15,2);not null;default:0" json:"subtotal"`
	Total               float64                    `gorm:"type:numeric(15,2);not null;default:0" json:"total"`
	CreatedAt           time.Time                  `gorm:"autoCreateTime" json:"created_at"`
}

// PrescriptionCompoundItem represents an active ingredient tablet/powder used in a compound racikan
type PrescriptionCompoundItem struct {
	ID                 string         `gorm:"type:varchar(15);primaryKey" json:"id"`
	PrescriptionItemID string         `gorm:"type:varchar(15);not null;index" json:"prescription_item_id"`
	MedicineID         string         `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine           *Medicine      `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID            *string        `gorm:"type:varchar(15);index" json:"batch_id"`
	Batch              *MedicineBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	UnitID             string         `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit               *Unit          `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	DosePerPack        float64        `gorm:"type:numeric(10,3);not null;default:1" json:"dose_per_pack"` // e.g. 0.5 tab per bungkus
	TotalQty           float64        `gorm:"type:numeric(12,2);not null;default:1" json:"total_qty"`     // e.g. 0.5 * 10 = 5 tab
	UnitPrice          float64        `gorm:"type:numeric(15,2);not null;default:0" json:"unit_price"`
	Subtotal           float64        `gorm:"type:numeric(15,2);not null;default:0" json:"subtotal"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// DTOs
type CreatePrescriptionCompoundItemRequest struct {
	MedicineID  string  `json:"medicine_id" validate:"required"`
	BatchID     *string `json:"batch_id"`
	UnitID      string  `json:"unit_id" validate:"required"`
	DosePerPack float64 `json:"dose_per_pack" validate:"required,gt=0"`
	TotalQty    float64 `json:"total_qty" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"gte=0"`
}

type CreatePrescriptionItemRequest struct {
	ItemType            PrescriptionItemType                    `json:"item_type" validate:"required"` // NON_RACIK / RACIKAN
	MedicineID          *string                                 `json:"medicine_id"`
	BatchID             *string                                 `json:"batch_id"`
	UnitID              *string                                 `json:"unit_id"`
	Qty                 float64                                 `json:"qty"`
	UnitPrice           float64                                 `json:"unit_price"`
	RacikanName         string                                  `json:"racikan_name"`
	RacikanDoseForm     string                                  `json:"racikan_dose_form"`
	RacikanQty          int                                     `json:"racikan_qty"`
	RacikanUnit         string                                  `json:"racikan_unit"`
	RacikanMethod       RacikanMethod                           `json:"racikan_method"`
	EmbalasePrice       float64                                 `json:"embalase_price"`
	TuslahPrice         float64                                 `json:"tuslah_price"`
	CompoundItems       []CreatePrescriptionCompoundItemRequest `json:"compound_items"`
	LabelType           string                                  `json:"label_type"`
	Signa               string                                  `json:"signa" validate:"required"`
	UsageTime           string                                  `json:"usage_time"`
	DoseMorning         string                                  `json:"dose_morning"`
	DoseNoon            string                                  `json:"dose_noon"`
	DoseNight           string                                  `json:"dose_night"`
	WarningTags         string                                  `json:"warning_tags"`
	SpecialInstructions string                                  `json:"special_instructions"`
	BUDDays             int                                     `json:"bud_days"`
	DeturStatus         string                                  `json:"detur_status"`
}

type CreatePrescriptionRequest struct {
	DoctorID          *string                         `json:"doctor_id"`
	DoctorName        string                          `json:"doctor_name" validate:"required"`
	DoctorSIP         string                          `json:"doctor_sip"`
	PatientName       string                          `json:"patient_name" validate:"required"`
	PatientAge        int                             `json:"patient_age"`
	PatientWeight     float64                         `json:"patient_weight"`
	PatientGender     string                          `json:"patient_gender"`
	PatientPhone      string                          `json:"patient_phone"`
	PatientAddress    string                          `json:"patient_address"`
	PrescriptionDate  string                          `json:"prescription_date"` // YYYY-MM-DD
	ScreeningAdmin    bool                            `json:"screening_admin"`
	ScreeningPharma   bool                            `json:"screening_pharma"`
	ScreeningClinical bool                            `json:"screening_clinical"`
	ScreeningNotes    string                          `json:"screening_notes"`
	Notes             string                          `json:"notes"`
	Items             []CreatePrescriptionItemRequest `json:"items" validate:"required,min=1"`
}

type UpdatePrescriptionStatusRequest struct {
	Status            PrescriptionStatus `json:"status"`
	IsPaid            *bool              `json:"is_paid"`
	SaleID            *string            `json:"sale_id"`
	ScreeningAdmin    *bool              `json:"screening_admin"`
	ScreeningPharma   *bool              `json:"screening_pharma"`
	ScreeningClinical *bool              `json:"screening_clinical"`
	ScreeningPassed   *bool              `json:"screening_passed"`
	ScreeningNotes    *string            `json:"screening_notes"`
	PharmacistID      *string            `json:"pharmacist_id"`
	Notes             *string            `json:"notes"`
}

type PrescriptionFilter struct {
	BranchID  string `json:"branch_id"`
	Status    string `json:"status"`
	Search    string `json:"search"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsPaid    *bool  `json:"is_paid"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
}

type PaginatedPrescriptions struct {
	Data       []Prescription `json:"data"`
	TotalItems int64          `json:"total_items"`
	TotalPages int            `json:"total_pages"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// Repository & Usecase Interfaces
type PrescriptionRepository interface {
	Create(ctx context.Context, prescription *Prescription) error
	FindByID(ctx context.Context, id string) (*Prescription, error)
	FindByNumber(ctx context.Context, number string) (*Prescription, error)
	FindAll(ctx context.Context, filter PrescriptionFilter) ([]Prescription, int64, error)
	Update(ctx context.Context, prescription *Prescription) error
	UpdateStatus(ctx context.Context, id string, status PrescriptionStatus, pharmacistID *string, notes string) error
	GetNextPrescriptionNumber(ctx context.Context, branchID string, date time.Time) (string, error)
}

type PrescriptionUsecase interface {
	CreatePrescription(ctx context.Context, branchID string, req *CreatePrescriptionRequest) (*Prescription, error)
	GetPrescriptions(ctx context.Context, filter PrescriptionFilter) (*PaginatedPrescriptions, error)
	GetPrescriptionByID(ctx context.Context, id string) (*Prescription, error)
	UpdateStatus(ctx context.Context, id string, req *UpdatePrescriptionStatusRequest) (*Prescription, error)
	DispensePrescription(ctx context.Context, id string, userID string) (*Prescription, error)
}
