package domain

import (
	"context"
	"time"
)

// SPType represents types of Indonesian pharmacy official Surat Pesanan
type SPType string

const (
	SPTypeReguler      SPType = "REGULER"       // Obat bebas, bebas terbatas, obat keras standar, alkes
	SPTypePrecursor    SPType = "PREKURSOR"     // Mengandung zat prekursor (Pseudoephedrine, Ephedrine, PPA)
	SPTypeOOT          SPType = "OOT"           // Obat-Obat Tertentu (Tramadol, Trihexyphenidyl, dll.)
	SPTypePsychotropic SPType = "PSIKOTROPIKA"  // Psikotropika (Alprazolam, Diazepam, dll.)
	SPTypeNarcotic     SPType = "NARKOTIKA"     // Narkotika (Kodein, Morfin, Fentanil - 1 item per SP)
)

// POStatus represents the lifecycle of a Purchase Order
type POStatus string

const (
	POStatusDraft     POStatus = "DRAFT"     // Rancangan pesanan
	POStatusOrdered   POStatus = "ORDERED"   // Telah dicetak/dikirimkan ke PBF
	POStatusPartial   POStatus = "PARTIAL"   // Sebagian barang telah diterima via faktur
	POStatusCompleted POStatus = "COMPLETED" // Seluruh pesanan telah diterima lengkap
	POStatusCancelled POStatus = "CANCELLED" // Pesanan dibatalkan
)

// PurchaseOrder represents an official pharmacy Purchase Order / Surat Pesanan (SP)
type PurchaseOrder struct {
	ID                    string              `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID              string              `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch                *Branch             `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	PONumber              string              `gorm:"type:varchar(100);not null;uniqueIndex" json:"po_number"` // Format: SP/REG/202611/0001
	OrderDate             time.Time           `gorm:"type:date;not null" json:"order_date"`
	EstimatedDeliveryDate *time.Time          `gorm:"type:date" json:"estimated_delivery_date,omitempty"`
	SupplierID            string              `gorm:"type:varchar(15);not null;index" json:"supplier_id"`
	Supplier              *Supplier           `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SPType                SPType              `gorm:"type:varchar(20);not null;default:'REGULER'" json:"sp_type"`
	Status                POStatus            `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"`
	PharmacistName        string              `gorm:"type:varchar(150);not null" json:"pharmacist_name"` // Nama Apoteker Penanggung Jawab (APJ)
	PharmacistSIPA        string              `gorm:"type:varchar(100);not null" json:"pharmacist_sipa"` // Nomor SIPA resmi APJ
	TotalItems            int                 `gorm:"not null;default:0" json:"total_items"`
	TotalEstimatedAmount  float64             `gorm:"type:decimal(14,2);not null;default:0" json:"total_estimated_amount"`
	Notes                 string              `gorm:"type:text" json:"notes"`
	CreatedByID           string              `gorm:"type:varchar(15);not null" json:"created_by_id"`
	CreatedBy             *User               `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Items                 []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID" json:"items,omitempty"`
	CreatedAt             time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
}

// PurchaseOrderItem represents single drug line in an SP
type PurchaseOrderItem struct {
	ID              string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseOrderID string    `gorm:"type:varchar(15);not null;index" json:"purchase_order_id"`
	MedicineID      string    `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine        *Medicine `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	UnitID          string    `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit            *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	QtyOrdered      float64   `gorm:"type:decimal(10,2);not null" json:"qty_ordered"`
	QtyReceived     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"qty_received"`
	EstimatedPrice  float64   `gorm:"type:decimal(14,2);not null;default:0" json:"estimated_price"`
	Subtotal        float64   `gorm:"type:decimal(14,2);not null;default:0" json:"subtotal"`
	Notes           string    `gorm:"type:text" json:"notes"` // e.g. "Bentuk sediaan, kekuatan, dll"
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// -----------------------------------------------------------------------
// DTOs for Procurement Operations
// -----------------------------------------------------------------------

type CreatePurchaseOrderRequest struct {
	SupplierID            string                   `json:"supplier_id" validate:"required"`
	OrderDate             string                   `json:"order_date" validate:"required"` // Format: YYYY-MM-DD
	EstimatedDeliveryDate string                   `json:"estimated_delivery_date"`        // Format: YYYY-MM-DD
	SPType                SPType                   `json:"sp_type" validate:"required"`
	PharmacistName        string                   `json:"pharmacist_name" validate:"required"`
	PharmacistSIPA        string                   `json:"pharmacist_sipa" validate:"required"`
	Notes                 string                   `json:"notes"`
	Items                 []PurchaseOrderItemInput `json:"items" validate:"required,min=1"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID            string                   `json:"supplier_id"`
	OrderDate             string                   `json:"order_date"`
	EstimatedDeliveryDate string                   `json:"estimated_delivery_date"`
	SPType                SPType                   `json:"sp_type"`
	PharmacistName        string                   `json:"pharmacist_name"`
	PharmacistSIPA        string                   `json:"pharmacist_sipa"`
	Notes                 string                   `json:"notes"`
	Items                 []PurchaseOrderItemInput `json:"items"`
}

type UpdatePOStatusRequest struct {
	Status POStatus `json:"status" validate:"required"`
	Notes  string   `json:"notes"`
}

type PurchaseOrderItemInput struct {
	MedicineID     string  `json:"medicine_id" validate:"required"`
	UnitID         string  `json:"unit_id" validate:"required"`
	QtyOrdered     float64 `json:"qty_ordered" validate:"required,gt=0"`
	EstimatedPrice float64 `json:"estimated_price"`
	Notes          string  `json:"notes"`
}

type POFilter struct {
	Search     string
	SupplierID string
	SPType     SPType
	Status     POStatus
	StartDate  string
	EndDate    string
	Page       int
	Limit      int
}

type PaginatedPurchaseOrders struct {
	Data       []PurchaseOrder `json:"data"`
	TotalItems int64           `json:"total_items"`
	TotalPages int             `json:"total_pages"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// -----------------------------------------------------------------------
// Repository & Usecase Interfaces
// -----------------------------------------------------------------------

type ProcurementRepository interface {
	CreatePurchaseOrder(ctx context.Context, po *PurchaseOrder) error
	GetPurchaseOrderByID(ctx context.Context, id string) (*PurchaseOrder, error)
	FindAllPurchaseOrders(ctx context.Context, branchID string, filter POFilter) ([]PurchaseOrder, int64, error)
	UpdatePurchaseOrder(ctx context.Context, po *PurchaseOrder, items []PurchaseOrderItem) error
	UpdatePOStatus(ctx context.Context, id string, status POStatus) error
	DeletePurchaseOrder(ctx context.Context, id string) error
	GetNextPONumber(ctx context.Context, branchID string, spType SPType) (string, error)
}

type ProcurementUsecase interface {
	CreatePurchaseOrder(ctx context.Context, branchID, userID string, req *CreatePurchaseOrderRequest) (*PurchaseOrder, error)
	GetPurchaseOrderByID(ctx context.Context, id string) (*PurchaseOrder, error)
	ListPurchaseOrders(ctx context.Context, branchID string, filter POFilter) (*PaginatedPurchaseOrders, error)
	UpdatePurchaseOrder(ctx context.Context, branchID string, id string, req *UpdatePurchaseOrderRequest) (*PurchaseOrder, error)
	UpdatePOStatus(ctx context.Context, branchID string, id string, req *UpdatePOStatusRequest) (*PurchaseOrder, error)
	DeletePurchaseOrder(ctx context.Context, branchID string, id string) error
}
