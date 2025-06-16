package gemini

import "github.com/benidevo/ascentio/internal/config"

// Config holds the configuration for the Gemini LLM client.
type Config struct {
	// APIKey is the API key for authenticating with the Gemini API.
	APIKey string
	// MaxTokens is the maximum number of tokens to generate in the response.
	MaxTokens int
	// Model is the model to use for generating responses. It should be a valid Gemini
	// model identifier, such as "gemini-1.5-flash".
	Model string
	// Temperature controls the randomness of the output. Higher values (e.g., 0.8) make the output more random,
	// while lower values (e.g., 0.2) make it more focused and deterministic.
	Temperature *float32
}

// NewConfig creates a new Config instance with the provided API key and optional parameters.
func NewConfig(cfg *config.Settings) *Config {
	defaultTemp := float32(0.4)

	return &Config{
		APIKey:      cfg.GeminiAPIKey,
		MaxTokens:   1024,
		Model:       cfg.GeminiModel,
		Temperature: &defaultTemp,
	}
}
