package job

import (
	"context"
	"time"

	"github.com/benidevo/prospector/internal/config"
	"github.com/benidevo/prospector/internal/job/interfaces"
	"github.com/benidevo/prospector/internal/job/models"
	"github.com/benidevo/prospector/internal/logger"
	"github.com/rs/zerolog"
)

// JobService provides business logic for job management.
type JobService struct {
	jobRepo interfaces.JobRepository
	cfg     *config.Settings
	log     zerolog.Logger
}

// NewJobService creates a new JobService instance.
func NewJobService(jobRepo interfaces.JobRepository, cfg *config.Settings) *JobService {
	return &JobService{
		jobRepo: jobRepo,
		cfg:     cfg,
		log:     logger.GetLogger("job"),
	}
}

// CreateJob creates a new job with the given title, description, and company name.
// Additional job options can be provided.
func (s *JobService) CreateJob(ctx context.Context, title, description, companyName string, options ...models.JobOption) (*models.Job, error) {
	s.log.Debug().
		Str("title", title).
		Str("company", companyName).
		Msg("Creating new job")

	company := models.Company{Name: companyName}
	job := models.NewJob(title, description, company, options...)

	if err := job.Validate(); err != nil {
		s.log.Error().
			Str("title", title).
			Str("company", companyName).
			Err(err).
			Msg("Job validation failed")
		return nil, err
	}

	createdJob, err := s.jobRepo.Create(ctx, job)
	if err != nil {
		s.log.Error().
			Str("title", title).
			Str("company", companyName).
			Err(err).
			Msg("Failed to create job")
		return nil, err
	}

	s.log.Info().
		Int("job_id", createdJob.ID).
		Str("title", createdJob.Title).
		Str("company", createdJob.Company.Name).
		Msg("Job created successfully")

	return createdJob, nil
}

// GetJob retrieves a job by its ID.
func (s *JobService) GetJob(ctx context.Context, id int) (*models.Job, error) {
	s.log.Debug().Int("job_id", id).Msg("Getting job by ID")

	if id <= 0 {
		s.log.Error().Int("job_id", id).Msg("Invalid job ID")
		return nil, models.ErrInvalidJobID
	}

	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().
			Int("job_id", id).
			Err(err).
			Msg("Failed to get job")
		return nil, err
	}

	s.log.Debug().
		Int("job_id", job.ID).
		Str("title", job.Title).
		Msg("Job retrieved successfully")

	return job, nil
}

// GetJobs retrieves jobs based on the provided filter.
func (s *JobService) GetJobs(ctx context.Context, filter models.JobFilter) ([]*models.Job, error) {
	logEvent := s.log.Debug()
	if filter.Search != "" {
		logEvent = logEvent.Str("search", filter.Search)
	}
	if filter.CompanyID != nil {
		logEvent = logEvent.Int("company_id", *filter.CompanyID)
	}
	if filter.Status != nil {
		logEvent = logEvent.Str("status", (*filter.Status).String())
	}
	if filter.JobType != nil {
		logEvent = logEvent.Str("job_type", (*filter.JobType).String())
	}
	if filter.Limit > 0 {
		logEvent = logEvent.Int("limit", filter.Limit)
	}
	if filter.Offset > 0 {
		logEvent = logEvent.Int("offset", filter.Offset)
	}
	logEvent.Msg("Getting jobs with filter")

	jobs, err := s.jobRepo.GetAll(ctx, filter)
	if err != nil {
		s.log.Error().
			Err(err).
			Msg("Failed to get jobs with filter")
		return nil, err
	}

	s.log.Debug().
		Int("count", len(jobs)).
		Msg("Jobs retrieved successfully")

	return jobs, nil
}

// UpdateJob updates a job's details.
func (s *JobService) UpdateJob(ctx context.Context, job *models.Job) error {
	if job == nil {
		s.log.Error().Msg("Attempted to update nil job")
		return models.ErrInvalidJobID
	}

	if job.ID <= 0 {
		s.log.Error().Int("job_id", job.ID).Msg("Invalid job ID")
		return models.ErrInvalidJobID
	}

	s.log.Debug().
		Int("job_id", job.ID).
		Str("title", job.Title).
		Msg("Updating job")

	if err := job.Validate(); err != nil {
		s.log.Error().
			Int("job_id", job.ID).
			Err(err).
			Msg("Job validation failed")
		return err
	}

	job.UpdatedAt = time.Now()

	err := s.jobRepo.Update(ctx, job)
	if err != nil {
		s.log.Error().
			Int("job_id", job.ID).
			Err(err).
			Msg("Failed to update job")
		return err
	}

	s.log.Info().
		Int("job_id", job.ID).
		Str("title", job.Title).
		Msg("Job updated successfully")

	return nil
}

// DeleteJob removes a job by its ID.
func (s *JobService) DeleteJob(ctx context.Context, id int) error {
	s.log.Debug().Int("job_id", id).Msg("Deleting job")

	if id <= 0 {
		s.log.Error().Int("job_id", id).Msg("Invalid job ID")
		return models.ErrInvalidJobID
	}

	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().
			Int("job_id", id).
			Err(err).
			Msg("Job not found for deletion")
		return err
	}

	err = s.jobRepo.Delete(ctx, id)
	if err != nil {
		s.log.Error().
			Int("job_id", id).
			Err(err).
			Msg("Failed to delete job")
		return err
	}

	s.log.Info().
		Int("job_id", id).
		Str("title", job.Title).
		Msg("Job deleted successfully")

	return nil
}

// UpdateJobStatus updates only the status of a job.
func (s *JobService) UpdateJobStatus(ctx context.Context, id int, status models.JobStatus) error {
	s.log.Debug().
		Int("job_id", id).
		Str("status", status.String()).
		Msg("Updating job status")

	if id <= 0 {
		s.log.Error().Int("job_id", id).Msg("Invalid job ID")
		return models.ErrInvalidJobID
	}

	if status < models.INTERESTED || status > models.NOT_INTERESTED {
		s.log.Error().
			Int("job_id", id).
			Int("status_value", int(status)).
			Msg("Invalid job status")
		return models.ErrInvalidJobStatus
	}

	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		s.log.Error().
			Int("job_id", id).
			Err(err).
			Msg("Job not found for status update")
		return err
	}

	err = s.jobRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		s.log.Error().
			Int("job_id", id).
			Str("status", status.String()).
			Err(err).
			Msg("Failed to update job status")
		return err
	}

	s.log.Info().
		Int("job_id", id).
		Str("title", job.Title).
		Str("old_status", job.Status.String()).
		Str("new_status", status.String()).
		Msg("Job status updated successfully")

	return nil
}
