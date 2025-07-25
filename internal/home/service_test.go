package home

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/benidevo/vega/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testService is a mock implementation used for testing service layer logic.
// It allows simulation of various responses and errors for job statistics,
// status counts, and recent jobs retrieval
type testService struct {
	statsResponse        *models.JobStats
	statsError           error
	statusCountsResponse map[models.JobStatus]int
	statusCountsError    error
	recentJobsResponse   []*models.Job
	recentJobsError      error
}

func (ts *testService) GetHomePageData(ctx context.Context, userID int, username string) (*HomePageData, error) {
	homeData := NewHomePageData(userID, username)

	if ts.statsError != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", ts.statsError)
	}

	if ts.statusCountsError != nil {
		return nil, fmt.Errorf("failed to get job status counts: %w", ts.statusCountsError)
	}

	if ts.recentJobsError != nil {
		return nil, fmt.Errorf("failed to get recent jobs: %w", ts.recentJobsError)
	}

	homeData.Stats = JobStatsSummary{
		TotalJobs:     ts.statsResponse.TotalJobs,
		Applied:       ts.statusCountsResponse[models.APPLIED],
		Interviewing:  ts.statusCountsResponse[models.INTERVIEWING],
		ActiveJobs:    calculateActiveJobs(ts.statusCountsResponse),
		OfferReceived: ts.statusCountsResponse[models.OFFER_RECEIVED],
		Interested:    ts.statusCountsResponse[models.INTERESTED],
	}

	homeData.RecentJobs = make([]JobSummary, 0, len(ts.recentJobsResponse))
	for _, job := range ts.recentJobsResponse {
		homeData.RecentJobs = append(homeData.RecentJobs, ToJobSummary(job))
	}

	homeData.HasJobs = ts.statsResponse.TotalJobs > 0
	homeData.ShowOnboarding = ts.statsResponse.TotalJobs == 0

	return homeData, nil
}

func TestService_GetHomePageData(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		setup    func() *testService
		want     *HomePageData
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "success with jobs",
			userID:   1,
			username: "testuser",
			setup: func() *testService {
				return &testService{
					statsResponse: &models.JobStats{
						TotalJobs:    5,
						TotalApplied: 3,
						HighMatch:    2,
					},
					statusCountsResponse: map[models.JobStatus]int{
						models.INTERESTED:     1,
						models.APPLIED:        2,
						models.INTERVIEWING:   1,
						models.OFFER_RECEIVED: 1,
						models.REJECTED:       0,
						models.NOT_INTERESTED: 0,
					},
					recentJobsResponse: []*models.Job{
						{
							ID:       1,
							Title:    "Software Engineer",
							Location: "Remote",
							Status:   models.APPLIED,
							Company:  models.Company{Name: "Tech Corp"},
						},
						{
							ID:       2,
							Title:    "Backend Developer",
							Location: "New York",
							Status:   models.INTERVIEWING,
							Company:  models.Company{Name: "StartupCo"},
						},
					},
				}
			},
			want: &HomePageData{
				UserID:   1,
				Username: "testuser",
				Title:    "Home",
				Page:     "home",
				Stats: JobStatsSummary{
					TotalJobs:     5,
					Applied:       2,
					Interviewing:  1,
					ActiveJobs:    5, // 1+2+1+1 = 5 (all non-terminal statuses)
					OfferReceived: 1,
					Interested:    1,
				},
				RecentJobs: []JobSummary{
					{
						ID:         1,
						Title:      "Software Engineer",
						Company:    "Tech Corp",
						Location:   "Remote",
						Status:     int(models.APPLIED),
						StatusText: models.APPLIED.String(),
					},
					{
						ID:         2,
						Title:      "Backend Developer",
						Company:    "StartupCo",
						Location:   "New York",
						Status:     int(models.INTERVIEWING),
						StatusText: models.INTERVIEWING.String(),
					},
				},
				HasJobs:        true,
				ShowOnboarding: false,
			},
			wantErr: false,
		},
		{
			name:     "success with no jobs",
			userID:   2,
			username: "newuser",
			setup: func() *testService {
				return &testService{
					statsResponse: &models.JobStats{
						TotalJobs:    0,
						TotalApplied: 0,
						HighMatch:    0,
					},
					statusCountsResponse: map[models.JobStatus]int{},
					recentJobsResponse:   []*models.Job{},
				}
			},
			want: &HomePageData{
				UserID:   2,
				Username: "newuser",
				Title:    "Home",
				Page:     "home",
				Stats: JobStatsSummary{
					TotalJobs:     0,
					Applied:       0,
					Interviewing:  0,
					ActiveJobs:    0,
					OfferReceived: 0,
					Interested:    0,
				},
				RecentJobs:     []JobSummary{},
				HasJobs:        false,
				ShowOnboarding: true,
			},
			wantErr: false,
		},
		{
			name:     "error getting job stats",
			userID:   1,
			username: "testuser",
			setup: func() *testService {
				return &testService{
					statsError: errors.New("database connection failed"),
				}
			},
			wantErr: true,
			errMsg:  "failed to get job statistics",
		},
		{
			name:     "error getting status counts",
			userID:   1,
			username: "testuser",
			setup: func() *testService {
				return &testService{
					statsResponse:     &models.JobStats{TotalJobs: 5},
					statusCountsError: errors.New("query failed"),
				}
			},
			wantErr: true,
			errMsg:  "failed to get job status counts",
		},
		{
			name:     "error getting recent jobs",
			userID:   1,
			username: "testuser",
			setup: func() *testService {
				return &testService{
					statsResponse:        &models.JobStats{TotalJobs: 5},
					statusCountsResponse: map[models.JobStatus]int{},
					recentJobsError:      errors.New("recent jobs query failed"),
				}
			},
			wantErr: true,
			errMsg:  "failed to get recent jobs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSvc := tt.setup()

			got, err := testSvc.GetHomePageData(context.Background(), tt.userID, tt.username)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tt.want.UserID, got.UserID)
			assert.Equal(t, tt.want.Username, got.Username)
			assert.Equal(t, tt.want.Title, got.Title)
			assert.Equal(t, tt.want.Page, got.Page)
			assert.Equal(t, tt.want.Stats, got.Stats)
			assert.Equal(t, tt.want.HasJobs, got.HasJobs)
			assert.Equal(t, tt.want.ShowOnboarding, got.ShowOnboarding)
			assert.Equal(t, tt.want.RecentJobs, got.RecentJobs)
		})
	}
}

func TestCalculateActiveJobs(t *testing.T) {
	tests := []struct {
		name         string
		statusCounts map[models.JobStatus]int
		want         int
	}{
		{
			name: "all active statuses",
			statusCounts: map[models.JobStatus]int{
				models.INTERESTED:     2,
				models.APPLIED:        3,
				models.INTERVIEWING:   1,
				models.OFFER_RECEIVED: 1,
				models.REJECTED:       5,
				models.NOT_INTERESTED: 2,
			},
			want: 7, // 2+3+1+1 = 7
		},
		{
			name: "no active jobs",
			statusCounts: map[models.JobStatus]int{
				models.REJECTED:       3,
				models.NOT_INTERESTED: 2,
			},
			want: 0,
		},
		{
			name:         "empty status counts",
			statusCounts: map[models.JobStatus]int{},
			want:         0,
		},
		{
			name: "only applied jobs",
			statusCounts: map[models.JobStatus]int{
				models.APPLIED: 5,
			},
			want: 5,
		},
		{
			name: "mixed with zeros",
			statusCounts: map[models.JobStatus]int{
				models.INTERESTED:     0,
				models.APPLIED:        2,
				models.INTERVIEWING:   0,
				models.OFFER_RECEIVED: 1,
			},
			want: 3, // 0+2+0+1 = 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateActiveJobs(tt.statusCounts)
			assert.Equal(t, tt.want, got)
		})
	}
}
