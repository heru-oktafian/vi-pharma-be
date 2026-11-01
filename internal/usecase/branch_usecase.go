package usecase

import (
	"context"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type BranchUsecase interface {
	CreateBranch(ctx context.Context, branch *domain.Branch) (*domain.Branch, error)
	GetBranchByID(ctx context.Context, id string) (*domain.Branch, error)
	ListBranches(ctx context.Context, page, pageSize int) ([]domain.Branch, int64, error)
	AssignUserBranch(ctx context.Context, userID, branchID string, isDefault bool) error
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
