package gemini

import (
	"testing"

	"github.com/benidevo/vega/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		settings *config.Settings
		validate func(*testing.T, *Config)
	}{
		{
			name: "creates config with default values",
			settings: &config.Settings{
				GeminiAPIKey: "test-api-key",
				GeminiModel:  "gemini-1.5-flash",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "test-api-key", cfg.APIKey)
				assert.Equal(t, "gemini-1.5-flash", cfg.Model)
				assert.Equal(t, 8192, cfg.MaxTokens)
				assert.NotNil(t, cfg.Temperature)
				assert.Equal(t, float32(0.4), *cfg.Temperature)

				// Retry configuration
				assert.Equal(t, 3, cfg.MaxRetries)
				assert.Equal(t, 1, cfg.BaseRetryDelay)
				assert.Equal(t, 30, cfg.MaxRetryDelay)

				// Response configuration
				assert.Equal(t, "application/json", cfg.ResponseMIMEType)

				// Cover letter configuration
				assert.Equal(t, "250-400", cfg.DefaultWordRange)

				// Match score configuration
				assert.Equal(t, 0, cfg.MinMatchScore)
				assert.Equal(t, 100, cfg.MaxMatchScore)

				// Default fallback messages
				assert.Equal(t, "No specific strengths identified", cfg.DefaultStrengthsMsg)
				assert.Equal(t, "No specific weaknesses identified", cfg.DefaultWeaknessMsg)
				assert.Equal(t, "No specific highlights identified", cfg.DefaultHighlightMsg)
				assert.Equal(t, "Unable to provide detailed feedback at this time.", cfg.DefaultFeedbackMsg)

				// Advanced generation parameters
				assert.Equal(t, int32(8192), cfg.MaxOutputTokens)
				assert.NotNil(t, cfg.TopP)
				assert.Equal(t, float32(0.9), *cfg.TopP)
				assert.NotNil(t, cfg.TopK)
				assert.Equal(t, float32(40), *cfg.TopK)
				assert.Contains(t, cfg.SystemInstruction, "professional career advisor")
			},
		},
		{
			name: "handles empty API key",
			settings: &config.Settings{
				GeminiAPIKey: "",
				GeminiModel:  "gemini-1.5-pro",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "", cfg.APIKey)
				assert.Equal(t, "gemini-1.5-pro", cfg.Model)
			},
		},
		{
			name: "handles different model",
			settings: &config.Settings{
				GeminiAPIKey: "another-key",
				GeminiModel:  "gemini-1.5-pro-latest",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "another-key", cfg.APIKey)
				assert.Equal(t, "gemini-1.5-pro-latest", cfg.Model)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.settings)

			assert.NotNil(t, cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestFloatPtr(t *testing.T) {
	t.Run("creates float32 pointer", func(t *testing.T) {
		value := float32(0.5)
		ptr := floatPtr(value)

		assert.NotNil(t, ptr)
		assert.Equal(t, value, *ptr)
	})

	t.Run("different values create different pointers", func(t *testing.T) {
		ptr1 := floatPtr(0.1)
		ptr2 := floatPtr(0.2)

		assert.NotEqual(t, ptr1, ptr2)
		assert.Equal(t, float32(0.1), *ptr1)
		assert.Equal(t, float32(0.2), *ptr2)
	})
}
