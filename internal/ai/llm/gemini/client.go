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
	case llm.ResponseTypeCVParsing:
		return g.parseCVContent(ctx, request.Prompt, start)
	default:
		return llm.GenerateResponse{}, fmt.Errorf("unsupported response type: %s", request.ResponseType)
	}
}

// generateCoverLetter generates a cover letter based on the provided prompt.
func (g *Gemini) generateCoverLetter(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	coverLetterPrompt := prompt.ToCoverLetterPrompt(g.cfg.DefaultWordRange)

	temperature := prompt.GetOptimalTemperature("cover_letter")

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		model := g.cfg.GetModelForTask("cover_letter")
		resp, err := g.client.Models.GenerateContent(ctx, model, genai.Text(coverLetterPrompt), &genai.GenerateContentConfig{
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
			"model":       g.cfg.GetModelForTask("cover_letter"),
			"task_type":   "cover_letter",
		},
	}, nil
}

// generateMatchResult analyzes the given prompt using the Gemini model and returns a match result.
func (g *Gemini) generateMatchResult(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	matchPrompt := prompt.ToMatchAnalysisPrompt(g.cfg.MinMatchScore, g.cfg.MaxMatchScore)

	temperature := prompt.GetOptimalTemperature("job_match")

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		model := g.cfg.GetModelForTask("job_analysis")
		resp, err := g.client.Models.GenerateContent(ctx, model, genai.Text(matchPrompt), &genai.GenerateContentConfig{
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
			"model":       g.cfg.GetModelForTask("job_analysis"),
			"task_type":   "job_analysis",
		},
	}, nil
}

// parseCVContent parses CV text and extracts structured information
func (g *Gemini) parseCVContent(ctx context.Context, prompt models.Prompt, start time.Time) (llm.GenerateResponse, error) {
	cvPrompt := g.buildCVParsingPrompt(prompt)

	temperature := float32(0.1) // low temperature for consistent parsing

	result, err := g.executeWithRetry(ctx, func() (string, error) {
		model := g.cfg.GetModelForTask("cv_parsing")
		resp, err := g.client.Models.GenerateContent(ctx, model, genai.Text(cvPrompt), &genai.GenerateContentConfig{
			Temperature:       &temperature,
			ResponseMIMEType:  g.cfg.ResponseMIMEType,
			ResponseSchema:    g.getCVParsingSchema(),
			MaxOutputTokens:   g.cfg.MaxOutputTokens,
			TopP:              g.cfg.TopP,
			TopK:              g.cfg.TopK,
			SystemInstruction: g.buildCVParsingSystemInstruction(),
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

	cvResult, err := g.parseCVJSON(result)
	if err != nil {
		return llm.GenerateResponse{}, err
	}

	return llm.GenerateResponse{
		Data:     cvResult,
		Duration: time.Since(start),
		Tokens:   0,
		Metadata: map[string]any{
			"temperature": temperature,
			"model":       g.cfg.GetModelForTask("cv_parsing"),
			"task_type":   "cv_parsing",
			"method":      "gemini_cv_parsing",
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

func (g *Gemini) buildCVParsingPrompt(prompt models.Prompt) string {
	cvText := prompt.CVText

	return fmt.Sprintf(`You are an expert CV/Resume parser and validator. First, determine if the provided text is actually a CV/Resume document. Then extract structured information if valid.

VALIDATION RULES:
- The document MUST be a CV, Resume, or professional profile
- It should contain career-related information (work experience, education, or skills)
- Reject documents that are: police reports, medical records, legal documents, news articles, fiction, academic papers, manuals, or any non-career documents
- If the document is NOT a valid CV/Resume, return: {"isValid": false, "reason": "explanation"}

PARSING INSTRUCTIONS (only if document is valid):
- Extract personal information (name, contact details, location, professional title)
- Parse work experience entries with company, title, dates, and descriptions
- Parse education entries with institution, degree, field of study, and dates
- Extract skills as a list
- For dates, use formats: "YYYY-MM" for month precision, "YYYY" for year precision, "Present" for current positions
- Be precise and don't make up information that's not clearly stated
- If information is ambiguous or missing, use empty strings rather than guessing
- Always include: {"isValid": true} in your response for valid CVs

Document Text:
%s

Please return the information in the exact JSON schema format specified.`, cvText)
}

func (g *Gemini) buildCVParsingSystemInstruction() *genai.Content {
	instruction := `You are a precise CV/Resume parsing and validation system. Your primary task is to first validate that the document is actually a CV/Resume, then extract structured information if valid. Always include an "isValid" field in your response. Reject any documents that are not career-related (police reports, medical records, etc.). For valid CVs, focus on accuracy and completeness. When dates are unclear, prefer broader ranges (year-only) over specific months. Do not hallucinate or guess information that is not explicitly stated.`

	contents := genai.Text(instruction)
	if len(contents) > 0 {
		return contents[0]
	}
	return nil
}

func (g *Gemini) getCVParsingSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"isValid": {
				Type:        genai.TypeBoolean,
				Description: "Whether the document is a valid CV/Resume",
			},
			"reason": {
				Type:        genai.TypeString,
				Description: "Reason for rejection if document is not valid (only required when isValid is false)",
			},
			"personalInfo": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"firstName": {Type: genai.TypeString, Description: "First name"},
					"lastName":  {Type: genai.TypeString, Description: "Last name"},
					"email":     {Type: genai.TypeString, Description: "Email address"},
					"phone":     {Type: genai.TypeString, Description: "Phone number"},
					"location":  {Type: genai.TypeString, Description: "Location/Address"},
					"title":     {Type: genai.TypeString, Description: "Professional title or role"},
				},
				Required: []string{"firstName", "lastName"},
			},
			"workExperience": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"company":     {Type: genai.TypeString, Description: "Company name"},
						"title":       {Type: genai.TypeString, Description: "Job title"},
						"location":    {Type: genai.TypeString, Description: "Job location"},
						"startDate":   {Type: genai.TypeString, Description: "Start date (YYYY-MM or YYYY format)"},
						"endDate":     {Type: genai.TypeString, Description: "End date (YYYY-MM, YYYY, or 'Present')"},
						"description": {Type: genai.TypeString, Description: "Job description/responsibilities"},
					},
					Required: []string{"company", "title", "startDate"},
				},
			},
			"education": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"institution":  {Type: genai.TypeString, Description: "Educational institution name"},
						"degree":       {Type: genai.TypeString, Description: "Degree type (e.g., BA, BS, Master)"},
						"fieldOfStudy": {Type: genai.TypeString, Description: "Field of study/major"},
						"startDate":    {Type: genai.TypeString, Description: "Start date (YYYY-MM or YYYY format)"},
						"endDate":      {Type: genai.TypeString, Description: "End date (YYYY-MM or YYYY format)"},
					},
					Required: []string{"institution", "degree", "startDate"},
				},
			},
			"skills": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
				Description: "List of skills and technologies",
			},
		},
		PropertyOrdering: []string{"isValid", "reason", "personalInfo", "workExperience", "education", "skills"},
		Required:         []string{"isValid"},
	}
}

func (g *Gemini) parseCVJSON(jsonResponse string) (models.CVParsingResult, error) {
	cleanJSON := g.extractJSON(jsonResponse)

	var result models.CVParsingResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return models.CVParsingResult{}, WrapError(ErrResponseParseFailed, err)
	}

	// Check if document is valid first
	if !result.IsValid {
		reason := result.Reason
		if reason == "" {
			reason = "Document is not a valid CV/Resume"
		}
		return models.CVParsingResult{}, fmt.Errorf("invalid document: %s", reason)
	}

	// For valid CVs, validate required fields
	if result.PersonalInfo.FirstName == "" && result.PersonalInfo.LastName == "" {
		return models.CVParsingResult{}, fmt.Errorf("no name found in CV")
	}

	// Ensure arrays are never nil for valid CVs
	if result.WorkExperience == nil {
		result.WorkExperience = []models.WorkExperience{}
	}
	if result.Education == nil {
		result.Education = []models.Education{}
	}
	if result.Skills == nil {
		result.Skills = []string{}
	}

	return result, nil
}
