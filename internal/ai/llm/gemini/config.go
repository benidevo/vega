package gemini

import (
	"github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/config"
)

// Config holds the configuration for the Gemini LLM client.
type Config struct {
	// APIKey is the API key for authenticating with the Gemini API.
	APIKey string
	// MaxTokens is the maximum number of tokens to generate in the response.
	MaxTokens int
	// Model is the default model to use for generating responses. It should be a valid Gemini
	// model identifier, such as "gemini-1.5-flash".
	Model string
	// Task-specific models for optimal performance and cost
	ModelCVParsing   string // Fast model for CV parsing (e.g., gemini-1.5-flash)
	ModelJobAnalysis string // Advanced model for job analysis (e.g., gemini-2.5-flash)
	ModelCoverLetter string // Advanced model for cover letter generation (e.g., gemini-2.5-flash)
	// Temperature controls the randomness of the output. Higher values (e.g., 0.8) make the output more random,
	// while lower values (e.g., 0.2) make it more focused and deterministic.
	Temperature *float32

	// Retry configuration
	MaxRetries     int
	BaseRetryDelay int // seconds
	MaxRetryDelay  int // seconds

	// Response configuration
	ResponseMIMEType string

	// Cover letter configuration
	DefaultWordRange string

	// Match score configuration
	MinMatchScore int
	MaxMatchScore int

	// Default values for fallbacks
	DefaultStrengthsMsg string
	DefaultWeaknessMsg  string
	DefaultHighlightMsg string
	DefaultFeedbackMsg  string

	// Advanced generation parameters
	MaxOutputTokens   int32
	TopP              *float32
	TopK              *float32
	SystemInstruction string
}

// NewConfig creates a new Config instance with the provided API key and optional parameters.
func NewConfig(cfg *config.Settings) *Config {
	defaultTemp := float32(0.4)

	return &Config{
		APIKey:           cfg.GeminiAPIKey,
		MaxTokens:        8192,
		Model:            cfg.GeminiModel,
		ModelCVParsing:   cfg.GeminiModelCVParsing,
		ModelJobAnalysis: cfg.GeminiModelJobAnalysis,
		ModelCoverLetter: cfg.GeminiModelCoverLetter,
		Temperature:      &defaultTemp,

		// Retry configuration
		MaxRetries:     3,
		BaseRetryDelay: 1,  // 1 second
		MaxRetryDelay:  30, // 30 seconds

		// Response configuration
		ResponseMIMEType: "application/json",

		// Cover letter configuration
		DefaultWordRange: "150-250",

		// Match score configuration
		MinMatchScore: 0,
		MaxMatchScore: 100,

		// Default fallback messages
		DefaultStrengthsMsg: "No specific strengths identified",
		DefaultWeaknessMsg:  "No specific weaknesses identified",
		DefaultHighlightMsg: "No specific highlights identified",
		DefaultFeedbackMsg:  "Unable to provide detailed feedback at this time.",

		// Advanced generation parameters
		MaxOutputTokens:   6000,
		TopP:              floatPtr(0.9),
		TopK:              floatPtr(40),
		SystemInstruction: "You are a professional career advisor and expert writer. Always provide helpful, accurate, and constructive feedback. IMPORTANT: For job matching, use experience-based evaluation - candidates with 2+ years experience should be evaluated primarily on work history and practical skills, with education as secondary. Entry-level candidates (<2 years) should be evaluated with education carrying more weight. BE MODERATELY LENIENT: Value similar and transferable skills, not just exact matches. Award modest bonuses for related technologies and cross-domain experience. When responding with JSON, output ONLY valid JSON without any preamble, explanation, or additional text. Do not include phrases like 'Here is the JSON' or any other text before or after the JSON object.",
	}
}

// GetModelForTask returns the appropriate model for the given task type
func (c *Config) GetModelForTask(taskType string) string {
	switch models.AITaskType(taskType) {
	case models.TaskTypeCVParsing:
		if c.ModelCVParsing != "" {
			return c.ModelCVParsing
		}
	case models.TaskTypeJobAnalysis, models.TaskTypeMatchResult:
		if c.ModelJobAnalysis != "" {
			return c.ModelJobAnalysis
		}
	case models.TaskTypeCoverLetter:
		if c.ModelCoverLetter != "" {
			return c.ModelCoverLetter
		}
	case models.TaskTypeCVGeneration:
		if c.ModelCoverLetter != "" {
			return c.ModelCoverLetter
		}
	}

	return c.Model
}

// Helper function to create float32 pointers
func floatPtr(f float32) *float32 {
	return &f
}
