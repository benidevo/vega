package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobStatus_FromString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    JobStatus
		wantErr bool
	}{
		{"Interested", "Interested", INTERESTED, false},
		{"Applied", "Applied", APPLIED, false},
		{"Interviewing", "Interviewing", INTERVIEWING, false},
		{"Offer Received", "Offer Received", OFFER_RECEIVED, false},
		{"Rejected", "Rejected", REJECTED, false},
		{"Not Interested", "Not Interested", NOT_INTERESTED, false},
		{"Invalid", "Invalid", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s JobStatus
			got, err := s.FromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   string
	}{
		{INTERESTED, "Interested"},
		{APPLIED, "Applied"},
		{INTERVIEWING, "Interviewing"},
		{OFFER_RECEIVED, "Offer Received"},
		{REJECTED, "Rejected"},
		{NOT_INTERESTED, "Not Interested"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJobType_FromString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    JobType
		wantErr bool
	}{
		{"Full Time", "Full Time", FULL_TIME, false},
		{"Part Time", "Part Time", PART_TIME, false},
		{"Contract", "Contract", CONTRACT, false},
		{"Intern", "Intern", INTERN, false},
		{"Remote", "Remote", REMOTE, false},
		{"Freelance", "Freelance", FREELANCE, false},
		{"Other", "Other", OTHER, false},
		{"Invalid", "Invalid", OTHER, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JobType
			got, err := j.FromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExperienceLevel_FromString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    ExperienceLevel
		wantErr bool
	}{
		{"Entry", "entry", ENTRY, false},
		{"Entry Level", "entry level", ENTRY, false},
		{"Junior", "junior", ENTRY, false},
		{"Mid", "mid", MID_LEVEL, false},
		{"Mid Level", "mid level", MID_LEVEL, false},
		{"Intermediate", "intermediate", MID_LEVEL, false},
		{"Senior", "senior", SENIOR, false},
		{"Senior Level", "senior level", SENIOR, false},
		{"Executive", "executive", EXECUTIVE, false},
		{"Leadership", "leadership", EXECUTIVE, false},
		{"Invalid", "invalid", MID_LEVEL, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e ExperienceLevel
			got, err := e.FromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewJob_WithOptions(t *testing.T) {
	title := "Software Engineer"
	description := "Build awesome software"
	company := Company{Name: "Acme Corp"}
	deadline := time.Now().Add(7 * 24 * time.Hour)
	postedAt := time.Now().Add(-24 * time.Hour)
	requiredSkills := []string{"Go", "SQL", "AWS"}

	// Create job with all options
	job := NewJob(
		title,
		description,
		company,
		WithJobType(FULL_TIME),
		WithLocation("Remote"),
		WithSourceURL("https://example.com"),
		WithSalaryRange("$100k-150k"),
		WithRequiredSkills(requiredSkills),
		WithApplicationDeadline(deadline),
		WithApplicationURL("https://apply.example.com"),
		WithStatus(APPLIED),
		WithExperienceLevel(SENIOR),
		WithContactPerson("John Doe"),
		WithNotes("Great opportunity"),
		WithPostedAt(postedAt),
	)

	assert.Equal(t, title, job.Title)
	assert.Equal(t, description, job.Description)
	assert.Equal(t, company.Name, job.Company.Name)
	assert.Equal(t, FULL_TIME, job.JobType)
	assert.Equal(t, "Remote", job.Location)
	assert.Equal(t, "https://example.com", job.SourceURL)
	assert.Equal(t, "$100k-150k", job.SalaryRange)
	assert.Equal(t, requiredSkills, job.RequiredSkills)
	assert.Equal(t, deadline.Unix(), job.ApplicationDeadline.Unix())
	assert.Equal(t, "https://apply.example.com", job.ApplicationURL)
	assert.Equal(t, APPLIED, job.Status)
	assert.Equal(t, SENIOR, job.ExperienceLevel)
	assert.Equal(t, "John Doe", job.ContactPerson)
	assert.Equal(t, "Great opportunity", job.Notes)
	assert.Equal(t, postedAt.Unix(), job.PostedAt.Unix())
	assert.NotZero(t, job.CreatedAt)
	assert.NotZero(t, job.UpdatedAt)
}

func TestNewJob_DefaultValues(t *testing.T) {
	title := "Software Engineer"
	description := "Build awesome software"
	company := Company{Name: "Acme Corp"}

	// Create job with only required fields
	job := NewJob(title, description, company)

	// Verify default values
	assert.Equal(t, title, job.Title)
	assert.Equal(t, description, job.Description)
	assert.Equal(t, company.Name, job.Company.Name)
	assert.Equal(t, INTERESTED, job.Status)
	assert.Empty(t, job.Location)
	assert.Empty(t, job.SourceURL)
	assert.Empty(t, job.SalaryRange)
	assert.Empty(t, job.RequiredSkills)
	assert.Nil(t, job.ApplicationDeadline)
	assert.Empty(t, job.ApplicationURL)
	assert.Zero(t, job.ExperienceLevel)
	assert.Empty(t, job.ContactPerson)
	assert.Empty(t, job.Notes)
	assert.Nil(t, job.PostedAt)
	assert.NotZero(t, job.CreatedAt)
	assert.NotZero(t, job.UpdatedAt)
}

func TestJob_Validate(t *testing.T) {
	tests := []struct {
		name    string
		job     *Job
		wantErr bool
	}{
		{
			name: "Valid job",
			job: &Job{
				Title:       "Software Engineer",
				Description: "Build awesome software",
				Company:     Company{Name: "Acme Corp"},
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			job: &Job{
				Title:       "",
				Description: "Build awesome software",
				Company:     Company{Name: "Acme Corp"},
			},
			wantErr: true,
		},
		{
			name: "Missing description",
			job: &Job{
				Title:       "Software Engineer",
				Description: "",
				Company:     Company{Name: "Acme Corp"},
			},
			wantErr: true,
		},
		{
			name: "Missing company name",
			job: &Job{
				Title:       "Software Engineer",
				Description: "Build awesome software",
				Company:     Company{Name: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.job.Title == "" {
					assert.True(t, strings.Contains(err.Error(), "title"), "Expected error about title")
				} else if tt.job.Description == "" {
					assert.True(t, strings.Contains(err.Error(), "description"), "Expected error about description")
				} else if tt.job.Company.Name == "" {
					assert.True(t, strings.Contains(err.Error(), "company"), "Expected error about company")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJob_IsActive(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name     string
		deadline *time.Time
		want     bool
	}{
		{"No deadline", nil, true},
		{"Future deadline", &future, true},
		{"Past deadline", &past, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Title:               "Software Engineer",
				Description:         "Build awesome software",
				Company:             Company{Name: "Acme Corp"},
				ApplicationDeadline: tt.deadline,
			}
			got := job.IsActive()
			assert.Equal(t, tt.want, got)
		})
	}
}
