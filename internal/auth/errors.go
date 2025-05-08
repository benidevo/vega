package auth

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrInvalidRole              = errors.New("invalid role")
	ErrUserCreationFailed       = errors.New("user creation failed")
	ErrUserUpdateFailed         = errors.New("user update failed")
	ErrUserDeletionFailed       = errors.New("user deletion failed")
	ErrUserRetrievalFailed      = errors.New("user retrieval failed")
	ErrUserListRetrievalFailed  = errors.New("user list retrieval failed")
	ErrUserPasswordChangeFailed = errors.New("user password change failed")
	ErrUserRoleChangeFailed     = errors.New("user role change failed")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrTokenExpired             = errors.New("token expired")
	ErrTokenInvalid             = errors.New("token invalid")
	ErrTokenCreationFailed      = errors.New("token creation failed")
	ErrInvalidToken             = errors.New("invalid token")
)
