package helpers

import (
	"maps"
	"time"

	"github.com/rs/zerolog"

	"github.com/benidevo/vega/internal/ai/constants"
	"github.com/benidevo/vega/internal/ai/models"
)

// ServiceHelper provides common logging and error handling utilities for AI services
type ServiceHelper struct {
	log zerolog.Logger
}

// NewServiceHelper creates a new service helper with the provided logger
func NewServiceHelper(log zerolog.Logger) *ServiceHelper {
	return &ServiceHelper{
		log: log,
	}
}

// LogOperationStart logs the start of an AI operation
func (h *ServiceHelper) LogOperationStart(operation, applicantName string) {
	h.log.Info().
		Str("applicant", applicantName).
		Str("operation", operation).
		Msg("Starting AI operation")
}

// LogOperationSuccess logs successful completion of an AI operation
func (h *ServiceHelper) LogOperationSuccess(operation, applicantName string, duration time.Duration, enhanced bool, metadata map[string]interface{}) {
	event := h.log.Info().
		Str("applicant", applicantName).
		Str("operation", operation).
		Dur("duration", duration).
		Bool("enhanced", enhanced).
		Bool("success", true)

	for key, value := range metadata {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case float64:
			event = event.Float64(key, v)
		case bool:
			event = event.Bool(key, v)
		}
	}

	event.Msg("AI operation completed successfully")
}

// LogValidationError logs validation errors with context
func (h *ServiceHelper) LogValidationError(operation, applicantName string, err error) error {
	h.log.Error().
		Err(err).
		Str("applicant", applicantName).
		Str("operation", operation).
		Str("error_type", constants.ErrorTypeValidationFailed).
		Msg("Validation failed")
	return err
}

// LogOperationError logs operation failures with context
func (h *ServiceHelper) LogOperationError(operation, applicantName, errorType string, duration time.Duration, err error) error {
	h.log.Error().
		Err(err).
		Str("applicant", applicantName).
		Str("operation", operation).
		Str("error_type", errorType).
		Dur("duration", duration).
		Msg("AI operation failed")
	return err
}

// WrapValidationError wraps validation errors with consistent messaging
func (h *ServiceHelper) WrapValidationError(err error) error {
	return models.WrapError(models.ErrValidationFailed, err)
}

// CreateOperationMetadata creates standard metadata for operations
func (h *ServiceHelper) CreateOperationMetadata(temperature float32, enhanced bool, additionalData map[string]interface{}) map[string]interface{} {
	metadata := map[string]interface{}{
		"temperature": temperature,
		"enhanced":    enhanced,
	}

	maps.Copy(metadata, additionalData)

	return metadata
}
