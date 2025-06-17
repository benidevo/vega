package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"google.golang.org/genai"

	"github.com/benidevo/ascentio/internal/ai/llm"
	"github.com/benidevo/ascentio/internal/ai/models"
)

// GeminiFlash represents a client for interacting with the Gemini AI service.
//
// It holds a reference to the underlying genai.Client and configuration settings.
type GeminiFlash struct {
	client *genai.Client
	cfg    *Config
}

// New creates and initializes a new GeminiFlash client using the provided context and configuration.
// It returns a pointer to the GeminiFlash instance or an error if the client initialization fails.
func New(ctx context.Context, cfg *Config) (*GeminiFlash, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI})

	if err != nil {
		return nil, WrapError(ErrClientInitFailed, err)
	}

	return &GeminiFlash{
		client: client,
		cfg:    cfg,
	}, nil
}

// Generate implements the Provider interface for the GeminiFlash client.
// It processes requests based on the ResponseType and returns appropriate data.
func (g *GeminiFlash) Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error) {
	start := time.Now()

	switch request.ResponseType {
	case llm.ResponseTypeCoverLetter:
		return g.generateCoverLetter(ctx, request.Prompt, start)
	case llm.ResponseTypeMatchResult:
		return g.generateMatchResult(ctx, request.Prompt, start)
	default:
		return llm.GenerateResponse{}, fmt.Errorf("unsupported response type: %s", request.ResponseType)
	}
}

// generateCoverLetter generates a cover letter based on the provided prompt.
func (g *GeminiFlash) generateCoverLetter(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	coverLetterPrompt := prompt.ToCoverLetterPrompt(g.cfg.DefaultWordRange)

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		resp, err := g.client.Models.GenerateContent(ctx, g.cfg.Model, genai.Text(coverLetterPrompt), &genai.GenerateContentConfig{
			Temperature:       g.cfg.Temperature,
			ResponseMIMEType:  g.cfg.ResponseMIMEType,
			ResponseSchema:    g.getCoverLetterSchema(),
			MaxOutputTokens:   g.cfg.MaxOutputTokens,
			TopP:              g.cfg.TopP,
			TopK:              g.cfg.TopK,
			StopSequences:     g.cfg.CoverLetterStopSeqs,
			SystemInstruction: g.buildSystemInstruction(),
		})
		if err != nil {
			return "", err
		}
		return resp.Text(), nil
	})

	if err != nil {
		return llm.GenerateResponse{}, WrapError(ErrCoverLetterGenFailed, err)
	}

	coverLetter, err := g.parseCoverLetterJSON(result)
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	return llm.GenerateResponse{
		Data:     coverLetter,
		Duration: time.Since(start),
		Tokens:   0,
	}, nil
}

// generateMatchResult analyzes the given prompt using the GeminiFlash model and returns a match result.
func (g *GeminiFlash) generateMatchResult(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	matchPrompt := prompt.ToMatchAnalysisPrompt(g.cfg.MinMatchScore, g.cfg.MaxMatchScore)

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		resp, err := g.client.Models.GenerateContent(ctx, g.cfg.Model, genai.Text(matchPrompt), &genai.GenerateContentConfig{
			Temperature:       g.cfg.Temperature,
			ResponseMIMEType:  g.cfg.ResponseMIMEType,
			ResponseSchema:    g.getMatchAnalysisSchema(),
			MaxOutputTokens:   g.cfg.MaxOutputTokens,
			TopP:              g.cfg.TopP,
			TopK:              g.cfg.TopK,
			StopSequences:     g.cfg.MatchAnalysisStopSeqs,
			SystemInstruction: g.buildSystemInstruction(),
		})
		if err != nil {
			return "", err
		}
		return resp.Text(), nil
	})

	if err != nil {
		return llm.GenerateResponse{}, WrapError(ErrMatchAnalysisFailed, err)
	}

	matchResult, err := g.parseMatchResultJSON(result)
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	return llm.GenerateResponse{
		Data:     matchResult,
		Duration: time.Since(start),
		Tokens:   0,
	}, nil
}

func (g *GeminiFlash) executeWithRetry(ctx context.Context, operation func() (string, error)) (string, error) {
	maxRetries := g.cfg.MaxRetries
	baseDelay := time.Duration(g.cfg.BaseRetryDelay) * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			maxDelay := time.Duration(g.cfg.MaxRetryDelay) * time.Second
			if delay > maxDelay {
				delay = maxDelay
			}

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !IsRetryableError(err) || attempt == maxRetries {
			break
		}
	}

	if lastErr != nil {
		return "", WrapError(ErrMaxRetriesExceeded, lastErr)
	}
	return "", lastErr
}

func (g *GeminiFlash) getMatchAnalysisSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"matchScore": {
				Type:        genai.TypeInteger,
				Description: fmt.Sprintf("Match score from %d-%d", g.cfg.MinMatchScore, g.cfg.MaxMatchScore),
			},
			"strengths": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				Description: "List of candidate strengths",
			},
			"weaknesses": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				Description: "List of areas for improvement",
			},
			"highlights": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				Description: "List of standout qualifications",
			},
			"feedback": {
				Type:        genai.TypeString,
				Description: "Overall assessment and recommendations",
			},
		},
		PropertyOrdering: []string{"matchScore", "strengths", "weaknesses", "highlights", "feedback"},
		Required:         []string{"matchScore", "strengths", "weaknesses", "highlights", "feedback"},
	}
}

func (g *GeminiFlash) parseMatchResultJSON(jsonResponse string) (models.MatchResult, error) {
	var result models.MatchResult
	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		return models.MatchResult{}, WrapError(ErrResponseParseFailed, err)
	}

	if result.MatchScore < g.cfg.MinMatchScore || result.MatchScore > g.cfg.MaxMatchScore {
		result.MatchScore = g.cfg.MinMatchScore
	}

	if len(result.Strengths) == 0 {
		result.Strengths = []string{g.cfg.DefaultStrengthsMsg}
	}

	if len(result.Weaknesses) == 0 {
		result.Weaknesses = []string{g.cfg.DefaultWeaknessMsg}
	}

	if len(result.Highlights) == 0 {
		result.Highlights = []string{g.cfg.DefaultHighlightMsg}
	}

	if result.Feedback == "" {
		result.Feedback = g.cfg.DefaultFeedbackMsg
	}

	return result, nil
}

func (g *GeminiFlash) getCoverLetterSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"content": {
				Type:        genai.TypeString,
				Description: "The complete cover letter content",
			},
		},
		PropertyOrdering: []string{"content"},
		Required:         []string{"content"},
	}
}

func (g *GeminiFlash) parseCoverLetterJSON(jsonResponse string) (models.CoverLetter, error) {
	var result models.CoverLetter
	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		return models.CoverLetter{}, WrapError(ErrResponseParseFailed, err)
	}

	if result.Content == "" {
		return models.CoverLetter{}, ErrEmptyResponse
	}

	result.Format = models.CoverLetterTypePlainText

	return result, nil
}

func (g *GeminiFlash) buildSystemInstruction() *genai.Content {
	if g.cfg.SystemInstruction == "" {
		return nil
	}

	contents := genai.Text(g.cfg.SystemInstruction)
	if len(contents) > 0 {
		return contents[0]
	}

	return nil
}
