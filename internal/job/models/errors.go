package models

import (
	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
)

var (
	// Model validation errors
	ErrInvalidJobStatus        = commonerrors.New("invalid job status")
	ErrInvalidStatusTransition = commonerrors.New("job status can only move forward")
	ErrInvalidExperienceLevel  = commonerrors.New("invalid experience level")
	ErrJobTitleRequired        = commonerrors.New("job title is required")
	ErrJobDescriptionRequired  = commonerrors.New("job description is required")
	ErrCompanyRequired         = commonerrors.New("job company is required")
	ErrInvalidFieldParam       = commonerrors.New("invalid field parameter")
	ErrFieldRequired           = commonerrors.New("field parameter is required")
	ErrInvalidURLFormat        = commonerrors.New("invalid URL format")
	ErrSkillsRequired          = commonerrors.New("at least one valid skill is required")
	ErrStatusRequired          = commonerrors.New("status is required")
	ErrInvalidJobIDFormat      = commonerrors.New("invalid job ID format")

	// Repository errors
	ErrJobNotFound           = commonerrors.New("job not found")
	ErrCompanyNotFound       = commonerrors.New("company not found")
	ErrCompanyNameRequired   = commonerrors.New("company name is required")
	ErrDuplicateJob          = commonerrors.New("job with this URL already exists")
	ErrTransactionFailed     = commonerrors.New("database transaction failed")
	ErrInvalidJobID          = commonerrors.New("invalid job ID")
	ErrInvalidCompanyID      = commonerrors.New("invalid company ID")
	ErrFailedToCreateJob     = commonerrors.New("failed to create job")
	ErrFailedToUpdateJob     = commonerrors.New("failed to update job")
	ErrFailedToDeleteJob     = commonerrors.New("failed to delete job")
	ErrFailedToCreateCompany = commonerrors.New("failed to create company")
	ErrFailedToUpdateCompany = commonerrors.New("failed to update company")
	ErrFailedToDeleteCompany = commonerrors.New("failed to delete company")
	ErrFailedToGetJobStats   = commonerrors.New("failed to get job stats")
	ErrFailedToGetJob        = commonerrors.New("failed to get job")
)

// WrapError is a convenience function that calls the common errors package
func WrapError(sentinelErr, innerErr error) error {
	return commonerrors.WrapError(sentinelErr, innerErr)
}

// GetSentinelError is a convenience function that calls the common errors package
func GetSentinelError(err error) error {
	return commonerrors.GetSentinelError(err)
}
