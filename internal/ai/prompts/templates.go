package prompts

import (
	"fmt"
	"strings"
)

// PromptTemplate represents a reusable prompt template that adapts to any profession or industry.
// The templates are designed to be profession-agnostic, allowing the AI to understand context
// from the specific job description and candidate profile provided.
type PromptTemplate struct {
	Role        string
	Context     string
	Examples    []Example
	Task        string
	Constraints []string
	OutputSpec  string
}

type Example struct {
	Input  string
	Output string
}

// CoverLetterTemplate returns an enhanced template for cover letter generation
func CoverLetterTemplate() *PromptTemplate {
	return &PromptTemplate{
		Role:    "You are an elite career consultant and professional writer with expertise across all industries and professions. You excel at crafting compelling cover letters that highlight relevant experience and create genuine connections between candidates and employers, regardless of field - from healthcare to finance, education to engineering, arts to business.",
		Context: "You're helping a candidate craft a personalized cover letter that demonstrates their unique value proposition. The goal is to show how their specific skills, experiences, and passion align with the employer's needs and culture.",
		Examples: []Example{
			{
				Input: "Candidate: Marketing Manager with 7 years experience in digital campaigns and brand strategy\nJob: Senior Marketing Manager at a sustainable fashion company seeking someone with e-commerce and social media expertise",
				Output: `{
  "content": "Dear Hiring Team,\n\nYour mission to revolutionize sustainable fashion through innovative marketing strategies deeply resonates with my professional values and expertise. With seven years of driving digital transformation for consumer brands, I'm excited about the opportunity to amplify your impact in the sustainable fashion space.\n\nIn my current role at RetailCo, I spearheaded a digital campaign that increased online sales by 150% while reducing customer acquisition costs by 40%. My experience launching viral social media campaigns, including one that reached 10M organic impressions, aligns perfectly with your need for someone who can elevate your brand's digital presence.\n\nWhat particularly draws me to your company is the intersection of purpose and profit. Having led the rebranding of an eco-conscious product line that resulted in 200% growth, I understand how to communicate sustainability stories that resonate with modern consumers.\n\nI would love to discuss how my expertise in e-commerce optimization and passion for sustainable business can contribute to your continued growth.\n\nBest regards,\n[Candidate Name]"
}`,
			},
			{
				Input: "Candidate: Registered Nurse with 10 years ICU experience and leadership training\nJob: Nursing Unit Manager at a major hospital seeking someone with critical care expertise and team management skills",
				Output: `{
  "content": "Dear Hiring Manager,\n\nThe opportunity to lead your ICU nursing team and enhance patient care standards at your renowned facility immediately captured my attention. With a decade of critical care experience and a proven track record of mentoring high-performing teams, I'm eager to contribute to your mission of exceptional healthcare delivery.\n\nDuring my tenure at Regional Medical Center, I implemented a peer mentorship program that reduced nursing turnover by 35% and improved patient satisfaction scores to the 95th percentile. My hands-on experience managing complex cases while developing staff competencies uniquely positions me to balance the dual demands of clinical excellence and team leadership.\n\nYour hospital's commitment to innovation in critical care, particularly your recent advanced life support protocols, aligns with my passion for evidence-based practice. I recently led the implementation of a new sepsis protocol that reduced mortality rates by 20%.\n\nI would welcome the opportunity to discuss how my clinical expertise and leadership experience can strengthen your ICU team's performance and patient outcomes.\n\nSincerely,\n[Candidate Name]"
}`,
			},
		},
		Task: "Create a compelling cover letter that showcases the candidate's unique value proposition and demonstrates genuine interest in the role.",
		Constraints: []string{
			"Use active voice and powerful action verbs appropriate to the industry",
			"Include specific, quantifiable achievements when possible",
			"Mirror key language from the job description naturally",
			"Show understanding of the organization's mission, values, or industry position",
			"Maintain professional tone appropriate to the field while showing personality",
			"Avoid generic phrases like 'I am writing to apply for...'",
			"Keep paragraphs concise (3-4 sentences max)",
			"End with a clear, confident call to action",
			"Adapt formality level to match industry norms (e.g., more formal for law/finance, creative for marketing/design)",
			"No em dashes or overly casual language or ai-like phrases",
		},
		OutputSpec: "Return ONLY a valid JSON object with a 'content' field containing the cover letter text with proper formatting (\\n for line breaks)",
	}
}

// JobMatchTemplate returns an enhanced template for job matching analysis
func JobMatchTemplate() *PromptTemplate {
	return &PromptTemplate{
		Role:    "You are a senior talent acquisition specialist with expertise across all industries and professions. You excel at assessing candidate-job fit by analyzing skills, experience, potential, and cultural alignment across diverse fields - from technical roles to creative positions, healthcare to business, education to trades.",
		Context: "You're conducting a comprehensive analysis to determine how well a candidate matches a specific job opportunity. Your assessment considers industry-specific requirements, transferable skills, growth potential, and overall fit.",
		Examples: []Example{
			{
				Input: "Candidate: Financial Analyst with 5 years experience in corporate finance and MBA\nJob: Senior Financial Analyst at a tech startup requiring startup experience and financial modeling expertise",
				Output: `{
  "matchScore": 75,
  "strengths": [
    "Strong foundation in financial analysis and modeling",
    "MBA provides strategic business perspective",
    "Corporate finance experience brings structured approach",
    "Analytical skills directly transferable to startup environment"
  ],
  "weaknesses": [
    "No direct startup experience mentioned",
    "May need adjustment from corporate to startup pace",
    "Unclear if comfortable with ambiguity and rapid change"
  ],
  "highlights": [
    "Led financial planning for $50M business unit",
    "Developed complex financial models adopted company-wide",
    "MBA from top-tier program with entrepreneurship focus"
  ],
  "feedback": "Strong analytical foundation with room to grow in startup context. The candidate's corporate experience provides valuable structure, though they'll need support transitioning to a fast-paced environment. Their MBA entrepreneurship focus suggests genuine interest in startups."
}`,
			},
			{
				Input: "Candidate: Elementary teacher with 8 years experience and special education certification\nJob: Special Education Coordinator requiring leadership experience and IEP expertise",
				Output: `{
  "matchScore": 82,
  "strengths": [
    "Dual certification in elementary and special education",
    "8 years hands-on classroom experience with diverse learners",
    "Direct experience developing and implementing IEPs",
    "Demonstrated success improving outcomes for special needs students"
  ],
  "weaknesses": [
    "Limited formal leadership or coordination experience",
    "No mention of budget management skills"
  ],
  "highlights": [
    "Pioneered inclusive classroom model adopted district-wide",
    "Mentored 5 new special education teachers informally",
    "100% parent satisfaction rating in IEP meetings",
    "Reduced special education referrals by 30% through early intervention"
  ],
  "feedback": "Excellent match with strong practical experience and proven results. While formal leadership experience is limited, the candidate shows natural leadership through mentoring and program development. Their innovative approaches and parent engagement skills are particularly valuable."
}`,
			},
		},
		Task: "Provide a data-driven analysis of candidate-job fit using multiple evaluation criteria.",
		Constraints: []string{
			"Be objective and balanced in assessment across all industries",
			"Consider both current capabilities and growth potential",
			"Identify specific, actionable strengths and gaps relevant to the field",
			"Provide constructive feedback that helps both parties",
			"Look beyond surface-level keyword matching to assess true fit",
			"Consider soft skills, cultural indicators, and industry-specific requirements",
			"Be honest about concerns without being discouraging",
			"Account for transferable skills from other industries or roles",
			"Recognize industry-specific qualifications (certifications, licenses, etc.)",
		},
		OutputSpec: "Return ONLY a valid JSON object with: matchScore (0-100), strengths (array), weaknesses (array), highlights (array), and feedback (string)",
	}
}

// BuildPrompt constructs the final prompt from a template
func (t *PromptTemplate) BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext string, params map[string]any) string {
	var promptBuilder strings.Builder

	if systemInstruction != "" {
		promptBuilder.WriteString(systemInstruction)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("# Your Role\n%s\n\n", t.Role))
	promptBuilder.WriteString(fmt.Sprintf("# Context\n%s\n\n", t.Context))

	if len(t.Examples) > 0 {
		promptBuilder.WriteString("# Examples of Excellent Output\n")
		for i, example := range t.Examples {
			promptBuilder.WriteString(fmt.Sprintf("## Example %d\n", i+1))
			promptBuilder.WriteString(fmt.Sprintf("Input: %s\n", example.Input))
			promptBuilder.WriteString(fmt.Sprintf("Output: %s\n\n", example.Output))
		}
	}

	// Current task details
	promptBuilder.WriteString("# Current Task\n\n")
	promptBuilder.WriteString(fmt.Sprintf("**Applicant:** %s\n\n", applicantName))
	promptBuilder.WriteString(fmt.Sprintf("**Job Description:**\n%s\n\n", jobDescription))
	promptBuilder.WriteString(fmt.Sprintf("**Applicant Profile:**\n%s\n\n", applicantProfile))

	if extraContext != "" {
		promptBuilder.WriteString(fmt.Sprintf("**Additional Context:**\n%s\n\n", extraContext))
	}

	promptBuilder.WriteString(fmt.Sprintf("# Your Task\n%s\n\n", t.Task))

	if len(t.Constraints) > 0 {
		promptBuilder.WriteString("# Requirements and Constraints\n")
		for _, constraint := range t.Constraints {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", constraint))
		}
		promptBuilder.WriteString("\n")
	}

	// Add any custom parameters
	if wordRange, ok := params["wordRange"].(string); ok {
		promptBuilder.WriteString(fmt.Sprintf("**Word Count:** %s words\n\n", wordRange))
	}

	// Output specification
	promptBuilder.WriteString(fmt.Sprintf("# Output Format\n%s\n", t.OutputSpec))

	// Chain of thought instruction for complex analysis
	if _, ok := params["useChainOfThought"]; ok {
		promptBuilder.WriteString("\n# Thinking Process\nBefore providing your final answer, briefly analyze:\n")
		promptBuilder.WriteString("1. Key requirements from the job description (technical skills, soft skills, experience level, industry-specific needs)\n")
		promptBuilder.WriteString("2. Matching qualifications from the candidate (direct matches, transferable skills, relevant achievements)\n")
		promptBuilder.WriteString("3. Gaps or areas of concern (missing requirements, experience differences, potential challenges)\n")
		promptBuilder.WriteString("4. Overall fit assessment (considering the specific industry context and role requirements)\n\n")
		promptBuilder.WriteString("Then provide your final JSON response.\n")
	}

	return promptBuilder.String()
}

// PromptEnhancer provides methods to enhance existing prompts
type PromptEnhancer struct {
	templates map[string]*PromptTemplate
}

// NewPromptEnhancer creates a new prompt enhancer
func NewPromptEnhancer() *PromptEnhancer {
	return &PromptEnhancer{
		templates: map[string]*PromptTemplate{
			"cover_letter": CoverLetterTemplate(),
			"job_match":    JobMatchTemplate(),
		},
	}
}

// EnhanceCoverLetterPrompt enhances a cover letter prompt
func (pe *PromptEnhancer) EnhanceCoverLetterPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, wordRange string) string {
	template := pe.templates["cover_letter"]
	params := map[string]any{
		"wordRange": wordRange,
	}
	return template.BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, params)
}

// EnhanceJobMatchPrompt enhances a job matching prompt
func (pe *PromptEnhancer) EnhanceJobMatchPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext string, minScore, maxScore int) string {
	template := pe.templates["job_match"]
	params := map[string]any{
		"useChainOfThought": true,
		"minScore":          minScore,
		"maxScore":          maxScore,
	}

	// Add score range to output spec
	template.OutputSpec = fmt.Sprintf("Return ONLY a valid JSON object with: matchScore (%d-%d where %d is no match and %d is perfect match), strengths (array of 3-5 items), weaknesses (array of 2-4 items), highlights (array of 3-5 items), and feedback (2-3 sentences)",
		minScore, maxScore, minScore, maxScore)

	return template.BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, params)
}
