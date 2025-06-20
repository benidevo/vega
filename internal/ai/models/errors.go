package models

import (
	commonerrors "github.com/benidevo/vega/internal/common/errors"
)

var (
	// Validation errors
	ErrValidationFailed = commonerrors.New("request validation failed")

	// Setup errors
	ErrProviderInitFailed  = commonerrors.New("failed to initialize AI provider")
	ErrUnsupportedProvider = commonerrors.New("unsupported AI provider")
	ErrMissingAPIKey       = commonerrors.New("missing API key for AI provider")
)

// WrapError wraps the given innerErr with the provided sentinelErr.
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}
