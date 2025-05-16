package models

import (
	"strings"
	"time"
)

// JobStatus represents the status of a job application.
type JobStatus int

const (
	INTERESTED JobStatus = iota
	APPLIED
	INTERVIEWING
	OFFER_RECEIVED
	REJECTED
	NOT_INTERESTED
)

// FromString converts a string representation of a job status to its corresponding JobStatus value.
func JobStatusFromString(status string) (JobStatus, error) {
	switch strings.ToLower(status) {
	case "interested":
		return INTERESTED, nil
	case "applied":
		return APPLIED, nil
	case "interviewing":
		return INTERVIEWING, nil
	case "offered", "offer received", "offer_received":
		return OFFER_RECEIVED, nil
	case "rejected":
		return REJECTED, nil
	case "not_interested", "not interested":
		return NOT_INTERESTED, nil
	default:
		return -1, ErrInvalidJobStatus
	}
}

// String returns the string representation of the JobStatus value.
func (j JobStatus) String() string {
	switch j {
	case INTERESTED:
		return "Interested"
	case APPLIED:
		return "Applied"
	case INTERVIEWING:
		return "Interviewing"
	case OFFER_RECEIVED:
		return "Offer Received"
	case REJECTED:
		return "Rejected"
	case NOT_INTERESTED:
		return "Not Interested"
	default:
		return "Unknown"
	}
}

// IsValidTransition checks if a change from current status to new status is valid.
// Job status should only move forward in the application process, or to terminal states.
func IsValidTransition(currentStatus, newStatus JobStatus) bool {
	if currentStatus == newStatus {
		return true
	}

	// Terminal states can't transition except to themselves (handled above)
	if currentStatus == REJECTED || currentStatus == NOT_INTERESTED {
		return false
	}

	// Special case: OFFER_RECEIVED can transition to REJECTED
	if currentStatus == OFFER_RECEIVED && newStatus == REJECTED {
		return true
	}

	if newStatus > currentStatus {
		return true
	}

	// Special case: Any state can transition to NOT_INTERESTED
	if newStatus == NOT_INTERESTED {
		return true
	}

	return false
}

// JobType represents the type of job (e.g., full-time, part-time, etc.).
type JobType int

const (
	FULL_TIME JobType = iota
	PART_TIME
	CONTRACT
	INTERN
	REMOTE
	FREELANCE
	OTHER
)

// FromString converts a string representation of a job type to its corresponding JobType constant.
func JobTypeFromString(jobType string) (JobType, error) {
	switch strings.ToLower(jobType) {
	case "full_time":
		return FULL_TIME, nil
	case "part_time":
		return PART_TIME, nil
	case "contract":
		return CONTRACT, nil
	case "intern":
		return INTERN, nil
	case "remote":
		return REMOTE, nil
	case "freelance":
		return FREELANCE, nil
	case "other":
		return OTHER, nil
	default:
		return OTHER, ErrInvalidJobType
	}
}

// String returns the string representation of the JobType.
func (j JobType) String() string {
	switch j {
	case FULL_TIME:
		return "Full Time"
	case PART_TIME:
		return "Part Time"
	case CONTRACT:
		return "Contract"
	case INTERN:
		return "Intern"
	case REMOTE:
		return "Remote"
	case FREELANCE:
		return "Freelance"
	default:
		return "Other"
	}
}

// ExperienceLevel represents the level of experience required for a job.
type ExperienceLevel int

const (
	ENTRY ExperienceLevel = iota
	MID_LEVEL
	SENIOR
	EXECUTIVE
	NOT_SPECIFIED
)

// FromString parses a string and returns the corresponding ExperienceLevel.
// Accepts various synonyms for each level.
// Returns NOT_SPECIFIED if the string does not match any known level.
func ExperienceLevelFromString(experience string) ExperienceLevel {
	switch strings.ToLower(experience) {
	case "entry", "entry level", "junior":
		return ENTRY
	case "mid", "mid level", "intermediate":
		return MID_LEVEL
	case "senior", "senior level":
		return SENIOR
	case "executive", "leadership":
		return EXECUTIVE
	default:
		return NOT_SPECIFIED
	}
}

// String returns the string representation of the ExperienceLevel.
// It maps each ExperienceLevel constant to a human-readable string.
func (e ExperienceLevel) String() string {
	switch e {
	case ENTRY:
		return "Entry Level"
	case MID_LEVEL:
		return "Mid Level"
	case SENIOR:
		return "Senior Level"
	case EXECUTIVE:
		return "Executive Level"
	default:
		return "Not Specified"
	}
}

// Job represents a job posting with details such as title, description, location, company, and application information.
type Job struct {
	ID                  int             `json:"id" db:"id" sql:"primary_key;auto_increment"`
	Title               string          `json:"title" db:"title" sql:"type:text;not null;index"`
	Description         string          `json:"description" db:"description" sql:"type:text;not null"`
	Location            string          `json:"location" db:"location" sql:"type:text"`
	JobType             JobType         `json:"job_type" db:"job_type" sql:"type:integer;not null;default:0"`
	SourceURL           string          `json:"source_url" db:"source_url" sql:"type:text;unique;index"`
	SalaryRange         string          `json:"salary_range" db:"salary_range" sql:"type:text"`
	RequiredSkills      []string        `json:"required_skills" db:"required_skills" sql:"type:text"` // Stored as JSON
	ApplicationDeadline *time.Time      `json:"application_deadline,omitempty" db:"application_deadline" sql:"type:timestamp"`
	ApplicationURL      string          `json:"application_url" db:"application_url" sql:"type:text"`
	Company             Company         `json:"company" sql:"-"` // Not stored directly, company_id is used instead
	Status              JobStatus       `json:"status" db:"status" sql:"type:integer;not null;default:0;index"`
	ExperienceLevel     ExperienceLevel `json:"experience_level" db:"experience_level" sql:"type:integer;not null;default:0"`
	ContactPerson       string          `json:"contact_person,omitempty" db:"contact_person" sql:"type:text"`
	Notes               string          `json:"notes,omitempty" db:"notes" sql:"type:text"`
	PostedAt            *time.Time      `json:"posted_at,omitempty" db:"posted_at" sql:"type:timestamp"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at" sql:"type:timestamp;not null;default:current_timestamp"`

	// SQL-only fields
	CompanyID int `json:"-" db:"company_id" sql:"type:integer;not null;index;references:companies(id)"`
}

// JobOption defines a function for configuring a Job
type JobOption func(*Job)

// WithJobType sets the job type
func WithJobType(jobType JobType) JobOption {
	return func(j *Job) {
		j.JobType = jobType
	}
}

// WithLocation sets the job location
func WithLocation(location string) JobOption {
	return func(j *Job) {
		j.Location = location
	}
}

// WithSourceURL sets the job source URL
func WithSourceURL(url string) JobOption {
	return func(j *Job) {
		j.SourceURL = url
	}
}

// WithSalaryRange sets the job salary range
func WithSalaryRange(salaryRange string) JobOption {
	return func(j *Job) {
		j.SalaryRange = salaryRange
	}
}

// WithRequiredSkills sets the required skills for the job
func WithRequiredSkills(skills []string) JobOption {
	return func(j *Job) {
		j.RequiredSkills = skills
	}
}

// WithApplicationDeadline sets the application deadline
func WithApplicationDeadline(deadline time.Time) JobOption {
	return func(j *Job) {
		j.ApplicationDeadline = &deadline
	}
}

// WithApplicationURL sets the application URL
func WithApplicationURL(url string) JobOption {
	return func(j *Job) {
		j.ApplicationURL = url
	}
}

// WithStatus sets the job application status
func WithStatus(status JobStatus) JobOption {
	return func(j *Job) {
		j.Status = status
	}
}

// WithExperienceLevel sets the required experience level
func WithExperienceLevel(level ExperienceLevel) JobOption {
	return func(j *Job) {
		j.ExperienceLevel = level
	}
}

// WithContactPerson sets the contact person
func WithContactPerson(contact string) JobOption {
	return func(j *Job) {
		j.ContactPerson = contact
	}
}

// WithNotes sets the notes for the job
func WithNotes(notes string) JobOption {
	return func(j *Job) {
		j.Notes = notes
	}
}

// WithPostedAt sets when the job was posted
func WithPostedAt(postedAt time.Time) JobOption {
	return func(j *Job) {
		j.PostedAt = &postedAt
	}
}

// NewJob creates a new Job instance with the required fields and applies the provided options.
// Only title, description, and company are required. All other fields can be set using options.
func NewJob(title, description string, company Company, options ...JobOption) *Job {
	now := time.Now().UTC()

	job := &Job{
		Title:          title,
		Description:    description,
		Company:        company,
		Status:         INTERESTED,
		RequiredSkills: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	for _, option := range options {
		option(job)
	}

	return job
}

// Validate performs basic validation on the Job struct
func (j *Job) Validate() error {
	if j.Title == "" {
		return ErrJobTitleRequired
	}

	if j.Description == "" {
		return ErrJobDescriptionRequired
	}

	if j.Company.Name == "" {
		return ErrCompanyRequired
	}

	return nil
}

// IsActive checks if the job application is still active based on deadline
func (j *Job) IsActive() bool {
	if j.ApplicationDeadline == nil {
		return true
	}

	return j.ApplicationDeadline.After(time.Now().UTC())
}

// JobFilter defines filters for querying jobs
type JobFilter struct {
	CompanyID *int
	Status    *JobStatus
	JobType   *JobType
	Search    string
	Limit     int
	Offset    int
}

type JobStats struct {
	TotalJobs    int `json:"total_jobs"`
	TotalApplied int `json:"total_applied"`
	HighMatch    int `json:"high_match"`
}
