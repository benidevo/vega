package quota

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	ctxutil "github.com/benidevo/vega/internal/common/context"
	timeutil "github.com/benidevo/vega/internal/common/time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testMonthlyQuotaLimit = 5
)

// setupMockDB creates a new mock database for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return db, mock
}

// mockCloudModeQueries adds common mock expectations for cloud mode queries
func mockCloudModeQueries(mock sqlmock.Sqlmock, userID int, isAdmin bool) {
	// Only add these if not admin (admin skips these checks)
	if !isAdmin {
		// Mock quota config query
		rows := sqlmock.NewRows([]string{"quota_type", "free_limit", "description", "created_at", "updated_at"}).
			AddRow("ai_analysis_monthly", testMonthlyQuotaLimit, "AI job analysis per month", time.Now(), time.Now())
		mock.ExpectQuery("SELECT quota_type, free_limit, description, created_at, updated_at FROM quota_configs").
			WithArgs("ai_analysis_monthly").
			WillReturnRows(rows)
	}
}

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) GetByID(ctx context.Context, userID, jobID int) (*Job, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Job), args.Error(1)
}

func (m *MockJobRepository) SetFirstAnalyzedAt(ctx context.Context, jobID int) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func TestService_NonCloudMode(t *testing.T) {
	ctx := context.Background()
	userID := 1
	jobID := 100

	t.Run("CanAnalyzeJob returns unlimited access", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockJobRepo := new(MockJobRepository)
		service := NewService(db, mockJobRepo, false) // Non-cloud mode

		// Mock job that hasn't been analyzed
		job := &Job{ID: jobID, FirstAnalyzedAt: nil}
		mockJobRepo.On("GetByID", ctx, userID, jobID).Return(job, nil).Once()

		// Mock GetMonthlyUsage query - no existing usage
		monthYear := timeutil.GetCurrentMonthYear()
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnError(sql.ErrNoRows)

		result, err := service.CanAnalyzeJob(ctx, userID, jobID)

		assert.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, QuotaReasonOK, result.Reason)
		assert.Equal(t, -1, result.Status.Limit) // Unlimited
		assert.Equal(t, 0, result.Status.Used)
		assert.Equal(t, time.Time{}, result.Status.ResetDate)

		mockJobRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetQuotaStatus returns unlimited quota", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockJobRepo := new(MockJobRepository)
		service := NewService(db, mockJobRepo, false) // Non-cloud mode

		// Mock GetMonthlyUsage query - no existing usage
		monthYear := timeutil.GetCurrentMonthYear()
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnError(sql.ErrNoRows)

		status, err := service.GetQuotaStatus(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, -1, status.Limit) // Unlimited
		assert.Equal(t, 0, status.Used)
		assert.Equal(t, time.Time{}, status.ResetDate)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RecordAnalysis records usage but doesn't enforce", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockJobRepo := new(MockJobRepository)
		service := NewService(db, mockJobRepo, false) // Non-cloud mode

		// Mock SetFirstAnalyzedAt
		mockJobRepo.On("SetFirstAnalyzedAt", ctx, jobID).Return(nil).Once()

		monthYear := timeutil.GetCurrentMonthYear()

		// Begin transaction
		mock.ExpectBegin()

		// Expect UPSERT query - still records usage in non-cloud mode
		mock.ExpectExec("INSERT INTO user_quota_usage \\(user_id, month_year, jobs_analyzed, updated_at\\) VALUES \\(\\?, \\?, 1, CURRENT_TIMESTAMP\\) ON CONFLICT\\(user_id, month_year\\) DO UPDATE SET jobs_analyzed = jobs_analyzed \\+ 1, updated_at = CURRENT_TIMESTAMP").
			WithArgs(userID, monthYear).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Commit transaction
		mock.ExpectCommit()

		err := service.RecordAnalysis(ctx, userID, jobID)

		assert.NoError(t, err)
		mockJobRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetMonthlyUsage returns actual usage", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockJobRepo := new(MockJobRepository)
		service := NewService(db, mockJobRepo, false) // Non-cloud mode

		// Mock GetMonthlyUsage query - with some existing usage
		monthYear := timeutil.GetCurrentMonthYear()
		usedCount := 3
		rows := sqlmock.NewRows([]string{"user_id", "month_year", "jobs_analyzed", "updated_at"}).
			AddRow(userID, monthYear, usedCount, time.Now())
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnRows(rows)

		usage, err := service.GetMonthlyUsage(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, userID, usage.UserID)
		assert.Equal(t, usedCount, usage.JobsAnalyzed)
		assert.Equal(t, timeutil.GetCurrentMonthYear(), usage.MonthYear)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestService_CloudMode(t *testing.T) {
	ctx := context.Background()
	userID := 1
	jobID := 100

	t.Run("CanAnalyzeJob enforces quota for new analysis", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock job that hasn't been analyzed
		job := &Job{ID: jobID, FirstAnalyzedAt: nil}
		mockRepo.On("GetByID", ctx, userID, jobID).Return(job, nil).Once()

		// Mock GetMonthlyUsage query - no existing usage
		monthYear := timeutil.GetCurrentMonthYear()
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnError(sql.ErrNoRows)

		// Mock cloud mode queries (not admin)
		mockCloudModeQueries(mock, userID, false)

		result, err := service.CanAnalyzeJob(ctx, userID, jobID)

		assert.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, QuotaReasonOK, result.Reason)
		assert.Equal(t, testMonthlyQuotaLimit, result.Status.Limit)
		assert.Equal(t, 0, result.Status.Used)

		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CanAnalyzeJob blocks when limit reached", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock job that hasn't been analyzed
		job := &Job{ID: jobID, FirstAnalyzedAt: nil}
		mockRepo.On("GetByID", ctx, userID, jobID).Return(job, nil).Once()

		// Mock GetMonthlyUsage query - at limit
		monthYear := timeutil.GetCurrentMonthYear()
		rows := sqlmock.NewRows([]string{"user_id", "month_year", "jobs_analyzed", "updated_at"}).
			AddRow(userID, monthYear, testMonthlyQuotaLimit, time.Now())
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnRows(rows)

		// Mock cloud mode queries (not admin)
		mockCloudModeQueries(mock, userID, false)

		result, err := service.CanAnalyzeJob(ctx, userID, jobID)

		assert.NoError(t, err)
		assert.False(t, result.Allowed)
		assert.Equal(t, QuotaReasonLimitReached, result.Reason)
		assert.Equal(t, testMonthlyQuotaLimit, result.Status.Limit)
		assert.Equal(t, testMonthlyQuotaLimit, result.Status.Used)

		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CanAnalyzeJob allows reanalysis", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock job that has been analyzed before
		analyzedAt := time.Now()
		job := &Job{ID: jobID, FirstAnalyzedAt: &analyzedAt}
		mockRepo.On("GetByID", ctx, userID, jobID).Return(job, nil).Once()

		// Mock GetMonthlyUsage query - even if at limit, reanalysis is allowed
		monthYear := timeutil.GetCurrentMonthYear()
		rows := sqlmock.NewRows([]string{"user_id", "month_year", "jobs_analyzed", "updated_at"}).
			AddRow(userID, monthYear, testMonthlyQuotaLimit, time.Now())
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnRows(rows)

		// Mock cloud mode queries (not admin)
		mockCloudModeQueries(mock, userID, false)

		result, err := service.CanAnalyzeJob(ctx, userID, jobID)

		assert.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, QuotaReasonReanalysis, result.Reason)

		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RecordAnalysis creates new usage record", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock SetFirstAnalyzedAt
		mockRepo.On("SetFirstAnalyzedAt", ctx, jobID).Return(nil).Once()

		monthYear := timeutil.GetCurrentMonthYear()

		// Begin transaction
		mock.ExpectBegin()

		// Expect UPSERT query
		mock.ExpectExec("INSERT INTO user_quota_usage \\(user_id, month_year, jobs_analyzed, updated_at\\) VALUES \\(\\?, \\?, 1, CURRENT_TIMESTAMP\\) ON CONFLICT\\(user_id, month_year\\) DO UPDATE SET jobs_analyzed = jobs_analyzed \\+ 1, updated_at = CURRENT_TIMESTAMP").
			WithArgs(userID, monthYear).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Commit transaction
		mock.ExpectCommit()

		err := service.RecordAnalysis(ctx, userID, jobID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RecordAnalysis increments existing usage", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock SetFirstAnalyzedAt
		mockRepo.On("SetFirstAnalyzedAt", ctx, jobID).Return(nil).Once()

		monthYear := timeutil.GetCurrentMonthYear()

		// Begin transaction
		mock.ExpectBegin()

		// Expect UPSERT query (same as new record, but will update existing)
		mock.ExpectExec("INSERT INTO user_quota_usage \\(user_id, month_year, jobs_analyzed, updated_at\\) VALUES \\(\\?, \\?, 1, CURRENT_TIMESTAMP\\) ON CONFLICT\\(user_id, month_year\\) DO UPDATE SET jobs_analyzed = jobs_analyzed \\+ 1, updated_at = CURRENT_TIMESTAMP").
			WithArgs(userID, monthYear).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected (update path)

		// Commit transaction
		mock.ExpectCommit()

		err := service.RecordAnalysis(ctx, userID, jobID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetQuotaStatus returns correct usage", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Mock GetMonthlyUsage query
		monthYear := timeutil.GetCurrentMonthYear()
		usedCount := 5
		rows := sqlmock.NewRows([]string{"user_id", "month_year", "jobs_analyzed", "updated_at"}).
			AddRow(userID, monthYear, usedCount, time.Now())
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnRows(rows)

		// Mock cloud mode queries (not admin)
		mockCloudModeQueries(mock, userID, false)

		status, err := service.GetQuotaStatus(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, testMonthlyQuotaLimit, status.Limit)
		assert.Equal(t, usedCount, status.Used)
		assert.NotEqual(t, time.Time{}, status.ResetDate)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Admin users get unlimited quota in cloud mode", func(t *testing.T) {
		db, mock := setupMockDB(t)
		defer db.Close()

		mockRepo := new(MockJobRepository)
		service := NewService(db, mockRepo, true) // Cloud mode enabled

		// Create context with admin role
		adminCtx := ctxutil.WithRole(ctx, "Admin")

		// Mock job that hasn't been analyzed
		job := &Job{ID: jobID, FirstAnalyzedAt: nil}
		mockRepo.On("GetByID", adminCtx, userID, jobID).Return(job, nil).Once()

		// Mock GetMonthlyUsage query - even if at limit
		monthYear := timeutil.GetCurrentMonthYear()
		rows := sqlmock.NewRows([]string{"user_id", "month_year", "jobs_analyzed", "updated_at"}).
			AddRow(userID, monthYear, testMonthlyQuotaLimit+10, time.Now()) // Over limit
		mock.ExpectQuery("SELECT user_id, month_year, jobs_analyzed, updated_at FROM user_quota_usage").
			WithArgs(userID, monthYear).
			WillReturnRows(rows)

		// Mock cloud mode queries (admin user)
		mockCloudModeQueries(mock, userID, true)

		result, err := service.CanAnalyzeJob(adminCtx, userID, jobID)

		assert.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, QuotaReasonOK, result.Reason)
		assert.Equal(t, -1, result.Status.Limit) // Unlimited for admin
		assert.Equal(t, testMonthlyQuotaLimit+10, result.Status.Used)

		mockRepo.AssertExpectations(t)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
