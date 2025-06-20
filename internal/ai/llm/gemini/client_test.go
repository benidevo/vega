package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/benidevo/ascentio/internal/ai/llm"
	"github.com/benidevo/ascentio/internal/ai/models"
	"github.com/stretchr/testify/assert"
)

func TestGemini_extractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean JSON object",
			input:    `{"content": "test"}`,
			expected: `{"content": "test"}`,
		},
		{
			name:     "JSON with extra text before",
			input:    `Here is the JSON: {"content": "test"}`,
			expected: `{"content": "test"}`,
		},
		{
			name:     "JSON with extra text after",
			input:    `{"content": "test"} - This is the result`,
			expected: `{"content": "test"}`,
		},
		{
			name:     "nested JSON object",
			input:    `{"outer": {"inner": "value"}, "count": 42}`,
			expected: `{"outer": {"inner": "value"}, "count": 42}`,
		},
		{
			name:     "JSON with text before and after",
			input:    `Response: {"status": "success", "data": {"value": 123}} End of response`,
			expected: `{"status": "success", "data": {"value": 123}}`,
		},
		{
			name:     "no JSON content",
			input:    `This is just plain text without JSON`,
			expected: `This is just plain text without JSON`,
		},
		{
			name:     "malformed JSON - unmatched brace",
			input:    `{"content": "test"`,
			expected: `{"content": "test"`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "only opening brace",
			input:    `{`,
			expected: `{`,
		},
		{
			name:     "multiple JSON objects - returns first",
			input:    `{"first": "object"} {"second": "object"}`,
			expected: `{"first": "object"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gemini{}
			result := g.extractJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGemini_parseMatchResultJSON(t *testing.T) {
	cfg := &Config{
		MinMatchScore:       0,
		MaxMatchScore:       100,
		DefaultStrengthsMsg: "No strengths identified",
		DefaultWeaknessMsg:  "No weaknesses identified",
		DefaultHighlightMsg: "No highlights identified",
		DefaultFeedbackMsg:  "No feedback available",
	}
	g := &Gemini{cfg: cfg}

	tests := []struct {
		name          string
		input         string
		expected      models.MatchResult
		expectedError bool
	}{
		{
			name: "valid JSON response",
			input: `{
				"matchScore": 85,
				"strengths": ["Strong Go skills", "Good communication"],
				"weaknesses": ["Limited Docker experience"],
				"highlights": ["5 years experience", "Team lead"],
				"feedback": "Great candidate overall"
			}`,
			expected: models.MatchResult{
				MatchScore: 85,
				Strengths:  []string{"Strong Go skills", "Good communication"},
				Weaknesses: []string{"Limited Docker experience"},
				Highlights: []string{"5 years experience", "Team lead"},
				Feedback:   "Great candidate overall",
			},
			expectedError: false,
		},
		{
			name: "score out of range - too high",
			input: `{
				"matchScore": 150,
				"strengths": ["Good skills"],
				"weaknesses": ["Some gaps"],
				"highlights": ["Experience"],
				"feedback": "Good candidate"
			}`,
			expected: models.MatchResult{
				MatchScore: 0, // Should be corrected to min score
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Some gaps"},
				Highlights: []string{"Experience"},
				Feedback:   "Good candidate",
			},
			expectedError: false,
		},
		{
			name: "score out of range - too low",
			input: `{
				"matchScore": -10,
				"strengths": ["Good skills"],
				"weaknesses": ["Some gaps"],
				"highlights": ["Experience"],
				"feedback": "Good candidate"
			}`,
			expected: models.MatchResult{
				MatchScore: 0, // Should be corrected to min score
				Strengths:  []string{"Good skills"},
				Weaknesses: []string{"Some gaps"},
				Highlights: []string{"Experience"},
				Feedback:   "Good candidate",
			},
			expectedError: false,
		},
		{
			name: "empty arrays get defaults",
			input: `{
				"matchScore": 75,
				"strengths": [],
				"weaknesses": [],
				"highlights": [],
				"feedback": ""
			}`,
			expected: models.MatchResult{
				MatchScore: 75,
				Strengths:  []string{"No strengths identified"},
				Weaknesses: []string{"No weaknesses identified"},
				Highlights: []string{"No highlights identified"},
				Feedback:   "No feedback available",
			},
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			input:         `{"matchScore": 85, "strengths": [}`,
			expected:      models.MatchResult{},
			expectedError: true,
		},
		{
			name: "JSON with extra text",
			input: `Here is the analysis: {
				"matchScore": 90,
				"strengths": ["Excellent skills"],
				"weaknesses": ["Minor gaps"],
				"highlights": ["Leadership"],
				"feedback": "Top candidate"
			} End of analysis`,
			expected: models.MatchResult{
				MatchScore: 90,
				Strengths:  []string{"Excellent skills"},
				Weaknesses: []string{"Minor gaps"},
				Highlights: []string{"Leadership"},
				Feedback:   "Top candidate",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.parseMatchResultJSON(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.MatchScore, result.MatchScore)
				assert.Equal(t, tt.expected.Strengths, result.Strengths)
				assert.Equal(t, tt.expected.Weaknesses, result.Weaknesses)
				assert.Equal(t, tt.expected.Highlights, result.Highlights)
				assert.Equal(t, tt.expected.Feedback, result.Feedback)
			}
		})
	}
}

func TestGemini_parseCoverLetterJSON(t *testing.T) {
	g := &Gemini{}

	tests := []struct {
		name          string
		input         string
		expected      models.CoverLetter
		expectedError bool
	}{
		{
			name:  "valid cover letter JSON",
			input: `{"content": "Dear Hiring Manager,\n\nI am writing to express my interest..."}`,
			expected: models.CoverLetter{
				Content: "Dear Hiring Manager,\n\nI am writing to express my interest...",
				Format:  models.CoverLetterTypePlainText,
			},
			expectedError: false,
		},
		{
			name:  "cover letter with extra text",
			input: `Here is your cover letter: {"content": "Dear Sir/Madam,\n\nApplication for the position..."} Hope this helps!`,
			expected: models.CoverLetter{
				Content: "Dear Sir/Madam,\n\nApplication for the position...",
				Format:  models.CoverLetterTypePlainText,
			},
			expectedError: false,
		},
		{
			name:          "empty content",
			input:         `{"content": ""}`,
			expected:      models.CoverLetter{},
			expectedError: true,
		},
		{
			name:          "missing content field",
			input:         `{"title": "test"}`,
			expected:      models.CoverLetter{},
			expectedError: true,
		},
		{
			name:          "malformed JSON",
			input:         `{"content": "test"`,
			expected:      models.CoverLetter{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := g.parseCoverLetterJSON(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Content, result.Content)
				assert.Equal(t, tt.expected.Format, result.Format)
			}
		})
	}
}

func TestGemini_executeWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		operation     func() (string, error)
		expectedError bool
		expectRetries bool
	}{
		{
			name: "successful operation on first try",
			config: &Config{
				MaxRetries:     3,
				BaseRetryDelay: 1,
				MaxRetryDelay:  10,
			},
			operation: func() (string, error) {
				return "success", nil
			},
			expectedError: false,
			expectRetries: false,
		},
		{
			name: "retryable error eventually succeeds",
			config: &Config{
				MaxRetries:     2,
				BaseRetryDelay: 1,
				MaxRetryDelay:  10,
			},
			operation: func() func() (string, error) {
				attemptCount := 0
				return func() (string, error) {
					attemptCount++
					if attemptCount < 2 {
						return "", NewGeminiError(503, "service unavailable", nil)
					}
					return "success after retries", nil
				}
			}(),
			expectedError: false,
			expectRetries: true,
		},
		{
			name: "non-retryable error fails immediately",
			config: &Config{
				MaxRetries:     3,
				BaseRetryDelay: 1,
				MaxRetryDelay:  10,
			},
			operation: func() (string, error) {
				return "", NewGeminiError(400, "bad request", nil)
			},
			expectedError: true,
			expectRetries: false,
		},
		{
			name: "max retries exceeded",
			config: &Config{
				MaxRetries:     2,
				BaseRetryDelay: 1,
				MaxRetryDelay:  10,
			},
			operation: func() (string, error) {
				return "", NewGeminiError(503, "service unavailable", nil)
			},
			expectedError: true,
			expectRetries: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gemini{cfg: tt.config}
			ctx := context.Background()

			start := time.Now()
			result, err := g.executeWithRetry(ctx, tt.operation)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			if tt.expectRetries {
				// If retries were expected, operation should have taken some time
				duration := time.Since(start)
				if !tt.expectedError {
					// Only check duration if operation eventually succeeded
					assert.True(t, duration > time.Millisecond*500, "Expected retry delays")
				}
			}
		})
	}
}

func TestGemini_Generate_UnsupportedResponseType(t *testing.T) {
	cfg := &Config{
		APIKey: "test-key",
		Model:  "gemini-1.5-flash",
	}
	g := &Gemini{cfg: cfg}

	req := llm.GenerateRequest{
		ResponseType: "unsupported_type",
		Prompt: models.Prompt{
			Instructions: "Test",
			Request: models.Request{
				ApplicantName:    "Test User",
				ApplicantProfile: "Test Profile",
				JobDescription:   "Test Job",
			},
		},
	}

	_, err := g.Generate(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported response type")
}
