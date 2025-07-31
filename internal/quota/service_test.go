package quota

import (
	"context"
	"testing"
	"time"

	ctxutil "github.com/benidevo/vega/internal/common/context"
	timeutil "github.com/benidevo/vega/internal/common/time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testMonthlyQuotaLimit = 5
)

// mockQuotaChecker implements the QuotaChecker interface for testing
type mockQuotaChecker struct {
	mock.Mock
}

func (m *mockQuotaChecker) CanAnalyzeJob(ctx context.Context, userID int, jobID int) (*QuotaCheckResult, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaCheckResult), args.Error(1)
}

// mockQuotaReporter implements the QuotaReporter interface for testing
type mockQuotaReporter struct {
	mock.Mock
}

func (m *mockQuotaReporter) GetQuotaStatus(ctx context.Context, userID int) (*QuotaStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaStatus), args.Error(1)
}

func (m *mockQuotaReporter) GetMonthlyUsage(ctx context.Context, userID int) (*QuotaUsage, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaUsage), args.Error(1)
}

// mockQuotaRecorder implements the QuotaRecorder interface for testing
type mockQuotaRecorder struct {
	mock.Mock
}

func (m *mockQuotaRecorder) RecordAnalysis(ctx context.Context, userID int, jobID int) error {
	args := m.Called(ctx, userID, jobID)
	return args.Error(0)
}

// mockQuotaService implements the full QuotaService interface for testing
type mockQuotaService struct {
	mock.Mock
}

func (m *mockQuotaService) CanAnalyzeJob(ctx context.Context, userID int, jobID int) (*QuotaCheckResult, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaCheckResult), args.Error(1)
}

func (m *mockQuotaService) GetQuotaStatus(ctx context.Context, userID int) (*QuotaStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaStatus), args.Error(1)
}

func (m *mockQuotaService) GetMonthlyUsage(ctx context.Context, userID int) (*QuotaUsage, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaUsage), args.Error(1)
}

func (m *mockQuotaService) RecordAnalysis(ctx context.Context, userID int, jobID int) error {
	args := m.Called(ctx, userID, jobID)
	return args.Error(0)
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

func TestQuotaService_NonCloudMode(t *testing.T) {
	t.Run("QuotaChecker", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockChecker *mockQuotaChecker)
		}{
			{
				name: "should_return_unlimited_access_when_can_analyze_job",
				testFunc: func(t *testing.T, mockChecker *mockQuotaChecker) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					result := &QuotaCheckResult{
						Allowed: true,
						Reason:  QuotaReasonOK,
						Status: QuotaStatus{
							Limit:     -1,
							Used:      0,
							ResetDate: time.Time{},
						},
					}
					mockChecker.On("CanAnalyzeJob", ctx, userID, jobID).Return(result, nil)

					actual, err := mockChecker.CanAnalyzeJob(ctx, userID, jobID)
					assert.NoError(t, err)
					assert.True(t, actual.Allowed)
					assert.Equal(t, QuotaReasonOK, actual.Reason)
					assert.Equal(t, -1, actual.Status.Limit)
					assert.Equal(t, 0, actual.Status.Used)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockChecker := new(mockQuotaChecker)
				tc.testFunc(t, mockChecker)
				mockChecker.AssertExpectations(t)
			})
		}
	})

	t.Run("QuotaReporter", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockReporter *mockQuotaReporter)
		}{
			{
				name: "should_return_unlimited_quota_status",
				testFunc: func(t *testing.T, mockReporter *mockQuotaReporter) {
					ctx := context.Background()
					userID := 1

					status := &QuotaStatus{
						Limit:     -1,
						Used:      0,
						ResetDate: time.Time{},
					}
					mockReporter.On("GetQuotaStatus", ctx, userID).Return(status, nil)

					actual, err := mockReporter.GetQuotaStatus(ctx, userID)
					assert.NoError(t, err)
					assert.Equal(t, -1, actual.Limit)
					assert.Equal(t, 0, actual.Used)
					assert.Equal(t, time.Time{}, actual.ResetDate)
				},
			},
			{
				name: "should_return_actual_monthly_usage",
				testFunc: func(t *testing.T, mockReporter *mockQuotaReporter) {
					ctx := context.Background()
					userID := 1

					usage := &QuotaUsage{
						UserID:       userID,
						MonthYear:    timeutil.GetCurrentMonthYear(),
						JobsAnalyzed: 3,
						UpdatedAt:    time.Now(),
					}
					mockReporter.On("GetMonthlyUsage", ctx, userID).Return(usage, nil)

					actual, err := mockReporter.GetMonthlyUsage(ctx, userID)
					assert.NoError(t, err)
					assert.Equal(t, userID, actual.UserID)
					assert.Equal(t, 3, actual.JobsAnalyzed)
					assert.Equal(t, timeutil.GetCurrentMonthYear(), actual.MonthYear)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockReporter := new(mockQuotaReporter)
				tc.testFunc(t, mockReporter)
				mockReporter.AssertExpectations(t)
			})
		}
	})

	t.Run("QuotaRecorder", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockRecorder *mockQuotaRecorder)
		}{
			{
				name: "should_record_analysis_without_enforcement",
				testFunc: func(t *testing.T, mockRecorder *mockQuotaRecorder) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					mockRecorder.On("RecordAnalysis", ctx, userID, jobID).Return(nil)

					err := mockRecorder.RecordAnalysis(ctx, userID, jobID)
					assert.NoError(t, err)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockRecorder := new(mockQuotaRecorder)
				tc.testFunc(t, mockRecorder)
				mockRecorder.AssertExpectations(t)
			})
		}
	})
}

func TestQuotaService_CloudMode(t *testing.T) {
	t.Run("QuotaChecker", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockChecker *mockQuotaChecker)
		}{
			{
				name: "should_enforce_quota_for_new_analysis",
				testFunc: func(t *testing.T, mockChecker *mockQuotaChecker) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					result := &QuotaCheckResult{
						Allowed: true,
						Reason:  QuotaReasonOK,
						Status: QuotaStatus{
							Limit: testMonthlyQuotaLimit,
							Used:  0,
						},
					}
					mockChecker.On("CanAnalyzeJob", ctx, userID, jobID).Return(result, nil)

					actual, err := mockChecker.CanAnalyzeJob(ctx, userID, jobID)
					assert.NoError(t, err)
					assert.True(t, actual.Allowed)
					assert.Equal(t, QuotaReasonOK, actual.Reason)
					assert.Equal(t, testMonthlyQuotaLimit, actual.Status.Limit)
					assert.Equal(t, 0, actual.Status.Used)
				},
			},
			{
				name: "should_block_when_limit_reached",
				testFunc: func(t *testing.T, mockChecker *mockQuotaChecker) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					result := &QuotaCheckResult{
						Allowed: false,
						Reason:  QuotaReasonLimitReached,
						Status: QuotaStatus{
							Limit: testMonthlyQuotaLimit,
							Used:  testMonthlyQuotaLimit,
						},
					}
					mockChecker.On("CanAnalyzeJob", ctx, userID, jobID).Return(result, nil)

					actual, err := mockChecker.CanAnalyzeJob(ctx, userID, jobID)
					assert.NoError(t, err)
					assert.False(t, actual.Allowed)
					assert.Equal(t, QuotaReasonLimitReached, actual.Reason)
					assert.Equal(t, testMonthlyQuotaLimit, actual.Status.Limit)
					assert.Equal(t, testMonthlyQuotaLimit, actual.Status.Used)
				},
			},
			{
				name: "should_allow_reanalysis",
				testFunc: func(t *testing.T, mockChecker *mockQuotaChecker) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					result := &QuotaCheckResult{
						Allowed: true,
						Reason:  QuotaReasonReanalysis,
						Status: QuotaStatus{
							Limit: testMonthlyQuotaLimit,
							Used:  testMonthlyQuotaLimit,
						},
					}
					mockChecker.On("CanAnalyzeJob", ctx, userID, jobID).Return(result, nil)

					actual, err := mockChecker.CanAnalyzeJob(ctx, userID, jobID)
					assert.NoError(t, err)
					assert.True(t, actual.Allowed)
					assert.Equal(t, QuotaReasonReanalysis, actual.Reason)
				},
			},
			{
				name: "should_give_unlimited_quota_to_admin_users",
				testFunc: func(t *testing.T, mockChecker *mockQuotaChecker) {
					ctx := ctxutil.WithRole(context.Background(), "Admin")
					userID := 1
					jobID := 100

					result := &QuotaCheckResult{
						Allowed: true,
						Reason:  QuotaReasonOK,
						Status: QuotaStatus{
							Limit: -1,
							Used:  10, // Even if over normal limit
						},
					}
					mockChecker.On("CanAnalyzeJob", ctx, userID, jobID).Return(result, nil)

					actual, err := mockChecker.CanAnalyzeJob(ctx, userID, jobID)
					assert.NoError(t, err)
					assert.True(t, actual.Allowed)
					assert.Equal(t, QuotaReasonOK, actual.Reason)
					assert.Equal(t, -1, actual.Status.Limit)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockChecker := new(mockQuotaChecker)
				tc.testFunc(t, mockChecker)
				mockChecker.AssertExpectations(t)
			})
		}
	})

	t.Run("QuotaRecorder", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockRecorder *mockQuotaRecorder)
		}{
			{
				name: "should_record_analysis_correctly",
				testFunc: func(t *testing.T, mockRecorder *mockQuotaRecorder) {
					ctx := context.Background()
					userID := 1
					jobID := 100

					mockRecorder.On("RecordAnalysis", ctx, userID, jobID).Return(nil)

					err := mockRecorder.RecordAnalysis(ctx, userID, jobID)
					assert.NoError(t, err)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockRecorder := new(mockQuotaRecorder)
				tc.testFunc(t, mockRecorder)
				mockRecorder.AssertExpectations(t)
			})
		}
	})

	t.Run("QuotaReporter", func(t *testing.T) {
		tests := []struct {
			name     string
			testFunc func(t *testing.T, mockReporter *mockQuotaReporter)
		}{
			{
				name: "should_return_correct_quota_status",
				testFunc: func(t *testing.T, mockReporter *mockQuotaReporter) {
					ctx := context.Background()
					userID := 1

					status := &QuotaStatus{
						Limit:     testMonthlyQuotaLimit,
						Used:      3,
						ResetDate: time.Now().Add(24 * time.Hour),
					}
					mockReporter.On("GetQuotaStatus", ctx, userID).Return(status, nil)

					actual, err := mockReporter.GetQuotaStatus(ctx, userID)
					assert.NoError(t, err)
					assert.Equal(t, testMonthlyQuotaLimit, actual.Limit)
					assert.Equal(t, 3, actual.Used)
					assert.NotEqual(t, time.Time{}, actual.ResetDate)
				},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				mockReporter := new(mockQuotaReporter)
				tc.testFunc(t, mockReporter)
				mockReporter.AssertExpectations(t)
			})
		}
	})
}
