package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrompt_ToCoverLetterPrompt(t *testing.T) {
	tests := []struct {
		name             string
		prompt           Prompt
		defaultWordRange string
		expectedContains []string
	}{
		{
			name: "complete prompt with all fields",
			prompt: Prompt{
				Instructions: "Generate a professional cover letter",
				Request: Request{
					ApplicantName:    "John Doe",
					ApplicantProfile: "Senior Software Engineer with 5 years experience",
					JobDescription:   "Looking for a Go developer",
					ExtraContext:     "Remote position preferred",
				},
			},
			defaultWordRange: "250-400",
			expectedContains: []string{
				"Generate a professional cover letter",
				"John Doe",
				"Senior Software Engineer with 5 years experience",
				"Looking for a Go developer",
				"Remote position preferred",
				"250-400 words",
				"JSON object",
				"content:",
			},
		},
		{
			name: "prompt with empty extra context",
			prompt: Prompt{
				Instructions: "Create cover letter",
				Request: Request{
					ApplicantName:    "Jane Smith",
					ApplicantProfile: "Marketing Manager",
					JobDescription:   "Marketing role",
					ExtraContext:     "",
				},
			},
			defaultWordRange: "200-300",
			expectedContains: []string{
				"Create cover letter",
				"Jane Smith",
				"Marketing Manager",
				"Marketing role",
				"200-300 words",
			},
		},
		{
			name: "prompt with custom word range",
			prompt: Prompt{
				Instructions: "Write letter",
				Request: Request{
					ApplicantName:    "Bob Wilson",
					ApplicantProfile: "Designer",
					JobDescription:   "Design position",
				},
			},
			defaultWordRange: "100-200",
			expectedContains: []string{
				"100-200 words",
				"Bob Wilson",
				"Designer",
				"Design position",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prompt.ToCoverLetterPrompt(tt.defaultWordRange)

			assert.NotEmpty(t, result)
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}

			// Ensure JSON structure requirements are present
			assert.Contains(t, result, "JSON object")
			assert.Contains(t, result, "content:")
			assert.Contains(t, result, "professional")
		})
	}
}

func TestPrompt_ToMatchAnalysisPrompt(t *testing.T) {
	tests := []struct {
		name             string
		prompt           Prompt
		minMatchScore    int
		maxMatchScore    int
		expectedContains []string
	}{
		{
			name: "complete match analysis prompt",
			prompt: Prompt{
				Instructions: "Analyze the job match",
				Request: Request{
					ApplicantName:    "Alice Johnson",
					ApplicantProfile: "Full-stack developer with React and Node.js",
					JobDescription:   "React developer position at startup",
					ExtraContext:     "Startup environment experience preferred",
				},
			},
			minMatchScore: 0,
			maxMatchScore: 100,
			expectedContains: []string{
				"Analyze the job match",
				"Alice Johnson",
				"Full-stack developer with React and Node.js",
				"React developer position at startup",
				"Startup environment experience preferred",
				"integer from 0-100",
				"where 0 is no match and 100 is perfect match",
				"matchScore:",
				"strengths:",
				"weaknesses:",
				"highlights:",
				"feedback:",
				"JSON object",
			},
		},
		{
			name: "custom score range",
			prompt: Prompt{
				Instructions: "Match analysis",
				Request: Request{
					ApplicantName:    "Charlie Brown",
					ApplicantProfile: "Project Manager",
					JobDescription:   "PM role",
				},
			},
			minMatchScore: 10,
			maxMatchScore: 90,
			expectedContains: []string{
				"integer from 10-90",
				"where 10 is no match and 90 is perfect match",
				"Charlie Brown",
				"Project Manager",
				"PM role",
			},
		},
		{
			name: "prompt with empty extra context",
			prompt: Prompt{
				Instructions: "Analyze match",
				Request: Request{
					ApplicantName:    "David Lee",
					ApplicantProfile: "Data Scientist",
					JobDescription:   "ML Engineer role",
					ExtraContext:     "",
				},
			},
			minMatchScore: 0,
			maxMatchScore: 100,
			expectedContains: []string{
				"David Lee",
				"Data Scientist",
				"ML Engineer role",
				"Skills alignment",
				"Experience level",
				"Industry knowledge",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prompt.ToMatchAnalysisPrompt(tt.minMatchScore, tt.maxMatchScore)

			assert.NotEmpty(t, result)
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}

			// Ensure analysis criteria are present
			assert.Contains(t, result, "Skills alignment")
			assert.Contains(t, result, "Experience level")
			assert.Contains(t, result, "Industry knowledge")
			assert.Contains(t, result, "Cultural fit")
			assert.Contains(t, result, "Growth potential")

			// Ensure JSON structure requirements are present
			assert.Contains(t, result, "JSON object")
			assert.Contains(t, result, "EXACTLY this structure")
		})
	}
}

func TestPrompt_BoundaryConditions(t *testing.T) {
	t.Run("empty fields in cover letter prompt", func(t *testing.T) {
		prompt := Prompt{
			Instructions: "",
			Request: Request{
				ApplicantName:    "",
				ApplicantProfile: "",
				JobDescription:   "",
				ExtraContext:     "",
			},
		}

		result := prompt.ToCoverLetterPrompt("100-200")

		// Should still generate a valid prompt structure
		assert.Contains(t, result, "JSON object")
		assert.Contains(t, result, "content:")
		assert.Contains(t, result, "100-200 words")
	})

	t.Run("empty fields in match analysis prompt", func(t *testing.T) {
		prompt := Prompt{
			Instructions: "",
			Request: Request{
				ApplicantName:    "",
				ApplicantProfile: "",
				JobDescription:   "",
				ExtraContext:     "",
			},
		}

		result := prompt.ToMatchAnalysisPrompt(0, 100)

		assert.Contains(t, result, "JSON object")
		assert.Contains(t, result, "matchScore:")
		assert.Contains(t, result, "integer from 0-100")
	})

	t.Run("extreme score boundaries", func(t *testing.T) {
		prompt := Prompt{
			Instructions: "Test",
			Request: Request{
				ApplicantName:    "Test User",
				ApplicantProfile: "Test Profile",
				JobDescription:   "Test Job",
			},
		}

		result := prompt.ToMatchAnalysisPrompt(-50, 200)

		assert.Contains(t, result, "integer from -50-200")
		assert.Contains(t, result, "where -50 is no match and 200 is perfect match")
	})
}
