package quota

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ctxutil "github.com/benidevo/vega/internal/common/context"
)

// JobRepository interface defines methods the quota service needs from the job repository
type JobRepository interface {
	GetByID(ctx context.Context, userID, jobID int) (*Job, error)
	SetFirstAnalyzedAt(ctx context.Context, jobID int) error
}

// Service handles quota management
type Service struct {
	db          *sql.DB
	repo        Repository
	jobRepo     JobRepository
	isCloudMode bool
}

// NewService creates a new quota service
func NewService(db *sql.DB, jobRepo JobRepository, isCloudMode bool) *Service {
	return &Service{
		db:          db,
		repo:        NewRepository(db),
		jobRepo:     jobRepo,
		isCloudMode: isCloudMode,
	}
}

// isUserAdmin checks if the user in context has admin role
func (s *Service) isUserAdmin(ctx context.Context) bool {
	role, _ := ctxutil.GetRole(ctx)
	return role == "Admin"
}

// CanAnalyzeJob checks if a user can analyze a specific job
func (s *Service) CanAnalyzeJob(ctx context.Context, userID int, jobID int) (*QuotaCheckResult, error) {
	// Check if job was previously analyzed
	job, err := s.jobRepo.GetByID(ctx, userID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Get current usage for status
	monthYear := getCurrentMonthYear()
	usage, err := s.repo.GetMonthlyUsage(ctx, userID, monthYear)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}

	// Check if user is admin (admins have unlimited quota in cloud mode)
	if s.isCloudMode && s.isUserAdmin(ctx) {
		status := QuotaStatus{
			Used:      usage.JobsAnalyzed,
			Limit:     -1,          // -1 indicates unlimited
			ResetDate: time.Time{}, // No reset date for unlimited
		}

		// If job was previously analyzed, note it as re-analysis
		reason := QuotaReasonOK
		if job.FirstAnalyzedAt != nil {
			reason = QuotaReasonReanalysis
		}

		return &QuotaCheckResult{
			Allowed: true,
			Reason:  reason,
			Status:  status,
		}, nil
	}

	// In non-cloud mode, always allow but show actual usage
	if !s.isCloudMode {
		status := QuotaStatus{
			Used:      usage.JobsAnalyzed,
			Limit:     -1,          // -1 indicates unlimited
			ResetDate: time.Time{}, // No reset date for unlimited
		}

		// If job was previously analyzed, note it as re-analysis
		reason := QuotaReasonOK
		if job.FirstAnalyzedAt != nil {
			reason = QuotaReasonReanalysis
		}

		return &QuotaCheckResult{
			Allowed: true,
			Reason:  reason,
			Status:  status,
		}, nil
	}

	// Cloud mode: get quota configuration
	quotaConfig, err := s.repo.GetQuotaConfig(ctx, "ai_analysis_monthly")
	if err != nil {
		return nil, fmt.Errorf("failed to get quota config: %w", err)
	}

	limit := quotaConfig.FreeLimit

	// Cloud mode: enforce quotas
	status := QuotaStatus{
		Used:      usage.JobsAnalyzed,
		Limit:     limit,
		ResetDate: getNextMonthStart(),
	}

	// If job was previously analyzed, it's a re-analysis (always allowed)
	if job.FirstAnalyzedAt != nil {
		return &QuotaCheckResult{
			Allowed: true,
			Reason:  QuotaReasonReanalysis,
			Status:  status,
		}, nil
	}

	// Check monthly limit for new analyses
	if usage.JobsAnalyzed >= limit {
		return &QuotaCheckResult{
			Allowed: false,
			Reason:  QuotaReasonLimitReached,
			Status:  status,
		}, nil
	}

	return &QuotaCheckResult{
		Allowed: true,
		Reason:  QuotaReasonOK,
		Status:  status,
	}, nil
}

// RecordAnalysis records that a job has been analyzed
func (s *Service) RecordAnalysis(ctx context.Context, userID int, jobID int) error {
	// Always record usage for tracking purposes
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = s.jobRepo.SetFirstAnalyzedAt(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to set first analyzed at: %w", err)
	}

	monthYear := getCurrentMonthYear()

	// Use repository to increment usage
	err = s.repo.IncrementMonthlyUsage(ctx, userID, monthYear)
	if err != nil {
		return fmt.Errorf("failed to update quota usage: %w", err)
	}

	return tx.Commit()
}

// GetMonthlyUsage gets the current month's usage for a user
func (s *Service) GetMonthlyUsage(ctx context.Context, userID int) (*QuotaUsage, error) {
	// Always get actual usage data for tracking purposes
	monthYear := getCurrentMonthYear()
	return s.repo.GetMonthlyUsage(ctx, userID, monthYear)
}

// GetQuotaStatus returns the current quota status for a user
func (s *Service) GetQuotaStatus(ctx context.Context, userID int) (*QuotaStatus, error) {
	// Always get actual usage data
	usage, err := s.GetMonthlyUsage(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is admin (admins have unlimited quota in cloud mode)
	if s.isCloudMode && s.isUserAdmin(ctx) {
		return &QuotaStatus{
			Used:      usage.JobsAnalyzed,
			Limit:     -1,          // -1 indicates unlimited
			ResetDate: time.Time{}, // No reset date for unlimited
		}, nil
	}

	if !s.isCloudMode {
		// In self-hosted mode, return actual usage but unlimited limit
		return &QuotaStatus{
			Used:      usage.JobsAnalyzed,
			Limit:     -1,          // -1 indicates unlimited
			ResetDate: time.Time{}, // No reset date for unlimited
		}, nil
	}

	// Cloud mode: get quota configuration
	quotaConfig, err := s.repo.GetQuotaConfig(ctx, "ai_analysis_monthly")
	if err != nil {
		return nil, err
	}

	// Cloud mode: return actual quota status with limits
	return &QuotaStatus{
		Used:      usage.JobsAnalyzed,
		Limit:     quotaConfig.FreeLimit,
		ResetDate: getNextMonthStart(),
	}, nil
}

// getCurrentMonthYear returns the current month in "YYYY-MM" format (UTC)
func getCurrentMonthYear() string {
	return time.Now().UTC().Format("2006-01")
}

// getNextMonthStart returns the first day of next month (UTC)
func getNextMonthStart() time.Time {
	now := time.Now().UTC()
	// Get first day of current month
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Add one month
	return firstOfMonth.AddDate(0, 1, 0)
}
