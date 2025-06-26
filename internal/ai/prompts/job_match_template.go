package prompts

import "fmt"

// JobMatchTemplate returns a template for job matching analysis
func JobMatchTemplate() *PromptTemplate {
	return &PromptTemplate{
		Role:    "You are a senior talent acquisition specialist with expertise across all industries and professions. You excel at assessing candidate-job fit by analyzing skills, experience, potential, and cultural alignment across diverse fields - from technical roles to creative positions, healthcare to business, education to trades.",
		Context: "You're conducting a comprehensive analysis to determine how well a candidate matches a specific job opportunity. Your assessment considers industry-specific requirements, transferable skills, growth potential, and overall fit.",
		Examples: []Example{
			{
				Input: "Candidate: Name: John, Title: Developer, Career Summary: I am a developer in technology industry\nJob: Senior Python Engineer requiring 5+ years Python, distributed systems, and AI/ML experience",
				Output: `{
  "matchScore": 12,
  "strengths": [
    "You're in the technology field, which aligns with this role",
    "Your developer title suggests some technical background"
  ],
  "weaknesses": [
    "Your profile lacks any work experience details - can't assess your Python expertise",
    "No skills listed to verify your technical capabilities",
    "Missing education background to understand your foundational knowledge",
    "No information about distributed systems or AI/ML experience"
  ],
  "highlights": [
    "You identify as a developer in the tech industry"
  ],
  "feedback": "Your profile is incomplete. You have no work experience, no skills, and no education listed. This makes you unqualified for any senior engineering position. A one-line summary saying 'I am a developer' provides zero evidence of your capabilities."
}`,
			},
			{
				Input: "Candidate: Financial Analyst with 5 years experience in corporate finance and MBA\nJob: Senior Financial Analyst at a tech startup requiring startup experience and financial modeling expertise",
				Output: `{
  "matchScore": 75,
  "strengths": [
    "You have a solid foundation in financial analysis and modeling that translates well here",
    "Your MBA gives you that strategic perspective they're looking for",
    "Your corporate finance background brings valuable structure to a startup environment",
    "Your analytical skills are exactly what they need - just in a different setting"
  ],
  "weaknesses": [
    "You don't have direct startup experience yet, which they mentioned wanting",
    "You might need some time to adjust from corporate pace to startup speed",
    "It's unclear how comfortable you are with ambiguity and constant change"
  ],
  "highlights": [
    "You led financial planning for a $50M business unit - that's impressive scale",
    "You developed complex financial models that were adopted company-wide",
    "Your MBA from a top-tier program included entrepreneurship focus"
  ],
  "feedback": "Strong financial analysis background but zero startup experience. Your corporate background is the opposite of what they're looking for. You'll struggle with the pace and chaos of startup life. The entrepreneurship MBA focus doesn't compensate for lack of real startup experience."
}`,
			},
			{
				Input: "Candidate: Elementary teacher with 8 years experience and special education certification\nJob: Special Education Coordinator requiring leadership experience and IEP expertise",
				Output: `{
  "matchScore": 82,
  "strengths": [
    "You're dual-certified in both elementary and special education - that's exactly what they need",
    "Your 8 years of hands-on classroom experience with diverse learners is invaluable",
    "You have direct, real-world experience developing and implementing IEPs",
    "You've already proven you can improve outcomes for special needs students"
  ],
  "weaknesses": [
    "You don't have much formal leadership or coordination experience on paper",
    "Budget management skills aren't mentioned, which might come up in this role"
  ],
  "highlights": [
    "You pioneered an inclusive classroom model that was adopted district-wide - that's leadership!",
    "You've mentored 5 new special education teachers, even if it was informal",
    "You achieved 100% parent satisfaction in IEP meetings - that's incredible",
    "You reduced special education referrals by 30% through early intervention strategies"
  ],
  "feedback": "Good match overall. You have the certifications and classroom experience needed. Your lack of formal leadership experience is a weakness - mentoring isn't the same as managing. Budget management skills are completely absent from your profile, which is concerning for a coordinator role."
}`,
			},
		},
		Task: "Provide a data-driven analysis of candidate-job fit using multiple evaluation criteria.",
		Constraints: []string{
			"Use direct, uncompromising language with 'you' and 'your' - no sugar-coating",
			"NEVER mention the candidate's name and NEVER use 'the candidate' - use 'you/your'",
			"Be brutally honest about profile deficiencies - no patronizing or feelings-based feedback",
			"State facts bluntly: incomplete profiles are unqualified, period",
			"Profile completeness is CRITICAL - incomplete profiles must score VERY LOW (15% or less)",
			"A profile with ONLY name/title/one-line summary scores 10-15% MAX regardless of title match",
			"Missing work experience: cap at 20% | Missing work experience AND education: cap at 15%",
			"Missing skills when job requires them: reduce score by 20%+ | Minimal summaries (<50 words): cap at 25%",
			"To score above 50%, MUST have substantial work experience, skills, AND education/certifications",
			"Previous match history is supplementary only - score based on CURRENT profile content",
			"Be objective and clinical in assessment - no emotional language",
			"Point out gaps and weaknesses directly without softening the message",
			"Don't frame weaknesses as 'opportunities' - call them what they are: deficiencies",
			"If someone is unqualified, say they're unqualified - don't dance around it",
			"Focus on what's missing, not potential - employers hire based on evidence, not hope",
			"Account for transferable skills but don't overstate their value",
			"Recognize industry-specific qualifications but don't inflate their importance",
		},
		OutputSpec: "Return ONLY a valid JSON object with: matchScore (0-100), strengths (array), weaknesses (array), highlights (array), and feedback (string)",
	}
}

// EnhanceJobMatchPrompt enhances a job matching prompt
func EnhanceJobMatchPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext string, minScore, maxScore int) string {
	template := JobMatchTemplate()
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
