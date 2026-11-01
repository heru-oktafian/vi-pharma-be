package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"pharma-be/internal/domain"
)

type branchRepository struct {
	db *gorm.DB
}

// NewBranchRepository creates a new PostgreSQL BranchRepository instance
func NewBranchRepository(db *gorm.DB) domain.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *branchRepository) FindByID(ctx context.Context, id string) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.db.WithContext(ctx).First(&branch, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, err
	}
	return &branch, nil
}

func (r *branchRepository) FindByCode(ctx context.Context, code string) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.db.WithContext(ctx).First(&branch, "branch_code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, err
	}
	return &branch, nil
}

func (r *branchRepository) List(ctx context.Context, limit, offset int) ([]domain.Branch, int64, error) {
	var branches []domain.Branch
	var total int64

	if err := r.db.WithContext(ctx).Model(&domain.Branch{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at ASC").
		Find(&branches).Error
	if err != nil {
		return nil, 0, err
	}

	return branches, total, nil
}

func (r *branchRepository) AssignUserBranch(ctx context.Context, ub *domain.UserBranch) error {
	return r.db.WithContext(ctx).Create(ub).Error
}

func (r *branchRepository) GetUserBranches(ctx context.Context, userID string) ([]domain.BranchSummaryResponse, error) {
	// First, check if user is superadmin
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err == nil {
		if user.UserRole == domain.RoleSuperAdmin {
			// Superadmin has access to ALL active branches
			var allBranches []domain.Branch
			if err := r.db.WithContext(ctx).
				Where("branch_status = ?", domain.BranchStatusActive).
				Order("branch_name ASC").
				Find(&allBranches).Error; err == nil {

				var summaries []domain.BranchSummaryResponse
				for _, b := range allBranches {
					summaries = append(summaries, domain.BranchSummaryResponse{
						ID:         b.ID,
						BranchCode: b.BranchCode,
						BranchName: b.BranchName,
						Address:    b.Address,
						Phone:      b.Phone,
						Status:     b.BranchStatus,
						IsDefault:  false,
					})
				}
				if len(summaries) > 0 {
					summaries[0].IsDefault = true
				}
				return summaries, nil
			}
		}
	}

	// For normal users, query assigned user_branches
	var userBranches []domain.UserBranch
	err := r.db.WithContext(ctx).
		Preload("Branch").
		Where("user_id = ?", userID).
		Order("is_default DESC").
		Find(&userBranches).Error

	if err != nil {
		return nil, err
	}

	var results []domain.BranchSummaryResponse
	for _, ub := range userBranches {
		if ub.Branch != nil && ub.Branch.BranchStatus == domain.BranchStatusActive {
			results = append(results, domain.BranchSummaryResponse{
				ID:         ub.Branch.ID,
				BranchCode: ub.Branch.BranchCode,
				BranchName: ub.Branch.BranchName,
				Address:    ub.Branch.Address,
				Phone:      ub.Branch.Phone,
				Status:     ub.Branch.BranchStatus,
				IsDefault:  ub.IsDefault,
			})
		}
	}

	return results, nil
}

func (r *branchRepository) IsUserAssignedToBranch(ctx context.Context, userID, branchID string) (bool, error) {
	// Check if user is superadmin
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err == nil {
		if user.UserRole == domain.RoleSuperAdmin {
			return true, nil
		}
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserBranch{}).
		Where("user_id = ? AND branch_id = ?", userID, branchID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
