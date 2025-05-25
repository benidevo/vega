package models

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("linkedin", validateLinkedIn)
	validate.RegisterValidation("github", validateGitHub)
	validate.RegisterValidation("notfuture", validateNotFuture)
	validate.RegisterValidation("validindustry", validateIndustry)
}

// Custom validators
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Optional field
	}
	for _, r := range phone {
		if !strings.ContainsRune("0123456789 -+()", r) {
			return false
		}
	}
	return len(phone) >= 10 && len(phone) <= 20
}

func validateLinkedIn(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Optional field
	}
	return strings.HasPrefix(url, "https://www.linkedin.com/") ||
		strings.HasPrefix(url, "https://linkedin.com/") ||
		strings.HasPrefix(url, "http://www.linkedin.com/") ||
		strings.HasPrefix(url, "http://linkedin.com/")
}

func validateGitHub(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	if url == "" {
		return true // Optional field
	}
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "https://www.github.com/") ||
		strings.HasPrefix(url, "http://github.com/") ||
		strings.HasPrefix(url, "http://www.github.com/")
}

func validateNotFuture(fl validator.FieldLevel) bool {
	switch t := fl.Field().Interface().(type) {
	case time.Time:
		return !t.After(time.Now())
	case *time.Time:
		if t == nil {
			return true
		}
		return !t.After(time.Now())
	}
	return true
}

func validateIndustry(fl validator.FieldLevel) bool {
	industry := fl.Field().Interface().(Industry)
	return industry.IsValid()
}

// Profile represents user profile information
type Profile struct {
	ID              int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	UserID          int       `json:"user_id" db:"user_id" sql:"type:integer;not null;unique;index;references:users(id)" validate:"required,min=1"`
	FirstName       string    `json:"first_name" db:"first_name" sql:"type:text" validate:"required,min=1,max=100"`
	LastName        string    `json:"last_name" db:"last_name" sql:"type:text" validate:"required,min=1,max=100"`
	Title           string    `json:"title" db:"title" sql:"type:text" validate:"max=200"`
	Industry        Industry  `json:"industry" db:"industry" sql:"type:integer" validate:"validindustry"`
	CareerSummary   string    `json:"career_summary" db:"career_summary" sql:"type:text" validate:"max=2000"`
	Skills          []string  `json:"skills" db:"skills" sql:"type:text" validate:"max=50,dive,required,min=1,max=100"`
	PhoneNumber     string    `json:"phone_number" db:"phone_number" sql:"type:text" validate:"phone"`
	Location        string    `json:"location" db:"location" sql:"type:text" validate:"max=200"`
	LinkedInProfile string    `json:"linkedin_profile" db:"linkedin_profile" sql:"type:text" validate:"linkedin,url,max=500"`
	GitHubProfile   string    `json:"github_profile" db:"github_profile" sql:"type:text" validate:"github,url,max=500"`
	Website         string    `json:"website" db:"website" sql:"type:text" validate:"omitempty,url,max=500"`
	CreatedAt       time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`

	WorkExperience []WorkExperience `json:"work_experience,omitempty" db:"-"`
	Education      []Education      `json:"education,omitempty" db:"-"`
	Certifications []Certification  `json:"certifications,omitempty" db:"-"`
}

// Sanitize cleans up the Profile data
func (p *Profile) Sanitize() {
	p.FirstName = strings.TrimSpace(p.FirstName)
	p.LastName = strings.TrimSpace(p.LastName)
	p.Title = strings.TrimSpace(p.Title)
	p.Location = strings.TrimSpace(p.Location)
	p.CareerSummary = strings.TrimSpace(p.CareerSummary)
	p.PhoneNumber = strings.TrimSpace(p.PhoneNumber)
	p.LinkedInProfile = strings.TrimSpace(p.LinkedInProfile)
	p.GitHubProfile = strings.TrimSpace(p.GitHubProfile)
	p.Website = strings.TrimSpace(p.Website)

	cleanSkills := make([]string, 0, len(p.Skills))
	for _, skill := range p.Skills {
		if s := strings.TrimSpace(skill); s != "" {
			cleanSkills = append(cleanSkills, s)
		}
	}
	p.Skills = cleanSkills
}

// Validate validates the Profile struct
func (p *Profile) Validate() error {
	return validate.Struct(p)
}

// WorkExperience represents work experience entries in a user's profile
type WorkExperience struct {
	ID          int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID   int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profiles(id)" validate:"required,min=1"`
	Company     string     `json:"company" db:"company" sql:"type:text;not null" validate:"required,min=1,max=200"`
	Title       string     `json:"title" db:"title" sql:"type:text;not null" validate:"required,min=1,max=200"`
	Location    string     `json:"location" db:"location" sql:"type:text" validate:"max=200"`
	StartDate   time.Time  `json:"start_date" db:"start_date" sql:"type:timestamp;not null" validate:"required,notfuture"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date" sql:"type:timestamp" validate:"omitempty,notfuture,gtfield=StartDate"`
	Description string     `json:"description" db:"description" sql:"type:text" validate:"max=2000"`
	Current     bool       `json:"current" db:"current" sql:"type:boolean;default:false"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Sanitize cleans up the WorkExperience data
func (w *WorkExperience) Sanitize() {
	w.Company = strings.TrimSpace(w.Company)
	w.Title = strings.TrimSpace(w.Title)
	w.Location = strings.TrimSpace(w.Location)
	w.Description = strings.TrimSpace(w.Description)
}

// Validate validates the WorkExperience struct
func (w *WorkExperience) Validate() error {
	// Check that if Current is true, EndDate should be nil
	if w.Current && w.EndDate != nil {
		return ErrEndDateWithCurrent
	}
	return validate.Struct(w)
}

// Education represents education entries in a user's profile
type Education struct {
	ID           int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID    int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profiles(id)" validate:"required,min=1"`
	Institution  string     `json:"institution" db:"institution" sql:"type:text;not null" validate:"required,min=1,max=200"`
	Degree       string     `json:"degree" db:"degree" sql:"type:text;not null" validate:"required,min=1,max=200"`
	FieldOfStudy string     `json:"field_of_study" db:"field_of_study" sql:"type:text" validate:"max=200"`
	StartDate    time.Time  `json:"start_date" db:"start_date" sql:"type:timestamp;not null" validate:"required,notfuture"`
	EndDate      *time.Time `json:"end_date,omitempty" db:"end_date" sql:"type:timestamp" validate:"omitempty,notfuture,gtfield=StartDate"`
	Description  string     `json:"description" db:"description" sql:"type:text" validate:"max=2000"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Sanitize cleans up the Education data
func (e *Education) Sanitize() {
	e.Institution = strings.TrimSpace(e.Institution)
	e.Degree = strings.TrimSpace(e.Degree)
	e.FieldOfStudy = strings.TrimSpace(e.FieldOfStudy)
	e.Description = strings.TrimSpace(e.Description)
}

// Validate validates the Education struct
func (e *Education) Validate() error {
	return validate.Struct(e)
}

// Certification represents certification entries in a user's profile
type Certification struct {
	ID            int        `json:"id" db:"id" sql:"primary_key;auto_increment"`
	ProfileID     int        `json:"profile_id" db:"profile_id" sql:"type:integer;not null;index;references:profiles(id)" validate:"required,min=1"`
	Name          string     `json:"name" db:"name" sql:"type:text;not null" validate:"required,min=1,max=200"`
	IssuingOrg    string     `json:"issuing_org" db:"issuing_org" sql:"type:text;not null" validate:"required,min=1,max=200"`
	IssueDate     time.Time  `json:"issue_date" db:"issue_date" sql:"type:timestamp;not null" validate:"required,notfuture"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty" db:"expiry_date" sql:"type:timestamp" validate:"omitempty,gtfield=IssueDate"`
	CredentialID  string     `json:"credential_id" db:"credential_id" sql:"type:text" validate:"max=200"`
	CredentialURL string     `json:"credential_url" db:"credential_url" sql:"type:text" validate:"omitempty,url,max=500"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// Sanitize cleans up the Certification data
func (c *Certification) Sanitize() {
	c.Name = strings.TrimSpace(c.Name)
	c.IssuingOrg = strings.TrimSpace(c.IssuingOrg)
	c.CredentialID = strings.TrimSpace(c.CredentialID)
	c.CredentialURL = strings.TrimSpace(c.CredentialURL)
}

// Validate validates the Certification struct
func (c *Certification) Validate() error {
	return validate.Struct(c)
}
