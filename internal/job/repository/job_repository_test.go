package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/prospector/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupJobRepositoryTest(t *testing.T) (*SQLiteJobRepository, sqlmock.Sqlmock, *MockCompanyRepository) {
	db, mock := setupMockDB(t)
	mockCompanyRepo := NewMockCompanyRepository()
	repo := NewSQLiteJobRepository(db, mockCompanyRepo)
	return repo, mock, mockCompanyRepo
}

// MockCompanyRepository is a mock implementation of CompanyRepository
type MockCompanyRepository struct {
	companies       map[string]*models.Company
	nextID          int
	GetOrCreateFunc func(ctx context.Context, name string) (*models.Company, error)
}

func NewMockCompanyRepository() *MockCompanyRepository {
	repo := &MockCompanyRepository{
		companies: make(map[string]*models.Company),
		nextID:    1,
	}

	repo.GetOrCreateFunc = func(ctx context.Context, name string) (*models.Company, error) {
		if name == "" {
			return nil, ErrCompanyNameRequired
		}

		normalizedName := name
		if company, ok := repo.companies[normalizedName]; ok {
			return company, nil
		}

		now := time.Now()
		company := &models.Company{
			ID:        repo.nextID,
			Name:      normalizedName,
			CreatedAt: now,
			UpdatedAt: now,
		}
		repo.nextID++
		repo.companies[normalizedName] = company

		return company, nil
	}

	return repo
}

func (r *MockCompanyRepository) GetOrCreate(ctx context.Context, name string) (*models.Company, error) {
	return r.GetOrCreateFunc(ctx, name)
}

func (r *MockCompanyRepository) GetByID(ctx context.Context, id int) (*models.Company, error) {
	for _, company := range r.companies {
		if company.ID == id {
			return company, nil
		}
	}
	return nil, ErrCompanyNotFound
}

func (r *MockCompanyRepository) GetByName(ctx context.Context, name string) (*models.Company, error) {
	if name == "" {
		return nil, ErrCompanyNameRequired
	}

	normalizedName := name
	if company, ok := r.companies[normalizedName]; ok {
		return company, nil
	}

	return nil, ErrCompanyNotFound
}

func (r *MockCompanyRepository) GetAll(ctx context.Context) ([]*models.Company, error) {
	companies := make([]*models.Company, 0, len(r.companies))
	for _, company := range r.companies {
		companies = append(companies, company)
	}
	return companies, nil
}

func (r *MockCompanyRepository) Delete(ctx context.Context, id int) error {
	for name, company := range r.companies {
		if company.ID == id {
			delete(r.companies, name)
			return nil
		}
	}
	return ErrCompanyNotFound
}

func (r *MockCompanyRepository) Update(ctx context.Context, company *models.Company) error {
	if company == nil {
		return errors.New("company cannot be nil")
	}

	if company.ID == 0 {
		return errors.New("company ID is required")
	}

	for name, c := range r.companies {
		if c.ID == company.ID {
			delete(r.companies, name)
			company.UpdatedAt = time.Now()
			r.companies[company.Name] = company
			return nil
		}
	}

	return ErrCompanyNotFound
}

func TestSQLiteJobRepository_Create(t *testing.T) {
	t.Run("should create a job successfully", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		j := &models.Job{
			Title:          "Software Engineer",
			Description:    "Build awesome software",
			Location:       "Remote",
			JobType:        models.FULL_TIME,
			SalaryRange:    "$100k-150k",
			RequiredSkills: []string{"Go", "SQL"},
			Company: models.Company{
				Name: "Acme Corp",
			},
			Status:          models.INTERESTED,
			ExperienceLevel: models.SENIOR,
		}

		expectedCompany := &models.Company{
			ID:   1,
			Name: "Acme Corp",
		}

		skillsJSON, err := json.Marshal(j.RequiredSkills)
		require.NoError(t, err)

		mock.ExpectBegin()

		mock.ExpectExec("INSERT INTO jobs").
			WithArgs(
				j.Title,
				j.Description,
				j.Location,
				int(j.JobType),
				j.SourceURL,
				j.SalaryRange,
				skillsJSON,
				nil, // ApplicationDeadline
				j.ApplicationURL,
				expectedCompany.ID,
				int(j.Status),
				int(j.ExperienceLevel),
				j.ContactPerson,
				j.Notes,
				nil,              // PostedAt
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		ctx := context.Background()
		createdJob, err := repo.Create(ctx, j)

		require.NoError(t, err)
		require.NotNil(t, createdJob)
		assert.Equal(t, 1, createdJob.ID)
		assert.Equal(t, j.Title, createdJob.Title)
		assert.Equal(t, expectedCompany.ID, createdJob.Company.ID)
		assert.Equal(t, expectedCompany.Name, createdJob.Company.Name)
		assert.NotZero(t, createdJob.CreatedAt)
		assert.NotZero(t, createdJob.UpdatedAt)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when job validation fails", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		j := &models.Job{
			Description: "Build awesome software",
			Company: models.Company{
				Name: "Acme Corp",
			},
		}

		ctx := context.Background()
		createdJob, err := repo.Create(ctx, j)

		assert.Error(t, err)
		assert.Equal(t, ErrJobTitleRequired, err)
		assert.Nil(t, createdJob)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when company creation fails", func(t *testing.T) {
		// Setup
		repo, mock, mockCompanyRepo := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		// Override the GetOrCreate function to return an error
		mockCompanyRepo.GetOrCreateFunc = func(ctx context.Context, name string) (*models.Company, error) {
			return nil, errors.New("company creation failed")
		}

		j := &models.Job{
			Title:       "Software Engineer",
			Description: "Build awesome software",
			Company: models.Company{
				Name: "Acme Corp",
			},
		}

		ctx := context.Background()
		createdJob, err := repo.Create(ctx, j)

		assert.Error(t, err)
		assert.Equal(t, "company creation failed", err.Error())
		assert.Nil(t, createdJob)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLiteJobRepository_GetByID(t *testing.T) {
	t.Run("should return job when it exists", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		jobID := 1
		now := time.Now()
		companyID := 2

		rows := sqlmock.NewRows([]string{
			"j.id", "j.title", "j.description", "j.location", "j.job_type",
			"j.source_url", "j.salary_range", "j.required_skills", "j.application_deadline",
			"j.application_url", "j.company_id", "j.status", "j.experience_level",
			"j.contact_person", "j.notes", "j.posted_at", "j.created_at", "j.updated_at",
			"c.id", "c.name", "c.created_at", "c.updated_at",
		}).
			AddRow(
				jobID, "Software Engineer", "Build awesome software", "Remote", int(models.FULL_TIME),
				"https://example.com", "$100k-150k", `["Go","SQL"]`, now.Add(7*24*time.Hour),
				"https://apply.example.com", companyID, int(models.INTERESTED), int(models.SENIOR),
				"John Doe", "Great company", now.Add(-24*time.Hour), now, now,
				companyID, "Acme Corp", now, now,
			)

		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE j.id = ?").
			WithArgs(jobID).
			WillReturnRows(rows)

		ctx := context.Background()
		j, err := repo.GetByID(ctx, jobID)

		require.NoError(t, err)
		require.NotNil(t, j)
		assert.Equal(t, jobID, j.ID)
		assert.Equal(t, "Software Engineer", j.Title)
		assert.Equal(t, "Build awesome software", j.Description)
		assert.Equal(t, "Remote", j.Location)
		assert.Equal(t, models.FULL_TIME, j.JobType)
		assert.Equal(t, "https://example.com", j.SourceURL)
		assert.Equal(t, "$100k-150k", j.SalaryRange)
		assert.Equal(t, []string{"Go", "SQL"}, j.RequiredSkills)
		assert.Equal(t, "https://apply.example.com", j.ApplicationURL)
		assert.Equal(t, models.INTERESTED, j.Status)
		assert.Equal(t, models.SENIOR, j.ExperienceLevel)
		assert.Equal(t, "John Doe", j.ContactPerson)
		assert.Equal(t, "Great company", j.Notes)
		assert.NotNil(t, j.ApplicationDeadline)
		assert.NotNil(t, j.PostedAt)
		assert.Equal(t, companyID, j.Company.ID)
		assert.Equal(t, "Acme Corp", j.Company.Name)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return ErrJobNotFound when job does not exist", func(t *testing.T) {
		// Setup
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		jobID := 999 // Non-existent job

		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE j.id = ?").
			WithArgs(jobID).
			WillReturnError(sql.ErrNoRows)

		ctx := context.Background()
		j, err := repo.GetByID(ctx, jobID)

		assert.Error(t, err)
		assert.Equal(t, ErrJobNotFound, err)
		assert.Nil(t, j)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLiteJobRepository_UpdateStatus(t *testing.T) {
	t.Run("should update job status successfully", func(t *testing.T) {
		// Setup
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		jobID := 1
		newStatus := models.APPLIED

		mock.ExpectExec("UPDATE jobs SET status = \\?, updated_at = \\? WHERE id = \\?").
			WithArgs(int(newStatus), sqlmock.AnyArg(), jobID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		ctx := context.Background()
		err := repo.UpdateStatus(ctx, jobID, newStatus)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return ErrJobNotFound when job does not exist", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		jobID := 999 // Non-existent job
		newStatus := models.APPLIED

		// Expect update but no rows affected
		mock.ExpectExec("UPDATE jobs SET status = \\?, updated_at = \\? WHERE id = \\?").
			WithArgs(int(newStatus), sqlmock.AnyArg(), jobID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		ctx := context.Background()
		err := repo.UpdateStatus(ctx, jobID, newStatus)

		assert.Error(t, err)
		assert.Equal(t, ErrJobNotFound, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLiteJobRepository_Delete(t *testing.T) {
	t.Run("should delete job successfully", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		jobID := 1

		mock.ExpectExec("DELETE FROM jobs WHERE id = \\?").
			WithArgs(jobID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Execute
		ctx := context.Background()
		err := repo.Delete(ctx, jobID)

		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLiteJobRepository_GetAll(t *testing.T) {
	t.Run("should return all jobs with no filters", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"j.id", "j.title", "j.description", "j.location", "j.job_type",
			"j.source_url", "j.salary_range", "j.required_skills", "j.application_deadline",
			"j.application_url", "j.company_id", "j.status", "j.experience_level",
			"j.contact_person", "j.notes", "j.posted_at", "j.created_at", "j.updated_at",
			"c.id", "c.name", "c.created_at", "c.updated_at",
		})

		// Add first job
		rows.AddRow(
			1, "Software Engineer", "Build awesome software", "Remote", int(models.FULL_TIME),
			"https://example.com", "$100k-150k", `["Go","SQL"]`, now.Add(7*24*time.Hour),
			"https://apply.example.com", 1, int(models.INTERESTED), int(models.SENIOR),
			"John Doe", "Great company", now.Add(-24*time.Hour), now, now,
			1, "Acme Corp", now, now,
		)

		// Add second job
		rows.AddRow(
			2, "Frontend Developer", "Create beautiful UIs", "San Francisco", int(models.CONTRACT),
			"https://example2.com", "$80k-120k", `["React","CSS"]`, now.Add(14*24*time.Hour),
			"https://apply.example2.com", 2, int(models.APPLIED), int(models.MID_LEVEL),
			"Jane Smith", "Fast growing startup", now.Add(-48*time.Hour), now, now,
			2, "Beta Inc", now, now,
		)

		mock.ExpectQuery("SELECT.*FROM jobs.*ORDER BY").
			WillReturnRows(rows)

		// Execute with empty filter
		ctx := context.Background()
		jobs, err := repo.GetAll(ctx, JobFilter{})

		require.NoError(t, err)
		require.NotNil(t, jobs)
		require.Len(t, jobs, 2)

		// Check first job
		assert.Equal(t, 1, jobs[0].ID)
		assert.Equal(t, "Software Engineer", jobs[0].Title)
		assert.Equal(t, "Acme Corp", jobs[0].Company.Name)
		assert.Equal(t, models.INTERESTED, jobs[0].Status)

		// Check second job
		assert.Equal(t, 2, jobs[1].ID)
		assert.Equal(t, "Frontend Developer", jobs[1].Title)
		assert.Equal(t, "Beta Inc", jobs[1].Company.Name)
		assert.Equal(t, models.APPLIED, jobs[1].Status)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should filter by company ID", func(t *testing.T) {
		repo, mock, _ := setupJobRepositoryTest(t)
		defer mock.ExpectClose()

		now := time.Now()
		companyID := 1

		rows := sqlmock.NewRows([]string{
			"j.id", "j.title", "j.description", "j.location", "j.job_type",
			"j.source_url", "j.salary_range", "j.required_skills", "j.application_deadline",
			"j.application_url", "j.company_id", "j.status", "j.experience_level",
			"j.contact_person", "j.notes", "j.posted_at", "j.created_at", "j.updated_at",
			"c.id", "c.name", "c.created_at", "c.updated_at",
		}).
			AddRow(
				1, "Software Engineer", "Build awesome software", "Remote", int(models.FULL_TIME),
				"https://example.com", "$100k-150k", `["Go","SQL"]`, now.Add(7*24*time.Hour),
				"https://apply.example.com", companyID, int(models.INTERESTED), int(models.SENIOR),
				"John Doe", "Great company", now.Add(-24*time.Hour), now, now,
				companyID, "Acme Corp", now, now,
			)

		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE.*company_id.*ORDER BY").
			WithArgs(companyID).
			WillReturnRows(rows)

		ctx := context.Background()
		filter := JobFilter{
			CompanyID: &companyID,
		}
		jobs, err := repo.GetAll(ctx, filter)

		require.NoError(t, err)
		require.NotNil(t, jobs)
		require.Len(t, jobs, 1)
		assert.Equal(t, companyID, jobs[0].Company.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
