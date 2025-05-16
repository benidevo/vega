package job

import (
	"context"
	"net/url"
	"strconv"
	"strings"
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

// ValidateFieldName checks if a field name is valid for job updates.
// It returns an error if the field is empty or not one of the allowed values.
func (s *JobService) ValidateFieldName(field string) error {
	if field == "" {
		s.log.Error().Msg("Field parameter is required")
		return models.ErrFieldRequired
	}

	validFields := map[string]bool{
		"status": true,
		"notes":  true,
		"skills": true,
		"basic":  true,
	}

	if !validFields[field] {
		s.log.Error().Str("field", field).Msg("Invalid field parameter")
		return models.ErrInvalidFieldParam
	}

	return nil
}

// ValidateJobIDFormat validates that a job ID string can be parsed as an integer.
func (s *JobService) ValidateJobIDFormat(jobIDStr string) (int, error) {
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		s.log.Error().Str("job_id", jobIDStr).Msg("Invalid job ID format")
		return 0, models.ErrInvalidJobIDFormat
	}
	return jobID, nil
}

// ValidateAndFilterSkills processes a comma-separated string of skills
// and returns only the non-empty skills after trimming whitespace.
// It returns an error if no valid skills are found.
func (s *JobService) ValidateAndFilterSkills(skillsStr string) ([]string, error) {
	if skillsStr == "" {
		s.log.Error().Msg("Empty skills string provided")
		return nil, models.ErrSkillsRequired
	}

	// Split and clean up skills
	rawSkills := strings.Split(skillsStr, ",")
	skills := make([]string, 0, len(rawSkills))

	// Only add non-empty skills
	for _, skill := range rawSkills {
		trimmedSkill := strings.TrimSpace(skill)
		if trimmedSkill != "" {
			skills = append(skills, trimmedSkill)
		}
	}

	if len(skills) == 0 {
		s.log.Error().Msg("No valid skills found after processing")
		return nil, models.ErrSkillsRequired
	}

	return skills, nil
}

// ValidateURL checks if a URL string is valid
func (s *JobService) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return nil // Empty URL is allowed
	}

	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		s.log.Error().Str("url", urlStr).Msg("Invalid URL format")
		return models.ErrInvalidURLFormat
	}

	return nil
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

// GetJobStats retrieves job statistics from the repository.
//
// If an error occurs during retrieval, it logs the error and returns an empty JobStats struct.
func (s *JobService) GetJobStats(ctx context.Context) *models.JobStats {
	s.log.Debug().Msg("Getting job statistics")

	stats, err := s.jobRepo.GetStats(ctx)
	if err != nil {
		s.log.Error().
			Err(err).
			Msg("Failed to get job statistics")

		return &models.JobStats{}
	}

	s.log.Debug().
		Int("total_jobs", stats.TotalJobs).
		Int("applied_jobs", stats.TotalApplied).
		Int("high_match", stats.HighMatch).
		Msg("Job statistics retrieved successfully")

	return stats
}

// GetJobs retrieves jobs based on the provided filter.
func (s *JobService) GetJobs(ctx context.Context, filter models.JobFilter) ([]*models.Job, error) {
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

	currentJob, err := s.jobRepo.GetByID(ctx, job.ID)
	if err != nil {
		s.log.Error().
			Int("job_id", job.ID).
			Err(err).
			Msg("Failed to get current job state")
		return err
	}

	if currentJob.Status != job.Status {
		if !models.IsValidTransition(currentJob.Status, job.Status) {
			s.log.Error().
				Int("job_id", job.ID).
				Str("current_status", currentJob.Status.String()).
				Str("new_status", job.Status.String()).
				Msg("Invalid job status transition")
			return models.ErrInvalidStatusTransition
		}
	}

	job.UpdatedAt = time.Now().UTC()

	err = s.jobRepo.Update(ctx, job)
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

	if !models.IsValidTransition(job.Status, status) {
		s.log.Error().
			Int("job_id", id).
			Str("current_status", job.Status.String()).
			Str("new_status", status.String()).
			Msg("Invalid job status transition")
		return models.ErrInvalidStatusTransition
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
