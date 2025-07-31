package services

import (
	"context"

	"github.com/benidevo/vega/internal/ai/llm"
	"github.com/benidevo/vega/internal/ai/models"
)

// CVGenerator defines the interface for CV generation
type CVGenerator interface {
	Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error)
}

// CVParser defines the interface for CV parsing
type CVParser interface {
	Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error)
}

// JobMatcher defines the interface for job matching
type JobMatcher interface {
	Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error)
}

// LetterGenerator defines the interface for cover letter generation
type LetterGenerator interface {
	Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error)
}

// CVGeneratorServiceInterface defines the public interface of CVGeneratorService
type CVGeneratorServiceInterface interface {
	GenerateCV(ctx context.Context, req models.Request, jobID int, jobTitle string) (*models.GeneratedCV, error)
}

// CVParserServiceInterface defines the public interface of CVParserService
type CVParserServiceInterface interface {
	ParseCV(ctx context.Context, cvContent string) (*models.CVParsingResult, error)
}

// JobMatcherServiceInterface defines the public interface of JobMatcherService
type JobMatcherServiceInterface interface {
	AnalyzeMatch(ctx context.Context, req models.Request) (*models.MatchResult, error)
	GetMatchCategories(score int) (string, string)
}

// LetterGeneratorServiceInterface defines the public interface of LetterGeneratorService
type LetterGeneratorServiceInterface interface {
	GenerateCoverLetter(ctx context.Context, req models.Request) (*models.CoverLetter, error)
}
