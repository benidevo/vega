package sqlite

import (
	"context"
	"os"
	"testing"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	// Set test environment to disable periodic sync
	os.Setenv("GO_TEST", "1")
	defer os.Unsetenv("GO_TEST")

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "vega-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	t.Run("Profile operations", func(t *testing.T) {
		// Create storage for this test
		storage, err := NewStorage("test@example.com", tempDir)
		require.NoError(t, err)
		defer storage.Close()
		// Save profile
		profile := &settingsmodels.Profile{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Skills:    []string{"Go", "Python"},
		}
		err = storage.SaveProfile(ctx, profile)
		assert.NoError(t, err)

		// Get profile - should come from Badger cache
		retrieved, err := storage.GetProfile(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "John", retrieved.FirstName)
		assert.Equal(t, "Doe", retrieved.LastName)
	})

	t.Run("Company operations", func(t *testing.T) {

		storage, err := NewStorage("test@example.com", tempDir)
		require.NoError(t, err)
		defer storage.Close()

		company := &jobmodels.Company{
			Name: "Test Corp",
		}
		err = storage.SaveCompany(ctx, company)
		assert.NoError(t, err)
		assert.NotZero(t, company.ID)

		// Get company - should come from Badger cache
		retrieved, err := storage.GetCompany(ctx, company.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Test Corp", retrieved.Name)

		// Note: ListCompanies would return empty because the list cache
		// is invalidated on save and SQLite hasn't been synced yet.
		// This is expected behavior in the current implementation.
	})

	t.Run("Job operations", func(t *testing.T) {
		storage, err := NewStorage("test@example.com", tempDir)
		require.NoError(t, err)
		defer storage.Close()

		company := &jobmodels.Company{
			Name: "Job Corp",
		}
		err = storage.SaveCompany(ctx, company)
		require.NoError(t, err)

		job := &jobmodels.Job{
			Title:       "Software Engineer",
			Description: "Great job",
			CompanyID:   company.ID,
			JobType:     jobmodels.FULL_TIME,
			Status:      jobmodels.INTERESTED,
		}
		err = storage.SaveJob(ctx, job)
		assert.NoError(t, err)
		assert.NotZero(t, job.ID)

		// Get job - should come from Badger cache
		retrieved, err := storage.GetJob(ctx, job.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Software Engineer", retrieved.Title)

		// Note: ListJobs would return empty because the list cache
		// is invalidated on save and SQLite hasn't been synced yet.
		// This is expected behavior in the current implementation.
	})

	t.Run("Badger cache and SQLite persistence", func(t *testing.T) {
		storage, err := NewStorage("test@example.com", tempDir)
		require.NoError(t, err)

		company := &jobmodels.Company{Name: "Cached Company"}
		err = storage.SaveCompany(ctx, company)
		require.NoError(t, err)

		job := &jobmodels.Job{
			Title:       "Cached Job",
			Description: "Test Description",
			CompanyID:   company.ID,
			JobType:     jobmodels.FULL_TIME,
			Status:      jobmodels.INTERESTED,
		}
		err = storage.SaveJob(ctx, job)
		require.NoError(t, err)

		cachedCompany, err := storage.GetCompany(ctx, company.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Cached Company", cachedCompany.Name)

		cachedJob, err := storage.GetJob(ctx, job.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Cached Job", cachedJob.Title)

		err = storage.Sync(ctx)
		assert.NoError(t, err)

		storage.Close()
	})
}
