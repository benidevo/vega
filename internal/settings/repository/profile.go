package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/benidevo/vega/internal/settings/models"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// GetProfile retrieves a user's profile without related entities
func (r *ProfileRepository) GetProfile(ctx context.Context, userID int) (*models.Profile, error) {
	query := `
		SELECT id, user_id, first_name, last_name, title, industry,
		       career_summary, skills, phone_number, email, location,
		       linkedin_profile, github_profile, website, context,
		       created_at, updated_at
		FROM profiles
		WHERE user_id = ?`

	var profile models.Profile
	var skillsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName,
		&profile.Title, &profile.Industry, &profile.CareerSummary, &skillsJSON,
		&profile.PhoneNumber, &profile.Email, &profile.Location, &profile.LinkedInProfile,
		&profile.GitHubProfile, &profile.Website, &profile.Context,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal skills
	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &profile.Skills); err != nil {
			return nil, err
		}
	} else {
		profile.Skills = []string{}
	}

	// Initialize empty slices for related entities
	profile.WorkExperience = []models.WorkExperience{}
	profile.Education = []models.Education{}
	profile.Certifications = []models.Certification{}

	return &profile, nil
}

// GetProfileWithRelated retrieves a user's profile and loads related entities
// This is an optimized version that uses separate queries instead of complex JOINs
func (r *ProfileRepository) GetProfileWithRelated(ctx context.Context, userID int) (*models.Profile, error) {
	profile, err := r.GetProfile(ctx, userID)
	if err != nil || profile == nil {
		return profile, err
	}

	workExperiences, err := r.GetWorkExperiences(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	education, err := r.GetEducation(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	certifications, err := r.GetCertifications(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	profile.WorkExperience = workExperiences
	profile.Education = education
	profile.Certifications = certifications

	return profile, nil
}

// CreateProfileIfNotExists creates a profile if it doesn't exist for the user
func (r *ProfileRepository) CreateProfileIfNotExists(ctx context.Context, userID int) (*models.Profile, error) {
	// Check if profile exists
	profile, err := r.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profile != nil {
		return profile, nil
	}

	// Create minimal profile
	newProfile := &models.Profile{
		UserID:         userID,
		FirstName:      "",
		LastName:       "",
		Skills:         []string{},
		WorkExperience: []models.WorkExperience{},
		Education:      []models.Education{},
		Certifications: []models.Certification{},
	}

	skillsJSON, err := json.Marshal(newProfile.Skills)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO profiles (user_id, first_name, last_name, skills, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id`

	now := time.Now().UTC()
	err = r.db.QueryRowContext(ctx, query, userID, "", "", skillsJSON, now, now).Scan(&newProfile.ID)
	if err != nil {
		return nil, err
	}

	newProfile.CreatedAt = now
	newProfile.UpdatedAt = now

	return newProfile, nil
}

// UpdateProfile inserts or updates a user's profile using an upsert operation
func (r *ProfileRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	skillsJSON, err := json.Marshal(profile.Skills)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO profiles (
			user_id, first_name, last_name, title, industry, career_summary, skills,
			phone_number, email, location, linkedin_profile, github_profile, website, context, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			title = excluded.title,
			industry = excluded.industry,
			career_summary = excluded.career_summary,
			skills = excluded.skills,
			phone_number = excluded.phone_number,
			email = excluded.email,
			location = excluded.location,
			linkedin_profile = excluded.linkedin_profile,
			github_profile = excluded.github_profile,
			website = excluded.website,
			context = excluded.context,
			updated_at = excluded.updated_at
		RETURNING id, updated_at`

	now := time.Now().UTC()
	var profileID int
	var updatedAt time.Time

	err = r.db.QueryRowContext(ctx, query,
		profile.UserID, profile.FirstName, profile.LastName, profile.Title,
		profile.Industry, profile.CareerSummary, skillsJSON, profile.PhoneNumber,
		profile.Email, profile.Location, profile.LinkedInProfile, profile.GitHubProfile,
		profile.Website, profile.Context, now,
	).Scan(&profileID, &updatedAt)

	if err != nil {
		return err
	}

	profile.ID = profileID
	profile.UpdatedAt = updatedAt

	return nil
}

// GetEntityByID retrieves any entity by ID and type with ownership verification
func (r *ProfileRepository) GetEntityByID(ctx context.Context, entityID, profileID int, entityType string) (interface{}, error) {
	switch entityType {
	case "Experience":
		return r.getWorkExperienceByID(ctx, entityID, profileID)
	case "Education":
		return r.getEducationByID(ctx, entityID, profileID)
	case "Certification":
		return r.getCertificationByID(ctx, entityID, profileID)
	default:
		return nil, models.ErrSettingsNotFound
	}
}

// Private helper methods
func (r *ProfileRepository) getWorkExperienceByID(ctx context.Context, entityID, profileID int) (*models.WorkExperience, error) {
	query := `
		SELECT id, profile_id, company, title, location, start_date, end_date,
		       description, current, created_at, updated_at
		FROM work_experiences
		WHERE id = ? AND profile_id = ?`

	var exp models.WorkExperience
	var endDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, entityID, profileID).Scan(
		&exp.ID, &exp.ProfileID, &exp.Company, &exp.Title, &exp.Location,
		&exp.StartDate, &endDate, &exp.Description, &exp.Current,
		&exp.CreatedAt, &exp.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrWorkExperienceNotFound
		}
		return nil, err
	}

	exp.EndDate = fromNullTime(endDate)
	return &exp, nil
}

func (r *ProfileRepository) getEducationByID(ctx context.Context, entityID, profileID int) (*models.Education, error) {
	query := `
		SELECT id, profile_id, institution, degree, field_of_study,
		       start_date, end_date, description, created_at, updated_at
		FROM education
		WHERE id = ? AND profile_id = ?`

	var edu models.Education
	var endDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, entityID, profileID).Scan(
		&edu.ID, &edu.ProfileID, &edu.Institution, &edu.Degree, &edu.FieldOfStudy,
		&edu.StartDate, &endDate, &edu.Description, &edu.CreatedAt, &edu.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrEducationNotFound
		}
		return nil, err
	}

	edu.EndDate = fromNullTime(endDate)
	return &edu, nil
}

func (r *ProfileRepository) getCertificationByID(ctx context.Context, entityID, profileID int) (*models.Certification, error) {
	query := `
		SELECT id, profile_id, name, issuing_org, issue_date, expiry_date,
		       credential_id, credential_url, created_at, updated_at
		FROM certifications
		WHERE id = ? AND profile_id = ?`

	var cert models.Certification
	var expiryDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, entityID, profileID).Scan(
		&cert.ID, &cert.ProfileID, &cert.Name, &cert.IssuingOrg, &cert.IssueDate,
		&expiryDate, &cert.CredentialID, &cert.CredentialURL, &cert.CreatedAt, &cert.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrCertificationNotFound
		}
		return nil, err
	}

	cert.ExpiryDate = fromNullTime(expiryDate)
	return &cert, nil
}

// Continue with all the existing CRUD methods for WorkExperience, Education, and Certification...

// GetWorkExperiences retrieves a list of work experiences for the specified profile ID,
// ordered by start date in descending order
func (r *ProfileRepository) GetWorkExperiences(ctx context.Context, profileID int) ([]models.WorkExperience, error) {
	query := `
		SELECT id, profile_id, company, title, location, start_date, end_date,
		       description, current, created_at, updated_at
		FROM work_experiences
		WHERE profile_id = ?
		ORDER BY start_date DESC`

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []models.WorkExperience
	for rows.Next() {
		var exp models.WorkExperience
		var endDate sql.NullTime

		err := rows.Scan(
			&exp.ID,
			&exp.ProfileID,
			&exp.Company,
			&exp.Title,
			&exp.Location,
			&exp.StartDate,
			&endDate,
			&exp.Description,
			&exp.Current,
			&exp.CreatedAt,
			&exp.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		exp.EndDate = fromNullTime(endDate)

		experiences = append(experiences, exp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return experiences, nil
}

// AddWorkExperience inserts a new WorkExperience record into the database for the given profile.
func (r *ProfileRepository) AddWorkExperience(ctx context.Context, experience *models.WorkExperience) error {
	query := `
		INSERT INTO work_experiences (
			profile_id, company, title, location, start_date, end_date,
			description, current, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	endDate := toNullTime(experience.EndDate)
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		experience.ProfileID, experience.Company, experience.Title,
		experience.Location, experience.StartDate, endDate,
		experience.Description, experience.Current, now, now,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateWorkExperience updates an existing work experience record in the database.
func (r *ProfileRepository) UpdateWorkExperience(ctx context.Context, experience *models.WorkExperience) (*models.WorkExperience, error) {
	query := `
		UPDATE work_experiences
		SET company = ?, title = ?, location = ?, start_date = ?, end_date = ?,
		    description = ?, current = ?, updated_at = ?
		WHERE id = ?`

	endDate := toNullTime(experience.EndDate)
	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		experience.Company, experience.Title, experience.Location,
		experience.StartDate, endDate, experience.Description,
		experience.Current, now, experience.ID,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, models.ErrWorkExperienceNotFound
	}

	experience.UpdatedAt = now
	return experience, nil
}

// DeleteWorkExperience deletes a work experience entry by its ID.
func (r *ProfileRepository) DeleteWorkExperience(ctx context.Context, id int) error {
	query := "DELETE FROM work_experiences WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrWorkExperienceNotFound
	}

	return nil
}

// GetEducation retrieves a list of education entries for the specified profile ID.
func (r *ProfileRepository) GetEducation(ctx context.Context, profileID int) ([]models.Education, error) {
	query := `
		SELECT id, profile_id, institution, degree, field_of_study,
		       start_date, end_date, description, created_at, updated_at
		FROM education
		WHERE profile_id = ?
		ORDER BY start_date DESC`

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var educationEntries []models.Education
	for rows.Next() {
		var edu models.Education
		var endDate sql.NullTime

		err := rows.Scan(
			&edu.ID,
			&edu.ProfileID,
			&edu.Institution,
			&edu.Degree,
			&edu.FieldOfStudy,
			&edu.StartDate,
			&endDate,
			&edu.Description,
			&edu.CreatedAt,
			&edu.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		edu.EndDate = fromNullTime(endDate)

		educationEntries = append(educationEntries, edu)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return educationEntries, nil
}

// AddEducation inserts a new education record into the database for the given profile.
func (r *ProfileRepository) AddEducation(ctx context.Context, education *models.Education) error {
	query := `
		INSERT INTO education (
			profile_id, institution, degree, field_of_study,
			start_date, end_date, description, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	var endDate sql.NullTime
	if education.EndDate != nil {
		endDate.Time = *education.EndDate
		endDate.Valid = true
	}

	now := time.Now().UTC()

	return r.db.QueryRowContext(ctx, query,
		education.ProfileID, education.Institution, education.Degree,
		education.FieldOfStudy, education.StartDate, endDate,
		education.Description, now, now,
	).Scan(&education.ID, &education.CreatedAt, &education.UpdatedAt)
}

// UpdateEducation updates an existing education record in the database with the provided information.
func (r *ProfileRepository) UpdateEducation(ctx context.Context, education *models.Education) (*models.Education, error) {
	query := `
		UPDATE education
		SET institution = ?, degree = ?, field_of_study = ?, start_date = ?,
		    end_date = ?, description = ?, updated_at = ?
		WHERE id = ?`

	var endDate sql.NullTime
	if education.EndDate != nil {
		endDate.Time = *education.EndDate
		endDate.Valid = true
	}

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		education.Institution, education.Degree, education.FieldOfStudy,
		education.StartDate, endDate, education.Description, now, education.ID,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, models.ErrEducationNotFound
	}

	education.UpdatedAt = now
	return education, nil
}

// DeleteEducation deletes an education record by its ID from the database.
func (r *ProfileRepository) DeleteEducation(ctx context.Context, id int) error {
	query := "DELETE FROM education WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrEducationNotFound
	}

	return nil
}

// GetCertifications retrieves all certifications associated with the given profile ID,
// ordered by issue date in descending order.
func (r *ProfileRepository) GetCertifications(ctx context.Context, profileID int) ([]models.Certification, error) {
	query := `
		SELECT id, profile_id, name, issuing_org, issue_date, expiry_date,
		       credential_id, credential_url, created_at, updated_at
		FROM certifications
		WHERE profile_id = ?
		ORDER BY issue_date DESC`

	rows, err := r.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certifications []models.Certification
	for rows.Next() {
		var cert models.Certification
		var expiryDate sql.NullTime

		err := rows.Scan(
			&cert.ID,
			&cert.ProfileID,
			&cert.Name,
			&cert.IssuingOrg,
			&cert.IssueDate,
			&expiryDate,
			&cert.CredentialID,
			&cert.CredentialURL,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		cert.ExpiryDate = fromNullTime(expiryDate)

		certifications = append(certifications, cert)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return certifications, nil
}

// AddCertification inserts a new certification record into the database for the given profile.
func (r *ProfileRepository) AddCertification(ctx context.Context, certification *models.Certification) error {
	query := `
		INSERT INTO certifications (
			profile_id, name, issuing_org, issue_date, expiry_date,
			credential_id, credential_url, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	var expiryDate sql.NullTime
	if certification.ExpiryDate != nil {
		expiryDate.Time = *certification.ExpiryDate
		expiryDate.Valid = true
	}

	now := time.Now().UTC()

	return r.db.QueryRowContext(ctx, query,
		certification.ProfileID, certification.Name, certification.IssuingOrg,
		certification.IssueDate, expiryDate, certification.CredentialID,
		certification.CredentialURL, now, now,
	).Scan(&certification.ID, &certification.CreatedAt, &certification.UpdatedAt)
}

// UpdateCertification updates an existing certification record in the database with the provided certification details.
func (r *ProfileRepository) UpdateCertification(ctx context.Context, certification *models.Certification) (*models.Certification, error) {
	query := `
		UPDATE certifications
		SET name = ?, issuing_org = ?, issue_date = ?, expiry_date = ?,
		    credential_id = ?, credential_url = ?, updated_at = ?
		WHERE id = ?`

	var expiryDate sql.NullTime
	if certification.ExpiryDate != nil {
		expiryDate.Time = *certification.ExpiryDate
		expiryDate.Valid = true
	}

	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		certification.Name, certification.IssuingOrg, certification.IssueDate,
		expiryDate, certification.CredentialID, certification.CredentialURL,
		now, certification.ID,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, models.ErrCertificationNotFound
	}

	certification.UpdatedAt = now
	return certification, nil
}

// DeleteCertification deletes a certification record by its ID from the database.
func (r *ProfileRepository) DeleteCertification(ctx context.Context, id int) error {
	query := "DELETE FROM certifications WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return models.ErrCertificationNotFound
	}

	return nil
}

// DeleteAllWorkExperience deletes all work experience entries for a profile
func (r *ProfileRepository) DeleteAllWorkExperience(ctx context.Context, profileID int) error {
	query := "DELETE FROM work_experiences WHERE profile_id = ?"

	_, err := r.db.ExecContext(ctx, query, profileID)
	return err
}

// DeleteAllEducation deletes all education entries for a profile
func (r *ProfileRepository) DeleteAllEducation(ctx context.Context, profileID int) error {
	query := "DELETE FROM education WHERE profile_id = ?"

	_, err := r.db.ExecContext(ctx, query, profileID)
	return err
}
