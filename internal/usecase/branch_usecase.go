package usecase

import (
	"context"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type BranchUsecase interface {
	CreateBranch(ctx context.Context, branch *domain.Branch) (*domain.Branch, error)
	UpdateBranch(ctx context.Context, id string, req *domain.UpdateBranchRequest) (*domain.Branch, error)
	GetBranchByID(ctx context.Context, id string) (*domain.Branch, error)
	ListBranches(ctx context.Context, page, pageSize int) ([]domain.Branch, int64, error)
	AssignUserBranch(ctx context.Context, userID, branchID string, isDefault bool) error
	GetUserBranches(ctx context.Context, userID string) ([]domain.BranchSummaryResponse, error)
}

type branchUsecase struct {
	branchRepo domain.BranchRepository
}

func NewBranchUsecase(branchRepo domain.BranchRepository) BranchUsecase {
	return &branchUsecase{branchRepo: branchRepo}
}

func (u *branchUsecase) CreateBranch(ctx context.Context, branch *domain.Branch) (*domain.Branch, error) {
	if branch.ID == "" {
		branch.ID = utils.GenerateBranchID()
	}
	if branch.BranchStatus == "" {
		branch.BranchStatus = domain.BranchStatusActive
	}
	if branch.JournalMethod == "" {
		branch.JournalMethod = domain.JournalMethodPerpetual
	}
	if branch.SubscriptionType == "" {
		branch.SubscriptionType = domain.SubscriptionTrial
	}
	if branch.RealAsset == "" {
		branch.RealAsset = domain.RealAssetOwned
	}

	if err := u.branchRepo.Create(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (u *branchUsecase) GetBranchByID(ctx context.Context, id string) (*domain.Branch, error) {
	return u.branchRepo.FindByID(ctx, id)
}

func (u *branchUsecase) ListBranches(ctx context.Context, page, pageSize int) ([]domain.Branch, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return u.branchRepo.List(ctx, pageSize, offset)
}

func (u *branchUsecase) AssignUserBranch(ctx context.Context, userID, branchID string, isDefault bool) error {
	ub := &domain.UserBranch{
		ID:        utils.GenerateUserBranchID(),
		UserID:    userID,
		BranchID:  branchID,
		IsDefault: isDefault,
	}
	return u.branchRepo.AssignUserBranch(ctx, ub)
}

func (u *branchUsecase) UpdateBranch(ctx context.Context, id string, req *domain.UpdateBranchRequest) (*domain.Branch, error) {
	branch, err := u.branchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.BranchName != "" {
		branch.BranchName = req.BranchName
	}
	if req.Address != "" {
		branch.Address = req.Address
	}
	if req.Phone != "" {
		branch.Phone = req.Phone
	}
	if req.Email != "" {
		branch.Email = req.Email
	}
	if req.SiaID != "" {
		branch.SiaID = req.SiaID
	}
	if req.SiaName != "" {
		branch.SiaName = req.SiaName
	}
	if req.Sipa != "" {
		branch.Sipa = req.Sipa
	}
	if req.SipaName != "" {
		branch.SipaName = req.SipaName
	}
	if req.PsaID != "" {
		branch.PsaID = req.PsaID
	}
	if req.PsaName != "" {
		branch.PsaName = req.PsaName
	}
	if req.BankName != "" {
		branch.BankName = req.BankName
	}
	if req.AccountName != "" {
		branch.AccountName = req.AccountName
	}
	if req.AccountNumber != "" {
		branch.AccountNumber = req.AccountNumber
	}
	if req.TaxPercentage >= 0 {
		branch.TaxPercentage = req.TaxPercentage
	}
	if req.JournalMethod != "" {
		branch.JournalMethod = req.JournalMethod
	}
	if req.BranchStatus != "" {
		branch.BranchStatus = req.BranchStatus
	}

	if err := u.branchRepo.Update(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

func (u *branchUsecase) GetUserBranches(ctx context.Context, userID string) ([]domain.BranchSummaryResponse, error) {
	return u.branchRepo.GetUserBranches(ctx, userID)
}
