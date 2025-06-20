package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/benidevo/vega/internal/ai/llm"
	"github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/common/logger"
)

const (
	// Match categories
	MatchCategoryExcellent = "Excellent Match"
	MatchCategoryStrong    = "Strong Match"
	MatchCategoryGood      = "Good Match"
	MatchCategoryFair      = "Fair Match"
	MatchCategoryPartial   = "Partial Match"
	MatchCategoryPoor      = "Poor Match"

	// Match descriptions
	MatchDescExcellent = "You are an outstanding fit for the role with minimal gaps."
	MatchDescStrong    = "You have strong qualifications with only minor areas for development."
	MatchDescGood      = "You meet most requirements with some skill gaps that can be addressed."
	MatchDescFair      = "You have potential but may need significant development in key areas."
	MatchDescPartial   = "You have some relevant qualifications but significant gaps exist."
	MatchDescPoor      = "You do not meet the core requirements for this position."
)

// JobMatcherService provides services for matching jobs using a language model provider.
// The 'model' field represents the LLM provider used for job matching operations.
type JobMatcherService struct {
	model llm.Provider
	log   zerolog.Logger
}

// NewJobMatcherService creates and returns a new instance of JobMatcherService
// using the provided llm.Provider as the model.
// The returned JobMatcherService can be used to perform job matching operations.
func NewJobMatcherService(model llm.Provider) *JobMatcherService {
	return &JobMatcherService{
		model: model,
		log:   logger.GetLogger("ai_job_matcher"),
	}
}

// AnalyzeMatch analyzes the match between a job applicant and a job.
func (j *JobMatcherService) AnalyzeMatch(ctx context.Context, req models.Request) (*models.MatchResult, error) {
	start := time.Now()

	j.log.Info().
		Str("applicant", req.ApplicantName).
		Str("operation", "match_analysis").
		Msg("Starting job match analysis")

	if req.ApplicantName == "" || req.ApplicantProfile == "" || req.JobDescription == "" {
		err := models.WrapError(models.ErrValidationFailed, fmt.Errorf("missing required fields: applicant name, profile, and job description are required"))
		j.log.Error().
			Err(err).
			Msg("Match analysis validation failed")
		return nil, err
	}

	prompt := models.Prompt{
		Instructions: "Analyze the job match between the candidate and position",
		Request:      req,
	}

	response, err := j.model.Generate(ctx, llm.GenerateRequest{
		Prompt:       prompt,
		ResponseType: llm.ResponseTypeMatchResult,
	})
	if err != nil {
		j.log.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("Match analysis failed")
		return nil, err
	}

	result, ok := response.Data.(models.MatchResult)
	if !ok {
		err := fmt.Errorf("unexpected response type: expected MatchResult, got %T", response.Data)
		j.log.Error().Err(err).Msg("Type assertion failed")
		return nil, err
	}

	j.validateMatchResult(&result)

	j.log.Info().
		Dur("duration", time.Since(start)).
		Int("match_score", result.MatchScore).
		Int("strengths_count", len(result.Strengths)).
		Int("weaknesses_count", len(result.Weaknesses)).
		Msg("Match analysis completed")

	return &result, nil
}

func (j *JobMatcherService) validateMatchResult(result *models.MatchResult) {
	if result.MatchScore < 0 || result.MatchScore > 100 {
		result.MatchScore = 0
	}

	if len(result.Strengths) == 0 {
		result.Strengths = []string{"No specific strengths identified"}
	}

	if len(result.Weaknesses) == 0 {
		result.Weaknesses = []string{"No specific weaknesses identified"}
	}

	if len(result.Highlights) == 0 {
		result.Highlights = []string{"No specific highlights identified"}
	}

	if result.Feedback == "" {
		result.Feedback = "Unable to provide detailed feedback at this time."
	}
}

// GetMatchCategories returns the match category and its description based on the provided score.
// It evaluates the score and maps it to a predefined category and description:
//   - 90 and above: Excellent
//   - 80 to 89: Strong
//   - 70 to 79: Good
//   - 60 to 69: Fair
//   - 50 to 59: Partial
//   - Below 50: Poor
func (j *JobMatcherService) GetMatchCategories(score int) (string, string) {

	switch {
	case score >= 90:
		return MatchCategoryExcellent, MatchDescExcellent
	case score >= 80:
		return MatchCategoryStrong, MatchDescStrong
	case score >= 70:
		return MatchCategoryGood, MatchDescGood
	case score >= 60:
		return MatchCategoryFair, MatchDescFair
	case score >= 50:
		return MatchCategoryPartial, MatchDescPartial
	default:
		return MatchCategoryPoor, MatchDescPoor
	}
}
