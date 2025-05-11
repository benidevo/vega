package job

import (
	"errors"
	"fmt"
)

var (
	// Model validation errors
	ErrInvalidJobType         = errors.New("invalid job type")
	ErrInvalidJobStatus       = errors.New("invalid job status")
	ErrInvalidExperienceLevel = errors.New("invalid experience level")
	ErrJobTitleRequired       = errors.New("job title is required")
	ErrJobDescriptionRequired = errors.New("job description is required")
	ErrCompanyRequired        = errors.New("job company is required")

	// Repository errors
	ErrJobNotFound           = errors.New("job not found")
	ErrCompanyNotFound       = errors.New("company not found")
	ErrCompanyNameRequired   = errors.New("company name is required")
	ErrDuplicateJob          = errors.New("job with this URL already exists")
	ErrTransactionFailed     = errors.New("database transaction failed")
	ErrInvalidJobID          = errors.New("invalid job ID")
	ErrInvalidCompanyID      = errors.New("invalid company ID")
	ErrFailedToCreateJob     = errors.New("failed to create job")
	ErrFailedToUpdateJob     = errors.New("failed to update job")
	ErrFailedToDeleteJob     = errors.New("failed to delete job")
	ErrFailedToCreateCompany = errors.New("failed to create company")
	ErrFailedToUpdateCompany = errors.New("failed to update company")
	ErrFailedToDeleteCompany = errors.New("failed to delete company")
)

// RepositoryError wraps a sentinel error with the underlying error details
// This allows for user-friendly error messages while preserving the technical details
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
