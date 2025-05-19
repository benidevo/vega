package models

import (
	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
)

var (
	// Model validation errors
	ErrInvalidSettingsType = commonerrors.New("invalid settings type")
	ErrFieldRequired       = commonerrors.New("field is required")
	ErrInvalidFieldParam   = commonerrors.New("invalid field parameter")

	// Repository errors
	ErrSettingsNotFound       = commonerrors.New("settings not found")
	ErrFailedToGetSettings    = commonerrors.New("failed to get settings")
	ErrFailedToUpdateSettings = commonerrors.New("failed to update settings")

	ErrProfileNotFound = commonerrors.New("profile not found")
)

// WrapError is a convenience function that calls the common errors package
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}

// GetSentinelError is a convenience function that calls the common errors package
func GetSentinelError(err error) error {
	return commonerrors.GetSentinelError(err)
}
