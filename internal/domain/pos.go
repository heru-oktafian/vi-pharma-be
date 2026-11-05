package domain

import (
	"context"
	"time"
)

// PaymentMethod represents accepted POS cashier payment methods
type PaymentMethod string

const (
	PaymentMethodTunai    PaymentMethod = "TUNAI"
	PaymentMethodQRIS     PaymentMethod = "QRIS"
	PaymentMethodDebit    PaymentMethod = "DEBIT"
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
)

// SaleType distinguishes standard OTC sales from prescription dispensing
type SaleType string

const (
	SaleTypeOTC   SaleType = "OTC"
	SaleTypeResep SaleType = "RESEP"
)

// Sale represents a POS sales transaction record
type Sale struct {
	ID              string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID        string        `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch          *Branch       `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	InvoiceNumber   string        `gorm:"type:varchar(50);not null;uniqueIndex" json:"invoice_number"`
	CashierID       string        `gorm:"type:varchar(15);not null;index" json:"cashier_id"`
	Cashier         *User         `gorm:"foreignKey:CashierID" json:"cashier,omitempty"`
	CustomerName    string        `gorm:"type:varchar(100);not null;default:'Pelanggan Umum'" json:"customer_name"`
	CustomerPhone   string        `gorm:"type:varchar(30)" json:"customer_phone"`
	TransactionType SaleType      `gorm:"type:varchar(20);not null;default:'OTC'" json:"transaction_type"`
	Subtotal        float64       `gorm:"type:numeric(15,2);not null;default:0" json:"subtotal"`
	DiscountAmount  float64       `gorm:"type:numeric(15,2);not null;default:0" json:"discount_amount"`
	TaxAmount       float64       `gorm:"type:numeric(15,2);not null;default:0" json:"tax_amount"`
	TotalAmount     float64       `gorm:"type:numeric(15,2);not null;default:0" json:"total_amount"`
	PaymentMethod   PaymentMethod `gorm:"type:varchar(20);not null;default:'TUNAI'" json:"payment_method"`
	AmountPaid      float64       `gorm:"type:numeric(15,2);not null;default:0" json:"amount_paid"`
	ChangeAmount    float64       `gorm:"type:numeric(15,2);not null;default:0" json:"change_amount"`
	PaymentStatus   PaymentStatus `gorm:"type:varchar(20);not null;default:'PAID'" json:"payment_status"`
	Notes           string        `gorm:"type:text" json:"notes"`
	Items           []SaleItem    `gorm:"foreignKey:SaleID" json:"items,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// SaleItem represents individual line items within a sales transaction
type SaleItem struct {
	ID               string         `gorm:"type:varchar(15);primaryKey" json:"id"`
	SaleID           string         `gorm:"type:varchar(15);not null;index" json:"sale_id"`
	MedicineID       string         `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine         *Medicine      `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID          string         `gorm:"type:varchar(15);not null;index" json:"batch_id"`
	Batch            *MedicineBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	UnitID           string         `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit             *Unit          `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	ConversionFactor int            `gorm:"not null;default:1" json:"conversion_factor"`
	Qty              float64        `gorm:"type:numeric(12,2);not null;default:1" json:"qty"`
	QtyBase          float64        `gorm:"type:numeric(12,2);not null;default:1" json:"qty_base"` // Qty in base unit
	UnitPrice        float64        `gorm:"type:numeric(15,2);not null;default:0" json:"unit_price"`
	Subtotal         float64        `gorm:"type:numeric(15,2);not null;default:0" json:"subtotal"`
	Discount         float64        `gorm:"type:numeric(15,2);not null;default:0" json:"discount"`
	Total            float64        `gorm:"type:numeric(15,2);not null;default:0" json:"total"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// DTOs
type CreateSaleItemRequest struct {
	MedicineID       string  `json:"medicine_id" validate:"required"`
	BatchID          string  `json:"batch_id"` // Optional: if empty, backend uses FEFO to pick batch
	UnitID           string  `json:"unit_id" validate:"required"`
	ConversionFactor int     `json:"conversion_factor"`
	Qty              float64 `json:"qty" validate:"required,gt=0"`
	UnitPrice        float64 `json:"unit_price" validate:"required,gte=0"`
	Discount         float64 `json:"discount"`
}

type CreateSaleRequest struct {
	CustomerName    string                  `json:"customer_name"`
	CustomerPhone   string                  `json:"customer_phone"`
	TransactionType SaleType                `json:"transaction_type"`
	DiscountAmount  float64                 `json:"discount_amount"`
	PaymentMethod   PaymentMethod           `json:"payment_method" validate:"required"`
	AmountPaid      float64                 `json:"amount_paid"`
	Notes           string                  `json:"notes"`
	Items           []CreateSaleItemRequest `json:"items" validate:"required,min=1"`
}

type SaleFilter struct {
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	PaymentMethod string `json:"payment_method"`
	Search        string `json:"search"`
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
}

type PaginatedSales struct {
	Data       []Sale `json:"data"`
	TotalItems int64  `json:"total_items"`
	TotalPages int    `json:"total_pages"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type DailyCashierSummaryResponse struct {
	Date              string  `json:"date"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalCash         float64 `json:"total_cash"`
	TotalNonCash      float64 `json:"total_non_cash"`
	TotalItemsSold    float64 `json:"total_items_sold"`
}

// Interfaces
type POSRepository interface {
	CreateSale(ctx context.Context, sale *Sale) error
	FindSales(ctx context.Context, branchID string, filter SaleFilter) ([]Sale, int64, error)
	FindSaleByID(ctx context.Context, id string) (*Sale, error)
	GetTodaySummary(ctx context.Context, branchID string) (*DailyCashierSummaryResponse, error)
	GenerateInvoiceNumber(ctx context.Context, branchID string) (string, error)
}

type POSUsecase interface {
	Checkout(ctx context.Context, branchID, cashierID string, req CreateSaleRequest) (*Sale, error)
	GetSales(ctx context.Context, branchID string, filter SaleFilter) (*PaginatedSales, error)
	GetSaleByID(ctx context.Context, id string) (*Sale, error)
	GetTodaySummary(ctx context.Context, branchID string) (*DailyCashierSummaryResponse, error)
}
