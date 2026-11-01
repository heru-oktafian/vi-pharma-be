package usecase

import (
	"context"

	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
	"pharma-be/pkg/utils"
)

type authUsecase struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

// NewAuthUsecase creates a new AuthUsecase instance
func NewAuthUsecase(userRepo domain.UserRepository, cfg *config.Config) domain.AuthUsecase {
	return &authUsecase{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (u *authUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := u.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Check status
	if user.UserStatus == domain.StatusInactive {
		return nil, domain.ErrUserInactive
	}
	if user.UserStatus == domain.StatusSuspended {
		return nil, domain.ErrUserSuspended
	}

	// Determine available divisions based on role
	availableDivisions := getDivisionsForRole(user.UserRole)

	// Generate initial JWT token (division empty until selected)
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		string(user.UserRole),
		"",
		u.cfg.JWTSecret,
		u.cfg.JWTExpirationHours,
	)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	return &domain.LoginResponse{
		Token:              token,
		AvailableDivisions: availableDivisions,
		User: domain.UserResponse{
			ID:         user.ID,
			Username:   user.Username,
			Name:       user.Name,
			UserRole:   user.UserRole,
			UserStatus: user.UserStatus,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		},
	}, nil
}

func (u *authUsecase) SelectDivision(ctx context.Context, userID, division string) (*domain.SelectDivisionResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if user.UserStatus != domain.StatusActive {
		return nil, domain.ErrUserInactive
	}

	allowedDivisions := getDivisionsForRole(user.UserRole)
	isAllowed := false
	for _, d := range allowedDivisions {
		if d == division {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return nil, domain.ErrUnauthorizedDivision
	}

	// Generate new JWT containing selected division
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		string(user.UserRole),
		division,
		u.cfg.JWTSecret,
		u.cfg.JWTExpirationHours,
	)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	return &domain.SelectDivisionResponse{
		Token:            token,
		SelectedDivision: division,
	}, nil
}

func (u *authUsecase) GetProfile(ctx context.Context, userID string) (*domain.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
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

// getDivisionsForRole maps user roles to their accessible divisions
func getDivisionsForRole(role domain.UserRole) []string {
	switch role {
	case domain.RoleSuperAdmin:
		return []string{"apotek", "kasir", "gudang", "purchasing", "manajemen"}
	case domain.RoleAdmin:
		return []string{"apotek", "kasir", "gudang", "purchasing"}
	case domain.RolePharmacist:
		return []string{"apotek", "gudang"}
	case domain.RoleCashier:
		return []string{"kasir"}
	case domain.RoleInventory:
		return []string{"gudang", "purchasing"}
	default:
		return []string{"apotek"}
	}
}
