package repository

import (
	"github.com/benidevo/prospector/internal/job/models"
)

var (
	// Company errors
	ErrCompanyNotFound     = models.ErrCompanyNotFound
	ErrCompanyNameRequired = models.ErrCompanyNameRequired

	// Transaction errors
	ErrTransactionFailed = models.ErrTransactionFailed

	// Job errors
	ErrJobNotFound            = models.ErrJobNotFound
	ErrJobTitleRequired       = models.ErrJobTitleRequired
	ErrJobDescriptionRequired = models.ErrJobDescriptionRequired
	ErrDuplicateJob           = models.ErrDuplicateJob
	ErrInvalidJobID           = models.ErrInvalidJobID
	ErrInvalidCompanyID       = models.ErrInvalidCompanyID
	ErrFailedToCreateJob      = models.ErrFailedToCreateJob
	ErrFailedToUpdateJob      = models.ErrFailedToUpdateJob
	ErrFailedToDeleteJob      = models.ErrFailedToDeleteJob
	ErrFailedToCreateCompany  = models.ErrFailedToCreateCompany
	ErrFailedToUpdateCompany  = models.ErrFailedToUpdateCompany
	ErrFailedToDeleteCompany  = models.ErrFailedToDeleteCompany
)
