package models

import (
	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
)

var (
	// User related errors
	ErrUserNotFound             = commonerrors.New("user not found")
	ErrUserAlreadyExists        = commonerrors.New("user already exists")
	ErrInvalidRole              = commonerrors.New("invalid role")
	ErrUserCreationFailed       = commonerrors.New("user creation failed")
	ErrUserUpdateFailed         = commonerrors.New("user update failed")
	ErrUserDeletionFailed       = commonerrors.New("user deletion failed")
	ErrUserRetrievalFailed      = commonerrors.New("user retrieval failed")
	ErrUserListRetrievalFailed  = commonerrors.New("user list retrieval failed")
	ErrUserPasswordChangeFailed = commonerrors.New("user password change failed")
	ErrUserRoleChangeFailed     = commonerrors.New("user role change failed")
	ErrInvalidCredentials       = commonerrors.New("invalid credentials")

	// Token related errors
	ErrTokenExpired        = commonerrors.New("token expired")
	ErrTokenInvalid        = commonerrors.New("token invalid")
	ErrTokenCreationFailed = commonerrors.New("token creation failed")
	ErrInvalidToken        = commonerrors.New("invalid token")

	// Google Auth related errors
	ErrGoogleCredentialsReadFailed   = commonerrors.New("failed to read Google credentials file")
	ErrGoogleCredentialsInvalid      = commonerrors.New("invalid Google credentials format")
	ErrGoogleCodeExchangeFailed      = commonerrors.New("failed to exchange Google auth code")
	ErrGoogleUserInfoFailed          = commonerrors.New("failed to retrieve Google user info")
	ErrGoogleAuthTokenCreationFailed = commonerrors.New("failed to create auth token for Google user")
	ErrGoogleUserCreationFailed      = commonerrors.New("failed to create user from Google account")
	ErrGoogleDriveServiceFailed      = commonerrors.New("failed to create Google Drive service")
)

// WrapError is a convenience function that calls the common errors package
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}

// GetSentinelError is a convenience function that calls the common errors package
func GetSentinelError(err error) error {
	return commonerrors.GetSentinelError(err)
}
