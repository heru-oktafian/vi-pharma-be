package domain

import (
	"context"
	"time"
)

// Custom PostgreSQL ENUM Types for Branch
type JournalMethod string

const (
	JournalMethodPerpetual JournalMethod = "perpetual"
	JournalMethodPeriodic  JournalMethod = "periodic"
)

type BranchStatus string

const (
	BranchStatusActive    BranchStatus = "active"
	BranchStatusInactive  BranchStatus = "inactive"
	BranchStatusSuspended BranchStatus = "suspended"
)

type SubscriptionType string

const (
	SubscriptionTrial      SubscriptionType = "trial"
	SubscriptionBasic      SubscriptionType = "basic"
	SubscriptionPro        SubscriptionType = "pro"
	SubscriptionEnterprise SubscriptionType = "enterprise"
)

type RealAsset string

const (
	RealAssetOwned       RealAsset = "owned"
	RealAssetRented      RealAsset = "rented"
	RealAssetPartnership RealAsset = "partnership"
)

// Branch represents the branches table entity in PostgreSQL
type Branch struct {
	ID               string           `gorm:"type:varchar(15);primaryKey" json:"id"`
	BranchCode       string           `gorm:"type:varchar(50);uniqueIndex;not null" json:"branch_code"`
	BranchName       string           `gorm:"type:text;not null" json:"branch_name"`
	Address          string           `gorm:"type:text" json:"address"`
	Phone            string           `gorm:"type:varchar(100)" json:"phone"`
	Email            string           `gorm:"type:varchar(100)" json:"email"`
	SiaID            string           `gorm:"type:varchar(100)" json:"sia_id"`
	SiaName          string           `gorm:"type:varchar(255)" json:"sia_name"`
	PsaID            string           `gorm:"type:varchar(100)" json:"psa_id"`
	PsaName          string           `gorm:"type:varchar(255)" json:"psa_name"`
	Sipa             string           `gorm:"type:varchar(100)" json:"sipa"`
	SipaName         string           `gorm:"type:varchar(255)" json:"sipa_name"`
	ApingID          string           `gorm:"type:varchar(100)" json:"aping_id"`
	ApingName        string           `gorm:"type:varchar(255)" json:"aping_name"`
	BankName         string           `gorm:"type:varchar(255)" json:"bank_name"`
	AccountName      string           `gorm:"type:varchar(255)" json:"account_name"`
	AccountNumber    string           `gorm:"type:varchar(100)" json:"account_number"`
	TaxPercentage    int64            `gorm:"type:int8;default:0" json:"tax_percentage"`
	JournalMethod    JournalMethod    `gorm:"type:journal_method;not null;default:'perpetual'" json:"journal_method"`
	BranchStatus     BranchStatus     `gorm:"type:branch_status;not null;default:'active'" json:"branch_status"`
	LicenseDate      *time.Time       `gorm:"type:timestamptz" json:"license_date"`
	DefaultMember    string           `gorm:"type:varchar(15)" json:"default_member"`
	SubscriptionType SubscriptionType `gorm:"type:subscription_type;not null;default:'trial'" json:"subscription_type"`
	Quota            int32            `gorm:"type:int4;default:0" json:"quota"`
	RealAsset        RealAsset        `gorm:"type:real_asset;not null;default:'owned'" json:"real_asset"`
	DailyMarginTarget float64         `gorm:"type:numeric(15,2);not null;default:250000" json:"daily_margin_target"`
	DailyBonusReward  float64         `gorm:"type:numeric(15,2);not null;default:50000" json:"daily_bonus_reward"`
	CreatedAt        time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Branch) TableName() string {
	return "branches"
}

// UserBranch represents the pivot table between users and branches
type UserBranch struct {
	ID        string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(15);index;not null" json:"user_id"`
	BranchID  string    `gorm:"type:varchar(15);index;not null" json:"branch_id"`
	IsDefault bool      `gorm:"type:boolean;default:false" json:"is_default"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Branch *Branch `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
}

func (UserBranch) TableName() string {
	return "user_branches"
}

// DTOs
type BranchSummaryResponse struct {
	ID         string       `json:"id"`
	BranchCode string       `json:"branch_code"`
	BranchName string       `json:"branch_name"`
	Address    string       `json:"address"`
	Phone      string       `json:"phone"`
	Status     BranchStatus `json:"branch_status"`
	IsDefault  bool         `json:"is_default"`
}

// BranchRepository interface
type BranchRepository interface {
	Create(ctx context.Context, branch *Branch) error
	FindByID(ctx context.Context, id string) (*Branch, error)
	FindByCode(ctx context.Context, code string) (*Branch, error)
	List(ctx context.Context, limit, offset int) ([]Branch, int64, error)
	AssignUserBranch(ctx context.Context, ub *UserBranch) error
	GetUserBranches(ctx context.Context, userID string) ([]BranchSummaryResponse, error)
	IsUserAssignedToBranch(ctx context.Context, userID, branchID string) (bool, error)
}
