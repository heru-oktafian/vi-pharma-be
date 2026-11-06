package domain

import (
	"context"
	"time"
)

type SalesReturnStatus string

const (
	SalesReturnStatusCompleted SalesReturnStatus = "COMPLETED"
	SalesReturnStatusCancelled SalesReturnStatus = "CANCELLED"
)

type ReturnCondition string

const (
	ReturnConditionGood    ReturnCondition = "KONDISI_BAIK"  // Restorasi ke Rak Kasir (Bisa dijual kembali)
	ReturnConditionDamaged ReturnCondition = "KONDISI_RUSAK" // Masuk Gudang Karantina
)

type SalesReturnRefundType string

const (
	RefundCash     SalesReturnRefundType = "REFUND_KAS"      // Uang tunai kasir dikembalikan ke pasien
	RefundTransfer SalesReturnRefundType = "REFUND_TRANSFER" // Transfer kembali ke rekening pasien
	RefundExchange SalesReturnRefundType = "TUKAR_BARANG"    // Tukar barang seharga
)

// SalesReturn represents customer return slip for POS transactions
type SalesReturn struct {
	ID            string                `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID      string                `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch        *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	ReturnNumber  string                `gorm:"type:varchar(50);not null;uniqueIndex" json:"return_number"` // Format: RET/POS/20261106/0001
	SaleID        string                `gorm:"type:varchar(15);not null;index" json:"sale_id"`
	Sale          *Sale                 `gorm:"foreignKey:SaleID" json:"sale,omitempty"`
	InvoiceNumber string                `gorm:"type:varchar(50);not null;index" json:"invoice_number"` // Nomor struk POS asal
	ReturnDate    time.Time             `gorm:"type:date;not null" json:"return_date"`
	CustomerName  string                `gorm:"type:varchar(100);not null" json:"customer_name"`
	Reason        string                `gorm:"type:varchar(100);not null" json:"reason"` // SALAH_BELI, PEMBATALAN_RESEP, KEMASAN_RUSAK, TIDAK_COCOK, LAINNYA
	RefundType    SalesReturnRefundType `gorm:"type:varchar(30);not null;default:'REFUND_KAS'" json:"refund_type"`
	TotalItems    int                   `gorm:"not null;default:0" json:"total_items"`
	TotalAmount   float64               `gorm:"type:numeric(15,2);not null;default:0" json:"total_amount"` // Total dana refund
	Notes         string                `gorm:"type:text" json:"notes"`
	CashierID     string                `gorm:"type:varchar(15);not null" json:"cashier_id"`
	Cashier       *User                 `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	Status        SalesReturnStatus     `gorm:"type:varchar(20);not null;default:'COMPLETED'" json:"status"`
	Items         []SalesReturnItem     `gorm:"foreignKey:SalesReturnID" json:"items,omitempty"`
	CreatedAt     time.Time             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
}

// SalesReturnItem represents an individual medicine line being returned by the customer
type SalesReturnItem struct {
	ID               string          `gorm:"type:varchar(15);primaryKey" json:"id"`
	SalesReturnID    string          `gorm:"type:varchar(15);not null;index" json:"sales_return_id"`
	SaleItemID       string          `gorm:"type:varchar(15);not null;index" json:"sale_item_id"`
	SaleItem         *SaleItem       `gorm:"foreignKey:SaleItemID" json:"sale_item,omitempty"`
	MedicineID       string          `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine         *Medicine       `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID          string          `gorm:"type:varchar(15);not null;index" json:"batch_id"`
	Batch            *MedicineBatch  `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	UnitID           string          `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit             *Unit           `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	ConversionFactor int             `gorm:"not null;default:1" json:"conversion_factor"`
	Qty              float64         `gorm:"type:numeric(12,2);not null" json:"qty"`
	QtyBase          float64         `gorm:"type:numeric(12,2);not null" json:"qty_base"`
	UnitPrice        float64         `gorm:"type:numeric(15,2);not null;default:0" json:"unit_price"`
	Subtotal         float64         `gorm:"type:numeric(15,2);not null;default:0" json:"subtotal"`
	Condition        ReturnCondition `gorm:"type:varchar(20);not null;default:'KONDISI_BAIK'" json:"condition"` // KONDISI_BAIK (masuk Rak) atau KONDISI_RUSAK (masuk Gudang Karantina)
	ItemNotes        string          `gorm:"type:varchar(255)" json:"item_notes,omitempty"`
	CreatedAt        time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

// DTOs
type CreateSalesReturnRequest struct {
	SaleID     string                 `json:"sale_id" validate:"required"`
	ReturnDate string                 `json:"return_date" validate:"required"` // Format: YYYY-MM-DD
	Reason     string                 `json:"reason" validate:"required"`      // SALAH_BELI, PEMBATALAN_RESEP, KEMASAN_RUSAK, TIDAK_COCOK, LAINNYA
	RefundType SalesReturnRefundType  `json:"refund_type" validate:"required"` // REFUND_KAS, REFUND_TRANSFER, TUKAR_BARANG
	Notes      string                 `json:"notes"`
	Items      []SalesReturnItemInput `json:"items" validate:"required,min=1"`
}

type SalesReturnItemInput struct {
	SaleItemID string          `json:"sale_item_id" validate:"required"`
	Qty        float64         `json:"qty" validate:"required,gt=0"`
	Condition  ReturnCondition `json:"condition"` // KONDISI_BAIK (default) atau KONDISI_RUSAK
	ItemNotes  string          `json:"item_notes"`
}

type SalesReturnFilter struct {
	Search     string                `json:"search"`
	RefundType SalesReturnRefundType `json:"refund_type"`
	StartDate  string                `json:"start_date"`
	EndDate    string                `json:"end_date"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
}

type PaginatedSalesReturns struct {
	Data       []SalesReturn `json:"data"`
	TotalItems int64         `json:"total_items"`
	TotalPages int           `json:"total_pages"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
}

type SaleReturnableResponse struct {
	Sale           *Sale              `json:"sale"`
	ReturnedQtys   map[string]float64 `json:"returned_qtys"`   // sale_item_id -> total qty already returned
	RemainingQtys  map[string]float64 `json:"remaining_qtys"`  // sale_item_id -> max qty eligible for return
	IsFullyReturned bool              `json:"is_fully_returned"`
}

// Interfaces
type SalesReturnRepository interface {
	CreateSalesReturn(ctx context.Context, ret *SalesReturn) error
	FindSalesReturnByID(ctx context.Context, id string) (*SalesReturn, error)
	FindSalesReturns(ctx context.Context, branchID string, filter SalesReturnFilter) ([]SalesReturn, int64, error)
	GetReturnedQtyBySaleItem(ctx context.Context, saleItemID string) (float64, error)
	GetReturnedQtysBySaleID(ctx context.Context, saleID string) (map[string]float64, error)
	GenerateReturnNumber(ctx context.Context, branchID string) (string, error)
}

type SalesReturnUsecase interface {
	ProcessReturn(ctx context.Context, branchID, cashierID string, req CreateSalesReturnRequest) (*SalesReturn, error)
	GetSalesReturns(ctx context.Context, branchID string, filter SalesReturnFilter) (*PaginatedSalesReturns, error)
	GetSalesReturnByID(ctx context.Context, id string) (*SalesReturn, error)
	GetSaleReturnableInfo(ctx context.Context, branchID, saleID string) (*SaleReturnableResponse, error)
}
