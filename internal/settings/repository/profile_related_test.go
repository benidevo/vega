package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/vega/internal/settings/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfileWithRelated(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	userID := 1
	profileID := 1
	now := time.Now()
	skillsJSON, _ := json.Marshal([]string{"Go", "Python"})

	t.Run("profile with all related entities", func(t *testing.T) {
		// Profile query
		profileRows := sqlmock.NewRows([]string{
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "email", "location",
			"linkedin_profile", "github_profile", "website", "context",
			"created_at", "updated_at",
		}).AddRow(
			profileID, userID, "John", "Doe", "Software Engineer", models.IndustryTechnology,
			"Experienced developer", skillsJSON, "+1234567890", "john@example.com", "New York",
			"https://linkedin.com/in/johndoe", "https://github.com/johndoe", "https://johndoe.com",
			"Context about John", now, now,
		)

		// Work experience query
		expRows := sqlmock.NewRows([]string{
			"id", "profile_id", "company", "title", "location", "start_date",
			"end_date", "description", "current", "created_at", "updated_at",
		}).AddRow(
			1, profileID, "Acme Corp", "Senior Engineer", "Remote", now.Add(-365*24*time.Hour),
			nil, "Building software", true, now, now,
		)

		// Education query
		eduRows := sqlmock.NewRows([]string{
			"id", "profile_id", "institution", "degree", "field_of_study",
			"start_date", "end_date", "description", "created_at", "updated_at",
		}).AddRow(
			1, profileID, "MIT", "BS Computer Science", "Computer Science",
			now.Add(-4*365*24*time.Hour), now.Add(-2*365*24*time.Hour), "Graduated with honors", now, now,
		)

		// Certifications query
		certRows := sqlmock.NewRows([]string{
			"id", "profile_id", "name", "issuing_org", "issue_date",
			"expiry_date", "credential_id", "credential_url", "created_at", "updated_at",
		}).AddRow(
			1, profileID, "AWS Solutions Architect", "Amazon Web Services", now.Add(-180*24*time.Hour),
			now.Add(185*24*time.Hour), "AWS-123456", "https://aws.amazon.com/verify/AWS-123456", now, now,
		)

		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(profileRows)

		mock.ExpectQuery("SELECT(.+)FROM work_experiences(.+)WHERE profile_id = \\?").
			WithArgs(profileID).
			WillReturnRows(expRows)

		mock.ExpectQuery("SELECT(.+)FROM education(.+)WHERE profile_id = \\?").
			WithArgs(profileID).
			WillReturnRows(eduRows)

		mock.ExpectQuery("SELECT(.+)FROM certifications(.+)WHERE profile_id = \\?").
			WithArgs(profileID).
			WillReturnRows(certRows)

		profile, err := repo.GetProfileWithRelated(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "John", profile.FirstName)
		assert.Len(t, profile.WorkExperience, 1)
		assert.Len(t, profile.Education, 1)
		assert.Len(t, profile.Certifications, 1)

		// Verify work experience data
		assert.Equal(t, "Acme Corp", profile.WorkExperience[0].Company)
		assert.True(t, profile.WorkExperience[0].Current)

		// Verify education data
		assert.Equal(t, "MIT", profile.Education[0].Institution)
		assert.Equal(t, "BS Computer Science", profile.Education[0].Degree)

		// Verify certification data
		assert.Equal(t, "AWS Solutions Architect", profile.Certifications[0].Name)
		assert.Equal(t, "AWS-123456", profile.Certifications[0].CredentialID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profile not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows(nil))

		profile, err := repo.GetProfileWithRelated(ctx, 999)
		assert.NoError(t, err)
		assert.Nil(t, profile)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profile exists but related entities queries fail", func(t *testing.T) {
		profileRows := sqlmock.NewRows([]string{
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "email", "location",
			"linkedin_profile", "github_profile", "website", "context",
			"created_at", "updated_at",
		}).AddRow(
			profileID, userID, "John", "Doe", "Software Engineer", models.IndustryTechnology,
			"Experienced developer", skillsJSON, "+1234567890", "john@example.com", "New York",
			"", "", "", "", now, now,
		)

		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(profileRows)

		mock.ExpectQuery("SELECT(.+)FROM work_experiences(.+)WHERE profile_id = \\?").
			WithArgs(profileID).
			WillReturnError(errors.New("work experience query failed"))

		profile, err := repo.GetProfileWithRelated(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "work experience query failed")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestCreateProfileIfNotExists tests the critical profile creation logic
func TestCreateProfileIfNotExists(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	userID := 1
	now := time.Now()

	t.Run("profile does not exist: creates new", func(t *testing.T) {
		// First query returns no profile
		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows(nil))

		// Insert new profile
		mock.ExpectQuery("INSERT INTO profiles \\(user_id, first_name, last_name, skills, created_at, updated_at\\) VALUES \\(\\?, \\?, \\?, \\?, \\?, \\?\\) RETURNING id").
			WithArgs(userID, "", "", []byte("[]"), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		profile, err := repo.CreateProfileIfNotExists(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, userID, profile.UserID)
		assert.Equal(t, 1, profile.ID)
		assert.Empty(t, profile.Skills)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profile already exists: returns existing", func(t *testing.T) {
		skillsJSON, _ := json.Marshal([]string{"Go", "Python"})
		existingRows := sqlmock.NewRows([]string{
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "email", "location",
			"linkedin_profile", "github_profile", "website", "context",
			"created_at", "updated_at",
		}).AddRow(
			1, userID, "John", "Doe", "Software Engineer", models.IndustryTechnology,
			"Experienced developer", skillsJSON, "+1234567890", "john@example.com", "New York",
			"", "", "", "", now, now,
		)

		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(existingRows)

		profile, err := repo.CreateProfileIfNotExists(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "John", profile.FirstName)
		assert.Equal(t, "Doe", profile.LastName)
		assert.Len(t, profile.Skills, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on profile check", func(t *testing.T) {
		dbErr := errors.New("database connection failed")
		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnError(dbErr)

		profile, err := repo.CreateProfileIfNotExists(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "database connection failed")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on profile creation", func(t *testing.T) {
		// Profile doesn't exist
		mock.ExpectQuery("SELECT(.+)FROM profiles(.+)WHERE user_id = \\?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows(nil))

		// Insert fails
		insertErr := errors.New("insert failed")
		mock.ExpectQuery("INSERT INTO profiles \\(user_id, first_name, last_name, skills, created_at, updated_at\\) VALUES \\(\\?, \\?, \\?, \\?, \\?, \\?\\) RETURNING id").
			WithArgs(userID, "", "", []byte("[]"), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(insertErr)

		profile, err := repo.CreateProfileIfNotExists(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "insert failed")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
