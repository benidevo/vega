package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/ascentio/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupJobRepositoryTest(t *testing.T) (*SQLiteJobRepository, sqlmock.Sqlmock, *MinimalMockCompanyRepository) {
	db, mock := setupMockDB(t)
	mockCompanyRepo := NewMinimalMockCompanyRepository()
	repo := NewSQLiteJobRepository(db, mockCompanyRepo)
	return repo, mock, mockCompanyRepo
}

// MinimalMockCompanyRepository is a simplified mock for testing job repository
type MinimalMockCompanyRepository struct {
	companies map[string]*models.Company
	nextID    int
}

func NewMinimalMockCompanyRepository() *MinimalMockCompanyRepository {
	return &MinimalMockCompanyRepository{
		companies: make(map[string]*models.Company),
		nextID:    1,
	}
}

func (r *MinimalMockCompanyRepository) GetOrCreate(ctx context.Context, name string) (*models.Company, error) {
	if name == "" {
		return nil, models.ErrCompanyNameRequired
	}

	if company, ok := r.companies[name]; ok {
		return company, nil
	}

	now := time.Now()
	company := &models.Company{
		ID:        r.nextID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.nextID++
	r.companies[name] = company
	return company, nil
}

func (r *MinimalMockCompanyRepository) GetByID(ctx context.Context, id int) (*models.Company, error) {
	for _, company := range r.companies {
		if company.ID == id {
			return company, nil
		}
	}
	return nil, models.ErrCompanyNotFound
}

func (r *MinimalMockCompanyRepository) GetByName(ctx context.Context, name string) (*models.Company, error) {
	if company, ok := r.companies[name]; ok {
		return company, nil
	}
	return nil, models.ErrCompanyNotFound
}

func (r *MinimalMockCompanyRepository) GetAll(ctx context.Context) ([]*models.Company, error) {
	companies := make([]*models.Company, 0, len(r.companies))
	for _, company := range r.companies {
		companies = append(companies, company)
	}
	return companies, nil
}

func (r *MinimalMockCompanyRepository) Delete(ctx context.Context, id int) error {
	for name, company := range r.companies {
		if company.ID == id {
			delete(r.companies, name)
			return nil
		}
	}
	return models.ErrCompanyNotFound
}

func (r *MinimalMockCompanyRepository) Update(ctx context.Context, company *models.Company) error {
	if company == nil || company.ID == 0 {
		return errors.New("invalid company")
	}

	for name, c := range r.companies {
		if c.ID == company.ID {
			delete(r.companies, name)
			company.UpdatedAt = time.Now()
			r.companies[company.Name] = company
			return nil
		}
	}
	return models.ErrCompanyNotFound
}

func TestSQLiteJobRepository_Create(t *testing.T) {
	tests := []struct {
		name         string
		job          *models.Job
		setupMock    func(sqlmock.Sqlmock, *models.Job)
		setupCompany func(*MinimalMockCompanyRepository)
		wantErr      error
		validateJob  func(*testing.T, *models.Job)
	}{
		{
			name: "successful creation",
			job: &models.Job{
				Title:          "Software Engineer",
				Description:    "Build awesome software",
				Location:       "Remote",
				JobType:        models.FULL_TIME,
				RequiredSkills: []string{"Go", "SQL"},
				Company:        models.Company{Name: "Acme Corp"},
				Status:         models.INTERESTED,
			},
			setupMock: func(mock sqlmock.Sqlmock, j *models.Job) {
				skillsJSON, _ := json.Marshal(j.RequiredSkills)
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO jobs").
					WithArgs(
						j.Title, j.Description, j.Location, int(j.JobType),
						j.SourceURL, skillsJSON, j.ApplicationURL, 1,
						int(j.Status), j.Notes,
						sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			validateJob: func(t *testing.T, j *models.Job) {
				assert.Equal(t, 1, j.ID)
				assert.Equal(t, "Acme Corp", j.Company.Name)
				assert.NotZero(t, j.CreatedAt)
				assert.NotZero(t, j.UpdatedAt)
			},
		},
		{
			name: "validation error - missing title",
			job: &models.Job{
				Description: "Build awesome software",
				Company:     models.Company{Name: "Acme Corp"},
			},
			wantErr: models.ErrJobTitleRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock, mockCompanyRepo := setupJobRepositoryTest(t)
			defer mock.ExpectClose()

			if tt.setupCompany != nil {
				tt.setupCompany(mockCompanyRepo)
			}

			if tt.setupMock != nil {
				tt.setupMock(mock, tt.job)
			}

			createdJob, err := repo.Create(context.Background(), tt.job)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, createdJob)
			} else {
				require.NoError(t, err)
				require.NotNil(t, createdJob)
				if tt.validateJob != nil {
					tt.validateJob(t, createdJob)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQLiteJobRepository_GetByID(t *testing.T) {
	repo, mock, _ := setupJobRepositoryTest(t)
	defer mock.ExpectClose()

	t.Run("existing job", func(t *testing.T) {
		jobID := 1
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"j.id", "j.title", "j.description", "j.location", "j.job_type",
			"j.source_url", "j.required_skills",
			"j.application_url", "j.company_id", "j.status",
			"j.notes", "j.created_at", "j.updated_at",
			"c.id", "c.name", "c.created_at", "c.updated_at",
		}).AddRow(
			jobID, "Software Engineer", "Build awesome software", "Remote", int(models.FULL_TIME),
			"https://example.com", `["Go","SQL"]`,
			"https://apply.example.com", 2, int(models.INTERESTED),
			"Great company", now.Add(-24*time.Hour), now,
			2, "Acme Corp", now, now,
		)

		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE j.id = ?").
			WithArgs(jobID).
			WillReturnRows(rows)

		job, err := repo.GetByID(context.Background(), jobID)

		require.NoError(t, err)
		require.NotNil(t, job)
		assert.Equal(t, jobID, job.ID)
		assert.Equal(t, "Software Engineer", job.Title)
		assert.Equal(t, "Acme Corp", job.Company.Name)
	})

	t.Run("non-existent job", func(t *testing.T) {
		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE j.id = ?").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		job, err := repo.GetByID(context.Background(), 999)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, models.ErrJobNotFound))
		assert.Nil(t, job)
	})
}

func TestSQLiteJobRepository_UpdateStatus(t *testing.T) {
	repo, mock, _ := setupJobRepositoryTest(t)
	defer mock.ExpectClose()

	tests := []struct {
		name         string
		jobID        int
		status       models.JobStatus
		rowsAffected int64
		wantErr      error
	}{
		{
			name:         "successful update",
			jobID:        1,
			status:       models.APPLIED,
			rowsAffected: 1,
		},
		{
			name:         "job not found",
			jobID:        999,
			status:       models.APPLIED,
			rowsAffected: 0,
			wantErr:      models.ErrJobNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectExec("UPDATE jobs SET status = \\?, updated_at = \\? WHERE id = \\?").
				WithArgs(int(tt.status), sqlmock.AnyArg(), tt.jobID).
				WillReturnResult(sqlmock.NewResult(0, tt.rowsAffected))

			err := repo.UpdateStatus(context.Background(), tt.jobID, tt.status)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSQLiteJobRepository_Delete(t *testing.T) {
	repo, mock, _ := setupJobRepositoryTest(t)
	defer mock.ExpectClose()

	mock.ExpectExec("DELETE FROM jobs WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLiteJobRepository_GetAll(t *testing.T) {
	repo, mock, _ := setupJobRepositoryTest(t)
	defer mock.ExpectClose()

	t.Run("filter by company", func(t *testing.T) {
		companyID := 1
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"j.id", "j.title", "j.description", "j.location", "j.job_type",
			"j.source_url", "j.required_skills",
			"j.application_url", "j.company_id", "j.status",
			"j.notes", "j.created_at", "j.updated_at",
			"c.id", "c.name", "c.created_at", "c.updated_at",
		}).AddRow(
			1, "Software Engineer", "Build awesome software", "Remote", int(models.FULL_TIME),
			"https://example.com", `["Go","SQL"]`,
			"https://apply.example.com", companyID, int(models.INTERESTED),
			"Great company", now.Add(-24*time.Hour), now,
			companyID, "Acme Corp", now, now,
		)

		mock.ExpectQuery("SELECT.*FROM jobs.*WHERE.*company_id.*ORDER BY").
			WithArgs(companyID).
			WillReturnRows(rows)

		filter := models.JobFilter{CompanyID: &companyID}
		jobs, err := repo.GetAll(context.Background(), filter)

		require.NoError(t, err)
		require.Len(t, jobs, 1)
		assert.Equal(t, companyID, jobs[0].Company.ID)
	})
}
