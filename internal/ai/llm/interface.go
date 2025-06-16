package llm

import (
	"context"

	"github.com/benidevo/ascentio/internal/ai/models"
)

// LLM defines the interface for interacting with a Large Language Model (LLM).
// It provides methods for generating cover letters and analyzing match results
// based on the provided prompt.
//
// GenerateCoverLetter generates a cover letter using the given prompt and returns
// the generated cover letter or an error.
//
// AnalyzeMatch analyzes the match between the prompt and some criteria, returning
// a match result or an error.
type LLM interface {
	GenerateCoverLetter(ctx context.Context, prompt models.Prompt) (models.CoverLetter, error)
	AnalyzeMatch(ctx context.Context, prompt models.Prompt) (models.MatchResult, error)
}
