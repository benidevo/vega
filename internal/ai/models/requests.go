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
	PreviousMatches  []PreviousMatch
}

// PreviousMatch represents a summary of a previous match result for context
type PreviousMatch struct {
	JobTitle    string
	Company     string
	MatchScore  int
	KeyInsights string
	DaysAgo     int
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

	previousMatchContext := ""
	if len(p.PreviousMatches) > 0 {
		previousMatchContext = "\n\nPrevious Match History (for context):\n"
		for i, match := range p.PreviousMatches {
			if i >= 3 { // Limit to 3 previous matches
				break
			}
			previousMatchContext += fmt.Sprintf("- %s at %s (%d days ago): Score %d/100. %s\n",
				match.JobTitle, match.Company, match.DaysAgo, match.MatchScore, match.KeyInsights)
		}
		previousMatchContext += "\nNote: These are for minor context only - base your score on the CURRENT profile content, not historical scores.\n"
	}

	return fmt.Sprintf(`%s

Analyze the match between this applicant and job opportunity:

Job Description: %s
Applicant Profile: %s
%s%s

CRITICAL SCORING GUIDELINES:
- Profile completeness is ESSENTIAL - incomplete profiles MUST receive VERY LOW scores (15%% or less)
- A profile with ONLY name, title, and a one-line summary should score 10-15%% MAX
- Missing work experience section: automatic cap at 20%% (even with good title match)
- Missing BOTH work experience AND education: automatic cap at 15%%
- Missing skills section when job lists required skills: reduce score by at least 20%%
- Previous match history is ONLY supplementary context - base your score primarily on the CURRENT profile content
- Empty or minimal career summaries (under 50 words) should cap score at 25%%
- To score above 50%%, profile MUST have substantial work experience, skills, AND either education or certifications

Provide a comprehensive analysis focusing on:
- Skills alignment with job requirements
- Experience level and relevance  
- Industry knowledge fit
- Cultural fit indicators
- Growth potential
- Any concerns or red flags

IMPORTANT: In your feedback, do NOT mention the applicant's name. Use "you" and "your" directly. Be brutally honest and direct - no sugar-coating or patronizing language. State facts bluntly about what's missing or inadequate.

Return the analysis as a JSON object with EXACTLY this structure:
- matchScore: integer from %d-%d where %d is no match and %d is perfect match
- strengths: array of 3-5 key strengths that align with job requirements
- weaknesses: array of 2-4 areas for improvement or skill gaps
- highlights: array of 3-5 standout qualifications that make this candidate attractive
- feedback: overall assessment and recommendations in 2-3 sentences (do NOT include the applicant's name)`,
		p.Instructions,
		p.JobDescription,
		p.ApplicantProfile,
		p.ExtraContext,
		previousMatchContext,
		minMatchScore,
		maxMatchScore,
		minMatchScore,
		maxMatchScore)
}
