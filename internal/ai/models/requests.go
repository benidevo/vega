package models

import (
	"fmt"

	"github.com/benidevo/vega/internal/ai/prompts"
)

// Request represents a generic request containing information needed for AI operations.
type Request struct {
	ApplicantName    string
	ApplicantProfile string
	JobDescription   string
	ExtraContext     string
}

// Prompt represents the structure for a prompt used in the application.
type Prompt struct {
	Instructions string
	Request
	UseEnhancedTemplates bool
	Temperature          *float32
	promptEnhancer       *prompts.PromptEnhancer
}

// NewPrompt creates a new prompt with optional enhanced features
func NewPrompt(instructions string, request Request, useEnhanced bool) *Prompt {
	p := &Prompt{
		Instructions:         instructions,
		Request:              request,
		UseEnhancedTemplates: useEnhanced,
	}

	if useEnhanced {
		p.promptEnhancer = prompts.NewPromptEnhancer()
	}

	return p
}

// SetTemperature sets a custom temperature for this prompt
func (p *Prompt) SetTemperature(temp float32) {
	p.Temperature = &temp
}

// GetOptimalTemperature returns the optimal temperature for the prompt type
func (p *Prompt) GetOptimalTemperature(promptType string) float32 {
	if p.Temperature != nil {
		return *p.Temperature
	}

	// Dynamic temperature based on task type
	switch promptType {
	case "cover_letter":
		return 0.65 // Higher creativity for writing
	case "job_match":
		return 0.2 // Lower for analytical consistency
	default:
		return 0.4 // Default balanced temperature
	}
}

// ToCoverLetterPrompt builds a cover letter generation prompt.
func (p Prompt) ToCoverLetterPrompt(defaultWordRange string) string {
	if p.UseEnhancedTemplates && p.promptEnhancer != nil {
		return p.promptEnhancer.EnhanceCoverLetterPrompt(
			p.Instructions,
			p.ApplicantName,
			p.JobDescription,
			p.ApplicantProfile,
			p.ExtraContext,
			defaultWordRange,
		)
	}

	return fmt.Sprintf(`%s

Generate a professional cover letter with the following details:

Applicant: %s
Job Description: %s
Applicant Profile: %s
%s

Requirements:
- Write in a professional but personalized tone that reflects the candidate's personality
- Highlight relevant skills and experiences from the applicant's profile that directly match the job requirements
- Address specific requirements mentioned in the job description
- Keep it concise (%s words)
- Use proper business letter format without date/address headers
- Include a strong opening that captures attention
- Provide specific examples of achievements when possible
- End with a compelling call to action
- Avoid generic phrases and clichés
- Do not include placeholder text like [Company Name] or [Your Name]

Return a JSON object with ONLY this field:
- content: the complete cover letter text (properly formatted with \n for line breaks)`,
		p.Instructions,
		p.ApplicantName,
		p.JobDescription,
		p.ApplicantProfile,
		p.ExtraContext,
		defaultWordRange)
}

// ToMatchAnalysisPrompt builds a job match analysis prompt from this Prompt
func (p Prompt) ToMatchAnalysisPrompt(minMatchScore, maxMatchScore int) string {
	if p.UseEnhancedTemplates && p.promptEnhancer != nil {
		return p.promptEnhancer.EnhanceJobMatchPrompt(
			p.Instructions,
			p.ApplicantName,
			p.JobDescription,
			p.ApplicantProfile,
			p.ExtraContext,
			minMatchScore,
			maxMatchScore,
		)
	}

	return fmt.Sprintf(`%s

Analyze the match between this applicant and job opportunity:

Applicant: %s
Job Description: %s
Applicant Profile: %s
%s

Provide a comprehensive analysis focusing on:
- Skills alignment with job requirements
- Experience level and relevance
- Industry knowledge fit
- Cultural fit indicators
- Growth potential
- Any concerns or red flags

Return the analysis as a JSON object with EXACTLY this structure:
- matchScore: integer from %d-%d where %d is no match and %d is perfect match
- strengths: array of 3-5 key strengths that align with job requirements
- weaknesses: array of 2-4 areas for improvement or skill gaps
- highlights: array of 3-5 standout qualifications that make this candidate attractive
- feedback: overall assessment and recommendations in 2-3 sentences`,
		p.Instructions,
		p.ApplicantName,
		p.JobDescription,
		p.ApplicantProfile,
		p.ExtraContext,
		minMatchScore,
		maxMatchScore,
		minMatchScore,
		maxMatchScore)
}
