package monitoring

import (
	"context"

	"github.com/benidevo/vega/internal/quota"
)

// QuotaServiceWithMetrics wraps the quota service to add monitoring
type QuotaServiceWithMetrics struct {
	*quota.Service
	monitor *Monitor
}

// NewQuotaServiceWithMetrics creates a new quota service wrapper with metrics
func NewQuotaServiceWithMetrics(service *quota.Service, monitor *Monitor) *QuotaServiceWithMetrics {
	if monitor == nil || service == nil {
		return &QuotaServiceWithMetrics{Service: service}
	}

	return &QuotaServiceWithMetrics{
		Service: service,
		monitor: monitor,
	}
}

// GetMonthlyUsage gets monthly usage and records metrics
func (q *QuotaServiceWithMetrics) GetMonthlyUsage(ctx context.Context, userID int) (*quota.QuotaUsage, error) {
	usage, err := q.Service.GetMonthlyUsage(ctx, userID)
	if err != nil {
		return nil, err
	}

	if q.monitor != nil && usage != nil {
		// Calculate usage percentage (assuming 100 jobs per month limit)
		percentage := float64(usage.JobsAnalyzed) / 100.0 * 100.0
		if percentage > 100 {
			percentage = 100
		}

		q.monitor.RecordQuotaUsage(ctx, usage.MonthYear, percentage)
	}

	return usage, nil
}

// RecordAnalysis records analysis and updates metrics
func (q *QuotaServiceWithMetrics) RecordAnalysis(ctx context.Context, userID int, jobID int) error {
	err := q.Service.RecordAnalysis(ctx, userID, jobID)
	if err != nil {
		return err
	}

	if q.monitor != nil {
		usage, _ := q.Service.GetMonthlyUsage(ctx, userID)
		if usage != nil {
			percentage := float64(usage.JobsAnalyzed) / 100.0 * 100.0
			if percentage > 100 {
				percentage = 100
			}
			q.monitor.RecordQuotaUsage(ctx, usage.MonthYear, percentage)
		}
	}

	return nil
}
