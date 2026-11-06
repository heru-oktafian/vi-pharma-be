package usecase

import (
	"context"

	"pharma-be/internal/domain"
	"pharma-be/pkg/utils"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

// NewUserUsecase creates a new UserUsecase instance
func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Validate role
	if !isValidRole(req.UserRole) {
		return nil, domain.ErrInvalidRole
	}

	// Default status to active if not provided
	status := req.UserStatus
	if status == "" {
		status = domain.StatusActive
	} else if !isValidStatus(status) {
		return nil, domain.ErrInvalidStatus
	}

	// Check if username already exists
	existing, _ := u.userRepo.FindByUsername(ctx, req.Username)
	if existing != nil {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	// Generate 15-character ID
	userID := utils.GenerateUserID()

	newUser := &domain.User{
		ID:         userID,
		Username:   req.Username,
		Password:   hashedPassword,
		Name:       req.Name,
		UserRole:   req.UserRole,
		UserStatus: status,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:         newUser.ID,
		Username:   newUser.Username,
		Name:       newUser.Name,
		UserRole:   newUser.UserRole,
		UserStatus: newUser.UserStatus,
		CreatedAt:  newUser.CreatedAt,
		UpdatedAt:  newUser.UpdatedAt,
	}, nil
}

func (u *userUsecase) GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Name:       user.Name,
		UserRole:   user.UserRole,
		UserStatus: user.UserStatus,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func (u *userUsecase) ListUsers(ctx context.Context, page, pageSize int) ([]domain.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, total, err := u.userRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var responseList []domain.UserResponse
	for _, user := range users {
		responseList = append(responseList, domain.UserResponse{
			ID:         user.ID,
			Username:   user.Username,
			Name:       user.Name,
			UserRole:   user.UserRole,
			UserStatus: user.UserStatus,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		})
	}

	return responseList, total, nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.UserRole != "" && isValidRole(req.UserRole) {
		user.UserRole = req.UserRole
	}
	if req.UserStatus != "" && isValidStatus(req.UserStatus) {
		user.UserStatus = req.UserStatus
	}
	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return nil, domain.ErrInternalServer
		}
		user.Password = hashed
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Name:       user.Name,
		UserRole:   user.UserRole,
		UserStatus: user.UserStatus,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}, nil
}

func isValidRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleAdmin, domain.RolePharmacist, domain.RoleCashier, domain.RoleInventory:
		return true
	default:
		return false
	}
}

func isValidStatus(status domain.UserStatus) bool {
	switch status {
	case domain.StatusActive, domain.StatusInactive, domain.StatusSuspended:
		return true
	default:
		return false
	}
}
