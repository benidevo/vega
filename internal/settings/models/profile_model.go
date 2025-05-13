package models

import (
	"time"
)

// ProfileSettings represents user profile information
type ProfileSettings struct {
	ID              int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	UserID          int       `json:"user_id" db:"user_id" sql:"type:integer;not null;unique;index;references:users(id)"`
	FirstName       string    `json:"first_name" db:"first_name" sql:"type:text"`
	LastName        string    `json:"last_name" db:"last_name" sql:"type:text"`
	Title           string    `json:"title" db:"title" sql:"type:text"`
	Industry        string    `json:"industry" db:"industry" sql:"type:text"`
	CareerSummary   string    `json:"career_summary" db:"career_summary" sql:"type:text"`
	Skills          []string  `json:"skills" db:"skills" sql:"type:text"` // Stored as JSON
	PhoneNumber     string    `json:"phone_number" db:"phone_number" sql:"type:text"`
	Location        string    `json:"location" db:"location" sql:"type:text"`
	LinkedInProfile string    `json:"linkedin_profile" db:"linkedin_profile" sql:"type:text"`
	GitHubProfile   string    `json:"github_profile" db:"github_profile" sql:"type:text"`
	Website         string    `json:"website" db:"website" sql:"type:text"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// WorkExperience represents work experience entries in a user's profile
type WorkExperience struct {
	ID          int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID   int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profile_settings(id)"`
	Company     string     `json:"company" db:"company" sql:"type:text;not null"`
	Title       string     `json:"title" db:"title" sql:"type:text;not null"`
	Location    string     `json:"location" db:"location" sql:"type:text"`
	StartDate   time.Time  `json:"start_date" db:"start_date" sql:"type:timestamp;not null"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date" sql:"type:timestamp"`
	Description string     `json:"description" db:"description" sql:"type:text"`
	Current     bool       `json:"current" db:"current" sql:"type:boolean;default:false"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Education represents education entries in a user's profile
type Education struct {
	ID           int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID    int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profile_settings(id)"`
	Institution  string     `json:"institution" db:"institution" sql:"type:text;not null"`
	Degree       string     `json:"degree" db:"degree" sql:"type:text;not null"`
	FieldOfStudy string     `json:"field_of_study" db:"field_of_study" sql:"type:text"`
	StartDate    time.Time  `json:"start_date" db:"start_date" sql:"type:timestamp;not null"`
	EndDate      *time.Time `json:"end_date,omitempty" db:"end_date" sql:"type:timestamp"`
	Description  string     `json:"description" db:"description" sql:"type:text"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Certification represents certification entries in a user's profile
type Certification struct {
	ID            int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID     int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profile_settings(id)"`
	Name          string     `json:"name" db:"name" sql:"type:text;not null"`
	IssuingOrg    string     `json:"issuing_org" db:"issuing_org" sql:"type:text;not null"`
	IssueDate     time.Time  `json:"issue_date" db:"issue_date" sql:"type:timestamp;not null"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty" db:"expiry_date" sql:"type:timestamp"`
	CredentialID  string     `json:"credential_id" db:"credential_id" sql:"type:text"`
	CredentialURL string     `json:"credential_url" db:"credential_url" sql:"type:text"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Award represents awards and recognitions in a user's profile
type Award struct {
	ID          int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID   int       `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profile_settings(id)"`
	Title       string    `json:"title" db:"title" sql:"type:text;not null"`
	Issuer      string    `json:"issuer" db:"issuer" sql:"type:text;not null"`
	IssueDate   time.Time `json:"issue_date" db:"issue_date" sql:"type:timestamp;not null"`
	Description string    `json:"description" db:"description" sql:"type:text"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}
