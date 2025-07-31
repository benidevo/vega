package quota

import (
	"context"
	"fmt"
	"time"

	timeutil "github.com/benidevo/vega/internal/common/time"
)

// JobCaptureService handles job capture tracking from browser extension
type JobCaptureService struct {
	repo        Repository
	isCloudMode bool
}

// NewJobCaptureService creates a new job capture service
func NewJobCaptureService(repo Repository, isCloudMode bool) *JobCaptureService {
	return &JobCaptureService{
		repo:        repo,
		isCloudMode: isCloudMode,
	}
}

// CanCaptureJobs checks if a user can capture more jobs (always returns true)
func (s *JobCaptureService) CanCaptureJobs(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := timeutil.GetCurrentDate()
	usage, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsCaptured)
	if err != nil {
		return nil, fmt.Errorf("failed to get job capture usage: %w", err)
	}

	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status: QuotaStatus{
			Used:      usage,
			Limit:     -1,
			ResetDate: time.Time{},
		},
	}, nil
}

// RecordJobsCaptured records that jobs were captured via extension
func (s *JobCaptureService) RecordJobsCaptured(ctx context.Context, userID int, count int) error {
	today := timeutil.GetCurrentDate()
	return s.repo.IncrementDailyUsage(ctx, userID, today, QuotaKeyJobsCaptured, count)
}

// GetStatus returns the current job capture status
func (s *JobCaptureService) GetStatus(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := timeutil.GetCurrentDate()
	jobsCaptured, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsCaptured)
	if err != nil {
		return nil, fmt.Errorf("failed to get job capture usage: %w", err)
	}

	// Job captures are unlimited for everyone
	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status: QuotaStatus{
			Used:      jobsCaptured,
			Limit:     -1,
			ResetDate: time.Time{},
		},
	}, nil
}
