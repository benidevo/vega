package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	aimodels "github.com/benidevo/ascentio/internal/ai/models"
	"github.com/benidevo/ascentio/internal/job/models"
	settingsmodels "github.com/benidevo/ascentio/internal/settings/models"
)

// AnalyzeJobMatch performs AI-powered job matching analysis between a user profile and a job.
func (s *JobService) AnalyzeJobMatch(ctx context.Context, userID, jobID int) (*models.JobMatchAnalysis, error) {
	userRef := fmt.Sprintf("user_%d", userID)

	s.log.Debug().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Str("operation", "job_match_analysis").
		Msg("Starting job match analysis")

	if s.aiService == nil {
		s.log.Error().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "ai_service_unavailable").
			Msg("AI service not available")
		return nil, models.ErrAIServiceUnavailable
	}

	if s.settingsService == nil {
		s.log.Error().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "settings_service_unavailable").
			Msg("Settings service not available")
		return nil, models.ErrProfileServiceRequired
	}

	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "job_not_found").
			Msg("Job not found for match analysis")
		return nil, err
	}

	profile, err := s.settingsService.GetProfileSettings(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_fetch_failed").
			Msg("Failed to get user profile for match analysis")
		return nil, err
	}

	if err := s.validateProfileForAI(profile); err != nil {
		s.log.Warn().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_incomplete").
			Msg("Profile incomplete for AI analysis")
		return nil, err
	}

	aiRequest := s.buildAIRequest(job, profile)
	aiResult, err := s.aiService.JobMatcher.AnalyzeMatch(ctx, aiRequest)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "ai_analysis_failed").
			Msg("Job match analysis failed")
		return nil, err
	}

	result := s.convertToJobMatchAnalysis(aiResult, userID, jobID)

	s.log.Info().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Int("match_score", result.MatchScore).
		Str("operation", "job_match_analysis").
		Bool("success", true).
		Msg("Job match analysis completed")

	return result, nil
}

// GenerateCoverLetter generates an AI-powered cover letter for a specific job application.
func (s *JobService) GenerateCoverLetter(ctx context.Context, userID, jobID int) (*models.CoverLetter, error) {
	userRef := fmt.Sprintf("user_%d", userID)

	s.log.Debug().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Str("operation", "cover_letter_generation").
		Msg("Starting cover letter generation")

	if s.aiService == nil {
		s.log.Error().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "ai_service_unavailable").
			Msg("AI service not available")
		return nil, models.ErrAIServiceUnavailable
	}

	if s.settingsService == nil {
		s.log.Error().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "settings_service_unavailable").
			Msg("Settings service not available")
		return nil, models.ErrProfileServiceRequired
	}

	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "job_not_found").
			Msg("Job not found for cover letter generation")
		return nil, err
	}

	profile, err := s.settingsService.GetProfileSettings(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_fetch_failed").
			Msg("Failed to get user profile for cover letter generation")
		return nil, err
	}

	if err := s.validateProfileForAI(profile); err != nil {
		s.log.Warn().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_incomplete").
			Msg("Profile incomplete for AI cover letter generation")
		return nil, err
	}

	aiRequest := s.buildAIRequest(job, profile)
	aiResult, err := s.aiService.CoverLetterGenerator.GenerateCoverLetter(ctx, aiRequest)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "ai_generation_failed").
			Msg("Cover letter generation failed")
		return nil, err
	}

	result := s.convertToCoverLetter(aiResult, userID, jobID)

	s.log.Info().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Str("operation", "cover_letter_generation").
		Bool("success", true).
		Msg("Cover letter generation completed")

	return result, nil
}

// buildAIRequest creates an AI request from job and profile data.
func (s *JobService) buildAIRequest(job *models.Job, profile *settingsmodels.Profile) aimodels.Request {
	applicantName := "Applicant"
	if profile.FirstName != "" || profile.LastName != "" {
		applicantName = fmt.Sprintf("%s %s", profile.FirstName, profile.LastName)
	}

	profileSummary := s.buildProfileSummary(profile)

	jobDescription := fmt.Sprintf("Position: %s\nCompany: %s\n\nDescription:\n%s",
		job.Title, job.Company.Name, job.Description)

	if len(job.RequiredSkills) > 0 {
		jobDescription += fmt.Sprintf("\n\nRequired Skills: %s", strings.Join(job.RequiredSkills, ", "))
	}

	return aimodels.Request{
		ApplicantName:    applicantName,
		ApplicantProfile: profileSummary,
		JobDescription:   jobDescription,
		ExtraContext:     profile.Context,
	}
}

// buildProfileSummary creates a comprehensive profile summary from user profile data.
func (s *JobService) buildProfileSummary(profile *settingsmodels.Profile) string {
	var summary strings.Builder

	if profile.Title != "" {
		summary.WriteString(fmt.Sprintf("Current Title: %s\n", profile.Title))
	}

	if profile.Industry.String() != "" {
		summary.WriteString(fmt.Sprintf("Industry: %s\n", profile.Industry.String()))
	}

	if profile.Location != "" {
		summary.WriteString(fmt.Sprintf("Location: %s\n", profile.Location))
	}

	if profile.CareerSummary != "" {
		summary.WriteString(fmt.Sprintf("\nCareer Summary:\n%s\n", profile.CareerSummary))
	}

	if len(profile.Skills) > 0 {
		summary.WriteString(fmt.Sprintf("\nSkills: %s\n", strings.Join(profile.Skills, ", ")))
	}

	if len(profile.WorkExperience) > 0 {
		summary.WriteString("\nWork Experience:\n")
		for i, exp := range profile.WorkExperience {
			if i >= 5 { // Limit to most recent 5 experiences
				break
			}

			endDate := "Present"
			if exp.EndDate != nil {
				endDate = exp.EndDate.Format("2006")
			}

			summary.WriteString(fmt.Sprintf("- %s at %s (%s - %s)\n",
				exp.Title, exp.Company, exp.StartDate.Format("2006"), endDate))

			if exp.Description != "" {
				// Truncate long descriptions
				desc := exp.Description
				if len(desc) > 200 {
					desc = desc[:200] + "..."
				}
				summary.WriteString(fmt.Sprintf("  %s\n", desc))
			}
		}
	}

	if len(profile.Education) > 0 {
		summary.WriteString("\nEducation:\n")
		for i, edu := range profile.Education {
			if i >= 3 { // Limit to most recent 3 education entries
				break
			}

			summary.WriteString(fmt.Sprintf("- %s in %s from %s (%s)\n",
				edu.Degree, edu.FieldOfStudy, edu.Institution, edu.StartDate.Format("2006")))
		}
	}

	if len(profile.Certifications) > 0 {
		summary.WriteString("\nCertifications:\n")
		for i, cert := range profile.Certifications {
			if i >= 5 { // Limit to most recent 5 certifications
				break
			}

			summary.WriteString(fmt.Sprintf("- %s from %s (%s)\n",
				cert.Name, cert.IssuingOrg, cert.IssueDate.Format("2006")))
		}
	}

	if profile.Context != "" {
		summary.WriteString(fmt.Sprintf("\nAdditional Context:\n%s\n", profile.Context))
	}

	return summary.String()
}

// convertToJobMatchAnalysis converts AI match result to job domain model.
func (s *JobService) convertToJobMatchAnalysis(aiResult *aimodels.MatchResult, userID, jobID int) *models.JobMatchAnalysis {
	now := time.Now().UTC()

	return &models.JobMatchAnalysis{
		JobID:      jobID,
		UserID:     userID,
		MatchScore: aiResult.MatchScore,
		Strengths:  aiResult.Strengths,
		Weaknesses: aiResult.Weaknesses,
		Highlights: aiResult.Highlights,
		Feedback:   aiResult.Feedback,
		AnalyzedAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// convertToCoverLetter converts AI cover letter result to job domain model.
func (s *JobService) convertToCoverLetter(aiResult *aimodels.CoverLetter, userID, jobID int) *models.CoverLetter {
	now := time.Now().UTC()

	return &models.CoverLetter{
		JobID:       jobID,
		UserID:      userID,
		Content:     aiResult.Content,
		Format:      string(aiResult.Format),
		GeneratedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// validateProfileForAI validates that the user profile has sufficient data for AI operations.
func (s *JobService) validateProfileForAI(profile *settingsmodels.Profile) error {
	if profile.FirstName == "" && profile.LastName == "" {
		return models.ErrProfileIncomplete
	}

	if len(profile.Skills) == 0 {
		return models.ErrProfileSkillsRequired
	}

	hasCareerInfo := profile.CareerSummary != "" ||
		len(profile.WorkExperience) > 0 ||
		len(profile.Education) > 0

	if !hasCareerInfo {
		return models.ErrProfileSummaryRequired
	}

	if len(profile.WorkExperience) > 0 {
		hasDetailedExperience := false
		for _, exp := range profile.WorkExperience {
			if exp.Title != "" && exp.Company != "" {
				hasDetailedExperience = true
				break
			}
		}
		if !hasDetailedExperience && profile.CareerSummary == "" {
			return models.ErrProfileSummaryRequired
		}
	}

	return nil
}
