package job

import (
	"context"
	"testing"
	"time"

	"github.com/benidevo/ascentio/internal/config"
	"github.com/benidevo/ascentio/internal/job/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	// Disable logs for tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// MockJobRepository mocks the JobRepository interface
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(ctx context.Context, job *models.Job) (*models.Job, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetBySourceURL(ctx context.Context, sourceURL string) (*models.Job, error) {
	args := m.Called(ctx, sourceURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetOrCreate(ctx context.Context, job *models.Job) (*models.Job, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetByID(ctx context.Context, id int) (*models.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetAll(ctx context.Context, filter models.JobFilter) ([]*models.Job, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockJobRepository) Update(ctx context.Context, job *models.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockJobRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateStatus(ctx context.Context, id int, status models.JobStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockJobRepository) GetStats(ctx context.Context) (*models.JobStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobStats), args.Error(1)
}

func setupTestConfig() *config.Settings {
	return &config.Settings{
		IsTest:   true,
		LogLevel: "disabled",
	}
}

func createTestCompany() models.Company {
	return models.Company{
		ID:        1,
		Name:      "Test Company",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestJob(id int, title string, company models.Company) *models.Job {
	return &models.Job{
		ID:          id,
		Title:       title,
		Description: "Test description",
		Company:     company,
		Status:      models.INTERESTED,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestJobService(t *testing.T) {
	ctx := context.Background()
	company := createTestCompany()
	job := createTestJob(1, "Software Engineer", company)
	cfg := setupTestConfig()

	t.Run("should create job successfully", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetOrCreate", ctx, mock.AnythingOfType("*models.Job")).Return(job, nil)

		service := NewJobService(mockRepo, cfg)
		createdJob, err := service.CreateJob(ctx, job.Title, job.Description, company.Name)

		require.NoError(t, err)
		assert.Equal(t, job.ID, createdJob.ID)
		assert.Equal(t, job.Title, createdJob.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return job when valid ID", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, 1).Return(job, nil)

		service := NewJobService(mockRepo, cfg)
		foundJob, err := service.GetJob(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, job.ID, foundJob.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when invalid ID", func(t *testing.T) {
		service := NewJobService(nil, cfg) // No repo calls expected
		_, err := service.GetJob(ctx, 0)
		assert.Equal(t, models.ErrInvalidJobID, err)

		_, err = service.GetJob(ctx, -1)
		assert.Equal(t, models.ErrInvalidJobID, err)
	})

	t.Run("should filter jobs with different criteria", func(t *testing.T) {
		jobs := []*models.Job{
			createTestJob(1, "Software Engineer", company),
			createTestJob(2, "Product Manager", company),
		}

		t.Run("should filter by search term", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			searchFilter := models.JobFilter{
				Search: "software",
				Limit:  10,
			}
			mockRepo.On("GetAll", ctx, searchFilter).Return(jobs[:1], nil)

			service := NewJobService(mockRepo, cfg)
			result, err := service.GetJobs(ctx, searchFilter)

			require.NoError(t, err)
			assert.Len(t, result, 1)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by status", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			status := models.APPLIED
			statusFilter := models.JobFilter{
				Status: &status,
			}
			mockRepo.On("GetAll", ctx, statusFilter).Return(jobs, nil)

			service := NewJobService(mockRepo, cfg)
			result, err := service.GetJobs(ctx, statusFilter)

			require.NoError(t, err)
			assert.Len(t, result, 2)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by company ID", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			companyID := 1
			companyFilter := models.JobFilter{
				CompanyID: &companyID,
			}
			mockRepo.On("GetAll", ctx, companyFilter).Return(jobs, nil)

			service := NewJobService(mockRepo, cfg)
			result, err := service.GetJobs(ctx, companyFilter)

			require.NoError(t, err)
			assert.Len(t, result, 2)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by job type", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			jobType := models.FULL_TIME
			typeFilter := models.JobFilter{
				JobType: &jobType,
			}
			mockRepo.On("GetAll", ctx, typeFilter).Return(jobs, nil)

			service := NewJobService(mockRepo, cfg)
			result, err := service.GetJobs(ctx, typeFilter)

			require.NoError(t, err)
			assert.Len(t, result, 2)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should apply complex filter with multiple criteria", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			status := models.APPLIED
			jobType := models.FULL_TIME
			complexFilter := models.JobFilter{
				Search:  "engineer",
				Status:  &status,
				JobType: &jobType,
				Limit:   5,
				Offset:  10,
			}
			mockRepo.On("GetAll", ctx, complexFilter).Return(jobs[:1], nil)

			service := NewJobService(mockRepo, cfg)
			result, err := service.GetJobs(ctx, complexFilter)

			require.NoError(t, err)
			assert.Len(t, result, 1)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("should validate URLs for XSS prevention", func(t *testing.T) {
		mockJobRepo := new(MockJobRepository)
		cfg := &config.Settings{}
		service := NewJobService(mockJobRepo, cfg)

		testCases := []struct {
			name    string
			url     string
			wantErr bool
		}{
			{"empty URL is valid", "", false},
			{"valid http URL", "http://example.com/job", false},
			{"valid https URL", "https://example.com/job", false},
			{"javascript URL is blocked", "javascript:alert('XSS')", true},
			{"data URL is blocked", "data:text/html,<script>alert('XSS')</script>", true},
			{"file URL is blocked", "file:///etc/passwd", true},
			{"ftp URL is blocked", "ftp://example.com/file", true},
			{"invalid URL format", "not a url", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := service.ValidateURL(tc.url)
				if tc.wantErr {
					assert.Error(t, err)
					assert.Equal(t, models.ErrInvalidURLFormat, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("should update job successfully", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, job.ID).Return(job, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Job")).Return(nil)

		service := NewJobService(mockRepo, cfg)
		err := service.UpdateJob(ctx, job)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should validate job before updating", func(t *testing.T) {
		service := NewJobService(nil, cfg) // No repo calls expected

		// Test nil job
		err := service.UpdateJob(ctx, nil)
		assert.Equal(t, models.ErrInvalidJobID, err)

		// Test job with invalid ID
		invalidJob := createTestJob(0, "Invalid Job", company)
		err = service.UpdateJob(ctx, invalidJob)
		assert.Equal(t, models.ErrInvalidJobID, err)
	})

	t.Run("should delete job successfully", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, 1).Return(job, nil)
		mockRepo.On("Delete", ctx, 1).Return(nil)

		service := NewJobService(mockRepo, cfg)
		err := service.DeleteJob(ctx, 1)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when trying to delete with invalid ID", func(t *testing.T) {
		service := NewJobService(nil, cfg) // No repo calls expected
		err := service.DeleteJob(ctx, 0)
		assert.Equal(t, models.ErrInvalidJobID, err)
	})

	t.Run("should update job status successfully", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, 1).Return(job, nil).Once()
		updatedJob := createTestJob(1, "Software Engineer", company)
		updatedJob.Status = models.APPLIED
		mockRepo.On("GetByID", ctx, 1).Return(job, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Job")).Return(nil)

		service := NewJobService(mockRepo, cfg)
		err := service.UpdateJobStatus(ctx, 1, models.APPLIED)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should validate inputs when updating status", func(t *testing.T) {
		service := NewJobService(nil, cfg) // No repo calls expected

		// Test invalid ID
		err := service.UpdateJobStatus(ctx, 0, models.APPLIED)
		assert.Equal(t, models.ErrInvalidJobID, err)

		// Test invalid status
		invalidStatus := models.JobStatus(999)
		err = service.UpdateJobStatus(ctx, 1, invalidStatus)
		assert.Equal(t, models.ErrInvalidJobStatus, err)
	})
}
