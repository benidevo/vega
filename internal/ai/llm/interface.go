package llm

import (
	"context"
	"time"

	"github.com/benidevo/vega/internal/ai/models"
)

// Provider defines the interface for interacting with a Large Language Model (LLM).
type Provider interface {
	// Generate sends a request to the LLM and returns a typed response.
	// The concrete type T should be one of the supported response types.
	Generate(ctx context.Context, request GenerateRequest) (GenerateResponse, error)
}

// GenerateRequest encapsulates all LLM request parameters
type GenerateRequest struct {
	Prompt       models.Prompt
	ResponseType ResponseType
	Options      map[string]any // Provider-specific options
}

// ResponseType indicates the expected response format
type ResponseType string

const (
	ResponseTypeCoverLetter ResponseType = "cover_letter"
	ResponseTypeMatchResult ResponseType = "match_result"
	ResponseTypeCVParsing   ResponseType = "cv_parsing"
	ResponseTypeCV          ResponseType = "cv_generation"
)

// GenerateResponse wraps the LLM response with metadata
type GenerateResponse struct {
	Data     any // Will be CoverLetter, MatchResult, etc
	Tokens   int
	Duration time.Duration
	Metadata map[string]any
}
