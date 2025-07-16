package quota

import (
	"context"

	"github.com/benidevo/vega/internal/job/interfaces"
)

// JobRepositoryAdapter adapts the job repository interface for quota service
type JobRepositoryAdapter struct {
	jobRepo interfaces.JobRepository
}

// NewJobRepositoryAdapter creates a new adapter
func NewJobRepositoryAdapter(jobRepo interfaces.JobRepository) JobRepository {
	return &JobRepositoryAdapter{
		jobRepo: jobRepo,
	}
}

// GetByID adapts the job repository GetByID method for quota service
func (a *JobRepositoryAdapter) GetByID(ctx context.Context, userID, jobID int) (*Job, error) {
	job, err := a.jobRepo.GetByID(ctx, userID, jobID)
	if err != nil {
		return nil, err
	}

	// Convert to quota Job struct
	return &Job{
		ID:              job.ID,
		FirstAnalyzedAt: job.FirstAnalyzedAt,
	}, nil
}

// GetMonthlyAnalysisCount delegates to the job repository
func (a *JobRepositoryAdapter) GetMonthlyAnalysisCount(ctx context.Context, userID int) (int, error) {
	return a.jobRepo.GetMonthlyAnalysisCount(ctx, userID)
}

// SetFirstAnalyzedAt delegates to the job repository
func (a *JobRepositoryAdapter) SetFirstAnalyzedAt(ctx context.Context, jobID int) error {
	return a.jobRepo.SetFirstAnalyzedAt(ctx, jobID)
}
