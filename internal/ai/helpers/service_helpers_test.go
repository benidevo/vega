package helpers

import (
	"errors"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/ai/constants"
	"github.com/benidevo/vega/internal/common/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestNewServiceHelper(t *testing.T) {
	log := logger.GetPrivacyLogger("test")
	helper := NewServiceHelper(log)

	assert.NotNil(t, helper)
	assert.NotNil(t, helper.log)
}

func TestServiceHelper_LogOperationStart(t *testing.T) {
	log := logger.GetPrivacyLogger("test")
	helper := NewServiceHelper(log)

	assert.NotPanics(t, func() {
		helper.LogOperationStart("test_operation", "John Doe")
	})
}

func TestServiceHelper_LogOperationSuccess(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		applicant string
		duration  time.Duration
		enhanced  bool
		metadata  map[string]interface{}
	}{
		{
			name:      "should_log_success_with_string_metadata",
			operation: "cv_generation",
			applicant: "Jane Smith",
			duration:  2 * time.Second,
			enhanced:  true,
			metadata: map[string]interface{}{
				"model":     "gpt-4",
				"task_type": "cv",
			},
		},
		{
			name:      "should_log_success_with_numeric_metadata",
			operation: "job_matching",
			applicant: "Bob Wilson",
			duration:  1 * time.Second,
			enhanced:  false,
			metadata: map[string]interface{}{
				"match_score": 85,
				"confidence":  0.92,
			},
		},
		{
			name:      "should_log_success_with_bool_metadata",
			operation: "cover_letter",
			applicant: "Alice Brown",
			duration:  3 * time.Second,
			enhanced:  true,
			metadata: map[string]interface{}{
				"cached":    true,
				"formatted": false,
			},
		},
		{
			name:      "should_log_success_with_empty_metadata",
			operation: "analysis",
			applicant: "Test User",
			duration:  500 * time.Millisecond,
			enhanced:  false,
			metadata:  map[string]interface{}{},
		},
		{
			name:      "should_log_success_with_nil_metadata",
			operation: "parsing",
			applicant: "Another User",
			duration:  1 * time.Second,
			enhanced:  true,
			metadata:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.GetPrivacyLogger("test")
			helper := NewServiceHelper(log)

			assert.NotPanics(t, func() {
				helper.LogOperationSuccess(tt.operation, tt.applicant, tt.duration, tt.enhanced, tt.metadata)
			})
		})
	}
}

func TestServiceHelper_LogValidationError(t *testing.T) {
	log := logger.GetPrivacyLogger("test")
	helper := NewServiceHelper(log)

	originalErr := errors.New("invalid input")
	err := helper.LogValidationError("test_operation", "John Doe", originalErr)

	assert.Equal(t, originalErr, err)
}

func TestServiceHelper_LogOperationError(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		applicant   string
		errorType   string
		duration    time.Duration
		originalErr error
	}{
		{
			name:        "should_log_ai_analysis_error",
			operation:   "job_analysis",
			applicant:   "Test User",
			errorType:   constants.ErrorTypeAIAnalysisFailed,
			duration:    2 * time.Second,
			originalErr: errors.New("AI service unavailable"),
		},
		{
			name:        "should_log_response_parse_error",
			operation:   "cv_parsing",
			applicant:   "Another User",
			errorType:   constants.ErrorTypeResponseParseFailed,
			duration:    1 * time.Second,
			originalErr: errors.New("invalid JSON response"),
		},
		{
			name:        "should_log_validation_error",
			operation:   "cover_letter",
			applicant:   "Third User",
			errorType:   constants.ErrorTypeValidationFailed,
			duration:    500 * time.Millisecond,
			originalErr: errors.New("missing required fields"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.GetPrivacyLogger("test")
			helper := NewServiceHelper(log)

			err := helper.LogOperationError(tt.operation, tt.applicant, tt.errorType, tt.duration, tt.originalErr)

			assert.Equal(t, tt.originalErr, err)
		})
	}
}

func TestServiceHelper_WrapValidationError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectNil bool
	}{
		{
			name:      "should_wrap_error_with_validation_failed_when_error_provided",
			err:       errors.New("missing required field: name"),
			expectNil: false,
		},
		{
			name:      "should_handle_nil_error_when_nil_provided",
			err:       nil,
			expectNil: true,
		},
		{
			name:      "should_wrap_complex_error_when_provided",
			err:       errors.New("multiple validation failures: field1, field2, field3"),
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := logger.GetPrivacyLogger("test")
			helper := NewServiceHelper(log)

			wrappedErr := helper.WrapValidationError(tt.err)

			if tt.expectNil {
				assert.Nil(t, wrappedErr)
			} else {
				assert.NotNil(t, wrappedErr)
				if tt.err != nil {
					assert.Contains(t, wrappedErr.Error(), tt.err.Error())
				}
			}
		})
	}
}

func TestServiceHelper_CreateOperationMetadata(t *testing.T) {
	log := logger.GetPrivacyLogger("test")
	helper := NewServiceHelper(log)

	tests := []struct {
		name           string
		temperature    float32
		enhanced       bool
		additionalData map[string]interface{}
		expectedKeys   []string
	}{
		{
			name:        "should_create_metadata_with_additional_data",
			temperature: 0.7,
			enhanced:    true,
			additionalData: map[string]interface{}{
				"model":      "gpt-4",
				"max_tokens": 1000,
			},
			expectedKeys: []string{"temperature", "enhanced", "model", "max_tokens"},
		},
		{
			name:           "should_create_metadata_without_additional_data",
			temperature:    0.5,
			enhanced:       false,
			additionalData: nil,
			expectedKeys:   []string{"temperature", "enhanced"},
		},
		{
			name:           "should_create_metadata_with_empty_additional_data",
			temperature:    0.9,
			enhanced:       true,
			additionalData: map[string]interface{}{},
			expectedKeys:   []string{"temperature", "enhanced"},
		},
		{
			name:        "should_override_base_keys_with_additional_data",
			temperature: 0.7,
			enhanced:    true,
			additionalData: map[string]interface{}{
				"temperature": 0.9,
				"enhanced":    false,
				"extra":       "value",
			},
			expectedKeys: []string{"temperature", "enhanced", "extra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := helper.CreateOperationMetadata(tt.temperature, tt.enhanced, tt.additionalData)

			assert.Len(t, metadata, len(tt.expectedKeys))

			for _, key := range tt.expectedKeys {
				assert.Contains(t, metadata, key)
			}

			if tt.additionalData != nil && tt.additionalData["temperature"] != nil {
				assert.Equal(t, tt.additionalData["temperature"], metadata["temperature"])
			} else {
				assert.Equal(t, tt.temperature, metadata["temperature"])
			}

			if tt.additionalData != nil && tt.additionalData["enhanced"] != nil {
				assert.Equal(t, tt.additionalData["enhanced"], metadata["enhanced"])
			} else {
				assert.Equal(t, tt.enhanced, metadata["enhanced"])
			}
		})
	}
}
