package prompts

import "fmt"

// JobMatchTemplate returns a template for job matching analysis
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
  "feedback": "You've got a really strong analytical foundation that would serve you well here, though there's definitely room to grow in the startup context. Your corporate experience actually brings valuable structure, but you'll want to be ready for a much faster pace and more ambiguity. The good news? Your MBA's entrepreneurship focus shows you're genuinely interested in this world."
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
  "feedback": "This looks like an excellent match! Your practical experience and proven results really stand out. Sure, you might not have formal leadership titles, but you've already been leading through mentoring and developing programs that others adopted. Your ability to connect with parents and get results with students is exactly what they're looking for in this coordinator role."
}`,
			},
		},
		Task: "Provide a data-driven analysis of candidate-job fit using multiple evaluation criteria.",
		Constraints: []string{
			"Write in a conversational, second-person tone addressing the candidate directly",
			"Be objective and balanced in assessment across all industries",
			"Consider both current capabilities and growth potential",
			"Identify specific, actionable strengths and gaps relevant to the field",
			"Provide constructive feedback that helps the candidate understand their fit",
			"Look beyond surface-level keyword matching to assess true fit",
			"Consider soft skills, cultural indicators, and industry-specific requirements",
			"Be honest about concerns without being discouraging - frame as opportunities",
			"Account for transferable skills from other industries or roles",
			"Recognize industry-specific qualifications (certifications, licenses, etc.)",
			"Use encouraging, supportive language while maintaining professional honesty",
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
