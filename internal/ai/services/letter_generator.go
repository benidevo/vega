package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/benidevo/ascentio/internal/ai/llm"
	"github.com/benidevo/ascentio/internal/ai/models"
	"github.com/benidevo/ascentio/internal/common/logger"
)

// CoverLetterService provides methods to generate cover letters using a specified LLM provider.
type CoverLetterService struct {
	model llm.Provider
	log   zerolog.Logger
}

// NewCoverLetterService creates and returns a new instance of CoverLetterService
// using the provided llm.Provider as the underlying model.
func NewCoverLetterService(model llm.Provider) *CoverLetterService {
	return &CoverLetterService{
		model: model,
		log:   logger.GetLogger("ai_cover_letter"),
	}
}

// GenerateCoverLetter generates a cover letter based on the provided request.
func (c *CoverLetterService) GenerateCoverLetter(ctx context.Context, req models.Request) (*models.CoverLetter, error) {
	start := time.Now()

	c.log.Info().
		Str("applicant", req.ApplicantName).
		Str("operation", "cover_letter_generation").
		Msg("Starting cover letter generation")

	if req.ApplicantName == "" || req.ApplicantProfile == "" || req.JobDescription == "" {
		err := models.WrapError(models.ErrValidationFailed, fmt.Errorf("missing required fields: applicant name, profile, and job description are required"))
		c.log.Error().
			Err(err).
			Msg("Cover letter generation validation failed")
		return nil, err
	}

	prompt := models.Prompt{
		Instructions: "You are a professional career advisor and expert cover letter writer.",
		Request:      req,
	}

	response, err := c.model.Generate(ctx, llm.GenerateRequest{
		Prompt:       prompt,
		ResponseType: llm.ResponseTypeCoverLetter,
	})
	if err != nil {
		c.log.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Cover letter generation failed")
		return nil, err
	}

	result, ok := response.Data.(models.CoverLetter)
	if !ok {
		err := fmt.Errorf("unexpected response type: expected CoverLetter, got %T", response.Data)
		c.log.Error().Err(err).Msg("Type assertion failed")
		return nil, err
	}

	if err := c.validateCoverLetter(&result); err != nil {
		c.log.Error().
			Err(err).
			Msg("Cover letter validation failed")
		return nil, err
	}

	c.log.Info().
		Dur("duration", time.Since(start)).
		Int("content_length", len(result.Content)).
		Str("format", string(result.Format)).
		Msg("Cover letter generation completed")

	return &result, nil
}

func (c *CoverLetterService) validateCoverLetter(letter *models.CoverLetter) error {
	if letter.Content == "" {
		return models.WrapError(models.ErrValidationFailed, fmt.Errorf("generated cover letter is empty"))
	}

	if letter.Format == "" {
		letter.Format = models.CoverLetterTypePlainText
	}

	return nil
}
