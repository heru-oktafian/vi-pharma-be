package domain

import "errors"

var (
	ErrNotFound                   = errors.New("record not found")
	ErrUserNotFound               = errors.New("user not found")
	ErrInvalidCredentials         = errors.New("invalid username or password")
	ErrUsernameAlreadyExists      = errors.New("username is already taken")
	ErrUserInactive               = errors.New("user account is inactive")
	ErrUserSuspended              = errors.New("user account is suspended")
	ErrInvalidRole                = errors.New("invalid user role")
	ErrInvalidStatus              = errors.New("invalid user status")
	ErrUnauthorizedDivision       = errors.New("user does not have access to this division")
	ErrBranchNotFound             = errors.New("branch not found")
	ErrUnauthorizedBranch         = errors.New("user does not have access to this branch")
	ErrQRSessionNotFound          = errors.New("qr session not found or already consumed")
	ErrQRSessionExpired           = errors.New("qr session has expired")
	ErrQRSessionAlreadyAuthorized = errors.New("qr session is already authorized")
	ErrQRSessionNotAuthorized     = errors.New("qr session has not been authorized by mobile")
	ErrInternalServer             = errors.New("internal server error")
)
