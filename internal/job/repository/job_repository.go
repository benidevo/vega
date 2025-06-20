package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	commonerrors "github.com/benidevo/ascentio/internal/common/errors"
	"github.com/benidevo/ascentio/internal/job/interfaces"
	"github.com/benidevo/ascentio/internal/job/models"
)

// SQLiteJobRepository is a SQLite implementation of JobRepository
type SQLiteJobRepository struct {
	db                *sql.DB
	companyRepository interfaces.CompanyRepository
}

// NewSQLiteJobRepository creates a new SQLiteJobRepository instance
func NewSQLiteJobRepository(db *sql.DB, companyRepository interfaces.CompanyRepository) *SQLiteJobRepository {
	return &SQLiteJobRepository{
		db:                db,
		companyRepository: companyRepository,
	}
}

// validateJob performs basic validation on a job
func validateJob(jobModel *models.Job) error {
	if jobModel == nil {
		return models.ErrInvalidJobID
	}

	return jobModel.Validate()
}

// GetOrCreate retrieves a job by its SourceURL or creates it if it does not exist.
// Returns the existing or newly created job, or an error if the operation fails.
func (r *SQLiteJobRepository) GetOrCreate(ctx context.Context, jobModel *models.Job) (*models.Job, error) {
	if err := validateJob(jobModel); err != nil {
		return nil, err
	}

	if jobModel.SourceURL == "" {
		return nil, models.ErrInvalidJobID
	}

	existingJob, err := r.GetBySourceURL(ctx, jobModel.SourceURL)
	if err == nil {
		return existingJob, nil
	}

	if !errors.Is(err, models.ErrJobNotFound) {
		return nil, err
	}

	return r.Create(ctx, jobModel)
}

// GetBySourceURL retrieves a job by its source URL
func (r *SQLiteJobRepository) GetBySourceURL(ctx context.Context, sourceURL string) (*models.Job, error) {
	if sourceURL == "" {
		return nil, models.ErrInvalidJobID
	}

	query := `
		SELECT
			j.id, j.title, j.description, j.location, j.job_type,
			j.source_url, j.required_skills,
			j.application_url, j.company_id, j.status, j.match_score,
			j.notes, j.created_at, j.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM jobs j
		JOIN companies c ON j.company_id = c.id
		WHERE j.source_url = ?
	`

	row := r.db.QueryRowContext(ctx, query, sourceURL)

	var j models.Job
	var company models.Company
	var skillsJSON string
	var jobType, status int
	var matchScore sql.NullInt64
	var notes, jobSourceURL, applicationURL, location sql.NullString

	err := row.Scan(
		&j.ID, &j.Title, &j.Description, &location, &jobType,
		&jobSourceURL, &skillsJSON,
		&applicationURL, &company.ID, &status, &matchScore,
		&notes, &j.CreatedAt, &j.UpdatedAt,
		&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrJobNotFound
		}
		return nil, models.WrapError(models.ErrFailedToGetJob, err)
	}

	if location.Valid {
		j.Location = location.String
	}
	if jobSourceURL.Valid {
		j.SourceURL = jobSourceURL.String
	}
	if applicationURL.Valid {
		j.ApplicationURL = applicationURL.String
	}
	if notes.Valid {
		j.Notes = notes.String
	}
	if matchScore.Valid {
		score := int(matchScore.Int64)
		j.MatchScore = &score
	}

	if err := json.Unmarshal([]byte(skillsJSON), &j.RequiredSkills); err != nil {
		j.RequiredSkills = []string{}
	}

	j.JobType = models.JobType(jobType)
	j.Status = models.JobStatus(status)
	j.Company = company

	return &j, nil
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
		return nil, models.WrapError(models.ErrTransactionFailed, err)
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	if jobModel.CreatedAt.IsZero() {
		jobModel.CreatedAt = now
	}
	jobModel.UpdatedAt = now
	if jobModel.RequiredSkills == nil {
		jobModel.RequiredSkills = []string{}
	}

	skillsJSON, err := json.Marshal(jobModel.RequiredSkills)
	if err != nil {
		return nil, models.WrapError(models.ErrFailedToCreateJob, err)
	}

	query := `
		INSERT INTO jobs (
			title, description, location, job_type, source_url,
			required_skills, application_url,
			company_id, status, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		jobModel.Title,
		jobModel.Description,
		jobModel.Location,
		int(jobModel.JobType),
		jobModel.SourceURL,
		skillsJSON,
		jobModel.ApplicationURL,
		company.ID,
		int(jobModel.Status),
		jobModel.Notes,
		jobModel.CreatedAt,
		jobModel.UpdatedAt,
	)
	if err != nil {
		return nil, models.WrapError(models.ErrFailedToCreateJob, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, models.WrapError(models.ErrFailedToCreateJob, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, models.WrapError(models.ErrTransactionFailed, err)
	}

	tx = nil

	jobModel.ID = int(id)
	jobModel.Company = *company

	return jobModel, nil
}

// GetByID retrieves a job by its ID
func (r *SQLiteJobRepository) GetByID(ctx context.Context, id int) (*models.Job, error) {
	if id <= 0 {
		return nil, models.ErrInvalidJobID
	}

	query := `
		SELECT
			j.id, j.title, j.description, j.location, j.job_type,
			j.source_url, j.required_skills,
			j.application_url, j.company_id, j.status, j.match_score,
			j.notes, j.created_at, j.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM jobs j
		JOIN companies c ON j.company_id = c.id
		WHERE j.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var j models.Job
	var company models.Company
	var skillsJSON string
	var jobType, status int
	var matchScore sql.NullInt64
	var notes, sourceURL, applicationURL, location sql.NullString

	err := row.Scan(
		&j.ID, &j.Title, &j.Description, &location, &jobType,
		&sourceURL, &skillsJSON,
		&applicationURL, &company.ID, &status, &matchScore,
		&notes, &j.CreatedAt, &j.UpdatedAt,
		&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrJobNotFound
		}
		return nil, &commonerrors.RepositoryError{
			SentinelError: models.ErrJobNotFound,
			InnerError:    err,
		}
	}

	if location.Valid {
		j.Location = location.String
	}
	if sourceURL.Valid {
		j.SourceURL = sourceURL.String
	}
	if applicationURL.Valid {
		j.ApplicationURL = applicationURL.String
	}
	if notes.Valid {
		j.Notes = notes.String
	}
	if matchScore.Valid {
		score := int(matchScore.Int64)
		j.MatchScore = &score
	}

	if err := json.Unmarshal([]byte(skillsJSON), &j.RequiredSkills); err != nil {
		j.RequiredSkills = []string{}
	}

	j.JobType = models.JobType(jobType)
	j.Status = models.JobStatus(status)

	j.Company = company

	return &j, nil
}

// GetAll retrieves all jobs with optional filtering
func (r *SQLiteJobRepository) GetAll(ctx context.Context, filter models.JobFilter) ([]*models.Job, error) {
	query := `
		SELECT
			j.id, j.title, j.description, j.location, j.job_type,
			j.source_url, j.required_skills,
			j.application_url, j.company_id, j.status, j.match_score,
			j.notes, j.created_at, j.updated_at,
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

	if filter.Matched != nil {
		if *filter.Matched {
			conditions = append(conditions, "j.match_score >= 70")
		} else {
			conditions = append(conditions, "(j.match_score IS NULL OR j.match_score < 70)")
		}
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
		var jobType, status int
		var matchScore sql.NullInt64
		var notes, sourceURL, applicationURL, location sql.NullString

		err := rows.Scan(
			&j.ID, &j.Title, &j.Description, &location, &jobType,
			&sourceURL, &skillsJSON,
			&applicationURL, &company.ID, &status, &matchScore,
			&notes, &j.CreatedAt, &j.UpdatedAt,
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
		if applicationURL.Valid {
			j.ApplicationURL = applicationURL.String
		}
		if notes.Valid {
			j.Notes = notes.String
		}
		if matchScore.Valid {
			score := int(matchScore.Int64)
			j.MatchScore = &score
		}

		if err := json.Unmarshal([]byte(skillsJSON), &j.RequiredSkills); err != nil {
			j.RequiredSkills = []string{}
		}

		j.JobType = models.JobType(jobType)
		j.Status = models.JobStatus(status)

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
		return models.ErrInvalidJobID
	}

	if job.ID <= 0 {
		return models.ErrInvalidJobID
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
		return models.ErrTransactionFailed
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	job.UpdatedAt = time.Now().UTC()

	skillsJSON, err := json.Marshal(job.RequiredSkills)
	if err != nil {
		return models.WrapError(models.ErrFailedToUpdateJob, err)
	}

	query := `
		UPDATE jobs SET
			title = ?, description = ?, location = ?, job_type = ?,
			source_url = ?, required_skills = ?,
			application_url = ?, company_id = ?,
			status = ?, match_score = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		job.Title,
		job.Description,
		job.Location,
		int(job.JobType),
		job.SourceURL,
		skillsJSON,
		job.ApplicationURL,
		company.ID,
		int(job.Status),
		job.MatchScore,
		job.Notes,
		job.UpdatedAt,
		job.ID,
	)
	if err != nil {
		return models.WrapError(models.ErrFailedToUpdateJob, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.ErrFailedToUpdateJob
	}

	if rowsAffected == 0 {
		return models.ErrJobNotFound
	}

	if err = tx.Commit(); err != nil {
		return models.ErrTransactionFailed
	}

	tx = nil

	job.Company = *company

	return nil
}

// UpdateMatchScore updates only the match score for a job
func (r *SQLiteJobRepository) UpdateMatchScore(ctx context.Context, jobID int, matchScore *int) error {
	if jobID <= 0 {
		return models.ErrInvalidJobID
	}

	query := `UPDATE jobs SET match_score = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, matchScore, time.Now().UTC(), jobID)
	if err != nil {
		return models.WrapError(models.ErrFailedToUpdateJob, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.ErrFailedToUpdateJob
	}

	if rowsAffected == 0 {
		return models.ErrJobNotFound
	}

	return nil
}

// Delete removes a job from the database
func (r *SQLiteJobRepository) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return models.ErrInvalidJobID
	}

	query := "DELETE FROM jobs WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return models.WrapError(models.ErrFailedToDeleteJob, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.WrapError(models.ErrFailedToDeleteJob, err)
	}

	if rowsAffected == 0 {
		return models.ErrJobNotFound
	}

	return nil
}

// UpdateStatus updates the status of a job
func (r *SQLiteJobRepository) UpdateStatus(ctx context.Context, id int, status models.JobStatus) error {
	if id <= 0 {
		return models.ErrInvalidJobID
	}

	if status < models.INTERESTED || status > models.NOT_INTERESTED {
		return models.ErrInvalidJobStatus
	}

	query := "UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?"

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, query, int(status), now, id)
	if err != nil {
		return models.WrapError(models.ErrFailedToUpdateJob, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.WrapError(models.ErrFailedToUpdateJob, err)
	}

	if rowsAffected == 0 {
		return models.ErrJobNotFound
	}

	return nil
}

// GetCount returns the total count of jobs matching the given filter
func (r *SQLiteJobRepository) GetCount(ctx context.Context, filter models.JobFilter) (int, error) {
	query := `
		SELECT COUNT(*)
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

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, models.WrapError(models.ErrFailedToGetJobStats, err)
	}

	return count, nil
}

// GetStats returns aggregate statistics about jobs in the database.
// It returns the total number of jobs, jobs with status APPLIED, and jobs with high match scores (>=70).
func (r *SQLiteJobRepository) GetStats(ctx context.Context) (*models.JobStats, error) {
	query := `
        SELECT
            COUNT(*) AS total_jobs,
            COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS applied,
            COALESCE(SUM(CASE WHEN match_score >= 70 THEN 1 ELSE 0 END), 0) AS high_match
        FROM jobs
    `
	rows, err := r.db.QueryContext(ctx, query, int(models.APPLIED))
	if err != nil {
		return nil, models.WrapError(models.ErrFailedToGetJobStats, err)
	}
	defer rows.Close()

	var stats models.JobStats
	if rows.Next() {
		if err := rows.Scan(&stats.TotalJobs, &stats.TotalApplied, &stats.HighMatch); err != nil {
			return nil, models.WrapError(models.ErrFailedToGetJobStats, err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, models.WrapError(models.ErrFailedToGetJobStats, err)
	}
	return &stats, nil
}
