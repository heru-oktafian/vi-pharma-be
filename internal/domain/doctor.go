package domain

import (
	"context"
	"time"
)

// Doctor represents doctors and medical prescribers
type Doctor struct {
	ID             string    `gorm:"type:varchar(15);primaryKey" json:"id"`
	Name           string    `gorm:"type:varchar(255);not null" json:"name"`
	SIP            string    `gorm:"type:varchar(100);not null" json:"sip"` // Surat Izin Praktik
	Specialization string    `gorm:"type:varchar(100)" json:"specialization"`
	ClinicHospital string    `gorm:"type:varchar(255)" json:"clinic_hospital"`
	Phone          string    `gorm:"type:varchar(50)" json:"phone"`
	Address        string    `gorm:"type:text" json:"address"`
	IsActive       bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type CreateDoctorRequest struct {
	Name           string `json:"name" validate:"required,min=2,max=255"`
	SIP            string `json:"sip" validate:"required,min=2,max=100"`
	Specialization string `json:"specialization"`
	ClinicHospital string `json:"clinic_hospital"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	IsActive       *bool  `json:"is_active"`
}

type UpdateDoctorRequest struct {
	Name           string `json:"name" validate:"required,min=2,max=255"`
	SIP            string `json:"sip" validate:"required,min=2,max=100"`
	Specialization string `json:"specialization"`
	ClinicHospital string `json:"clinic_hospital"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	IsActive       *bool  `json:"is_active"`
}

type DoctorResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	SIP            string    `json:"sip"`
	Specialization string    `json:"specialization"`
	ClinicHospital string    `json:"clinic_hospital"`
	Phone          string    `json:"phone"`
	Address        string    `json:"address"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (d *Doctor) ToResponse() *DoctorResponse {
	return &DoctorResponse{
		ID:             d.ID,
		Name:           d.Name,
		SIP:            d.SIP,
		Specialization: d.Specialization,
		ClinicHospital: d.ClinicHospital,
		Phone:          d.Phone,
		Address:        d.Address,
		IsActive:       d.IsActive,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

type DoctorRepository interface {
	FindAll(ctx context.Context, search string, activeOnly bool) ([]Doctor, error)
	FindByID(ctx context.Context, id string) (*Doctor, error)
	Create(ctx context.Context, doctor *Doctor) error
	Update(ctx context.Context, doctor *Doctor) error
	Delete(ctx context.Context, id string) error
}

type DoctorUsecase interface {
	GetAllDoctors(ctx context.Context, search string, activeOnly bool) ([]DoctorResponse, error)
	GetDoctorByID(ctx context.Context, id string) (*DoctorResponse, error)
	CreateDoctor(ctx context.Context, req *CreateDoctorRequest) (*DoctorResponse, error)
	UpdateDoctor(ctx context.Context, id string, req *UpdateDoctorRequest) (*DoctorResponse, error)
	DeleteDoctor(ctx context.Context, id string) error
}
