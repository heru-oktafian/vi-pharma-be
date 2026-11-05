package domain

import (
	"context"
	"time"
)

// MutationType represents kinds of stock movements
type MutationType string

const (
	MutationPurchase          MutationType = "PEMBELIAN"       // Masuk dari PBF
	MutationSale              MutationType = "PENJUALAN"       // Keluar dari Kasir POS
	MutationPrescription      MutationType = "RESEP"           // Keluar dari Resep
	MutationRefillShelf       MutationType = "REFILL_RAK"      // Transfer internal: Gudang -> Rak
	MutationReturnToWarehouse MutationType = "KEMBALI_GUDANG"  // Transfer internal: Rak -> Gudang
	MutationReturnSupplier    MutationType = "RETUR_PBF"       // Keluar: Retur ke PBF
	MutationAdjustment        MutationType = "OPNAME"          // Koreksi Opname Fisik
)

// StockLocation represents the physical zone in the pharmacy
type StockLocation string

const (
	LocationWarehouse StockLocation = "GUDANG"
	LocationShelf     StockLocation = "RAK"
)

// PaymentType represents purchase payment terms
type PaymentType string

const (
	PaymentCash     PaymentType = "CASH"
	PaymentCredit   PaymentType = "CREDIT"
	PaymentTransfer PaymentType = "TRANSFER"
)

// PaymentStatus represents purchase payment settlement status
type PaymentStatus string

const (
	PaymentStatusPaid      PaymentStatus = "PAID"
	PaymentStatusUnpaid    PaymentStatus = "UNPAID"
	PaymentStatusPartial   PaymentStatus = "PARTIAL"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
)

// OpnameStatus represents cycle counting status
type OpnameStatus string

const (
	OpnameDraft     OpnameStatus = "DRAFT"
	OpnameCompleted OpnameStatus = "COMPLETED"
	OpnameCancelled OpnameStatus = "CANCELLED"
)

// MedicineBatch represents stock of a medicine segmented by Batch Number and Expired Date
type MedicineBatch struct {
	ID           string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	MedicineID   string    `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine     *Medicine `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BranchID     string    `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	BatchNumber  string    `gorm:"type:varchar(100);not null;index" json:"batch_number"`
	ExpiredDate  time.Time `gorm:"type:date;not null;index" json:"expired_date"`
	QtyWarehouse float64   `gorm:"type:decimal(12,2);not null;default:0" json:"qty_warehouse"` // Stok di gudang belakang
	QtyShelf     float64   `gorm:"type:decimal(12,2);not null;default:0" json:"qty_shelf"`     // Stok di rak kasir/dispensing
	TotalStock   float64   `gorm:"type:decimal(12,2);not null;default:0" json:"total_stock"`   // Total = Gudang + Rak
	BuyPrice     float64   `gorm:"type:decimal(14,2);not null;default:0" json:"buy_price"`     // HPP per base unit
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (b *MedicineBatch) CalculateTotal() {
	b.TotalStock = b.QtyWarehouse + b.QtyShelf
}

// StockMutation represents an immutable audit entry in the Digital Stock Card
type StockMutation struct {
	ID           string        `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID     string        `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	MedicineID   string        `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine     *Medicine     `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID      string        `gorm:"type:varchar(15);not null;index" json:"batch_id"`
	Batch        *MedicineBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	MutationType MutationType  `gorm:"type:varchar(30);not null;index" json:"mutation_type"`
	Location     StockLocation `gorm:"type:varchar(20);not null" json:"location"` // GUDANG atau RAK
	ReferenceNo  string        `gorm:"type:varchar(100);index" json:"reference_no"`
	QtyIn        float64       `gorm:"type:decimal(12,2);not null;default:0" json:"qty_in"`
	QtyOut       float64       `gorm:"type:decimal(12,2);not null;default:0" json:"qty_out"`
	BalanceAfter float64       `gorm:"type:decimal(12,2);not null;default:0" json:"balance_after"` // Saldo spesifik di lokasi tsb
	TotalAfter   float64       `gorm:"type:decimal(12,2);not null;default:0" json:"total_after"`   // Saldo gabungan (Gudang + Rak)
	Notes        string        `gorm:"type:text" json:"notes"`
	UserID       string        `gorm:"type:varchar(15);not null" json:"user_id"`
	User         *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt    time.Time     `gorm:"autoCreateTime;index" json:"created_at"`
}

// PurchaseInvoice represents incoming PBF invoice / Faktur Pembelian
type PurchaseInvoice struct {
	ID             string                `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID       string                `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	Branch         *Branch               `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	SupplierID     string                `gorm:"type:varchar(15);not null;index" json:"supplier_id"`
	Supplier       *Supplier             `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	InvoiceNumber  string                `gorm:"type:varchar(100);not null;index" json:"invoice_number"`
	InvoiceDate    time.Time             `gorm:"type:date;not null" json:"invoice_date"`
	DueDate        time.Time             `gorm:"type:date" json:"due_date"`
	PaymentType    PaymentType           `gorm:"type:varchar(20);not null;default:'CREDIT'" json:"payment_type"`
	PaymentStatus  PaymentStatus         `gorm:"type:varchar(20);not null;default:'UNPAID'" json:"payment_status"`
	Subtotal       float64               `gorm:"type:decimal(14,2);not null;default:0" json:"subtotal"`
	DiscountAmount float64               `gorm:"type:decimal(14,2);not null;default:0" json:"discount_amount"`
	TaxAmount      float64               `gorm:"type:decimal(14,2);not null;default:0" json:"tax_amount"`
	TotalAmount    float64               `gorm:"type:decimal(14,2);not null;default:0" json:"total_amount"`
	Notes          string                `gorm:"type:text" json:"notes"`
	ReceivedByID   string                `gorm:"type:varchar(15);not null" json:"received_by_id"`
	ReceivedBy     *User                 `gorm:"foreignKey:ReceivedByID" json:"received_by,omitempty"`
	Items          []PurchaseInvoiceItem `gorm:"foreignKey:PurchaseInvoiceID" json:"items,omitempty"`
	CreatedAt      time.Time             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
}

// PurchaseInvoiceItem represents single item line in PBF invoice
type PurchaseInvoiceItem struct {
	ID                 string         `gorm:"type:varchar(15);primaryKey" json:"id"`
	PurchaseInvoiceID  string         `gorm:"type:varchar(15);not null;index" json:"purchase_invoice_id"`
	MedicineID         string         `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine           *Medicine      `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	UnitID             string         `gorm:"type:varchar(15);not null" json:"unit_id"`
	Unit               *Unit          `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	ConversionFactor   float64        `gorm:"type:decimal(10,2);not null;default:1" json:"conversion_factor"`
	QtyPurchased       float64        `gorm:"type:decimal(10,2);not null" json:"qty_purchased"`
	QtyBaseUnit        float64        `gorm:"type:decimal(12,2);not null" json:"qty_base_unit"`
	BatchNumber        string         `gorm:"type:varchar(100);not null" json:"batch_number"`
	ExpiredDate        time.Time      `gorm:"type:date;not null" json:"expired_date"`
	BuyPrice           float64        `gorm:"type:decimal(14,2);not null" json:"buy_price"` // Harga beli per kemasan beli
	DiscountPercent    float64        `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	Subtotal           float64        `gorm:"type:decimal(14,2);not null" json:"subtotal"`
	PutToShelfDirectly bool           `gorm:"default:false" json:"put_to_shelf_directly"`
}

// StockOpname represents rolling cycle counting session
type StockOpname struct {
	ID          string            `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchID    string            `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	OpnameCode  string            `gorm:"type:varchar(50);not null;uniqueIndex" json:"opname_code"`
	OpnameDate  time.Time         `gorm:"type:date;not null" json:"opname_date"`
	Location    StockLocation     `gorm:"type:varchar(20);not null" json:"location"` // GUDANG atau RAK
	RackName    string            `gorm:"type:varchar(100)" json:"rack_name"`        // Misal: "Rak 01 - Obat Batuk"
	Status      OpnameStatus      `gorm:"type:varchar(20);not null;default:'DRAFT'" json:"status"`
	Notes       string            `gorm:"type:text" json:"notes"`
	CreatedByID string            `gorm:"type:varchar(15);not null" json:"created_by_id"`
	CreatedBy   *User             `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	Items       []StockOpnameItem `gorm:"foreignKey:StockOpnameID" json:"items,omitempty"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

// StockOpnameItem represents counting per batch in a cycle count session
type StockOpnameItem struct {
	ID            string         `gorm:"type:varchar(15);primaryKey" json:"id"`
	StockOpnameID string         `gorm:"type:varchar(15);not null;index" json:"stock_opname_id"`
	MedicineID    string         `gorm:"type:varchar(15);not null;index" json:"medicine_id"`
	Medicine      *Medicine      `gorm:"foreignKey:MedicineID" json:"medicine,omitempty"`
	BatchID       string         `gorm:"type:varchar(15);not null" json:"batch_id"`
	Batch         *MedicineBatch `gorm:"foreignKey:BatchID" json:"batch,omitempty"`
	SystemQty     float64        `gorm:"type:decimal(12,2);not null" json:"system_qty"`
	PhysicalQty   float64        `gorm:"type:decimal(12,2);not null" json:"physical_qty"`
	DifferenceQty float64        `gorm:"type:decimal(12,2);not null" json:"difference_qty"` // PhysicalQty - SystemQty
	Notes         string         `gorm:"type:text" json:"notes"`
}

// -----------------------------------------------------------------------
// DTOs for Inventory Operations
// -----------------------------------------------------------------------

type CreatePurchaseInvoiceRequest struct {
	SupplierID     string                     `json:"supplier_id" validate:"required"`
	InvoiceNumber  string                     `json:"invoice_number" validate:"required"`
	InvoiceDate    string                     `json:"invoice_date" validate:"required"` // Format: YYYY-MM-DD
	DueDate        string                     `json:"due_date"`                         // Format: YYYY-MM-DD
	PaymentType    PaymentType                `json:"payment_type" validate:"required"`
	DiscountAmount float64                    `json:"discount_amount"`
	TaxAmount      float64                    `json:"tax_amount"`
	Notes          string                     `json:"notes"`
	Items          []PurchaseInvoiceItemInput `json:"items" validate:"required,min=1"`
}

type PurchaseInvoiceItemInput struct {
	MedicineID         string  `json:"medicine_id" validate:"required"`
	UnitID             string  `json:"unit_id" validate:"required"`
	ConversionFactor   float64 `json:"conversion_factor" validate:"required,gt=0"`
	QtyPurchased       float64 `json:"qty_purchased" validate:"required,gt=0"`
	BatchNumber        string  `json:"batch_number" validate:"required"`
	ExpiredDate        string  `json:"expired_date" validate:"required"` // Format: YYYY-MM-DD
	BuyPrice           float64 `json:"buy_price" validate:"required,gt=0"`
	DiscountPercent    float64 `json:"discount_percent"`
	PutToShelfDirectly bool    `json:"put_to_shelf_directly"`
}

type RefillShelfRequest struct {
	BatchID string  `json:"batch_id" validate:"required"`
	Qty     float64 `json:"qty" validate:"required,gt=0"`
	Notes   string  `json:"notes"`
}

type ReturnToWarehouseRequest struct {
	BatchID string  `json:"batch_id" validate:"required"`
	Qty     float64 `json:"qty" validate:"required,gt=0"`
	Notes   string  `json:"notes"`
}

type CreateOpnameSessionRequest struct {
	Location StockLocation      `json:"location" validate:"required"` // GUDANG atau RAK
	RackName string             `json:"rack_name" validate:"required"`
	Notes    string             `json:"notes"`
	Items    []OpnameItemInput  `json:"items"`
}

type OpnameItemInput struct {
	MedicineID  string  `json:"medicine_id" validate:"required"`
	BatchID     string  `json:"batch_id"`
	BatchNumber string  `json:"batch_number"`
	ExpiredDate string  `json:"expired_date"`
	PhysicalQty float64 `json:"physical_qty" validate:"gte=0"`
	Notes       string  `json:"notes"`
}

type SubmitOpnameRequest struct {
	Items []OpnameItemInput `json:"items" validate:"required,min=1"`
}

// Responses
type StockSummaryResponse struct {
	MedicineID     string           `json:"medicine_id"`
	ItemCode       string           `json:"item_code"`
	Barcode        string           `json:"barcode"`
	Name           string           `json:"name"`
	BaseUnitID     string           `json:"base_unit_id"`
	BaseUnitName   string           `json:"base_unit_name"`
	SellPrice      float64          `json:"sell_price"`
	Prices         []MedicinePrice  `json:"prices,omitempty"`
	TotalWarehouse float64          `json:"total_warehouse"`
	TotalShelf     float64          `json:"total_shelf"`
	TotalStock     float64          `json:"total_stock"`
	MinStock       int              `json:"min_stock"`
	Batches        []*MedicineBatch `json:"batches"` // Sorted FEFO
}

type FefoAlertItem struct {
	BatchID      string    `json:"batch_id"`
	MedicineID   string    `json:"medicine_id"`
	MedicineName string    `json:"medicine_name"`
	ItemCode     string    `json:"item_code"`
	BatchNumber  string    `json:"batch_number"`
	ExpiredDate  time.Time `json:"expired_date"`
	DaysToExpiry int       `json:"days_to_expiry"`
	QtyWarehouse float64   `json:"qty_warehouse"`
	QtyShelf     float64   `json:"qty_shelf"`
	TotalStock   float64   `json:"total_stock"`
	AlertStatus  string    `json:"alert_status"` // EXPIRED, CRITICAL (<90d), WARNING (<180d), SAFE
}

// -----------------------------------------------------------------------
// Repository & Usecase Interfaces
// -----------------------------------------------------------------------

type InventoryRepository interface {
	// Batches
	CreateBatch(ctx context.Context, batch *MedicineBatch) error
	UpdateBatch(ctx context.Context, batch *MedicineBatch) error
	FindBatchByID(ctx context.Context, batchID string) (*MedicineBatch, error)
	FindBatchesByMedicine(ctx context.Context, branchID, medicineID string) ([]*MedicineBatch, error)
	FindBatchByNumberAndMedicine(ctx context.Context, branchID, medicineID, batchNumber string) (*MedicineBatch, error)
	FindBatchesFEFO(ctx context.Context, branchID string) ([]*MedicineBatch, error)
	GetStockSummaryByBranch(ctx context.Context, branchID string, search string) ([]StockSummaryResponse, error)

	// Stock Mutations
	CreateMutation(ctx context.Context, mutation *StockMutation) error
	GetMutationsByMedicine(ctx context.Context, branchID, medicineID string, startDate, endDate time.Time) ([]*StockMutation, error)

	// Invoices
	CreatePurchaseInvoice(ctx context.Context, invoice *PurchaseInvoice) error
	GetPurchaseInvoiceByID(ctx context.Context, invoiceID string) (*PurchaseInvoice, error)
	ListPurchaseInvoices(ctx context.Context, branchID string, page, limit int) ([]*PurchaseInvoice, int64, error)

	// Opname
	CreateOpname(ctx context.Context, opname *StockOpname) error
	GetOpnameByID(ctx context.Context, opnameID string) (*StockOpname, error)
	UpdateOpname(ctx context.Context, opname *StockOpname) error
	ListOpnames(ctx context.Context, branchID string, page, limit int) ([]*StockOpname, int64, error)
	AddOpnameItem(ctx context.Context, item *StockOpnameItem) error
}

type InventoryUsecase interface {
	ReceivePurchaseInvoice(ctx context.Context, branchID, userID string, req *CreatePurchaseInvoiceRequest) (*PurchaseInvoice, error)
	ListPurchaseInvoices(ctx context.Context, branchID string, page, limit int) ([]*PurchaseInvoice, int64, error)
	GetPurchaseInvoiceDetail(ctx context.Context, invoiceID string) (*PurchaseInvoice, error)

	RefillShelf(ctx context.Context, branchID, userID string, req *RefillShelfRequest) (*MedicineBatch, error)
	ReturnToWarehouse(ctx context.Context, branchID, userID string, req *ReturnToWarehouseRequest) (*MedicineBatch, error)

	GetStockSummary(ctx context.Context, branchID, medicineID string) (*StockSummaryResponse, error)
	ListStockSummaries(ctx context.Context, branchID string, search string) ([]StockSummaryResponse, error)
	GetStockCard(ctx context.Context, branchID, medicineID string, startDate, endDate string) ([]*StockMutation, error)
	GetFefoAlerts(ctx context.Context, branchID string) ([]FefoAlertItem, error)

	CreateOpnameSession(ctx context.Context, branchID, userID string, req *CreateOpnameSessionRequest) (*StockOpname, error)
	SubmitOpnameSession(ctx context.Context, branchID, userID, opnameID string, req *SubmitOpnameRequest) (*StockOpname, error)
	ListOpnames(ctx context.Context, branchID string, page, limit int) ([]*StockOpname, int64, error)
	GetOpnameDetail(ctx context.Context, opnameID string) (*StockOpname, error)
}
