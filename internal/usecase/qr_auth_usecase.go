package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"pharma-be/internal/delivery/http/websocket"
	"pharma-be/internal/domain"
	"pharma-be/pkg/config"
	"pharma-be/pkg/utils"
)

type QRAuthUsecase interface {
	GenerateSession(ctx context.Context) (*domain.GenerateQRResponse, error)
	AuthorizeSession(ctx context.Context, userID string, req *domain.AuthorizeQRRequest) (*domain.QRSession, error)
	SelectBranch(ctx context.Context, req *domain.SelectBranchRequest) (*domain.SelectBranchResponse, error)
	GetSession(ctx context.Context, sessionID string) (*domain.QRSession, error)
}

type qrAuthUsecase struct {
	qrStore    domain.QRSessionStore
	userRepo   domain.UserRepository
	branchRepo domain.BranchRepository
	wsHub      *websocket.Hub
	cfg        *config.Config
}

func NewQRAuthUsecase(
	qrStore domain.QRSessionStore,
	userRepo domain.UserRepository,
	branchRepo domain.BranchRepository,
	wsHub *websocket.Hub,
	cfg *config.Config,
) QRAuthUsecase {
	return &qrAuthUsecase{
		qrStore:    qrStore,
		userRepo:   userRepo,
		branchRepo: branchRepo,
		wsHub:      wsHub,
		cfg:        cfg,
	}
}

func (u *qrAuthUsecase) GenerateSession(ctx context.Context) (*domain.GenerateQRResponse, error) {
	sessionID := uuid.New().String()
	const ttlSeconds = 120
	expiresAt := time.Now().Add(ttlSeconds * time.Second)

	session := &domain.QRSession{
		SessionID: sessionID,
		Status:    domain.QRSessionStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	if err := u.qrStore.Save(ctx, session); err != nil {
		return nil, err
	}

	return &domain.GenerateQRResponse{
		SessionID:        sessionID,
		QRString:         "vipharma:qr:" + sessionID,
		ExpiresAt:        expiresAt,
		ExpiresInSeconds: ttlSeconds,
	}, nil
}

func (u *qrAuthUsecase) AuthorizeSession(ctx context.Context, userID string, req *domain.AuthorizeQRRequest) (*domain.QRSession, error) {
	// 1. Validate User
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if user.UserStatus == domain.StatusInactive {
		return nil, domain.ErrUserInactive
	}
	if user.UserStatus == domain.StatusSuspended {
		return nil, domain.ErrUserSuspended
	}

	// 2. Validate Session
	session, err := u.qrStore.Get(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status == domain.QRSessionStatusAuthorized {
		return nil, domain.ErrQRSessionAlreadyAuthorized
	}
	if session.Status != domain.QRSessionStatusPending {
		return nil, domain.ErrQRSessionNotFound
	}

	// 3. Query Accessible Branches for User
	branches, err := u.branchRepo.GetUserBranches(ctx, userID)
	if err != nil {
		return nil, err
	}

	userResp := &domain.UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Name:       user.Name,
		UserRole:   user.UserRole,
		UserStatus: user.UserStatus,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	// 4. Update Session in Memory
	session.Status = domain.QRSessionStatusAuthorized
	session.UserID = userID
	session.User = userResp
	session.Branches = branches
	session.DeviceID = req.DeviceID

	if err := u.qrStore.Update(ctx, session); err != nil {
		return nil, err
	}

	// 5. Broadcast to WebSocket Listener (vi-pharma Web)
	_ = u.wsHub.BroadcastToSession(session.SessionID, &domain.WSMessage{
		Event: domain.WSEventAuthorized,
		Data: map[string]interface{}{
			"session_id": session.SessionID,
			"user":       userResp,
			"branches":   branches,
		},
	})

	return session, nil
}

func (u *qrAuthUsecase) SelectBranch(ctx context.Context, req *domain.SelectBranchRequest) (*domain.SelectBranchResponse, error) {
	// 1. Get Session
	session, err := u.qrStore.Get(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != domain.QRSessionStatusAuthorized {
		return nil, domain.ErrQRSessionNotAuthorized
	}

	// 2. Verify User has access to requested branch
	hasAccess, err := u.branchRepo.IsUserAssignedToBranch(ctx, session.UserID, req.BranchID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, domain.ErrUnauthorizedBranch
	}

	// 3. Retrieve Branch Details
	branch, err := u.branchRepo.FindByID(ctx, req.BranchID)
	if err != nil {
		return nil, err
	}

	// 4. Generate Web Session JWT with BranchID claim
	webToken, err := utils.GenerateTokenWithBranch(
		session.User.ID,
		session.User.Username,
		string(session.User.UserRole),
		branch.ID,
		"",
		u.cfg.JWTSecret,
		u.cfg.JWTExpirationHours,
	)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	// 5. Mark Session as Consumed
	session.Status = domain.QRSessionStatusConsumed
	session.SelectedBranchID = branch.ID
	_ = u.qrStore.Update(ctx, session)

	// Notify WebSocket of final consumption
	_ = u.wsHub.BroadcastToSession(session.SessionID, &domain.WSMessage{
		Event: domain.WSEventConsumed,
		Data: map[string]interface{}{
			"session_id": session.SessionID,
			"branch_id":  branch.ID,
		},
	})

	branchSummary := domain.BranchSummaryResponse{
		ID:         branch.ID,
		BranchCode: branch.BranchCode,
		BranchName: branch.BranchName,
		Address:    branch.Address,
		Phone:      branch.Phone,
		Status:     branch.BranchStatus,
		IsDefault:  false,
	}

	return &domain.SelectBranchResponse{
		Token:  webToken,
		Branch: branchSummary,
		User:   *session.User,
	}, nil
}

func (u *qrAuthUsecase) GetSession(ctx context.Context, sessionID string) (*domain.QRSession, error) {
	return u.qrStore.Get(ctx, sessionID)
}
