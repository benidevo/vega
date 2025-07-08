package badger

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
	"github.com/benidevo/vega/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCache(t *testing.T) (*BadgerCache, func()) {
	tmpDir, err := os.MkdirTemp("", "badger-test-*")
	require.NoError(t, err)

	cache, err := NewBadgerCache(tmpDir)
	require.NoError(t, err)

	cleanup := func() {
		cache.Close()
		os.RemoveAll(tmpDir)
	}

	return cache, cleanup
}

func TestBadgerCache_Initialize(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"

	err := cache.Initialize(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, userID, cache.userID)
	assert.Equal(t, userID, cache.metadata.UserID)
	assert.False(t, cache.metadata.IsDirty)
}

func TestBadgerCache_Profile(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Test profile not found
	_, err = cache.GetProfile(ctx)
	assert.Equal(t, storage.ErrProfileNotFound, err)

	// Test save profile
	profile := &settingsmodels.Profile{
		ID:          1,
		UserID:      1,
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "123-456-7890",
	}

	err = cache.SaveProfile(ctx, profile)
	assert.NoError(t, err)
	assert.True(t, cache.metadata.IsDirty)

	// Test get profile
	retrieved, err := cache.GetProfile(ctx)
	assert.NoError(t, err)
	assert.Equal(t, profile.FirstName, retrieved.FirstName)
	assert.Equal(t, profile.LastName, retrieved.LastName)
	assert.Equal(t, profile.Email, retrieved.Email)
}

func TestBadgerCache_Companies(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Test empty list
	companies, err := cache.ListCompanies(ctx)
	assert.NoError(t, err)
	assert.Len(t, companies, 0)

	// Test save company
	company := &jobmodels.Company{
		ID:   1,
		Name: "Test Corp",
	}

	err = cache.SaveCompany(ctx, company)
	assert.NoError(t, err)

	// Test list companies
	companies, err = cache.ListCompanies(ctx)
	assert.NoError(t, err)
	assert.Len(t, companies, 1)
	assert.Equal(t, company.Name, companies[0].Name)

	// Test get company
	retrieved, err := cache.GetCompany(ctx, company.ID)
	assert.NoError(t, err)
	assert.Equal(t, company.Name, retrieved.Name)

	// Test update company
	company.Name = "Updated Corp"
	err = cache.SaveCompany(ctx, company)
	assert.NoError(t, err)

	retrieved, err = cache.GetCompany(ctx, company.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Corp", retrieved.Name)

	// Test delete company
	err = cache.DeleteCompany(ctx, company.ID)
	assert.NoError(t, err)

	companies, err = cache.ListCompanies(ctx)
	assert.NoError(t, err)
	assert.Len(t, companies, 0)

	// Test company not found
	_, err = cache.GetCompany(ctx, 999)
	assert.Equal(t, storage.ErrCompanyNotFound, err)
}

func TestBadgerCache_Jobs(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Setup company first
	company := &jobmodels.Company{
		ID:   1,
		Name: "Test Corp",
	}
	err = cache.SaveCompany(ctx, company)
	require.NoError(t, err)

	// Test empty job list
	jobs, err := cache.ListJobs(ctx, company.ID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	// Test save job
	job := &jobmodels.Job{
		ID:          1,
		CompanyID:   company.ID,
		Title:       "Software Engineer",
		Description: "A great job",
		CreatedAt:   time.Now(),
	}

	err = cache.SaveJob(ctx, job)
	assert.NoError(t, err)

	// Test list jobs
	jobs, err = cache.ListJobs(ctx, company.ID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, job.Title, jobs[0].Title)

	// Test get job
	retrieved, err := cache.GetJob(ctx, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job.Title, retrieved.Title)

	// Test update job
	job.Title = "Senior Software Engineer"
	err = cache.SaveJob(ctx, job)
	assert.NoError(t, err)

	retrieved, err = cache.GetJob(ctx, job.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Senior Software Engineer", retrieved.Title)

	// Test delete job
	err = cache.DeleteJob(ctx, job.ID)
	assert.NoError(t, err)

	jobs, err = cache.ListJobs(ctx, company.ID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)

	// Test job not found
	_, err = cache.GetJob(ctx, 999)
	assert.Equal(t, storage.ErrJobNotFound, err)
}

func TestBadgerCache_Matches(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Test no matches
	_, err = cache.GetMatchHistory(ctx, 0)
	assert.Equal(t, storage.ErrNoMatches, err)

	// Test save match
	match := &jobmodels.MatchResult{
		ID:         1,
		JobID:      1,
		MatchScore: 85,
		CreatedAt:  time.Now(),
		Strengths:  []string{"Python", "Go"},
		Weaknesses: []string{"Kubernetes"},
	}

	err = cache.SaveMatchResult(ctx, match)
	assert.NoError(t, err)

	// Test get match history
	matches, err := cache.GetMatchHistory(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, match.MatchScore, matches[0].MatchScore)

	// Test get specific match
	retrieved, err := cache.GetMatchResult(ctx, match.ID)
	assert.NoError(t, err)
	assert.Equal(t, match.MatchScore, retrieved.MatchScore)

	// Test limit in match history
	for i := 2; i <= 5; i++ {
		newMatch := &jobmodels.MatchResult{
			ID:         i,
			JobID:      i,
			MatchScore: 80 + i,
			CreatedAt:  time.Now(),
		}
		err = cache.SaveMatchResult(ctx, newMatch)
		require.NoError(t, err)
	}

	matches, err = cache.GetMatchHistory(ctx, 3)
	assert.NoError(t, err)
	assert.Len(t, matches, 3)

	// Test match not found
	_, err = cache.GetMatchResult(ctx, 999)
	assert.Equal(t, storage.ErrMatchNotFound, err)
}

func TestBadgerCache_Sync(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Initially not dirty
	assert.False(t, cache.metadata.IsDirty)

	// Save something to make it dirty
	profile := &settingsmodels.Profile{
		ID:        1,
		UserID:    1,
		FirstName: "Test",
		LastName:  "User",
	}
	err = cache.SaveProfile(ctx, profile)
	require.NoError(t, err)
	assert.True(t, cache.metadata.IsDirty)

	// Test sync
	lastSync := cache.GetLastSyncTime()
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	err = cache.Sync(ctx)
	assert.NoError(t, err)
	assert.False(t, cache.metadata.IsDirty)
	assert.True(t, cache.GetLastSyncTime().After(lastSync))
}

func TestBadgerCache_CascadeDelete(t *testing.T) {
	cache, cleanup := setupTestCache(t)
	defer cleanup()

	ctx := context.Background()
	userID := "test-user-123"
	err := cache.Initialize(ctx, userID)
	require.NoError(t, err)

	// Setup company and jobs
	company := &jobmodels.Company{
		ID:   1,
		Name: "Test Corp",
	}
	err = cache.SaveCompany(ctx, company)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		job := &jobmodels.Job{
			ID:          i,
			CompanyID:   company.ID,
			Title:       fmt.Sprintf("Job %d", i),
			Description: "Test job description",
		}
		err = cache.SaveJob(ctx, job)
		require.NoError(t, err)
	}

	// Verify jobs exist
	jobs, err := cache.ListJobs(ctx, company.ID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 3)

	// Delete company
	err = cache.DeleteCompany(ctx, company.ID)
	assert.NoError(t, err)

	// Verify jobs are also deleted
	jobs, err = cache.ListJobs(ctx, company.ID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 0)
}
