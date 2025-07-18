package models

import (
	"strings"
	"time"
)

// MaxAge constants for job search preferences
const (
	MaxAgeOneHour     = 3600    // 1 hour in seconds
	MaxAgeSixHours    = 21600   // 6 hours in seconds
	MaxAgeTwelveHours = 43200   // 12 hours in seconds
	MaxAgeOneDay      = 86400   // 1 day in seconds
	MaxAgeThreeDays   = 259200  // 3 days in seconds
	MaxAgeOneWeek     = 604800  // 1 week in seconds
	MaxAgeTwoWeeks    = 1209600 // 2 weeks in seconds
	MaxAgeThirtyDays  = 2592000 // 30 days in seconds

	MinMaxAge = MaxAgeOneHour    // Minimum allowed max age
	MaxMaxAge = MaxAgeThirtyDays // Maximum allowed max age
)

// JobSearchPreference represents a user's job search criteria
type JobSearchPreference struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	JobTitle  string    `db:"job_title" json:"job_title" validate:"required,min=1,max=100"`
	Location  string    `db:"location" json:"location" validate:"required,min=1,max=100"`
	MaxAge    int       `db:"max_age" json:"max_age" validate:"required,min=3600,max=2592000"` // MinMaxAge to MaxMaxAge
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Sanitize cleans up the JobSearchPreference data
func (j *JobSearchPreference) Sanitize() {
	j.JobTitle = strings.TrimSpace(j.JobTitle)
	j.Location = strings.TrimSpace(j.Location)
}

// Validate validates the JobSearchPreference struct
func (j *JobSearchPreference) Validate() error {
	return validate.Struct(j)
}

// GetMaxAgeDisplay returns a human-readable representation of MaxAge
func (j *JobSearchPreference) GetMaxAgeDisplay() string {
	switch j.MaxAge {
	case MaxAgeOneHour:
		return "1 hour"
	case MaxAgeSixHours:
		return "6 hours"
	case MaxAgeTwelveHours:
		return "12 hours"
	case MaxAgeOneDay:
		return "1 day"
	case MaxAgeThreeDays:
		return "3 days"
	case MaxAgeOneWeek:
		return "1 week"
	case MaxAgeTwoWeeks:
		return "2 weeks"
	case MaxAgeThirtyDays:
		return "30 days"
	default:
		// For any other value, calculate and display
		duration := time.Duration(j.MaxAge) * time.Second
		return duration.String()
	}
}
