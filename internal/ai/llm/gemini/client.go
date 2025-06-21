package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/benidevo/vega/internal/ai/llm"
	"github.com/benidevo/vega/internal/ai/models"
)

// Gemini represents a client for interacting with the Gemini AI service.
//
// It holds a reference to the underlying genai.Client and configuration settings.
type Gemini struct {
	client *genai.Client
	cfg    *Config
}

// New creates and initializes a new Gemini client using the provided context and configuration.
// It returns a pointer to the Gemini instance or an error if the client initialization fails.
func New(ctx context.Context, cfg *Config) (*Gemini, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI})

	if err != nil {
		return nil, WrapError(ErrClientInitFailed, err)
	}

	return &Gemini{
		client: client,
		cfg:    cfg,
	}, nil
}

// Generate implements the Provider interface for the Gemini client.
// It processes requests based on the ResponseType and returns appropriate data.
func (g *Gemini) Generate(ctx context.Context, request llm.GenerateRequest) (llm.GenerateResponse, error) {
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
func (g *Gemini) generateCoverLetter(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	coverLetterPrompt := prompt.ToCoverLetterPrompt(g.cfg.DefaultWordRange)

	temperature := prompt.GetOptimalTemperature("cover_letter")

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		resp, err := g.client.Models.GenerateContent(ctx, g.cfg.Model, genai.Text(coverLetterPrompt), &genai.GenerateContentConfig{
			Temperature:       &temperature,
			ResponseMIMEType:  g.cfg.ResponseMIMEType,
			ResponseSchema:    g.getCoverLetterSchema(),
			MaxOutputTokens:   g.cfg.MaxOutputTokens,
			TopP:              g.cfg.TopP,
			TopK:              g.cfg.TopK,
			SystemInstruction: g.buildSystemInstruction(),
		})
		if err != nil {
			return "", fmt.Errorf("generate content error: %w", err)
		}

		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("no candidates in response")
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			return "", fmt.Errorf("no content in response candidate")
		}

		var responseText string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				responseText += part.Text
			}
		}

		return responseText, nil
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
		Metadata: map[string]interface{}{
			"temperature": temperature,
			"enhanced":    prompt.UseEnhancedTemplates,
		},
	}, nil
}

// generateMatchResult analyzes the given prompt using the Gemini model and returns a match result.
func (g *Gemini) generateMatchResult(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	matchPrompt := prompt.ToMatchAnalysisPrompt(g.cfg.MinMatchScore, g.cfg.MaxMatchScore)

	temperature := prompt.GetOptimalTemperature("job_match")

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		resp, err := g.client.Models.GenerateContent(ctx, g.cfg.Model, genai.Text(matchPrompt), &genai.GenerateContentConfig{
			Temperature:       &temperature,
			ResponseMIMEType:  g.cfg.ResponseMIMEType,
			ResponseSchema:    g.getMatchAnalysisSchema(),
			MaxOutputTokens:   g.cfg.MaxOutputTokens,
			TopP:              g.cfg.TopP,
			TopK:              g.cfg.TopK,
			SystemInstruction: g.buildSystemInstruction(),
		})
		if err != nil {
			return "", fmt.Errorf("generate content error: %w", err)
		}

		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("no candidates in response")
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			return "", fmt.Errorf("no content in response candidate")
		}

		var responseText string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				responseText += part.Text
			}
		}

		return responseText, nil
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
		Metadata: map[string]interface{}{
			"temperature": temperature,
			"enhanced":    prompt.UseEnhancedTemplates,
		},
	}, nil
}

func (g *Gemini) executeWithRetry(ctx context.Context, operation func() (string, error)) (string, error) {
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

func (g *Gemini) getMatchAnalysisSchema() *genai.Schema {
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

func (g *Gemini) parseMatchResultJSON(jsonResponse string) (models.MatchResult, error) {
	cleanJSON := g.extractJSON(jsonResponse)

	var result models.MatchResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
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

func (g *Gemini) getCoverLetterSchema() *genai.Schema {
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

func (g *Gemini) parseCoverLetterJSON(jsonResponse string) (models.CoverLetter, error) {
	cleanJSON := g.extractJSON(jsonResponse)

	var result models.CoverLetter
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return models.CoverLetter{}, WrapError(ErrResponseParseFailed, err)
	}

	if result.Content == "" {
		return models.CoverLetter{}, ErrEmptyResponse
	}

	result.Format = models.CoverLetterTypePlainText

	return result, nil
}

func (g *Gemini) buildSystemInstruction() *genai.Content {
	if g.cfg.SystemInstruction == "" {
		return nil
	}

	contents := genai.Text(g.cfg.SystemInstruction)
	if len(contents) > 0 {
		return contents[0]
	}

	return nil
}

// extractJSON attempts to extract JSON content from a response that may contain extra text
func (g *Gemini) extractJSON(response string) string {
	response = strings.TrimSpace(response)

	// Look for JSON object boundaries
	startIdx := strings.Index(response, "{")
	if startIdx == -1 {
		return response // No JSON found, return as is
	}

	// Find the matching closing brace
	braceCount := 0
	endIdx := -1

	for i := startIdx; i < len(response) && endIdx == -1; i++ {
		switch response[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				endIdx = i
			}
		}
	}

	if endIdx == -1 {
		return response // No matching brace found, return as is
	}

	return response[startIdx : endIdx+1]
}
