package repository

import (
	"github.com/benidevo/prospector/internal/job"
)

var (
	// Company errors
	ErrCompanyNotFound     = job.ErrCompanyNotFound
	ErrCompanyNameRequired = job.ErrCompanyNameRequired

	// Transaction errors
	ErrTransactionFailed = job.ErrTransactionFailed

	// Job errors
	ErrJobNotFound            = job.ErrJobNotFound
	ErrJobTitleRequired       = job.ErrJobTitleRequired
	ErrJobDescriptionRequired = job.ErrJobDescriptionRequired
	ErrDuplicateJob           = job.ErrDuplicateJob
	ErrInvalidJobID           = job.ErrInvalidJobID
	ErrInvalidCompanyID       = job.ErrInvalidCompanyID
	ErrFailedToCreateJob      = job.ErrFailedToCreateJob
	ErrFailedToUpdateJob      = job.ErrFailedToUpdateJob
	ErrFailedToDeleteJob      = job.ErrFailedToDeleteJob
	ErrFailedToCreateCompany  = job.ErrFailedToCreateCompany
	ErrFailedToUpdateCompany  = job.ErrFailedToUpdateCompany
	ErrFailedToDeleteCompany  = job.ErrFailedToDeleteCompany
)
