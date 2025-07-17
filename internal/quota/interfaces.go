package quota

import "context"

// QuotaService defines the interface for quota management
type QuotaService interface {
	CanAnalyzeJob(ctx context.Context, userID int, jobID int) (*QuotaCheckResult, error)
	RecordAnalysis(ctx context.Context, userID int, jobID int) error
	GetMonthlyUsage(ctx context.Context, userID int) (*QuotaUsage, error)
	GetQuotaStatus(ctx context.Context, userID int) (*QuotaStatus, error)
}
