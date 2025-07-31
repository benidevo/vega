package quota

import (
	"context"
	"errors"
	"testing"
	"time"

	timeutil "github.com/benidevo/vega/internal/common/time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockJobCaptureRepository struct {
	mock.Mock
}

func (m *MockJobCaptureRepository) GetMonthlyUsage(ctx context.Context, userID int, monthYear string) (*QuotaUsage, error) {
	args := m.Called(ctx, userID, monthYear)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaUsage), args.Error(1)
}

func (m *MockJobCaptureRepository) IncrementMonthlyUsage(ctx context.Context, userID int, monthYear string) error {
	args := m.Called(ctx, userID, monthYear)
	return args.Error(0)
}

func (m *MockJobCaptureRepository) GetDailyUsage(ctx context.Context, userID int, date string, quotaKey string) (int, error) {
	args := m.Called(ctx, userID, date, quotaKey)
	return args.Int(0), args.Error(1)
}

func (m *MockJobCaptureRepository) IncrementDailyUsage(ctx context.Context, userID int, date string, quotaKey string, amount int) error {
	args := m.Called(ctx, userID, date, quotaKey, amount)
	return args.Error(0)
}

func (m *MockJobCaptureRepository) GetAllDailyUsage(ctx context.Context, userID int, date string) (map[string]int, error) {
	args := m.Called(ctx, userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockJobCaptureRepository) GetQuotaConfig(ctx context.Context, quotaType string) (*QuotaConfig, error) {
	args := m.Called(ctx, quotaType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaConfig), args.Error(1)
}

func TestJobCaptureService_CanCaptureJobs(t *testing.T) {
	today := timeutil.GetCurrentDate()

	tests := []struct {
		name          string
		userID        int
		isCloudMode   bool
		setupMock     func(*MockJobCaptureRepository)
		expectedAllow bool
		expectedUsed  int
		expectError   bool
		errorContains string
	}{
		{
			name:        "should_allow_capture_when_no_usage",
			userID:      1,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 1, today, QuotaKeyJobsCaptured).
					Return(0, nil)
			},
			expectedAllow: true,
			expectedUsed:  0,
		},
		{
			name:        "should_allow_capture_when_has_usage",
			userID:      2,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 2, today, QuotaKeyJobsCaptured).
					Return(50, nil)
			},
			expectedAllow: true,
			expectedUsed:  50,
		},
		{
			name:        "should_allow_capture_when_self_hosted",
			userID:      3,
			isCloudMode: false,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 3, today, QuotaKeyJobsCaptured).
					Return(100, nil)
			},
			expectedAllow: true,
			expectedUsed:  100,
		},
		{
			name:        "should_return_error_when_repository_fails",
			userID:      1,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 1, today, QuotaKeyJobsCaptured).
					Return(0, errors.New("database error"))
			},
			expectError:   true,
			errorContains: "failed to get job capture usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockJobCaptureRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewJobCaptureService(mockRepo, tt.isCloudMode)
			result, err := service.CanCaptureJobs(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedAllow, result.Allowed)
				assert.Equal(t, QuotaReasonOK, result.Reason)
				assert.Equal(t, tt.expectedUsed, result.Status.Used)
				assert.Equal(t, -1, result.Status.Limit)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobCaptureService_RecordJobsCaptured(t *testing.T) {
	today := timeutil.GetCurrentDate()

	tests := []struct {
		name          string
		userID        int
		count         int
		setupMock     func(*MockJobCaptureRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:   "should_record_jobs_captured_when_count_positive",
			userID: 1,
			count:  10,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("IncrementDailyUsage", mock.Anything, 1, today, QuotaKeyJobsCaptured, 10).
					Return(nil)
			},
		},
		{
			name:   "should_record_zero_jobs_captured",
			userID: 2,
			count:  0,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("IncrementDailyUsage", mock.Anything, 2, today, QuotaKeyJobsCaptured, 0).
					Return(nil)
			},
		},
		{
			name:   "should_return_error_when_repository_fails",
			userID: 3,
			count:  5,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("IncrementDailyUsage", mock.Anything, 3, today, QuotaKeyJobsCaptured, 5).
					Return(errors.New("database error"))
			},
			expectError:   true,
			errorContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockJobCaptureRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewJobCaptureService(mockRepo, true)
			err := service.RecordJobsCaptured(context.Background(), tt.userID, tt.count)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobCaptureService_GetStatus(t *testing.T) {
	today := timeutil.GetCurrentDate()

	tests := []struct {
		name          string
		userID        int
		isCloudMode   bool
		setupMock     func(*MockJobCaptureRepository)
		expectedUsed  int
		expectError   bool
		errorContains string
	}{
		{
			name:        "should_return_status_when_no_usage",
			userID:      1,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 1, today, QuotaKeyJobsCaptured).
					Return(0, nil)
			},
			expectedUsed: 0,
		},
		{
			name:        "should_return_status_when_has_usage",
			userID:      2,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 2, today, QuotaKeyJobsCaptured).
					Return(25, nil)
			},
			expectedUsed: 25,
		},
		{
			name:        "should_return_status_when_self_hosted",
			userID:      3,
			isCloudMode: false,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 3, today, QuotaKeyJobsCaptured).
					Return(100, nil)
			},
			expectedUsed: 100,
		},
		{
			name:        "should_return_error_when_repository_fails",
			userID:      1,
			isCloudMode: true,
			setupMock: func(m *MockJobCaptureRepository) {
				m.On("GetDailyUsage", mock.Anything, 1, today, QuotaKeyJobsCaptured).
					Return(0, errors.New("database error"))
			},
			expectError:   true,
			errorContains: "failed to get job capture usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockJobCaptureRepository{}
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			service := NewJobCaptureService(mockRepo, tt.isCloudMode)
			result, err := service.GetStatus(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.Allowed)
				assert.Equal(t, QuotaReasonOK, result.Reason)
				assert.Equal(t, tt.expectedUsed, result.Status.Used)
				assert.Equal(t, -1, result.Status.Limit)
				assert.Equal(t, time.Time{}, result.Status.ResetDate)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
