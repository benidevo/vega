package job

import (
	"context"
	"testing"
	"time"

	"github.com/benidevo/vega/internal/config"
	"github.com/benidevo/vega/internal/job/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	// Disable logs for tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

const testUserID = 1

// MockJobRepository mocks the JobRepository interface
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(ctx context.Context, userID int, job *models.Job) (*models.Job, error) {
	args := m.Called(ctx, userID, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) UpdateMatchScore(ctx context.Context, userID int, jobID int, matchScore *int) error {
	args := m.Called(ctx, userID, jobID, matchScore)
	return args.Error(0)
}

func (m *MockJobRepository) GetBySourceURL(ctx context.Context, userID int, sourceURL string) (*models.Job, error) {
	args := m.Called(ctx, userID, sourceURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetOrCreate(ctx context.Context, userID int, job *models.Job) (*models.Job, bool, error) {
	args := m.Called(ctx, userID, job)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*models.Job), args.Bool(1), args.Error(2)
}

func (m *MockJobRepository) GetByID(ctx context.Context, userID int, id int) (*models.Job, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetAll(ctx context.Context, userID int, filter models.JobFilter) ([]*models.Job, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetCount(ctx context.Context, userID int, filter models.JobFilter) (int, error) {
	args := m.Called(ctx, userID, filter)
	return args.Int(0), args.Error(1)
}

func (m *MockJobRepository) Update(ctx context.Context, userID int, job *models.Job) error {
	args := m.Called(ctx, userID, job)
	return args.Error(0)
}

func (m *MockJobRepository) Delete(ctx context.Context, userID int, id int) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateStatus(ctx context.Context, userID int, id int, status models.JobStatus) error {
	args := m.Called(ctx, userID, id, status)
	return args.Error(0)
}

func (m *MockJobRepository) GetStats(ctx context.Context, userID int) (*models.JobStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobStats), args.Error(1)
}

func (m *MockJobRepository) CreateMatchResult(ctx context.Context, userID int, matchResult *models.MatchResult) error {
	args := m.Called(ctx, userID, matchResult)
	return args.Error(0)
}

func (m *MockJobRepository) GetJobMatchHistory(ctx context.Context, userID int, jobID int) ([]*models.MatchResult, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchResult), args.Error(1)
}

func (m *MockJobRepository) GetRecentMatchResults(ctx context.Context, userID int, limit int) ([]*models.MatchResult, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchResult), args.Error(1)
}

func (m *MockJobRepository) GetRecentMatchResultsWithDetails(ctx context.Context, userID int, limit int, currentJobID int) ([]*models.MatchSummary, error) {
	args := m.Called(ctx, userID, limit, currentJobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchSummary), args.Error(1)
}

func (m *MockJobRepository) DeleteMatchResult(ctx context.Context, userID int, matchID int) error {
	args := m.Called(ctx, userID, matchID)
	return args.Error(0)
}

func (m *MockJobRepository) MatchResultBelongsToJob(ctx context.Context, userID int, matchID, jobID int) (bool, error) {
	args := m.Called(ctx, userID, matchID, jobID)
	return args.Bool(0), args.Error(1)
}

func (m *MockJobRepository) GetStatsByUserID(ctx context.Context, userID int) (*models.JobStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.JobStats), args.Error(1)
}

func (m *MockJobRepository) GetRecentJobsByUserID(ctx context.Context, userID int, limit int) ([]*models.Job, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Job), args.Error(1)
}

func (m *MockJobRepository) GetJobStatsByStatus(ctx context.Context, userID int) (map[models.JobStatus]int, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[models.JobStatus]int), args.Error(1)
}

func (m *MockJobRepository) SetFirstAnalyzedAt(ctx context.Context, jobID int) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func (m *MockJobRepository) GetMonthlyAnalysisCount(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
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
		UserID:      testUserID,
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
		mockRepo.On("GetOrCreate", ctx, testUserID, mock.AnythingOfType("*models.Job")).Return(job, true, nil)

		service := NewJobService(mockRepo, nil, nil, nil, cfg)
		createdJob, isNew, err := service.CreateJob(ctx, testUserID, job.Title, job.Description, company.Name)

		require.NoError(t, err)
		assert.True(t, isNew)
		assert.Equal(t, job.ID, createdJob.ID)
		assert.Equal(t, job.Title, createdJob.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return job when valid ID", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, testUserID, 1).Return(job, nil)

		service := NewJobService(mockRepo, nil, nil, nil, cfg)
		foundJob, err := service.GetJob(ctx, testUserID, 1)

		require.NoError(t, err)
		assert.Equal(t, job.ID, foundJob.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when invalid ID", func(t *testing.T) {
		service := NewJobService(nil, nil, nil, nil, cfg) // No repo calls expected
		_, err := service.GetJob(ctx, testUserID, 0)
		assert.Equal(t, models.ErrInvalidJobID, err)

		_, err = service.GetJob(ctx, testUserID, -1)
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
			mockRepo.On("GetAll", ctx, testUserID, searchFilter).Return(jobs[:1], nil)
			mockRepo.On("GetCount", ctx, testUserID, searchFilter).Return(1, nil)

			service := NewJobService(mockRepo, nil, nil, nil, cfg)
			result, err := service.GetJobsWithPagination(ctx, testUserID, searchFilter)

			require.NoError(t, err)
			assert.Len(t, result.Jobs, 1)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by status", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			status := models.APPLIED
			statusFilter := models.JobFilter{
				Status: &status,
				Limit:  12,
			}
			mockRepo.On("GetAll", ctx, testUserID, statusFilter).Return(jobs, nil)
			mockRepo.On("GetCount", ctx, testUserID, statusFilter).Return(2, nil)

			service := NewJobService(mockRepo, nil, nil, nil, cfg)
			result, err := service.GetJobsWithPagination(ctx, testUserID, statusFilter)

			require.NoError(t, err)
			assert.Len(t, result.Jobs, 2)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by company ID", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			companyID := 1
			companyFilter := models.JobFilter{
				CompanyID: &companyID,
				Limit:     12,
			}
			mockRepo.On("GetAll", ctx, testUserID, companyFilter).Return(jobs, nil)
			mockRepo.On("GetCount", ctx, testUserID, companyFilter).Return(2, nil)

			service := NewJobService(mockRepo, nil, nil, nil, cfg)
			result, err := service.GetJobsWithPagination(ctx, testUserID, companyFilter)

			require.NoError(t, err)
			assert.Len(t, result.Jobs, 2)
			mockRepo.AssertExpectations(t)
		})

		t.Run("should filter by job type", func(t *testing.T) {
			mockRepo := new(MockJobRepository)
			jobType := models.FULL_TIME
			typeFilter := models.JobFilter{
				JobType: &jobType,
				Limit:   12,
			}
			mockRepo.On("GetAll", ctx, testUserID, typeFilter).Return(jobs, nil)
			mockRepo.On("GetCount", ctx, testUserID, typeFilter).Return(2, nil)

			service := NewJobService(mockRepo, nil, nil, nil, cfg)
			result, err := service.GetJobsWithPagination(ctx, testUserID, typeFilter)

			require.NoError(t, err)
			assert.Len(t, result.Jobs, 2)
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
			mockRepo.On("GetAll", ctx, testUserID, complexFilter).Return(jobs[:1], nil)
			mockRepo.On("GetCount", ctx, testUserID, complexFilter).Return(1, nil)

			service := NewJobService(mockRepo, nil, nil, nil, cfg)
			result, err := service.GetJobsWithPagination(ctx, testUserID, complexFilter)

			require.NoError(t, err)
			assert.Len(t, result.Jobs, 1)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("should validate URLs for XSS prevention", func(t *testing.T) {
		mockJobRepo := new(MockJobRepository)
		cfg := &config.Settings{}
		service := NewJobService(mockJobRepo, nil, nil, nil, cfg)

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
		mockRepo.On("Update", ctx, testUserID, mock.AnythingOfType("*models.Job")).Return(nil)

		service := NewJobService(mockRepo, nil, nil, nil, cfg)
		err := service.UpdateJob(ctx, testUserID, job)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should validate job before updating", func(t *testing.T) {
		service := NewJobService(nil, nil, nil, nil, cfg) // No repo calls expected

		// Test nil job
		err := service.UpdateJob(ctx, testUserID, nil)
		assert.Equal(t, models.ErrInvalidJobID, err)

		// Test job with invalid ID
		invalidJob := createTestJob(0, "Invalid Job", company)
		err = service.UpdateJob(ctx, testUserID, invalidJob)
		assert.Equal(t, models.ErrInvalidJobID, err)
	})

	t.Run("should delete job successfully", func(t *testing.T) {
		mockRepo := new(MockJobRepository)
		mockRepo.On("GetByID", ctx, testUserID, 1).Return(job, nil)
		mockRepo.On("Delete", ctx, testUserID, 1).Return(nil)

		service := NewJobService(mockRepo, nil, nil, nil, cfg)
		err := service.DeleteJob(ctx, testUserID, 1)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when trying to delete with invalid ID", func(t *testing.T) {
		service := NewJobService(nil, nil, nil, nil, cfg) // No repo calls expected
		err := service.DeleteJob(ctx, testUserID, 0)
		assert.Equal(t, models.ErrInvalidJobID, err)
	})
}
