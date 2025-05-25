package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/benidevo/ascentio/internal/settings/models"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// GetProfile retrieves a user's profile and its related work experiences, education, and certifications
// from the database by the given userID.
func (r *ProfileRepository) GetProfile(ctx context.Context, userID int) (*models.Profile, error) {
	query := `
		SELECT
			p.id, p.user_id, p.first_name, p.last_name, p.title, p.industry,
			p.career_summary, p.skills, p.phone_number, p.location,
			p.linkedin_profile, p.github_profile, p.website, p.context, p.created_at, p.updated_at,

			we.id, we.profile_id, we.company, we.title, we.location,
			we.start_date, we.end_date, we.description, we.current,
			we.created_at, we.updated_at,

			ed.id, ed.profile_id, ed.institution, ed.degree, ed.field_of_study,
			ed.start_date, ed.end_date, ed.description, ed.created_at, ed.updated_at,

			cert.id, cert.profile_id, cert.name, cert.issuing_org, cert.issue_date,
			cert.expiry_date, cert.credential_id, cert.credential_url,
			cert.created_at, cert.updated_at
		FROM
			profiles p
		LEFT JOIN
			work_experiences we ON p.id = we.profile_id
		LEFT JOIN
			education ed ON p.id = ed.profile_id
		LEFT JOIN
			certifications cert ON p.id = cert.profile_id
		WHERE
			p.user_id = ?
		ORDER BY
			we.start_date DESC,
			ed.start_date DESC,
			cert.issue_date DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profile *models.Profile

	// Maps to track which related entities we've already added
	workExperiencesMap := make(map[int]struct{})
	educationMap := make(map[int]struct{})
	certificationsMap := make(map[int]struct{})

	// Iterate through all rows, which will include the profile repeated
	// for each related entity
	for rows.Next() {
		// Nullable fields for the work experience, education, and certification
		var (
			// Work experience fields
			weID          sql.NullInt64
			weProfileID   sql.NullInt64
			weCompany     sql.NullString
			weTitle       sql.NullString
			weLocation    sql.NullString
			weStartDate   sql.NullTime
			weEndDate     sql.NullTime
			weDescription sql.NullString
			weCurrent     sql.NullBool
			weCreatedAt   sql.NullTime
			weUpdatedAt   sql.NullTime

			// Education fields
			edID           sql.NullInt64
			edProfileID    sql.NullInt64
			edInstitution  sql.NullString
			edDegree       sql.NullString
			edFieldOfStudy sql.NullString
			edStartDate    sql.NullTime
			edEndDate      sql.NullTime
			edDescription  sql.NullString
			edCreatedAt    sql.NullTime
			edUpdatedAt    sql.NullTime

			// Certification fields
			certID         sql.NullInt64
			certProfileID  sql.NullInt64
			certName       sql.NullString
			certIssuingOrg sql.NullString
			certIssueDate  sql.NullTime
			certExpiryDate sql.NullTime
			certCredID     sql.NullString
			certCredURL    sql.NullString
			certCreatedAt  sql.NullTime
			certUpdatedAt  sql.NullTime

			// Profile fields (will be the same in all rows)
			profileID        int
			profileUserID    int
			firstName        string
			lastName         string
			profileTitle     string
			industry         models.Industry
			careerSummary    string
			skillsJSON       []byte
			phoneNumber      string
			profileLocation  string
			linkedInProfile  string
			gitHubProfile    string
			website          string
			context          string
			profileCreatedAt time.Time
			profileUpdatedAt time.Time
		)

		err := rows.Scan(
			// Profile fields
			&profileID, &profileUserID, &firstName, &lastName, &profileTitle,
			&industry, &careerSummary, &skillsJSON, &phoneNumber, &profileLocation,
			&linkedInProfile, &gitHubProfile, &website, &context, &profileCreatedAt, &profileUpdatedAt,

			// Work experience fields
			&weID, &weProfileID, &weCompany, &weTitle, &weLocation,
			&weStartDate, &weEndDate, &weDescription, &weCurrent,
			&weCreatedAt, &weUpdatedAt,

			// Education fields
			&edID, &edProfileID, &edInstitution, &edDegree, &edFieldOfStudy,
			&edStartDate, &edEndDate, &edDescription, &edCreatedAt, &edUpdatedAt,

			// Certification fields
			&certID, &certProfileID, &certName, &certIssuingOrg, &certIssueDate,
			&certExpiryDate, &certCredID, &certCredURL, &certCreatedAt, &certUpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Initialize the profile if this is the first row
		if profile == nil {
			profile = &models.Profile{
				ID:              profileID,
				UserID:          profileUserID,
				FirstName:       firstName,
				LastName:        lastName,
				Title:           profileTitle,
				Industry:        industry,
				CareerSummary:   careerSummary,
				PhoneNumber:     phoneNumber,
				Location:        profileLocation,
				LinkedInProfile: linkedInProfile,
				GitHubProfile:   gitHubProfile,
				Website:         website,
				Context:         context,
				CreatedAt:       profileCreatedAt,
				UpdatedAt:       profileUpdatedAt,
				WorkExperience:  []models.WorkExperience{},
				Education:       []models.Education{},
				Certifications:  []models.Certification{},
			}

			if len(skillsJSON) > 0 {
				if err := json.Unmarshal(skillsJSON, &profile.Skills); err != nil {
					return nil, err
				}
			}
		}

		// Add work experience if it exists and we haven't added it yet
		if weID.Valid {
			if _, exists := workExperiencesMap[int(weID.Int64)]; !exists {
				workExperiencesMap[int(weID.Int64)] = struct{}{}

				exp := models.WorkExperience{
					ID:          int(weID.Int64),
					ProfileID:   int(weProfileID.Int64),
					Company:     weCompany.String,
					Title:       weTitle.String,
					Location:    weLocation.String,
					StartDate:   weStartDate.Time,
					Description: weDescription.String,
					Current:     weCurrent.Bool,
					CreatedAt:   weCreatedAt.Time,
					UpdatedAt:   weUpdatedAt.Time,
				}

				exp.EndDate = fromNullTime(weEndDate)

				profile.WorkExperience = append(profile.WorkExperience, exp)
			}
		}

		// Add education if it exists and we haven't added it yet
		if edID.Valid {
			if _, exists := educationMap[int(edID.Int64)]; !exists {
				educationMap[int(edID.Int64)] = struct{}{}

				edu := models.Education{
					ID:           int(edID.Int64),
					ProfileID:    int(edProfileID.Int64),
					Institution:  edInstitution.String,
					Degree:       edDegree.String,
					FieldOfStudy: edFieldOfStudy.String,
					StartDate:    edStartDate.Time,
					Description:  edDescription.String,
					CreatedAt:    edCreatedAt.Time,
					UpdatedAt:    edUpdatedAt.Time,
				}

				edu.EndDate = fromNullTime(edEndDate)

				profile.Education = append(profile.Education, edu)
			}
		}

		// Add certification if it exists and we haven't added it yet
		if certID.Valid {
			if _, exists := certificationsMap[int(certID.Int64)]; !exists {
				certificationsMap[int(certID.Int64)] = struct{}{}

				cert := models.Certification{
					ID:            int(certID.Int64),
					ProfileID:     int(certProfileID.Int64),
					Name:          certName.String,
					IssuingOrg:    certIssuingOrg.String,
					IssueDate:     certIssueDate.Time,
					CredentialID:  certCredID.String,
					CredentialURL: certCredURL.String,
					CreatedAt:     certCreatedAt.Time,
					UpdatedAt:     certUpdatedAt.Time,
				}

				cert.ExpiryDate = fromNullTime(certExpiryDate)

				profile.Certifications = append(profile.Certifications, cert)
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if profile == nil {
		return nil, nil
	}

	return profile, nil
}

// UpdateProfile inserts or updates a user's profile in the database using an upsert operation.
func (r *ProfileRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	skillsJSON, err := json.Marshal(profile.Skills)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	upsertQuery := `
		INSERT INTO profiles (
			user_id, first_name, last_name, title, industry, career_summary, skills,
			phone_number, location, linkedin_profile, github_profile, website, context, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			title = excluded.title,
			industry = excluded.industry,
			career_summary = excluded.career_summary,
			skills = excluded.skills,
			phone_number = excluded.phone_number,
			location = excluded.location,
			linkedin_profile = excluded.linkedin_profile,
			github_profile = excluded.github_profile,
			website = excluded.website,
			context = excluded.context,
			updated_at = excluded.updated_at
		RETURNING id`

	now := time.Now().UTC()
	var profileID int

	err = tx.QueryRowContext(ctx, upsertQuery,
		profile.UserID, profile.FirstName, profile.LastName, profile.Title,
		profile.Industry, profile.CareerSummary, skillsJSON, profile.PhoneNumber,
		profile.Location, profile.LinkedInProfile, profile.GitHubProfile,
		profile.Website, profile.Context, now,
	).Scan(&profileID)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	profile.ID = profileID
	profile.UpdatedAt = now

	return nil
}

// GetWorkExperiences retrieves a list of work experiences for the specified profile ID,
// ordered by start date in descending order. Returns a slice of WorkExperience models
// or an error if the query fails.
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
// Returns an error if the operation fails or if no rows are affected (certification not found).
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
