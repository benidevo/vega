package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/benidevo/vega/internal/settings/models"
	"github.com/google/uuid"
)

// CreateJobSearchPreference creates a new job search preference
func (r *ProfileRepository) CreateJobSearchPreference(ctx context.Context, userID int, preference *models.JobSearchPreference) error {
	preference.ID = uuid.New().String()
	preference.UserID = userID
	preference.CreatedAt = time.Now()
	preference.UpdatedAt = time.Now()

	// Convert skills to JSON
	var skillsJSON []byte
	var err error
	if len(preference.Skills) > 0 {
		skillsJSON, err = json.Marshal(preference.Skills)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO job_search_preferences (
			id, user_id, job_title, location, skills, max_age, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		preference.ID, preference.UserID, preference.JobTitle, preference.Location,
		skillsJSON, preference.MaxAge, preference.IsActive, preference.CreatedAt, preference.UpdatedAt,
	)
	return err
}

// GetJobSearchPreferenceByID retrieves a single job search preference by ID
func (r *ProfileRepository) GetJobSearchPreferenceByID(ctx context.Context, userID int, preferenceID string) (*models.JobSearchPreference, error) {
	query := `
		SELECT id, user_id, job_title, location, skills, max_age, is_active, created_at, updated_at
		FROM job_search_preferences
		WHERE id = ? AND user_id = ?`

	var preference models.JobSearchPreference
	var skillsJSON sql.NullString

	err := r.db.QueryRowContext(ctx, query, preferenceID, userID).Scan(
		&preference.ID, &preference.UserID, &preference.JobTitle, &preference.Location,
		&skillsJSON, &preference.MaxAge, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse skills JSON
	if skillsJSON.Valid && skillsJSON.String != "" {
		if err := json.Unmarshal([]byte(skillsJSON.String), &preference.Skills); err != nil {
			return nil, err
		}
	}

	return &preference, nil
}

// GetJobSearchPreferencesByUserID retrieves all job search preferences for a user
func (r *ProfileRepository) GetJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*models.JobSearchPreference, error) {
	query := `
		SELECT id, user_id, job_title, location, skills, max_age, is_active, created_at, updated_at
		FROM job_search_preferences
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*models.JobSearchPreference
	for rows.Next() {
		var preference models.JobSearchPreference
		var skillsJSON sql.NullString

		err := rows.Scan(
			&preference.ID, &preference.UserID, &preference.JobTitle, &preference.Location,
			&skillsJSON, &preference.MaxAge, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse skills JSON
		if skillsJSON.Valid && skillsJSON.String != "" {
			if err := json.Unmarshal([]byte(skillsJSON.String), &preference.Skills); err != nil {
				return nil, err
			}
		}

		preferences = append(preferences, &preference)
	}

	return preferences, rows.Err()
}

// GetActiveJobSearchPreferencesByUserID retrieves only active job search preferences for a user
func (r *ProfileRepository) GetActiveJobSearchPreferencesByUserID(ctx context.Context, userID int) ([]*models.JobSearchPreference, error) {
	query := `
		SELECT id, user_id, job_title, location, skills, max_age, is_active, created_at, updated_at
		FROM job_search_preferences
		WHERE user_id = ? AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*models.JobSearchPreference
	for rows.Next() {
		var preference models.JobSearchPreference
		var skillsJSON sql.NullString

		err := rows.Scan(
			&preference.ID, &preference.UserID, &preference.JobTitle, &preference.Location,
			&skillsJSON, &preference.MaxAge, &preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse skills JSON
		if skillsJSON.Valid && skillsJSON.String != "" {
			if err := json.Unmarshal([]byte(skillsJSON.String), &preference.Skills); err != nil {
				return nil, err
			}
		}

		preferences = append(preferences, &preference)
	}

	return preferences, rows.Err()
}

// UpdateJobSearchPreference updates an existing job search preference
func (r *ProfileRepository) UpdateJobSearchPreference(ctx context.Context, userID int, preference *models.JobSearchPreference) error {
	preference.UpdatedAt = time.Now()

	// Convert skills to JSON
	var skillsJSON []byte
	var err error
	if len(preference.Skills) > 0 {
		skillsJSON, err = json.Marshal(preference.Skills)
		if err != nil {
			return err
		}
	}

	query := `
		UPDATE job_search_preferences
		SET job_title = ?, location = ?, skills = ?, max_age = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query,
		preference.JobTitle, preference.Location, skillsJSON, preference.MaxAge, preference.IsActive,
		preference.UpdatedAt, preference.ID, userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteJobSearchPreference deletes a job search preference
func (r *ProfileRepository) DeleteJobSearchPreference(ctx context.Context, userID int, preferenceID string) error {
	query := `DELETE FROM job_search_preferences WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, preferenceID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ToggleJobSearchPreferenceActive toggles the active status of a job search preference
func (r *ProfileRepository) ToggleJobSearchPreferenceActive(ctx context.Context, userID int, preferenceID string) error {
	query := `
		UPDATE job_search_preferences
		SET is_active = NOT is_active, updated_at = ?
		WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, time.Now(), preferenceID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
