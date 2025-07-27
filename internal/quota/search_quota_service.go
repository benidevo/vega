package quota

import (
	"context"
	"fmt"
	"time"

	timeutil "github.com/benidevo/vega/internal/common/time"
)

// SearchQuotaService handles search-related quota management
type SearchQuotaService struct {
	repo        Repository
	isCloudMode bool
}

// NewSearchQuotaService creates a new search quota service
func NewSearchQuotaService(repo Repository, isCloudMode bool) *SearchQuotaService {
	return &SearchQuotaService{
		repo:        repo,
		isCloudMode: isCloudMode,
	}
}

// CanSearchJobs checks if a user can search for more jobs
func (s *SearchQuotaService) CanSearchJobs(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := timeutil.GetCurrentDate()
	usage, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsFound)
	if err != nil {
		return nil, fmt.Errorf("failed to get job search usage: %w", err)
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

// RecordJobsFound records that jobs were found
func (s *SearchQuotaService) RecordJobsFound(ctx context.Context, userID int, count int) error {
	today := timeutil.GetCurrentDate()
	return s.repo.IncrementDailyUsage(ctx, userID, today, QuotaKeyJobsFound, count)
}

// GetStatus returns the current search quota status
func (s *SearchQuotaService) GetStatus(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := timeutil.GetCurrentDate()
	jobsFound, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsFound)
	if err != nil {
		return nil, fmt.Errorf("failed to get job search usage: %w", err)
	}

	// Job searches are unlimited for everyone
	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status: QuotaStatus{
			Used:      jobsFound,
			Limit:     -1,
			ResetDate: time.Time{},
		},
	}, nil
}
