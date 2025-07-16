package quota

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// JobRepository interface defines methods the quota service needs from the job repository
type JobRepository interface {
	GetByID(ctx context.Context, userID, jobID int) (*Job, error)
	SetFirstAnalyzedAt(ctx context.Context, jobID int) error
}

// Service handles quota management
type Service struct {
	db          *sql.DB
	jobRepo     JobRepository
	isCloudMode bool
}

// NewService creates a new quota service
func NewService(db *sql.DB, jobRepo JobRepository, isCloudMode bool) *Service {
	return &Service{
		db:          db,
		jobRepo:     jobRepo,
		isCloudMode: isCloudMode,
	}
}

// CanAnalyzeJob checks if a user can analyze a specific job
func (s *Service) CanAnalyzeJob(ctx context.Context, userID int, jobID int) (*QuotaCheckResult, error) {
	// In non-cloud mode, always allow unlimited access
	if !s.isCloudMode {
		return &QuotaCheckResult{
			Allowed: true,
			Reason:  QuotaReasonOK,
			Status: QuotaStatus{
				Used:      0,
				Limit:     -1,          // -1 indicates unlimited
				ResetDate: time.Time{}, // No reset date for unlimited
			},
		}, nil
	}

	// Cloud mode: enforce quotas
	// 1. Check if job was previously analyzed
	job, err := s.jobRepo.GetByID(ctx, userID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Get current usage for status
	usage, err := s.GetMonthlyUsage(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly usage: %w", err)
	}

	status := QuotaStatus{
		Used:      usage.JobsAnalyzed,
		Limit:     FreeUserMonthlyLimit,
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

	// 2. Check monthly limit for new analyses
	if usage.JobsAnalyzed >= FreeUserMonthlyLimit {
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
	// In non-cloud mode, don't track quota usage
	if !s.isCloudMode {
		return nil
	}

	// Cloud mode: track quota usage
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

	// Use UPSERT pattern to avoid race conditions
	upsertQuery := `
		INSERT INTO user_quota_usage (user_id, month_year, jobs_analyzed, updated_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, month_year) DO UPDATE SET
			jobs_analyzed = jobs_analyzed + 1,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = tx.ExecContext(ctx, upsertQuery, userID, monthYear)
	if err != nil {
		return fmt.Errorf("failed to update quota usage: %w", err)
	}

	return tx.Commit()
}

// GetMonthlyUsage gets the current month's usage for a user
func (s *Service) GetMonthlyUsage(ctx context.Context, userID int) (*QuotaUsage, error) {
	// In non-cloud mode, always return zero usage
	if !s.isCloudMode {
		return &QuotaUsage{
			UserID:       userID,
			MonthYear:    getCurrentMonthYear(),
			JobsAnalyzed: 0,
			UpdatedAt:    time.Now(),
		}, nil
	}

	monthYear := getCurrentMonthYear()

	usage := &QuotaUsage{
		UserID:       userID,
		MonthYear:    monthYear,
		JobsAnalyzed: 0,
		UpdatedAt:    time.Now(),
	}

	query := `
		SELECT user_id, month_year, jobs_analyzed, updated_at
		FROM user_quota_usage
		WHERE user_id = ? AND month_year = ?
	`

	row := s.db.QueryRowContext(ctx, query, userID, monthYear)
	err := row.Scan(&usage.UserID, &usage.MonthYear, &usage.JobsAnalyzed, &usage.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// No usage recorded yet for this month, return zero usage
			return usage, nil
		}
		return nil, fmt.Errorf("failed to get quota usage: %w", err)
	}

	return usage, nil
}

// GetQuotaStatus returns the current quota status for a user
func (s *Service) GetQuotaStatus(ctx context.Context, userID int) (*QuotaStatus, error) {
	// In non-cloud mode, return unlimited quota
	if !s.isCloudMode {
		return &QuotaStatus{
			Used:      0,
			Limit:     -1,          // -1 indicates unlimited
			ResetDate: time.Time{}, // No reset date for unlimited
		}, nil
	}

	// Cloud mode: return actual quota status
	usage, err := s.GetMonthlyUsage(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &QuotaStatus{
		Used:      usage.JobsAnalyzed,
		Limit:     FreeUserMonthlyLimit,
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
