package quota

import (
	"context"
	"database/sql"
	"fmt"
)

// UnifiedService provides a single interface for all quota operations
type UnifiedService struct {
	aiQuota     *Service
	searchQuota *SearchQuotaService
	repo        Repository
}

// NewUnifiedService creates a new unified quota service
func NewUnifiedService(db *sql.DB, jobRepo JobRepository, isCloudMode bool) *UnifiedService {
	repo := NewRepository(db)

	return &UnifiedService{
		aiQuota:     NewService(db, jobRepo, isCloudMode),
		searchQuota: NewSearchQuotaService(repo, isCloudMode),
		repo:        repo,
	}
}

// CheckQuota checks any quota type
func (s *UnifiedService) CheckQuota(ctx context.Context, userID int, quotaType string, metadata map[string]interface{}) (*QuotaCheckResult, error) {
	switch quotaType {
	case QuotaTypeAIAnalysis:
		jobID, ok := metadata["job_id"].(int)
		if !ok {
			return nil, fmt.Errorf("job_id required for AI analysis quota check")
		}
		return s.aiQuota.CanAnalyzeJob(ctx, userID, jobID)

	case QuotaTypeJobSearch:
		return s.searchQuota.CanSearchJobs(ctx, userID)

	case QuotaTypeSearchRuns:
		return s.searchQuota.CanRunSearch(ctx, userID)

	default:
		return nil, fmt.Errorf("unknown quota type: %s", quotaType)
	}
}

// RecordUsage records usage for any quota type
func (s *UnifiedService) RecordUsage(ctx context.Context, userID int, quotaType string, metadata map[string]interface{}) error {
	switch quotaType {
	case QuotaTypeAIAnalysis:
		jobID, ok := metadata["job_id"].(int)
		if !ok {
			return fmt.Errorf("job_id required for AI analysis recording")
		}
		return s.aiQuota.RecordAnalysis(ctx, userID, jobID)

	case QuotaTypeJobSearch:
		count, ok := metadata["count"].(int)
		if !ok {
			count = 1
		}
		return s.searchQuota.RecordJobsFound(ctx, userID, count)

	case QuotaTypeSearchRuns:
		return s.searchQuota.RecordSearchRun(ctx, userID)

	default:
		return fmt.Errorf("unknown quota type: %s", quotaType)
	}
}

// GetAllQuotaStatus gets all quota statuses for a user
func (s *UnifiedService) GetAllQuotaStatus(ctx context.Context, userID int) (interface{}, error) {
	status := &UnifiedQuotaStatus{}

	// Get AI analysis quota (monthly)
	aiStatus, err := s.aiQuota.GetQuotaStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI quota status: %w", err)
	}
	status.AIAnalysis = *aiStatus

	// Get job search quota (daily)
	searchResult, err := s.searchQuota.GetStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job search quota status: %w", err)
	}
	status.JobSearch = searchResult.Status

	// Get search runs quota (daily)
	searchRunsResult, err := s.searchQuota.GetSearchRunStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get search runs quota status: %w", err)
	}
	status.SearchRuns = searchRunsResult.Status

	return status, nil
}

// Expose underlying services for backward compatibility
func (s *UnifiedService) AIQuotaService() *Service {
	return s.aiQuota
}

func (s *UnifiedService) SearchQuotaService() *SearchQuotaService {
	return s.searchQuota
}
