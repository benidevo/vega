package models

import (
	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
)

var (
	// Validation errors
	ErrValidationFailed = commonerrors.New("request validation failed")
)

// WrapError wraps the given innerErr with the provided sentinelErr.
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}
