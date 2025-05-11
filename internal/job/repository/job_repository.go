package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/benidevo/prospector/internal/job"
	"github.com/benidevo/prospector/internal/job/models"
)

// JobRepository defines methods for interacting with job data
type JobRepository interface {
	Create(ctx context.Context, job *models.Job) (*models.Job, error)
	GetByID(ctx context.Context, id int) (*models.Job, error)
	GetAll(ctx context.Context, filter JobFilter) ([]*models.Job, error)
	Update(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, id int) error
	UpdateStatus(ctx context.Context, id int, status models.JobStatus) error
}

// JobFilter defines filters for querying jobs
type JobFilter struct {
	CompanyID *int
	Status    *models.JobStatus
	JobType   *models.JobType
	Search    string
	Limit     int
	Offset    int
}

// SQLiteJobRepository is a SQLite implementation of JobRepository
type SQLiteJobRepository struct {
	db                *sql.DB
	companyRepository CompanyRepository
}

// NewSQLiteJobRepository creates a new SQLiteJobRepository instance
func NewSQLiteJobRepository(db *sql.DB, companyRepository CompanyRepository) *SQLiteJobRepository {
	return &SQLiteJobRepository{
		db:                db,
		companyRepository: companyRepository,
	}
}

// validateJob performs basic validation on a job
func validateJob(jobModel *models.Job) error {
	if jobModel == nil {
		return job.ErrInvalidJobID
	}

	return jobModel.Validate()
}

// Create inserts a new job into the database
func (r *SQLiteJobRepository) Create(ctx context.Context, jobModel *models.Job) (*models.Job, error) {
	if err := validateJob(jobModel); err != nil {
		return nil, err
	}
	company, err := r.companyRepository.GetOrCreate(ctx, jobModel.Company.Name)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, wrapError(job.ErrTransactionFailed, err)
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now()
	if jobModel.CreatedAt.IsZero() {
		jobModel.CreatedAt = now
	}
	jobModel.UpdatedAt = now
	if jobModel.RequiredSkills == nil {
		jobModel.RequiredSkills = []string{}
	}

	skillsJSON, err := json.Marshal(jobModel.RequiredSkills)
	if err != nil {
		return nil, wrapError(job.ErrFailedToCreateJob, err)
	}

	query := `
		INSERT INTO jobs (
			title, description, location, job_type, source_url,
			salary_range, required_skills, application_deadline, application_url,
			company_id, status, experience_level, contact_person, notes,
			posted_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var postedAt *time.Time
	if jobModel.PostedAt != nil {
		postedAt = jobModel.PostedAt
	}

	var deadline *time.Time
	if jobModel.ApplicationDeadline != nil {
		deadline = jobModel.ApplicationDeadline
	}

	result, err := tx.ExecContext(
		ctx,
		query,
		jobModel.Title,
		jobModel.Description,
		jobModel.Location,
		int(jobModel.JobType),
		jobModel.SourceURL,
		jobModel.SalaryRange,
		skillsJSON,
		deadline,
		jobModel.ApplicationURL,
		company.ID,
		int(jobModel.Status),
		int(jobModel.ExperienceLevel),
		jobModel.ContactPerson,
		jobModel.Notes,
		postedAt,
		jobModel.CreatedAt,
		jobModel.UpdatedAt,
	)
	if err != nil {
		return nil, wrapError(job.ErrFailedToCreateJob, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, wrapError(job.ErrFailedToCreateJob, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, wrapError(job.ErrTransactionFailed, err)
	}

	tx = nil

	jobModel.ID = int(id)
	jobModel.Company = *company

	return jobModel, nil
}

// GetByID retrieves a job by its ID
func (r *SQLiteJobRepository) GetByID(ctx context.Context, id int) (*models.Job, error) {
	if id <= 0 {
		return nil, job.ErrInvalidJobID
	}

	query := `
		SELECT
			j.id, j.title, j.description, j.location, j.job_type,
			j.source_url, j.salary_range, j.required_skills, j.application_deadline,
			j.application_url, j.company_id, j.status, j.experience_level,
			j.contact_person, j.notes, j.posted_at, j.created_at, j.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM jobs j
		JOIN companies c ON j.company_id = c.id
		WHERE j.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var j models.Job
	var company models.Company
	var skillsJSON string
	var jobType, status, experienceLevel int
	var applicationDeadline, postedAt sql.NullTime
	var contactPerson, notes, sourceURL, applicationURL, salaryRange, location sql.NullString

	err := row.Scan(
		&j.ID, &j.Title, &j.Description, &location, &jobType,
		&sourceURL, &salaryRange, &skillsJSON, &applicationDeadline,
		&applicationURL, &company.ID, &status, &experienceLevel,
		&contactPerson, &notes, &postedAt, &j.CreatedAt, &j.UpdatedAt,
		&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, job.ErrJobNotFound
		}
		return nil, &job.RepositoryError{
			SentinelError: job.ErrJobNotFound,
			InnerError:    err,
		}
	}

	if location.Valid {
		j.Location = location.String
	}
	if sourceURL.Valid {
		j.SourceURL = sourceURL.String
	}
	if salaryRange.Valid {
		j.SalaryRange = salaryRange.String
	}
	if applicationURL.Valid {
		j.ApplicationURL = applicationURL.String
	}
	if contactPerson.Valid {
		j.ContactPerson = contactPerson.String
	}
	if notes.Valid {
		j.Notes = notes.String
	}
	if applicationDeadline.Valid {
		j.ApplicationDeadline = &applicationDeadline.Time
	}
	if postedAt.Valid {
		j.PostedAt = &postedAt.Time
	}

	if err := json.Unmarshal([]byte(skillsJSON), &j.RequiredSkills); err != nil {
		j.RequiredSkills = []string{}
	}

	j.JobType = models.JobType(jobType)
	j.Status = models.JobStatus(status)
	j.ExperienceLevel = models.ExperienceLevel(experienceLevel)

	j.Company = company

	return &j, nil
}

// GetAll retrieves all jobs with optional filtering
func (r *SQLiteJobRepository) GetAll(ctx context.Context, filter JobFilter) ([]*models.Job, error) {
	query := `
		SELECT
			j.id, j.title, j.description, j.location, j.job_type,
			j.source_url, j.salary_range, j.required_skills, j.application_deadline,
			j.application_url, j.company_id, j.status, j.experience_level,
			j.contact_person, j.notes, j.posted_at, j.created_at, j.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM jobs j
		JOIN companies c ON j.company_id = c.id
	`

	var conditions []string
	var args []interface{}

	if filter.CompanyID != nil {
		conditions = append(conditions, "j.company_id = ?")
		args = append(args, *filter.CompanyID)
	}

	if filter.Status != nil {
		conditions = append(conditions, "j.status = ?")
		args = append(args, int(*filter.Status))
	}

	if filter.JobType != nil {
		conditions = append(conditions, "j.job_type = ?")
		args = append(args, int(*filter.JobType))
	}

	if filter.Search != "" {
		conditions = append(conditions, "(j.title LIKE ? OR j.description LIKE ? OR c.name LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY j.updated_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job

	for rows.Next() {
		var j models.Job
		var company models.Company
		var skillsJSON string
		var jobType, status, experienceLevel int
		var applicationDeadline, postedAt sql.NullTime
		var contactPerson, notes, sourceURL, applicationURL, salaryRange, location sql.NullString

		err := rows.Scan(
			&j.ID, &j.Title, &j.Description, &location, &jobType,
			&sourceURL, &salaryRange, &skillsJSON, &applicationDeadline,
			&applicationURL, &company.ID, &status, &experienceLevel,
			&contactPerson, &notes, &postedAt, &j.CreatedAt, &j.UpdatedAt,
			&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if location.Valid {
			j.Location = location.String
		}
		if sourceURL.Valid {
			j.SourceURL = sourceURL.String
		}
		if salaryRange.Valid {
			j.SalaryRange = salaryRange.String
		}
		if applicationURL.Valid {
			j.ApplicationURL = applicationURL.String
		}
		if contactPerson.Valid {
			j.ContactPerson = contactPerson.String
		}
		if notes.Valid {
			j.Notes = notes.String
		}
		if applicationDeadline.Valid {
			j.ApplicationDeadline = &applicationDeadline.Time
		}
		if postedAt.Valid {
			j.PostedAt = &postedAt.Time
		}

		if err := json.Unmarshal([]byte(skillsJSON), &j.RequiredSkills); err != nil {
			j.RequiredSkills = []string{}
		}

		j.JobType = models.JobType(jobType)
		j.Status = models.JobStatus(status)
		j.ExperienceLevel = models.ExperienceLevel(experienceLevel)

		j.Company = company

		jobs = append(jobs, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// Update updates an existing job in the database
func (r *SQLiteJobRepository) Update(ctx context.Context, job *models.Job) error {
	if job == nil {
		return ErrInvalidJobID
	}

	if job.ID <= 0 {
		return ErrInvalidJobID
	}

	if err := validateJob(job); err != nil {
		return err
	}

	company, err := r.companyRepository.GetOrCreate(ctx, job.Company.Name)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ErrTransactionFailed
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	job.UpdatedAt = time.Now()

	skillsJSON, err := json.Marshal(job.RequiredSkills)
	if err != nil {
		return ErrFailedToUpdateJob
	}

	query := `
		UPDATE jobs SET
			title = ?, description = ?, location = ?, job_type = ?,
			source_url = ?, salary_range = ?, required_skills = ?,
			application_deadline = ?, application_url = ?, company_id = ?,
			status = ?, experience_level = ?, contact_person = ?,
			notes = ?, posted_at = ?, updated_at = ?
		WHERE id = ?
	`

	var postedAt *time.Time
	if job.PostedAt != nil {
		postedAt = job.PostedAt
	}

	var deadline *time.Time
	if job.ApplicationDeadline != nil {
		deadline = job.ApplicationDeadline
	}

	result, err := tx.ExecContext(
		ctx,
		query,
		job.Title,
		job.Description,
		job.Location,
		int(job.JobType),
		job.SourceURL,
		job.SalaryRange,
		skillsJSON,
		deadline,
		job.ApplicationURL,
		company.ID,
		int(job.Status),
		int(job.ExperienceLevel),
		job.ContactPerson,
		job.Notes,
		postedAt,
		job.UpdatedAt,
		job.ID,
	)
	if err != nil {
		return ErrFailedToUpdateJob
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ErrFailedToUpdateJob
	}

	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	if err = tx.Commit(); err != nil {
		return ErrTransactionFailed
	}

	tx = nil

	job.Company = *company

	return nil
}

// Delete removes a job from the database
func (r *SQLiteJobRepository) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidJobID
	}

	query := "DELETE FROM jobs WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return ErrFailedToDeleteJob
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ErrFailedToDeleteJob
	}

	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

// UpdateStatus updates the status of a job
func (r *SQLiteJobRepository) UpdateStatus(ctx context.Context, id int, status models.JobStatus) error {
	if id <= 0 {
		return ErrInvalidJobID
	}

	if status < models.INTERESTED || status > models.NOT_INTERESTED {
		return job.ErrInvalidJobStatus
	}

	query := "UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?"

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, int(status), now, id)
	if err != nil {
		return ErrFailedToUpdateJob
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ErrFailedToUpdateJob
	}

	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}
