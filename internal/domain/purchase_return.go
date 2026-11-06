package domain

import (
	"context"
	"time"
)

// PurchaseReturnStatus represents the lifecycle of a Supplier Purchase Return
type PurchaseReturnStatus string

const (
	ReturnStatusDraft     PurchaseReturnStatus = "DRAFT"     // Rancangan usulan retur
	ReturnStatusRequested PurchaseReturnStatus = "REQUESTED" // Telah diajukan / dikirim ke PBF
	ReturnStatusCompleted PurchaseReturnStatus = "COMPLETED" // Barang telah diambil PBF & stok apotek dipotong
	ReturnStatusCancelled PurchaseReturnStatus = "CANCELLED" // Retur dibatalkan
)

// CompensationType represents how PBF resolves the returned items
type CompensationType string

const (
	CompensationPotongFaktur CompensationType = "POTONG_FAKTUR" // Pemotongan saldo utang / invoice
	CompensationTukarBarang  CompensationType = "TUKAR_BARANG"  // Ganti barang batch baru
	CompensationRefundDana   CompensationType = "REFUND_DANA"   // Pengembalian dana tunai / transfer
)

// PurchaseReturn represents official pharmacy return slip to PBF
type PurchaseReturn struct {
	ID               string               `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID         string               `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch           *Branch              `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ReturnNumber     string               `gorm:"type:varchar(100);not null;uniqueIndex" json:"return_number"` // Format: RET/PBF/202611/0001
	ReturnDate       time.Time            `gorm:"type:date;not null" json:"return_date"`
	SupplierID       string               `gorm:"type:varchar(15);not null;index" json:"supplier_id"`
	Supplier         *Supplier            `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	InvoiceNumber    string               `gorm:"type:varchar(100)" json:"invoice_number,omitempty"` // Referensi nomor faktur asal (opsional)
	Status           PurchaseReturnStatus `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"`
	Reason           string               `gorm:"type:varchar(100);not null" json:"reason"` // RUSAK_KEMASAN, KADALUARSA_ED, SALAH_KIRIM, RECALL_BPOM, LAINNYA
	CompensationType CompensationType     `gorm:"type:varchar(30);not null;default:'POTONG_FAKTUR'" json:"compensation_type"`
	TotalItems       int                  `gorm:"not null;default:0" json:"total_items"`
	TotalAmount      float64              `gorm:"type:decimal(14,2);not null;default:0" json:"total_amount"`
	Notes            string               `gorm:"type:text" json:"notes"`
	CreatedByID      string               `gorm:"type:varchar(15);not null" json:"created_by_id"`
	CreatedBy        *User                `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Items            []PurchaseReturnItem `gorm:"foreignKey:PurchaseReturnID" json:"items,omitempty"`
	CreatedAt        time.Time            `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time            `gorm:"autoUpdateTime" json:"updated_at"`
}

// PurchaseReturnItem represents a single medicine batch line being returned to PBF
type PurchaseReturnItem struct {
	ID               string         `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseReturnID string         `gorm:"type:varchar(15);not null;index" json:"purchase_return_id"`
	MedicineID       string         `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine         *Medicine      `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID          string         `gorm:"type:varchar(15);not null;index" json:"batch_id"`
	Batch            *MedicineBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	UnitID           string         `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit             *Unit          `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	Location         StockLocation  `gorm:"type:varchar(20);not null;default:'GUDANG'" json:"location"` // GUDANG atau RAK
	ConversionFactor float64        `gorm:"type:decimal(10,2);not null;default:1" json:"conversion_factor"`
	Qty              float64        `gorm:"type:decimal(10,2);not null" json:"qty"`
	QtyBaseUnit      float64        `gorm:"type:decimal(12,2);not null" json:"qty_base_unit"`
	UnitPrice        float64        `gorm:"type:decimal(14,2);not null;default:0" json:"unit_price"`
	Subtotal         float64        `gorm:"type:decimal(14,2);not null;default:0" json:"subtotal"`
	ItemReason       string         `gorm:"type:varchar(255)" json:"item_reason,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// -----------------------------------------------------------------------
// DTOs for Purchase Return
// -----------------------------------------------------------------------

type CreatePurchaseReturnRequest struct {
	SupplierID       string                    `json:"supplier_id" validate:"required"`
	ReturnDate       string                    `json:"return_date" validate:"required"` // Format: YYYY-MM-DD
	InvoiceNumber    string                    `json:"invoice_number"`
	Reason           string                    `json:"reason" validate:"required"`
	CompensationType CompensationType          `json:"compensation_type" validate:"required"`
	Notes            string                    `json:"notes"`
	Items            []PurchaseReturnItemInput `json:"items" validate:"required,min=1"`
}

type PurchaseReturnItemInput struct {
	MedicineID       string        `json:"medicine_id" validate:"required"`
	BatchID          string        `json:"batch_id" validate:"required"`
	UnitID           string        `json:"unit_id" validate:"required"`
	Location         StockLocation `json:"location"` // default GUDANG
	ConversionFactor float64       `json:"conversion_factor"`
	Qty              float64       `json:"qty" validate:"required,gt=0"`
	UnitPrice        float64       `json:"unit_price" validate:"gte=0"`
	ItemReason       string        `json:"item_reason"`
}

type UpdatePurchaseReturnRequest struct {
	SupplierID       string                    `json:"supplier_id"`
	ReturnDate       string                    `json:"return_date"`
	InvoiceNumber    string                    `json:"invoice_number"`
	Reason           string                    `json:"reason"`
	CompensationType CompensationType          `json:"compensation_type"`
	Notes            string                    `json:"notes"`
	Items            []PurchaseReturnItemInput `json:"items"`
}

type UpdateReturnStatusRequest struct {
	Status PurchaseReturnStatus `json:"status" validate:"required"`
	Notes  string               `json:"notes"`
}

type PurchaseReturnFilter struct {
	Search           string               `json:"search"`
	SupplierID       string               `json:"supplier_id"`
	Status           PurchaseReturnStatus `json:"status"`
	Reason           string               `json:"reason"`
	CompensationType CompensationType     `json:"compensation_type"`
	StartDate        string               `json:"start_date"`
	EndDate          string               `json:"end_date"`
	Page             int                  `json:"page"`
	Limit            int                  `json:"limit"`
}

type PaginatedPurchaseReturns struct {
	Data       []PurchaseReturn `json:"data"`
	TotalItems int64            `json:"total_items"`
	TotalPages int              `json:"total_pages"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
}

// -----------------------------------------------------------------------
// Repository & Usecase Interfaces
// -----------------------------------------------------------------------

type PurchaseReturnRepository interface {
	CreatePurchaseReturn(ctx context.Context, ret *PurchaseReturn) error
	GetPurchaseReturnByID(ctx context.Context, id string) (*PurchaseReturn, error)
	FindAllPurchaseReturns(ctx context.Context, branchID string, filter PurchaseReturnFilter) ([]PurchaseReturn, int64, error)
	UpdatePurchaseReturn(ctx context.Context, ret *PurchaseReturn, items []PurchaseReturnItem) error
	UpdateReturnStatus(ctx context.Context, id string, status PurchaseReturnStatus, notes string) error
	ProcessStockDeduction(ctx context.Context, ret *PurchaseReturn, userID string) error
	DeletePurchaseReturn(ctx context.Context, id string) error
	GetNextReturnNumber(ctx context.Context, branchID string) (string, error)
}

type PurchaseReturnUsecase interface {
	CreatePurchaseReturn(ctx context.Context, branchID, userID string, req *CreatePurchaseReturnRequest) (*PurchaseReturn, error)
	GetPurchaseReturnByID(ctx context.Context, id string) (*PurchaseReturn, error)
	ListPurchaseReturns(ctx context.Context, branchID string, filter PurchaseReturnFilter) (*PaginatedPurchaseReturns, error)
	UpdatePurchaseReturn(ctx context.Context, branchID string, id string, req *UpdatePurchaseReturnRequest) (*PurchaseReturn, error)
	UpdateReturnStatus(ctx context.Context, branchID, userID string, id string, req *UpdateReturnStatusRequest) (*PurchaseReturn, error)
	DeletePurchaseReturn(ctx context.Context, branchID, id string) error
}
