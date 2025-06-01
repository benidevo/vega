package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/benidevo/ascentio/internal/settings/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestGetProfile(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	userID := 1
	now := time.Now()
	skillsJSON, _ := json.Marshal([]string{"Go", "Python", "JavaScript"})

	t.Run("profile with all related data", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			// Profile fields
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "location",
			"linkedin_profile", "github_profile", "website", "context", "created_at", "updated_at",
			// Work experience fields
			"we_id", "we_profile_id", "we_company", "we_title", "we_location",
			"we_start_date", "we_end_date", "we_description", "we_current",
			"we_created_at", "we_updated_at",
			// Education fields
			"ed_id", "ed_profile_id", "ed_institution", "ed_degree", "ed_field_of_study",
			"ed_start_date", "ed_end_date", "ed_description", "ed_created_at", "ed_updated_at",
			// Certification fields
			"cert_id", "cert_profile_id", "cert_name", "cert_issuing_org", "cert_issue_date",
			"cert_expiry_date", "cert_credential_id", "cert_credential_url",
			"cert_created_at", "cert_updated_at",
		}).AddRow(
			// Profile data
			1, userID, "John", "Doe", "Software Engineer", models.IndustryTechnology,
			"Experienced developer", skillsJSON, "+1234567890", "New York",
			"https://linkedin.com/in/johndoe", "https://github.com/johndoe", "https://johndoe.com",
			"", now, now,
			// Work experience data
			1, 1, "Acme Corp", "Senior Developer", "Remote",
			now.Add(-365*24*time.Hour), nil, "Building great software", true,
			now, now,
			// Education data
			1, 1, "MIT", "BS Computer Science", "Computer Science",
			now.Add(-4*365*24*time.Hour), now.Add(-365*24*time.Hour), "Graduated with honors",
			now, now,
			// Certification data
			1, 1, "AWS Solutions Architect", "Amazon Web Services", now.Add(-180*24*time.Hour),
			now.Add(365*24*time.Hour), "AWS-123456", "https://aws.amazon.com/verify/AWS-123456",
			now, now,
		)

		mock.ExpectQuery("SELECT(.+)FROM(.+)profiles(.+)WHERE(.+)p.user_id = ?").
			WithArgs(userID).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "John", profile.FirstName)
		assert.Equal(t, "Doe", profile.LastName)
		assert.Len(t, profile.Skills, 3)
		assert.Len(t, profile.WorkExperience, 1)
		assert.Len(t, profile.Education, 1)
		assert.Len(t, profile.Certifications, 1)

		// Verify work experience
		assert.Equal(t, "Acme Corp", profile.WorkExperience[0].Company)
		assert.True(t, profile.WorkExperience[0].Current)

		// Verify education
		assert.Equal(t, "MIT", profile.Education[0].Institution)

		// Verify certification
		assert.Equal(t, "AWS Solutions Architect", profile.Certifications[0].Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profile without related data", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			// Profile fields
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "location",
			"linkedin_profile", "github_profile", "website", "context", "created_at", "updated_at",
			// All nullable fields for related entities
			"we_id", "we_profile_id", "we_company", "we_title", "we_location",
			"we_start_date", "we_end_date", "we_description", "we_current",
			"we_created_at", "we_updated_at",
			"ed_id", "ed_profile_id", "ed_institution", "ed_degree", "ed_field_of_study",
			"ed_start_date", "ed_end_date", "ed_description", "ed_created_at", "ed_updated_at",
			"cert_id", "cert_profile_id", "cert_name", "cert_issuing_org", "cert_issue_date",
			"cert_expiry_date", "cert_credential_id", "cert_credential_url",
			"cert_created_at", "cert_updated_at",
		}).AddRow(
			// Profile data
			2, userID, "Jane", "Smith", "", models.IndustryUnspecified,
			"", []byte("[]"), "", "",
			"", "", "", "", now, now,
			// All nulls for related data
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		)

		mock.ExpectQuery("SELECT(.+)FROM(.+)profiles(.+)WHERE(.+)p.user_id = ?").
			WithArgs(userID).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "Jane", profile.FirstName)
		assert.Empty(t, profile.WorkExperience)
		assert.Empty(t, profile.Education)
		assert.Empty(t, profile.Certifications)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("profile not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT(.+)FROM(.+)profiles(.+)WHERE(.+)p.user_id = ?").
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows(nil))

		profile, err := repo.GetProfile(ctx, 999)
		assert.NoError(t, err)
		assert.Nil(t, profile)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT(.+)FROM(.+)profiles(.+)WHERE(.+)p.user_id = ?").
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		profile, err := repo.GetProfile(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, profile)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// Note: CreateProfile is handled by UpdateProfile with upsert logic

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	profile := &models.Profile{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe Updated",
		Skills:    []string{"Go", "Python", "Kubernetes"},
	}

	t.Run("successful update", func(t *testing.T) {
		skillsJSON, _ := json.Marshal(profile.Skills)

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO profiles").
			WithArgs(
				profile.UserID, profile.FirstName, profile.LastName,
				profile.Title, profile.Industry, profile.CareerSummary,
				skillsJSON, profile.PhoneNumber, profile.Location,
				profile.LinkedInProfile, profile.GitHubProfile, profile.Website,
				profile.Context, sqlmock.AnyArg(), // context, updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.UpdateProfile(ctx, profile)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWorkExperienceOperations(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	exp := &models.WorkExperience{
		ProfileID:   1,
		Company:     "Acme Corp",
		Title:       "Senior Developer",
		Location:    "Remote",
		StartDate:   time.Now().Add(-365 * 24 * time.Hour),
		Description: "Building great software",
		Current:     true,
	}

	t.Run("create work experience", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO work_experiences").
			WithArgs(
				exp.ProfileID, exp.Company, exp.Title, exp.Location,
				exp.StartDate, nil, exp.Description, exp.Current,
				sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.AddWorkExperience(ctx, exp)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update work experience", func(t *testing.T) {
		exp.ID = 1
		mock.ExpectExec("UPDATE work_experiences SET").
			WithArgs(
				exp.Company, exp.Title, exp.Location, exp.StartDate,
				nil, exp.Description, exp.Current, sqlmock.AnyArg(), exp.ID, // nil for EndDate, AnyArg for updated_at
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := repo.UpdateWorkExperience(ctx, exp)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete work experience", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM work_experiences WHERE").
			WithArgs(10).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteWorkExperience(ctx, 10)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete non-existent work experience", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM work_experiences WHERE").
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteWorkExperience(ctx, 999)
		assert.Error(t, err)
		assert.Equal(t, models.ErrWorkExperienceNotFound, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEducationOperations(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	endDate := time.Now().Add(-180 * 24 * time.Hour)
	edu := &models.Education{
		ProfileID:    1,
		Institution:  "MIT",
		Degree:       "BS Computer Science",
		FieldOfStudy: "Computer Science",
		StartDate:    time.Now().Add(-4 * 365 * 24 * time.Hour),
		EndDate:      &endDate,
		Description:  "Graduated with honors",
	}

	t.Run("create education", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO education").
			WithArgs(
				edu.ProfileID, edu.Institution, edu.Degree, edu.FieldOfStudy,
				edu.StartDate, sqlmock.AnyArg(), edu.Description,
				sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()))

		err := repo.AddEducation(ctx, edu)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update education", func(t *testing.T) {
		edu.ID = 1
		mock.ExpectExec("UPDATE education SET").
			WithArgs(
				edu.Institution, edu.Degree, edu.FieldOfStudy,
				edu.StartDate, sqlmock.AnyArg(), edu.Description,
				sqlmock.AnyArg(), edu.ID, // updated_at, id
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := repo.UpdateEducation(ctx, edu)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete education", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM education WHERE").
			WithArgs(20).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteEducation(ctx, 20)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCertificationOperations(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	expiryDate := time.Now().Add(365 * 24 * time.Hour)
	cert := &models.Certification{
		ProfileID:     1,
		Name:          "AWS Solutions Architect",
		IssuingOrg:    "Amazon Web Services",
		IssueDate:     time.Now().Add(-180 * 24 * time.Hour),
		ExpiryDate:    &expiryDate,
		CredentialID:  "AWS-123456",
		CredentialURL: "https://aws.amazon.com/verify/AWS-123456",
	}

	t.Run("create certification", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO certifications").
			WithArgs(
				cert.ProfileID, cert.Name, cert.IssuingOrg, cert.IssueDate,
				sqlmock.AnyArg(), cert.CredentialID, cert.CredentialURL,
				sqlmock.AnyArg(), sqlmock.AnyArg(), // created_at, updated_at
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()))

		err := repo.AddCertification(ctx, cert)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update certification", func(t *testing.T) {
		cert.ID = 1
		mock.ExpectExec("UPDATE certifications SET").
			WithArgs(
				cert.Name, cert.IssuingOrg, cert.IssueDate, sqlmock.AnyArg(),
				cert.CredentialID, cert.CredentialURL, sqlmock.AnyArg(), cert.ID, // updated_at, id
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := repo.UpdateCertification(ctx, cert)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete certification", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM certifications WHERE").
			WithArgs(30).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteCertification(ctx, 30)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestHandleNullValues(t *testing.T) {
	ctx := context.Background()
	db, mock := setupMockDB(t)
	repo := NewProfileRepository(db)

	// Test with various null scenarios
	now := time.Now()

	t.Run("profile with null optional fields", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			// Profile fields
			"id", "user_id", "first_name", "last_name", "title", "industry",
			"career_summary", "skills", "phone_number", "location",
			"linkedin_profile", "github_profile", "website", "context", "created_at", "updated_at",
			// Work experience fields (all nullable)
			"we_id", "we_profile_id", "we_company", "we_title", "we_location",
			"we_start_date", "we_end_date", "we_description", "we_current",
			"we_created_at", "we_updated_at",
			// Education fields (all nullable)
			"ed_id", "ed_profile_id", "ed_institution", "ed_degree", "ed_field_of_study",
			"ed_start_date", "ed_end_date", "ed_description", "ed_created_at", "ed_updated_at",
			// Certification fields (all nullable)
			"cert_id", "cert_profile_id", "cert_name", "cert_issuing_org", "cert_issue_date",
			"cert_expiry_date", "cert_credential_id", "cert_credential_url",
			"cert_created_at", "cert_updated_at",
		}).AddRow(
			// Profile with some null fields
			1, 1, "John", "Doe", "", models.IndustryUnspecified,
			"", []byte("null"), "", "",
			"", "", "", "", now, now,
			// All nulls for related data
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		)

		mock.ExpectQuery("SELECT(.+)FROM(.+)profiles(.+)WHERE(.+)p.user_id = ?").
			WithArgs(1).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(ctx, 1)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "John", profile.FirstName)
		assert.Equal(t, "Doe", profile.LastName)
		assert.Empty(t, profile.Title)
		assert.Empty(t, profile.Skills)
		assert.Empty(t, profile.PhoneNumber)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
