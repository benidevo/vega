package quota

import (
	"context"
	"fmt"
	"time"
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
	today := getCurrentDate()
	usage, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsFound)
	if err != nil {
		return nil, fmt.Errorf("failed to get job search usage: %w", err)
	}

	if !s.isCloudMode {
		// In self-hosted mode, always allow but show actual usage
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

	// Cloud mode: enforce limits
	limit := FreeUserDailyJobSearchLimit
	status := QuotaStatus{
		Used:      usage,
		Limit:     limit,
		ResetDate: getTomorrowStart(),
	}

	if usage >= limit {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Daily limit of %d jobs reached", limit),
			Status:  status,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status:  status,
	}, nil
}

// CanRunSearch checks if a user can run another search
func (s *SearchQuotaService) CanRunSearch(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := getCurrentDate()
	usage, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeySearchesRun)
	if err != nil {
		return nil, fmt.Errorf("failed to get search run usage: %w", err)
	}

	if !s.isCloudMode {
		// In self-hosted mode, always allow but show actual usage
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

	// Cloud mode: enforce limits
	limit := FreeUserDailySearchRunLimit
	status := QuotaStatus{
		Used:      usage,
		Limit:     limit,
		ResetDate: getTomorrowStart(),
	}

	if usage >= limit {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Daily limit of %d searches reached", limit),
			Status:  status,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status:  status,
	}, nil
}

// RecordJobsFound records that jobs were found
func (s *SearchQuotaService) RecordJobsFound(ctx context.Context, userID int, count int) error {
	// Always record usage for tracking purposes
	today := getCurrentDate()
	return s.repo.IncrementDailyUsage(ctx, userID, today, QuotaKeyJobsFound, count)
}

// RecordSearchRun records that a search was run
func (s *SearchQuotaService) RecordSearchRun(ctx context.Context, userID int) error {
	// Always record usage for tracking purposes
	today := getCurrentDate()
	return s.repo.IncrementDailyUsage(ctx, userID, today, QuotaKeySearchesRun, 1)
}

// GetStatus returns the current search quota status
func (s *SearchQuotaService) GetStatus(ctx context.Context, userID int) (*QuotaCheckResult, error) {
	today := getCurrentDate()
	jobsFound, err := s.repo.GetDailyUsage(ctx, userID, today, QuotaKeyJobsFound)
	if err != nil {
		return nil, fmt.Errorf("failed to get job search usage: %w", err)
	}

	if !s.isCloudMode {
		// In self-hosted mode, return actual usage but unlimited limit
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

	// Cloud mode: enforce limits
	limit := FreeUserDailyJobSearchLimit
	return &QuotaCheckResult{
		Allowed: jobsFound < limit,
		Reason:  QuotaReasonOK,
		Status: QuotaStatus{
			Used:      jobsFound,
			Limit:     limit,
			ResetDate: getTomorrowStart(),
		},
	}, nil
}

// getCurrentDate returns the current date in "2006-01-02" format (UTC)
func getCurrentDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// getTomorrowStart returns the start of tomorrow (UTC)
func getTomorrowStart() time.Time {
	now := time.Now().UTC()
	tomorrow := now.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
}
