package job

import (
	"github.com/benidevo/prospector/internal/job/models"
)

// Re-export all errors from models package
var (
	// Model validation errors
	ErrInvalidJobType         = models.ErrInvalidJobType
	ErrInvalidJobStatus       = models.ErrInvalidJobStatus
	ErrInvalidExperienceLevel = models.ErrInvalidExperienceLevel
	ErrJobTitleRequired       = models.ErrJobTitleRequired
	ErrJobDescriptionRequired = models.ErrJobDescriptionRequired
	ErrCompanyRequired        = models.ErrCompanyRequired

	// Repository errors
	ErrJobNotFound           = models.ErrJobNotFound
	ErrCompanyNotFound       = models.ErrCompanyNotFound
	ErrCompanyNameRequired   = models.ErrCompanyNameRequired
	ErrDuplicateJob          = models.ErrDuplicateJob
	ErrTransactionFailed     = models.ErrTransactionFailed
	ErrInvalidJobID          = models.ErrInvalidJobID
	ErrInvalidCompanyID      = models.ErrInvalidCompanyID
	ErrFailedToCreateJob     = models.ErrFailedToCreateJob
	ErrFailedToUpdateJob     = models.ErrFailedToUpdateJob
	ErrFailedToDeleteJob     = models.ErrFailedToDeleteJob
	ErrFailedToCreateCompany = models.ErrFailedToCreateCompany
	ErrFailedToUpdateCompany = models.ErrFailedToUpdateCompany
	ErrFailedToDeleteCompany = models.ErrFailedToDeleteCompany
)

// Re-export utility functions from models
var (
	WrapError        = models.WrapError
	GetSentinelError = models.GetSentinelError
)

// Re-export repository error type for backward compatibility
type RepositoryError = models.RepositoryError
