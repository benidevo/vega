package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "should_have_validation_failed_error_message",
			err:      ErrValidationFailed,
			expected: "request validation failed",
		},
		{
			name:     "should_have_provider_init_failed_error_message",
			err:      ErrProviderInitFailed,
			expected: "failed to initialize AI provider",
		},
		{
			name:     "should_have_unsupported_provider_error_message",
			err:      ErrUnsupportedProvider,
			expected: "unsupported AI provider",
		},
		{
			name:     "should_have_missing_api_key_error_message",
			err:      ErrMissingAPIKey,
			expected: "missing API key for AI provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		sentinelErr  error
		innerErr     error
		expectNil    bool
		containsText []string
	}{
		{
			name:        "should_wrap_error_with_sentinel_and_inner_when_both_provided",
			sentinelErr: ErrValidationFailed,
			innerErr:    errors.New("missing required field"),
			expectNil:   false,
			containsText: []string{
				"request validation failed",
				"missing required field",
			},
		},
		{
			name:         "should_wrap_different_errors_when_provider_init_fails",
			sentinelErr:  ErrProviderInitFailed,
			innerErr:     errors.New("connection timeout"),
			expectNil:    false,
			containsText: []string{"failed to initialize AI provider", "connection timeout"},
		},
		{
			name:         "should_wrap_error_with_unsupported_provider_when_provider_unknown",
			sentinelErr:  ErrUnsupportedProvider,
			innerErr:     errors.New("provider 'xyz' not found"),
			expectNil:    false,
			containsText: []string{"unsupported AI provider", "provider 'xyz' not found"},
		},
		{
			name:         "should_wrap_error_with_missing_api_key_when_key_absent",
			sentinelErr:  ErrMissingAPIKey,
			innerErr:     errors.New("OPENAI_API_KEY not set"),
			expectNil:    false,
			containsText: []string{"missing API key for AI provider", "OPENAI_API_KEY not set"},
		},
		{
			name:         "should_return_nil_when_inner_error_is_nil",
			sentinelErr:  ErrValidationFailed,
			innerErr:     nil,
			expectNil:    true,
			containsText: []string{},
		},
		{
			name:         "should_handle_custom_sentinel_error_when_wrapping",
			sentinelErr:  errors.New("custom error"),
			innerErr:     errors.New("inner details"),
			expectNil:    false,
			containsText: []string{"custom error", "inner details"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.sentinelErr, tt.innerErr)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)

				errorMsg := result.Error()
				for _, text := range tt.containsText {
					assert.Contains(t, errorMsg, text)
				}
			}
		})
	}
}

func TestErrorUsage(t *testing.T) {
	t.Run("should_be_able_to_use_errors_with_errors_is_when_checking_type", func(t *testing.T) {
		innerErr := errors.New("database connection failed")
		wrappedErr := WrapError(ErrProviderInitFailed, innerErr)

		assert.True(t, errors.Is(wrappedErr, ErrProviderInitFailed))

		assert.Contains(t, wrappedErr.Error(), innerErr.Error())
	})

	t.Run("should_maintain_error_info_when_wrapping_multiple_times", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrap1 := WrapError(ErrMissingAPIKey, baseErr)
		wrap2 := WrapError(ErrProviderInitFailed, wrap1)

		assert.True(t, errors.Is(wrap2, ErrProviderInitFailed))

		errorMsg := wrap2.Error()
		assert.Contains(t, errorMsg, "failed to initialize AI provider")
		assert.Contains(t, errorMsg, "missing API key")
		assert.Contains(t, errorMsg, "base error")
	})
}
