package auth

import (
	"errors"
	"fmt"
)

var (
	// User related errors
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

	// Token related errors
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenInvalid        = errors.New("token invalid")
	ErrTokenCreationFailed = errors.New("token creation failed")
	ErrInvalidToken        = errors.New("invalid token")

	// Google Auth related errors
	ErrGoogleCredentialsReadFailed   = errors.New("failed to read Google credentials file")
	ErrGoogleCredentialsInvalid      = errors.New("invalid Google credentials format")
	ErrGoogleCodeExchangeFailed      = errors.New("failed to exchange Google auth code")
	ErrGoogleUserInfoFailed          = errors.New("failed to retrieve Google user info")
	ErrGoogleAuthTokenCreationFailed = errors.New("failed to create auth token for Google user")
	ErrGoogleUserCreationFailed      = errors.New("failed to create user from Google account")
	ErrGoogleDriveServiceFailed      = errors.New("failed to create Google Drive service")
)

// RepositoryError wraps a sentinel error with the underlying error details
type RepositoryError struct {
	SentinelError error // The predefined user-friendly error
	InnerError    error // The underlying technical error
}

// Error implements the error interface
func (e *RepositoryError) Error() string {
	if e.InnerError != nil {
		return fmt.Sprintf("%s: %v", e.SentinelError.Error(), e.InnerError)
	}
	return e.SentinelError.Error()
}

// Unwrap implements the errors.Wrapper interface to support errors.Is and errors.As
func (e *RepositoryError) Unwrap() error {
	return e.SentinelError
}

// Is implements custom behavior for errors.Is
func (e *RepositoryError) Is(target error) bool {
	return errors.Is(e.SentinelError, target)
}

// WrapError wraps a technical error with a user-friendly sentinel error
func WrapError(sentinelErr, innerErr error) error {
	if innerErr == nil {
		return nil
	}
	return &RepositoryError{
		SentinelError: sentinelErr,
		InnerError:    innerErr,
	}
}

// GetSentinelError extracts the sentinel error from a wrapped error
func GetSentinelError(err error) error {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.SentinelError
	}
	return err
}
