package home

import (
	"context"
	"fmt"

	"github.com/benidevo/vega/internal/job"
	"github.com/benidevo/vega/internal/job/models"
	"github.com/benidevo/vega/internal/job/repository"
)

// Service handles business logic for homepage data aggregation
type Service struct {
	jobRepository *repository.SQLiteJobRepository
	jobService    *job.JobService
}

// NewService creates a new homepage service instance
func NewService(jobRepository *repository.SQLiteJobRepository, jobService *job.JobService) *Service {
	return &Service{
		jobRepository: jobRepository,
		jobService:    jobService,
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
		Interested:    statusCounts[models.INTERESTED],
	}

	homeData.RecentJobs = make([]JobSummary, 0, len(recentJobs))
	for _, job := range recentJobs {
		homeData.RecentJobs = append(homeData.RecentJobs, ToJobSummary(job))
	}

	homeData.HasJobs = jobStats.TotalJobs > 0
	homeData.ShowOnboarding = jobStats.TotalJobs == 0

	// Get quota status if job service is available
	if s.jobService != nil {
		quotaStatus, err := s.jobService.GetQuotaStatus(ctx, userID)
		if err == nil && quotaStatus != nil {
			remaining := 0
			percentage := 0

			if quotaStatus.Limit < 0 {
				// Unlimited quota
				remaining = -1
				percentage = 0
			} else if quotaStatus.Limit > 0 {
				// Calculate remaining, ensuring it's never negative
				remaining = quotaStatus.Limit - quotaStatus.Used
				if remaining < 0 {
					remaining = 0
				}

				// Calculate percentage, capping at 100%
				percentage = (quotaStatus.Used * 100) / quotaStatus.Limit
				if percentage > 100 {
					percentage = 100
				}
			}

			homeData.QuotaStatus = &QuotaStatus{
				Used:       quotaStatus.Used,
				Limit:      quotaStatus.Limit,
				Remaining:  remaining,
				ResetDate:  quotaStatus.ResetDate,
				Percentage: percentage,
			}
		}
	}

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
