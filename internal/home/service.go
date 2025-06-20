package home

import (
	"context"
	"fmt"

	"github.com/benidevo/vega/internal/job/models"
	"github.com/benidevo/vega/internal/job/repository"
)

// Service handles business logic for homepage data aggregation
type Service struct {
	jobRepository *repository.SQLiteJobRepository
}

// NewService creates a new homepage service instance
func NewService(jobRepository *repository.SQLiteJobRepository) *Service {
	return &Service{
		jobRepository: jobRepository,
	}
}

// GetHomePageData aggregates all data needed for the homepage display
func (s *Service) GetHomePageData(ctx context.Context, userID int, username string) (*HomePageData, error) {
	homeData := NewHomePageData(userID, username)

	jobStats, err := s.jobRepository.GetStatsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", err)
	}

	statusCounts, err := s.jobRepository.GetJobStatsByStatus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status counts: %w", err)
	}

	// Get recent jobs for activity display (last 3)
	recentJobs, err := s.jobRepository.GetRecentJobsByUserID(ctx, userID, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent jobs: %w", err)
	}

	homeData.Stats = JobStatsSummary{
		TotalJobs:     jobStats.TotalJobs,
		Applied:       statusCounts[models.APPLIED],
		Interviewing:  statusCounts[models.INTERVIEWING],
		ActiveJobs:    calculateActiveJobs(statusCounts),
		OfferReceived: statusCounts[models.OFFER_RECEIVED],
	}

	homeData.RecentJobs = make([]JobSummary, 0, len(recentJobs))
	for _, job := range recentJobs {
		homeData.RecentJobs = append(homeData.RecentJobs, ToJobSummary(job))
	}

	homeData.HasJobs = jobStats.TotalJobs > 0
	homeData.ShowOnboarding = jobStats.TotalJobs == 0

	return homeData, nil
}

// calculateActiveJobs determines the count of "active" jobs
// Active jobs are those that are not in terminal states (rejected/not interested)
func calculateActiveJobs(statusCounts map[models.JobStatus]int) int {
	active := 0

	// Sum non-terminal statuses
	active += statusCounts[models.INTERESTED]
	active += statusCounts[models.APPLIED]
	active += statusCounts[models.INTERVIEWING]
	active += statusCounts[models.OFFER_RECEIVED]

	return active
}
