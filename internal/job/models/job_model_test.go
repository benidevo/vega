package models

import (
	"testing"

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
			got, err := JobStatusFromString(tt.str)
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
		name string
		str  string
		want JobType
	}{
		{"Full Time", "full_time", FULL_TIME},
		{"Part Time", "part_time", PART_TIME},
		{"Contract", "contract", CONTRACT},
		{"Intern", "intern", INTERN},
		{"Remote", "remote", REMOTE},
		{"Freelance", "freelance", FREELANCE},
		{"Unknown input returns OTHER", "unknown", OTHER},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JobTypeFromString(tt.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJobType_String(t *testing.T) {
	tests := []struct {
		jobType JobType
		want    string
	}{
		{FULL_TIME, "Full Time"},
		{PART_TIME, "Part Time"},
		{CONTRACT, "Contract"},
		{INTERN, "Intern"},
		{REMOTE, "Remote"},
		{FREELANCE, "Freelance"},
		{OTHER, "Other"},
		{999, "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.jobType.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewJob(t *testing.T) {
	t.Run("With options", func(t *testing.T) {
		title := "Software Engineer"
		description := "Build awesome software"
		company := Company{Name: "Acme Corp"}
		requiredSkills := []string{"Go", "SQL", "AWS"}

		job := NewJob(
			title,
			description,
			company,
			WithJobType(FULL_TIME),
			WithLocation("Remote"),
			WithSourceURL("https://example.com"),
			WithRequiredSkills(requiredSkills),
			WithApplicationURL("https://apply.example.com"),
			WithStatus(APPLIED),
			WithNotes("Great opportunity"),
		)

		assert.Equal(t, title, job.Title)
		assert.Equal(t, description, job.Description)
		assert.Equal(t, company.Name, job.Company.Name)
		assert.Equal(t, FULL_TIME, job.JobType)
		assert.Equal(t, "Remote", job.Location)
		assert.Equal(t, "https://example.com", job.SourceURL)
		assert.Equal(t, requiredSkills, job.RequiredSkills)
		assert.Equal(t, "https://apply.example.com", job.ApplicationURL)
		assert.Equal(t, APPLIED, job.Status)
		assert.Equal(t, "Great opportunity", job.Notes)
		assert.NotZero(t, job.CreatedAt)
		assert.NotZero(t, job.UpdatedAt)
	})

	t.Run("Default values", func(t *testing.T) {
		title := "Software Engineer"
		description := "Build awesome software"
		company := Company{Name: "Acme Corp"}

		job := NewJob(title, description, company)

		assert.Equal(t, title, job.Title)
		assert.Equal(t, description, job.Description)
		assert.Equal(t, company.Name, job.Company.Name)
		assert.Equal(t, INTERESTED, job.Status)
		assert.Empty(t, job.Location)
		assert.Empty(t, job.SourceURL)
		assert.Empty(t, job.RequiredSkills)
		assert.Empty(t, job.ApplicationURL)
		assert.Empty(t, job.Notes)
		assert.NotZero(t, job.CreatedAt)
		assert.NotZero(t, job.UpdatedAt)
	})
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
				Description: "Build awesome software",
				Company:     Company{Name: "Acme Corp"},
			},
			wantErr: true,
		},
		{
			name: "Missing description",
			job: &Job{
				Title:   "Software Engineer",
				Company: Company{Name: "Acme Corp"},
			},
			wantErr: true,
		},
		{
			name: "Missing company name",
			job: &Job{
				Title:       "Software Engineer",
				Description: "Build awesome software",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name    string
		current JobStatus
		new     JobStatus
		want    bool
	}{
		{"Same status", INTERESTED, INTERESTED, true},
		{"Forward progression", INTERESTED, APPLIED, true},
		{"Multiple forward steps", INTERVIEWING, OFFER_RECEIVED, true},
		{"Backward progression", APPLIED, INTERESTED, false},
		{"Terminal state cannot transition", REJECTED, APPLIED, false},
		{"Offer can be rejected", OFFER_RECEIVED, REJECTED, true},
		{"Any state to not interested", APPLIED, NOT_INTERESTED, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(tt.current, tt.new)
			assert.Equal(t, tt.want, got)
		})
	}
}
