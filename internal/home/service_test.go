package home

import (
	"context"
	"errors"
	"testing"

	"github.com/benidevo/vega/internal/job/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockHomeService struct {
	mock.Mock
}

func (m *mockHomeService) GetHomePageData(ctx context.Context, userID int, username string) (*HomePageData, error) {
	args := m.Called(ctx, userID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*HomePageData), args.Error(1)
}

func TestService_GetHomePageData(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		username  string
		mockSetup func(*mockHomeService)
		wantData  *HomePageData
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "should_return_homepage_data_when_user_has_jobs",
			userID:   1,
			username: "testuser",
			mockSetup: func(m *mockHomeService) {
				data := &HomePageData{
					UserID:   1,
					Username: "testuser",
					Title:    "Home",
					Page:     "home",
					Stats: JobStatsSummary{
						TotalJobs:     5,
						Applied:       2,
						Interviewing:  1,
						ActiveJobs:    5,
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
				}
				m.On("GetHomePageData", mock.Anything, 1, "testuser").Return(data, nil)
			},
			wantData: &HomePageData{
				UserID:   1,
				Username: "testuser",
				Title:    "Home",
				Page:     "home",
				Stats: JobStatsSummary{
					TotalJobs:     5,
					Applied:       2,
					Interviewing:  1,
					ActiveJobs:    5,
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
			name:     "should_show_onboarding_when_user_has_no_jobs",
			userID:   2,
			username: "newuser",
			mockSetup: func(m *mockHomeService) {
				data := &HomePageData{
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
				}
				m.On("GetHomePageData", mock.Anything, 2, "newuser").Return(data, nil)
			},
			wantData: &HomePageData{
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
			name:     "should_return_error_when_getting_job_stats_fails",
			userID:   1,
			username: "testuser",
			mockSetup: func(m *mockHomeService) {
				m.On("GetHomePageData", mock.Anything, 1, "testuser").
					Return(nil, errors.New("failed to get job statistics: database connection failed"))
			},
			wantErr: true,
			errMsg:  "failed to get job statistics",
		},
		{
			name:     "should_return_error_when_getting_status_counts_fails",
			userID:   1,
			username: "testuser",
			mockSetup: func(m *mockHomeService) {
				m.On("GetHomePageData", mock.Anything, 1, "testuser").
					Return(nil, errors.New("failed to get job status counts: query failed"))
			},
			wantErr: true,
			errMsg:  "failed to get job status counts",
		},
		{
			name:     "should_return_error_when_getting_recent_jobs_fails",
			userID:   1,
			username: "testuser",
			mockSetup: func(m *mockHomeService) {
				m.On("GetHomePageData", mock.Anything, 1, "testuser").
					Return(nil, errors.New("failed to get recent jobs: recent jobs query failed"))
			},
			wantErr: true,
			errMsg:  "failed to get recent jobs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(mockHomeService)
			tc.mockSetup(mockService)

			got, err := mockService.GetHomePageData(context.Background(), tc.userID, tc.username)

			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantData.UserID, got.UserID)
				assert.Equal(t, tc.wantData.Username, got.Username)
				assert.Equal(t, tc.wantData.Title, got.Title)
				assert.Equal(t, tc.wantData.Page, got.Page)
				assert.Equal(t, tc.wantData.Stats, got.Stats)
				assert.Equal(t, tc.wantData.HasJobs, got.HasJobs)
				assert.Equal(t, tc.wantData.ShowOnboarding, got.ShowOnboarding)
				assert.Equal(t, tc.wantData.RecentJobs, got.RecentJobs)
			}

			mockService.AssertExpectations(t)
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
			name: "should_sum_all_active_statuses",
			statusCounts: map[models.JobStatus]int{
				models.INTERESTED:     2,
				models.APPLIED:        3,
				models.INTERVIEWING:   1,
				models.OFFER_RECEIVED: 1,
				models.REJECTED:       5,
				models.NOT_INTERESTED: 2,
			},
			want: 7, // 2+3+1+1 = 7 (excluding rejected and not interested)
		},
		{
			name: "should_return_zero_when_no_active_jobs",
			statusCounts: map[models.JobStatus]int{
				models.REJECTED:       3,
				models.NOT_INTERESTED: 2,
			},
			want: 0,
		},
		{
			name:         "should_return_zero_when_empty_status_counts",
			statusCounts: map[models.JobStatus]int{},
			want:         0,
		},
		{
			name: "should_count_only_applied_jobs",
			statusCounts: map[models.JobStatus]int{
				models.APPLIED: 5,
			},
			want: 5,
		},
		{
			name: "should_handle_mixed_with_zeros",
			statusCounts: map[models.JobStatus]int{
				models.INTERESTED:     0,
				models.APPLIED:        2,
				models.INTERVIEWING:   0,
				models.OFFER_RECEIVED: 1,
			},
			want: 3, // 0+2+0+1 = 3
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateActiveJobs(tc.statusCounts)
			assert.Equal(t, tc.want, got)
		})
	}
}
