package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileValidation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid profile",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Industry:  IndustryTechnology,
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			profile: Profile{
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing first name",
			profile: Profile{
				UserID:   1,
				LastName: "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing last name",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
			},
			wantErr: true,
		},
		{
			name: "first name too long",
			profile: Profile{
				UserID:    1,
				FirstName: string(make([]byte, 101)),
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "invalid phone number",
			profile: Profile{
				UserID:      1,
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "invalid!@#",
			},
			wantErr: true,
		},
		{
			name: "valid phone number",
			profile: Profile{
				UserID:      1,
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "+1 (555) 123-4567",
			},
			wantErr: false,
		},
		{
			name: "invalid email format",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "valid email format",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty email (should be valid)",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
			},
			wantErr: false,
		},
		{
			name: "invalid LinkedIn URL",
			profile: Profile{
				UserID:          1,
				FirstName:       "John",
				LastName:        "Doe",
				LinkedInProfile: "not-a-linkedin-url",
			},
			wantErr: true,
		},
		{
			name: "valid LinkedIn URL",
			profile: Profile{
				UserID:          1,
				FirstName:       "John",
				LastName:        "Doe",
				LinkedInProfile: "https://www.linkedin.com/in/johndoe",
			},
			wantErr: false,
		},
		{
			name: "invalid GitHub URL",
			profile: Profile{
				UserID:        1,
				FirstName:     "John",
				LastName:      "Doe",
				GitHubProfile: "not-a-github-url",
			},
			wantErr: true,
		},
		{
			name: "valid GitHub URL",
			profile: Profile{
				UserID:        1,
				FirstName:     "John",
				LastName:      "Doe",
				GitHubProfile: "https://github.com/johndoe",
			},
			wantErr: false,
		},
		{
			name: "invalid website URL",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Website:   "not a url",
			},
			wantErr: true,
		},
		{
			name: "too many skills",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Skills:    make([]string, 51),
			},
			wantErr: true,
		},
		{
			name: "invalid industry",
			profile: Profile{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Industry:  Industry(999),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileSanitize(t *testing.T) {
	profile := Profile{
		FirstName:       "  John  ",
		LastName:        "  Doe  ",
		Title:           "  Software Engineer  ",
		Location:        "  New York  ",
		CareerSummary:   "  Experienced developer  ",
		PhoneNumber:     "  +1234567890  ",
		Email:           "  john.doe@example.com  ",
		LinkedInProfile: "  https://linkedin.com/in/johndoe  ",
		GitHubProfile:   "  https://github.com/johndoe  ",
		Website:         "  https://johndoe.com  ",
		Skills:          []string{"  Go  ", "  Python  ", "", "  JavaScript  "},
	}

	profile.Sanitize()

	assert.Equal(t, "John", profile.FirstName)
	assert.Equal(t, "Doe", profile.LastName)
	assert.Equal(t, "Software Engineer", profile.Title)
	assert.Equal(t, "New York", profile.Location)
	assert.Equal(t, "Experienced developer", profile.CareerSummary)
	assert.Equal(t, "+1234567890", profile.PhoneNumber)
	assert.Equal(t, "john.doe@example.com", profile.Email)
	assert.Equal(t, "https://linkedin.com/in/johndoe", profile.LinkedInProfile)
	assert.Equal(t, "https://github.com/johndoe", profile.GitHubProfile)
	assert.Equal(t, "https://johndoe.com", profile.Website)
	assert.Equal(t, []string{"Go", "Python", "JavaScript"}, profile.Skills)
}

func TestWorkExperienceValidation(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		exp     WorkExperience
		wantErr bool
	}{
		{
			name: "valid work experience",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: past,
			},
			wantErr: false,
		},
		{
			name: "missing profile ID",
			exp: WorkExperience{
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: past,
			},
			wantErr: true,
		},
		{
			name: "missing company",
			exp: WorkExperience{
				ProfileID: 1,
				Title:     "Software Engineer",
				StartDate: past,
			},
			wantErr: true,
		},
		{
			name: "missing title",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				StartDate: past,
			},
			wantErr: true,
		},
		{
			name: "future start date",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: future,
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: now,
				EndDate:   &past,
			},
			wantErr: true,
		},
		{
			name: "current job with end date",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: past,
				EndDate:   &now,
				Current:   true,
			},
			wantErr: true,
		},
		{
			name: "valid current job",
			exp: WorkExperience{
				ProfileID: 1,
				Company:   "Acme Corp",
				Title:     "Software Engineer",
				StartDate: past,
				Current:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exp.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkExperienceSanitize(t *testing.T) {
	exp := WorkExperience{
		Company:     "  Acme Corp  ",
		Title:       "  Software Engineer  ",
		Location:    "  New York  ",
		Description: "  Building great software  ",
	}

	exp.Sanitize()

	assert.Equal(t, "Acme Corp", exp.Company)
	assert.Equal(t, "Software Engineer", exp.Title)
	assert.Equal(t, "New York", exp.Location)
	assert.Equal(t, "Building great software", exp.Description)
}

func TestEducationValidation(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		edu     Education
		wantErr bool
	}{
		{
			name: "valid education",
			edu: Education{
				ProfileID:   1,
				Institution: "MIT",
				Degree:      "BS Computer Science",
				StartDate:   past,
			},
			wantErr: false,
		},
		{
			name: "missing profile ID",
			edu: Education{
				Institution: "MIT",
				Degree:      "BS Computer Science",
				StartDate:   past,
			},
			wantErr: true,
		},
		{
			name: "future start date",
			edu: Education{
				ProfileID:   1,
				Institution: "MIT",
				Degree:      "BS Computer Science",
				StartDate:   future,
			},
			wantErr: true,
		},
		{
			name: "end date before start date",
			edu: Education{
				ProfileID:   1,
				Institution: "MIT",
				Degree:      "BS Computer Science",
				StartDate:   now,
				EndDate:     &past,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.edu.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEducationSanitize(t *testing.T) {
	edu := Education{
		Institution:  "  MIT  ",
		Degree:       "  BS Computer Science  ",
		FieldOfStudy: "  Computer Science  ",
		Description:  "  Graduated with honors  ",
	}

	edu.Sanitize()

	assert.Equal(t, "MIT", edu.Institution)
	assert.Equal(t, "BS Computer Science", edu.Degree)
	assert.Equal(t, "Computer Science", edu.FieldOfStudy)
	assert.Equal(t, "Graduated with honors", edu.Description)
}

func TestCertificationValidation(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		cert    Certification
		wantErr bool
	}{
		{
			name: "valid certification",
			cert: Certification{
				ProfileID:  1,
				Name:       "AWS Solutions Architect",
				IssuingOrg: "Amazon Web Services",
				IssueDate:  past,
			},
			wantErr: false,
		},
		{
			name: "missing profile ID",
			cert: Certification{
				Name:       "AWS Solutions Architect",
				IssuingOrg: "Amazon Web Services",
				IssueDate:  past,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			cert: Certification{
				ProfileID:  1,
				IssuingOrg: "Amazon Web Services",
				IssueDate:  past,
			},
			wantErr: true,
		},
		{
			name: "missing issuing org",
			cert: Certification{
				ProfileID: 1,
				Name:      "AWS Solutions Architect",
				IssueDate: past,
			},
			wantErr: true,
		},
		{
			name: "future issue date",
			cert: Certification{
				ProfileID:  1,
				Name:       "AWS Solutions Architect",
				IssuingOrg: "Amazon Web Services",
				IssueDate:  future,
			},
			wantErr: true,
		},
		{
			name: "expiry date before issue date",
			cert: Certification{
				ProfileID:  1,
				Name:       "AWS Solutions Architect",
				IssuingOrg: "Amazon Web Services",
				IssueDate:  now,
				ExpiryDate: &past,
			},
			wantErr: true,
		},
		{
			name: "invalid credential URL",
			cert: Certification{
				ProfileID:     1,
				Name:          "AWS Solutions Architect",
				IssuingOrg:    "Amazon Web Services",
				IssueDate:     past,
				CredentialURL: "not a url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cert.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCertificationSanitize(t *testing.T) {
	cert := Certification{
		Name:          "  AWS Solutions Architect  ",
		IssuingOrg:    "  Amazon Web Services  ",
		CredentialID:  "  ABC123  ",
		CredentialURL: "  https://aws.amazon.com/cert/ABC123  ",
	}

	cert.Sanitize()

	assert.Equal(t, "AWS Solutions Architect", cert.Name)
	assert.Equal(t, "Amazon Web Services", cert.IssuingOrg)
	assert.Equal(t, "ABC123", cert.CredentialID)
	assert.Equal(t, "https://aws.amazon.com/cert/ABC123", cert.CredentialURL)
}

func TestCustomValidators(t *testing.T) {
	t.Run("validatePhone", func(t *testing.T) {
		validPhones := []string{
			"",
			"1234567890",
			"+1-234-567-8900",
			"(123) 456-7890",
			"+1 (555) 123-4567",
		}

		for _, phone := range validPhones {
			profile := Profile{
				UserID:      1,
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: phone,
			}
			err := profile.Validate()
			assert.NoError(t, err, "Phone: %s should be valid", phone)
		}

		invalidPhones := []string{
			"123456789",             // too short (9 chars)
			"123456789012345678901", // too long (21 chars)
			"phone@number",          // invalid characters
		}

		for _, phone := range invalidPhones {
			profile := Profile{
				UserID:      1,
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: phone,
			}
			err := profile.Validate()
			assert.Error(t, err, "Phone: %s should be invalid", phone)
		}
	})

	t.Run("validateLinkedIn", func(t *testing.T) {
		validURLs := []string{
			"",
			"https://www.linkedin.com/in/johndoe",
			"https://linkedin.com/in/johndoe",
		}

		for _, url := range validURLs {
			profile := Profile{
				UserID:          1,
				FirstName:       "John",
				LastName:        "Doe",
				LinkedInProfile: url,
			}
			err := profile.Validate()
			assert.NoError(t, err, "LinkedIn URL: %s should be valid", url)
		}

		invalidURLs := []string{
			"linkedin.com/in/johndoe",
			"https://twitter.com/johndoe",
			"not-a-url",
		}

		for _, url := range invalidURLs {
			profile := Profile{
				UserID:          1,
				FirstName:       "John",
				LastName:        "Doe",
				LinkedInProfile: url,
			}
			err := profile.Validate()
			assert.Error(t, err, "LinkedIn URL: %s should be invalid", url)
		}
	})

	t.Run("validateGitHub", func(t *testing.T) {
		validURLs := []string{
			"",
			"https://github.com/johndoe",
			"https://www.github.com/johndoe",
		}

		for _, url := range validURLs {
			profile := Profile{
				UserID:        1,
				FirstName:     "John",
				LastName:      "Doe",
				GitHubProfile: url,
			}
			err := profile.Validate()
			assert.NoError(t, err, "GitHub URL: %s should be valid", url)
		}

		invalidURLs := []string{
			"github.com/johndoe",
			"https://gitlab.com/johndoe",
			"not-a-url",
		}

		for _, url := range invalidURLs {
			profile := Profile{
				UserID:        1,
				FirstName:     "John",
				LastName:      "Doe",
				GitHubProfile: url,
			}
			err := profile.Validate()
			assert.Error(t, err, "GitHub URL: %s should be invalid", url)
		}
	})

	t.Run("validateNotFuture", func(t *testing.T) {
		now := time.Now()
		past := now.Add(-24 * time.Hour)
		future := now.Add(24 * time.Hour)

		// Test with time.Time
		exp := WorkExperience{
			ProfileID: 1,
			Company:   "Acme Corp",
			Title:     "Software Engineer",
			StartDate: past,
		}
		require.NoError(t, exp.Validate())

		exp.StartDate = future
		require.Error(t, exp.Validate())

		// Test with *time.Time
		exp.StartDate = past
		pastEndDate := now.Add(-12 * time.Hour)
		exp.EndDate = &pastEndDate
		require.NoError(t, exp.Validate())

		exp.EndDate = &future
		require.Error(t, exp.Validate())

		// Test with nil *time.Time
		exp.EndDate = nil
		require.NoError(t, exp.Validate())
	})
}
