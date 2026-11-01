package domain

import (
	"context"
	"time"
)

// UserRole represents custom ENUM type in PostgreSQL: user_role
type UserRole string

const (
	RoleSuperAdmin UserRole = "superadmin"
	RoleAdmin      UserRole = "admin"
	RolePharmacist UserRole = "pharmacist"
	RoleCashier    UserRole = "cashier"
	RoleInventory  UserRole = "inventory"
)

// UserStatus represents custom ENUM type in PostgreSQL: user_status
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// User represents the users table entity in Clean Architecture
type User struct {
	ID         string     `gorm:"type:varchar(15);primaryKey" json:"id"`
	Username   string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Password   string     `gorm:"type:text;not null" json:"-"`
	Name       string     `gorm:"type:varchar(255);not null" json:"name"`
	UserRole   UserRole   `gorm:"type:user_role;not null;default:'pharmacist'" json:"user_role"`
	UserStatus UserStatus `gorm:"type:user_status;not null;default:'active'" json:"user_status"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName overrides the table name for GORM
func (User) TableName() string {
	return "users"
}

// Request and Response DTOs
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Name       string     `json:"name"`
	UserRole   UserRole   `json:"user_role"`
	UserStatus UserStatus `json:"user_status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type LoginResponse struct {
	Token              string       `json:"token"`
	User               UserResponse `json:"user"`
	AvailableDivisions []string     `json:"available_divisions"`
}

type SelectDivisionRequest struct {
	Division string `json:"division" validate:"required"`
}

type SelectDivisionResponse struct {
	Token            string `json:"token"`
	SelectedDivision string `json:"selected_division"`
}

type CreateUserRequest struct {
	Username   string     `json:"username" validate:"required,min=3,max=50"`
	Password   string     `json:"password" validate:"required,min=6"`
	Name       string     `json:"name" validate:"required"`
	UserRole   UserRole   `json:"user_role" validate:"required"`
	UserStatus UserStatus `json:"user_status"`
}

// UserRepository interface defines data access contracts
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]User, int64, error)
}

// AuthUsecase interface defines authentication and division selection contracts
type AuthUsecase interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	SelectDivision(ctx context.Context, userID, division string) (*SelectDivisionResponse, error)
	GetProfile(ctx context.Context, userID string) (*UserResponse, error)
}

// UserUsecase interface defines user administration business logic
type UserUsecase interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	ListUsers(ctx context.Context, page, pageSize int) ([]UserResponse, int64, error)
}
