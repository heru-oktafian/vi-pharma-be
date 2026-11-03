package usecase

import (
	"context"
	"time"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type doctorUsecase struct {
	doctorRepo domain.DoctorRepository
}

func NewDoctorUsecase(doctorRepo domain.DoctorRepository) domain.DoctorUsecase {
	return &doctorUsecase{doctorRepo: doctorRepo}
}

func (u *doctorUsecase) GetAllDoctors(ctx context.Context, search string, activeOnly bool) ([]domain.DoctorResponse, error) {
	doctors, err := u.doctorRepo.FindAll(ctx, search, activeOnly)
	if err != nil {
		return nil, err
	}

	res := make([]domain.DoctorResponse, len(doctors))
	for i, doc := range doctors {
		res[i] = *doc.ToResponse()
	}
	return res, nil
}

func (u *doctorUsecase) GetDoctorByID(ctx context.Context, id string) (*domain.DoctorResponse, error) {
	doc, err := u.doctorRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return doc.ToResponse(), nil
}

func (u *doctorUsecase) CreateDoctor(ctx context.Context, req *domain.CreateDoctorRequest) (*domain.DoctorResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	doc := &domain.Doctor{
		ID:             utils.GenerateDoctorID(),
		Name:           req.Name,
		SIP:            req.SIP,
		Specialization: req.Specialization,
		ClinicHospital: req.ClinicHospital,
		Phone:          req.Phone,
		Address:        req.Address,
		IsActive:       isActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.doctorRepo.Create(ctx, doc); err != nil {
		return nil, err
	}

	return doc.ToResponse(), nil
}

func (u *doctorUsecase) UpdateDoctor(ctx context.Context, id string, req *domain.UpdateDoctorRequest) (*domain.DoctorResponse, error) {
	doc, err := u.doctorRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	doc.Name = req.Name
	doc.SIP = req.SIP
	doc.Specialization = req.Specialization
	doc.ClinicHospital = req.ClinicHospital
	doc.Phone = req.Phone
	doc.Address = req.Address
	if req.IsActive != nil {
		doc.IsActive = *req.IsActive
	}
	doc.UpdatedAt = time.Now()

	if err := u.doctorRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	return doc.ToResponse(), nil
}

func (u *doctorUsecase) DeleteDoctor(ctx context.Context, id string) error {
	return u.doctorRepo.Delete(ctx, id)
}
