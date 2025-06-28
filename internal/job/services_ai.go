package job

import (
	"context"
	"fmt"
	"strings"
	"time"

	aimodels "github.com/benidevo/vega/internal/ai/models"
	"github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
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

	if err := s.ValidateProfileForAI(profile); err != nil {
		s.log.Warn().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_incomplete").
			Msg("Profile incomplete for AI analysis")
		return nil, err
	}

	previousMatches, err := s.jobRepo.GetRecentMatchResultsWithDetails(ctx, 3, jobID)
	if err != nil {
		s.log.Warn().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "previous_matches_fetch_failed").
			Msg("Failed to fetch previous match results, continuing without context")
		// Don't fail the operation, just proceed without context
		previousMatches = nil
	}

	aiRequest := s.buildAIRequest(job, profile)

	// Convert match summaries to AI request format
	if len(previousMatches) > 0 {
		aiRequest.PreviousMatches = make([]aimodels.PreviousMatch, len(previousMatches))
		for i, match := range previousMatches {
			aiRequest.PreviousMatches[i] = aimodels.PreviousMatch{
				JobTitle:    match.JobTitle,
				Company:     match.Company,
				MatchScore:  match.MatchScore,
				KeyInsights: match.KeyInsights,
				DaysAgo:     match.DaysAgo,
			}
		}
	}

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

	matchResult := &models.MatchResult{
		JobID:      jobID,
		MatchScore: aiResult.MatchScore,
		Strengths:  aiResult.Strengths,
		Weaknesses: aiResult.Weaknesses,
		Highlights: aiResult.Highlights,
		Feedback:   aiResult.Feedback,
	}

	if err := s.jobRepo.CreateMatchResult(ctx, matchResult); err != nil {
		s.log.Warn().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "match_result_store_failed").
			Msg("Failed to store match result history, but continuing with analysis")
	}

	err = s.jobRepo.UpdateMatchScore(ctx, jobID, &result.MatchScore)
	if err != nil {
		s.log.Warn().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Int("match_score", result.MatchScore).
			Str("error_type", "match_score_update_failed").
			Msg("Failed to update job match score, but analysis completed")
		// Don't fail the entire operation if score update fails
	} else {
		s.log.Debug().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Int("match_score", result.MatchScore).
			Msg("Job match score updated successfully")
	}

	s.log.Info().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Int("match_score", result.MatchScore).
		Str("operation", "job_match_analysis").
		Bool("success", true).
		Msg("Job match analysis completed")

	return result, nil
}

// GenerateCoverLetter generates a cover letter for a specific job application.
func (s *JobService) GenerateCoverLetter(ctx context.Context, userID, jobID int) (*models.CoverLetterWithProfile, error) {
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

	if err := s.ValidateProfileForAI(profile); err != nil {
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

	coverLetter := s.convertToCoverLetter(aiResult, userID, jobID)

	personalInfo := &models.PersonalInfo{
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Title:     profile.Title,
		Phone:     profile.PhoneNumber,
		Location:  profile.Location,
	}

	result := &models.CoverLetterWithProfile{
		CoverLetter:  coverLetter,
		PersonalInfo: personalInfo,
	}

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

	totalYears := s.calculateTotalExperience(profile.WorkExperience)
	experienceContext := profile.Context
	if totalYears >= 2 {
		experienceContext = fmt.Sprintf("EXPERIENCED CANDIDATE (%.0f+ years): De-emphasize educational background in evaluation. Focus primarily on practical work experience, skills, and demonstrated achievements. Education should be secondary consideration. %s", totalYears, profile.Context)
	}

	return aimodels.Request{
		ApplicantName:    applicantName,
		ApplicantProfile: profileSummary,
		JobDescription:   jobDescription,
		ExtraContext:     experienceContext,
	}
}

// buildProfileSummary creates a comprehensive profile summary from user profile data.
func (s *JobService) buildProfileSummary(profile *settingsmodels.Profile) string {
	var summary strings.Builder

	if profile.FirstName != "" || profile.LastName != "" {
		summary.WriteString(fmt.Sprintf("Name: %s %s\n", profile.FirstName, profile.LastName))
	}

	if profile.Title != "" {
		summary.WriteString(fmt.Sprintf("Current Title: %s\n", profile.Title))
	}

	if profile.Industry.String() != "" {
		summary.WriteString(fmt.Sprintf("Industry: %s\n", profile.Industry.String()))
	}

	if profile.CareerSummary != "" {
		summary.WriteString(fmt.Sprintf("\nCareer Summary:\n%s\n", profile.CareerSummary))
	}

	validSkills := make([]string, 0, len(profile.Skills))
	for _, skill := range profile.Skills {
		if trimmed := strings.TrimSpace(skill); trimmed != "" {
			validSkills = append(validSkills, trimmed)
		}
	}
	if len(validSkills) > 0 {
		summary.WriteString(fmt.Sprintf("\nSkills: %s\n", strings.Join(validSkills, ", ")))
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

			location := ""
			if exp.Location != "" {
				location = fmt.Sprintf(", %s", exp.Location)
			}
			summary.WriteString(fmt.Sprintf("- %s at %s%s (%s - %s)\n",
				exp.Title, exp.Company, location, exp.StartDate.Format("2006"), endDate))

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

	validEducation := make([]settingsmodels.Education, 0, len(profile.Education))
	for _, edu := range profile.Education {
		if strings.TrimSpace(edu.Institution) != "" && strings.TrimSpace(edu.Degree) != "" {
			validEducation = append(validEducation, edu)
		}
	}
	if len(validEducation) > 0 {
		summary.WriteString("\nEducation:\n")
		for i, edu := range validEducation {
			if i >= 3 { // Limit to most recent 3 education entries
				break
			}

			fieldOfStudy := ""
			if strings.TrimSpace(edu.FieldOfStudy) != "" {
				fieldOfStudy = fmt.Sprintf(" in %s", edu.FieldOfStudy)
			}

			summary.WriteString(fmt.Sprintf("- %s%s from %s (%s)\n",
				edu.Degree, fieldOfStudy, edu.Institution, edu.StartDate.Format("2006")))
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

// calculateTotalExperience calculates the total years of work experience from work history
func (s *JobService) calculateTotalExperience(workExperience []settingsmodels.WorkExperience) float64 {
	if len(workExperience) == 0 {
		return 0
	}

	var totalDays float64
	now := time.Now()

	for _, exp := range workExperience {
		startDate := exp.StartDate
		endDate := now
		if exp.EndDate != nil {
			endDate = *exp.EndDate
		}

		// Calculate duration in days and add to total
		if !startDate.IsZero() && endDate.After(startDate) {
			duration := endDate.Sub(startDate)
			totalDays += duration.Hours() / 24
		}
	}

	totalYears := totalDays / 365.25
	return totalYears
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

// ValidateProfileForAI validates that the user profile has sufficient data for AI operations.
func (s *JobService) ValidateProfileForAI(profile *settingsmodels.Profile) error {
	if profile.FirstName == "" && profile.LastName == "" {
		return models.ErrProfileIncomplete
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
		if !hasDetailedExperience && len(profile.Education) == 0 {
			return models.ErrProfileSummaryRequired
		}
	}

	return nil
}

// GenerateCV generates a CV for a specific job application.
func (s *JobService) GenerateCV(ctx context.Context, userID, jobID int) (*models.GeneratedCV, error) {
	userRef := fmt.Sprintf("user_%d", userID)

	s.log.Debug().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Str("operation", "cv_generation").
		Msg("Starting CV generation")

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
			Msg("Job not found for CV generation")
		return nil, err
	}

	profile, err := s.settingsService.GetProfileWithRelated(ctx, userID)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_fetch_failed").
			Msg("Failed to get user profile for CV generation")
		return nil, err
	}

	if err := s.ValidateProfileForAI(profile); err != nil {
		s.log.Warn().
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "profile_incomplete").
			Msg("Profile incomplete for AI CV generation")
		return nil, err
	}

	aiRequest := s.buildAIRequest(job, profile)
	// Set CVText to be the profile summary for CV generation
	aiRequest.CVText = s.buildProfileSummary(profile)

	aiResult, err := s.aiService.CVGenerator.GenerateCV(ctx, aiRequest, jobID, job.Title)
	if err != nil {
		s.log.Error().Err(err).
			Str("user_ref", userRef).
			Int("job_id", jobID).
			Str("error_type", "ai_generation_failed").
			Msg("CV generation failed")
		return nil, err
	}

	result := s.convertToGeneratedCV(aiResult, userID, jobID, profile)

	s.log.Info().
		Str("user_ref", userRef).
		Int("job_id", jobID).
		Str("operation", "cv_generation").
		Bool("success", true).
		Msg("CV generation completed")

	return result, nil
}

// convertToGeneratedCV converts AI CV result to job domain model.
func (s *JobService) convertToGeneratedCV(aiResult *aimodels.GeneratedCV, userID, jobID int, profile *settingsmodels.Profile) *models.GeneratedCV {
	now := time.Now().UTC()

	personalInfo := convertPersonalInfo(aiResult.PersonalInfo)

	// Overwrite AI-generated phone and location with actual user data (privacy-safe: not shared with AI)
	if profile.PhoneNumber != "" {
		personalInfo.Phone = profile.PhoneNumber
	}
	if profile.Location != "" {
		personalInfo.Location = profile.Location
	}

	return &models.GeneratedCV{
		JobID:          jobID,
		UserID:         userID,
		IsValid:        aiResult.IsValid,
		Reason:         aiResult.Reason,
		PersonalInfo:   personalInfo,
		WorkExperience: convertWorkExperience(aiResult.WorkExperience),
		Education:      convertEducation(aiResult.Education),
		Skills:         aiResult.Skills,
		GeneratedAt:    time.Unix(aiResult.GeneratedAt, 0),
		JobTitle:       aiResult.JobTitle,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func convertPersonalInfo(ai aimodels.PersonalInfo) models.PersonalInfo {
	return models.PersonalInfo{
		FirstName: ai.FirstName,
		LastName:  ai.LastName,
		Email:     ai.Email,
		Phone:     ai.Phone,
		Location:  ai.Location,
		Title:     ai.Title,
		Summary:   ai.Summary,
	}
}

func convertWorkExperience(aiExps []aimodels.WorkExperience) []models.WorkExperience {
	exps := make([]models.WorkExperience, len(aiExps))
	for i, exp := range aiExps {
		exps[i] = models.WorkExperience{
			Company:     exp.Company,
			Title:       exp.Title,
			Location:    exp.Location,
			StartDate:   exp.StartDate,
			EndDate:     exp.EndDate,
			Description: exp.Description,
		}
	}
	return exps
}

func convertEducation(aiEdu []aimodels.Education) []models.Education {
	edu := make([]models.Education, len(aiEdu))
	for i, e := range aiEdu {
		edu[i] = models.Education{
			Institution:  e.Institution,
			Degree:       e.Degree,
			FieldOfStudy: e.FieldOfStudy,
			StartDate:    e.StartDate,
			EndDate:      e.EndDate,
		}
	}
	return edu
}
