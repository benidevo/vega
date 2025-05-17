package models

import (
	"errors"
	"fmt"
)

var (
	// Model validation errors
	ErrInvalidSettingsType = errors.New("invalid settings type")
	ErrFieldRequired       = errors.New("field is required")
	ErrInvalidFieldParam   = errors.New("invalid field parameter")

	// Repository errors
	ErrSettingsNotFound       = errors.New("settings not found")
	ErrFailedToGetSettings    = errors.New("failed to get settings")
	ErrFailedToUpdateSettings = errors.New("failed to update settings")

	ErrProfileNotFound = errors.New("profile not found")
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
