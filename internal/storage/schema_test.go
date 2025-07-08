package storage

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserDocument(t *testing.T) {
	doc := NewUserDocument()

	assert.NotNil(t, doc)
	assert.NotZero(t, doc.UpdatedAt)
	assert.NotNil(t, doc.Data.Companies)
	assert.NotNil(t, doc.Data.Jobs)
	assert.NotNil(t, doc.Data.Matches)
	assert.Empty(t, doc.Data.Companies)
	assert.Empty(t, doc.Data.Jobs)
	assert.Empty(t, doc.Data.Matches)
}

func TestUserDocument_Validate(t *testing.T) {
	tests := []struct {
		name    string
		doc     *UserDocument
		wantErr bool
	}{
		{
			name: "valid document",
			doc: &UserDocument{
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name:    "empty document",
			doc:     &UserDocument{},
			wantErr: false, // No validation rules currently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserDocument_Checksum(t *testing.T) {
	doc := NewUserDocument()
	doc.Data.Profile = &Profile{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	err := doc.UpdateChecksum()
	require.NoError(t, err)
	assert.NotEmpty(t, doc.Checksum)

	err = doc.VerifyChecksum()
	assert.NoError(t, err)

	originalEmail := doc.Data.Profile.Email
	doc.Data.Profile.Email = "tampered@example.com"
	err = doc.VerifyChecksum()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")

	doc.Data.Profile.Email = originalEmail

	doc.Checksum = ""
	err = doc.VerifyChecksum()
	assert.NoError(t, err)
}

func TestUserDocument_ToJSON(t *testing.T) {
	doc := NewUserDocument()
	doc.Data.Companies = []*Company{
		{ID: 1, Name: "Test Company"},
	}
	doc.Data.Jobs = []*Job{
		{ID: 1, Title: "Software Engineer"},
	}

	data, err := doc.ToJSON(false)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)

	prettyData, err := doc.ToJSON(true)
	require.NoError(t, err)
	assert.NotEmpty(t, prettyData)
	assert.Contains(t, string(prettyData), "\n")
	assert.Greater(t, len(prettyData), len(data))
}

func TestFromJSON(t *testing.T) {
	// Create a document
	original := NewUserDocument()
	original.Data.Profile = &Profile{
		FirstName: "Jane",
		LastName:  "Smith",
	}
	original.Data.Companies = []*Company{
		{ID: 1, Name: "Acme Corp"},
	}

	data, err := original.ToJSON(false)
	require.NoError(t, err)

	parsed, err := FromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, original.UpdatedAt.Unix(), parsed.UpdatedAt.Unix())
	assert.NotNil(t, parsed.Data.Profile)
	assert.Equal(t, "Jane", parsed.Data.Profile.FirstName)
	assert.Len(t, parsed.Data.Companies, 1)
	assert.Equal(t, "Acme Corp", parsed.Data.Companies[0].Name)

	_, err = FromJSON([]byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")

	tampered := make([]byte, len(data))
	copy(tampered, data)
	// Tamper with the data but keep valid JSON structure
	tampered = []byte(strings.Replace(string(tampered), "Jane", "John", 1))
	_, err = FromJSON(tampered)
	assert.Error(t, err)
}

func TestUserDocument_GetByID(t *testing.T) {
	doc := NewUserDocument()

	// Add test data
	doc.Data.Companies = []*Company{
		{ID: 1, Name: "Company 1"},
		{ID: 2, Name: "Company 2"},
	}
	doc.Data.Jobs = []*Job{
		{ID: 1, Title: "Job 1", Company: Company{ID: 1}},
		{ID: 2, Title: "Job 2", Company: Company{ID: 2}},
	}
	doc.Data.Matches = []*MatchResult{
		{ID: 1, JobID: 1, MatchScore: 85},
		{ID: 2, JobID: 2, MatchScore: 90},
	}

	company := doc.GetCompanyByID(1)
	assert.NotNil(t, company)
	assert.Equal(t, "Company 1", company.Name)

	company = doc.GetCompanyByID(999)
	assert.Nil(t, company)

	job := doc.GetJobByID(2)
	assert.NotNil(t, job)
	assert.Equal(t, "Job 2", job.Title)

	job = doc.GetJobByID(999)
	assert.Nil(t, job)

	match := doc.GetMatchResultByID(1)
	assert.NotNil(t, match)
	assert.Equal(t, 85, match.MatchScore)

	match = doc.GetMatchResultByID(999)
	assert.Nil(t, match)
}

func TestUserDocument_GetJobsByCompany(t *testing.T) {
	doc := NewUserDocument()

	doc.Data.Companies = []*Company{
		{ID: 1, Name: "Company A"},
		{ID: 2, Name: "Company B"},
	}
	doc.Data.Jobs = []*Job{
		{ID: 1, Title: "Job 1", Company: Company{ID: 1}},
		{ID: 2, Title: "Job 2", Company: Company{ID: 1}},
		{ID: 3, Title: "Job 3", Company: Company{ID: 2}},
	}

	jobs := doc.GetJobsByCompany(1)
	assert.Len(t, jobs, 2)
	assert.Equal(t, "Job 1", jobs[0].Title)
	assert.Equal(t, "Job 2", jobs[1].Title)

	jobs = doc.GetJobsByCompany(2)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "Job 3", jobs[0].Title)

	jobs = doc.GetJobsByCompany(999)
	assert.Empty(t, jobs)
}

func TestUserDocument_GetMatchesByJob(t *testing.T) {
	doc := NewUserDocument()

	doc.Data.Matches = []*MatchResult{
		{ID: 1, JobID: 1, MatchScore: 80},
		{ID: 2, JobID: 1, MatchScore: 85},
		{ID: 3, JobID: 2, MatchScore: 90},
	}

	matches := doc.GetMatchesByJob(1)
	assert.Len(t, matches, 2)
	assert.Equal(t, 80, matches[0].MatchScore)
	assert.Equal(t, 85, matches[1].MatchScore)

	matches = doc.GetMatchesByJob(2)
	assert.Len(t, matches, 1)
	assert.Equal(t, 90, matches[0].MatchScore)

	matches = doc.GetMatchesByJob(999)
	assert.Empty(t, matches)
}

func TestUserDocument_RemoveCompany(t *testing.T) {
	doc := NewUserDocument()

	doc.Data.Companies = []*Company{
		{ID: 1, Name: "Company 1"},
		{ID: 2, Name: "Company 2"},
	}
	doc.Data.Jobs = []*Job{
		{ID: 1, Title: "Job 1", Company: Company{ID: 1}},
		{ID: 2, Title: "Job 2", Company: Company{ID: 1}},
		{ID: 3, Title: "Job 3", Company: Company{ID: 2}},
	}
	doc.Data.Matches = []*MatchResult{
		{ID: 1, JobID: 1, MatchScore: 80},
		{ID: 2, JobID: 2, MatchScore: 85},
		{ID: 3, JobID: 3, MatchScore: 90},
	}

	doc.RemoveCompany(1)

	assert.Len(t, doc.Data.Companies, 1)
	assert.Equal(t, 2, doc.Data.Companies[0].ID)

	assert.Len(t, doc.Data.Jobs, 1)
	assert.Equal(t, 3, doc.Data.Jobs[0].ID)

	assert.Len(t, doc.Data.Matches, 1)
	assert.Equal(t, 3, doc.Data.Matches[0].JobID)
}

func TestUserDocument_RemoveJob(t *testing.T) {
	doc := NewUserDocument()

	doc.Data.Jobs = []*Job{
		{ID: 1, Title: "Job 1"},
		{ID: 2, Title: "Job 2"},
	}
	doc.Data.Matches = []*MatchResult{
		{ID: 1, JobID: 1, MatchScore: 80},
		{ID: 2, JobID: 1, MatchScore: 85},
		{ID: 3, JobID: 2, MatchScore: 90},
	}

	doc.RemoveJob(1)

	assert.Len(t, doc.Data.Jobs, 1)
	assert.Equal(t, 2, doc.Data.Jobs[0].ID)

	assert.Len(t, doc.Data.Matches, 1)
	assert.Equal(t, 2, doc.Data.Matches[0].JobID)
}
