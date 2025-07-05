package prompts

import (
	"fmt"
	"time"
)

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
				Input: "Candidate: Software Engineer with 4 years Node.js, Express, MongoDB, and AWS experience\nJob: Backend Developer position requiring Python, Django, PostgreSQL, and cloud deployment skills",
				Output: `{
  "matchScore": 78,
  "strengths": [
    "Your 4 years of backend development experience directly translates to this role",
    "Node.js and Python are both server-side languages - your async programming skills transfer well",
    "Express and Django are similar web frameworks - you understand MVC patterns and API design",
    "You have solid database experience with MongoDB that applies to any database system",
    "Your AWS cloud experience covers the deployment and infrastructure requirements"
  ],
  "weaknesses": [
    "You'll need to learn Python syntax and Django specifics, though the concepts are familiar",
    "PostgreSQL differs from MongoDB in being relational vs document-based",
    "No direct experience with Django's ORM, but your database background helps"
  ],
  "highlights": [
    "You built scalable APIs handling high traffic - that's exactly what they need",
    "Your microservices architecture experience shows advanced backend skills",
    "AWS deployment experience means you can handle their infrastructure needs"
  ],
  "feedback": "Strong backend foundation with transferable skills. Your Node.js/Express experience translates well to Python/Django work. The core concepts are the same - just different syntax. Your cloud and database experience fills the infrastructure requirements perfectly."
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
		Task: "Provide a data-driven analysis of candidate-job fit using multiple evaluation criteria. Be moderately lenient - value similar skills and transferable experience, not just exact matches.",
		Constraints: []string{
			"Use direct, uncompromising language with 'you' and 'your' - no sugar-coating",
			"NEVER mention the candidate's name and NEVER use 'the candidate' - use 'you/your'",
			"Be brutally honest about profile deficiencies - no patronizing or feelings-based feedback",
			"State facts bluntly: incomplete profiles are unqualified, period",
			"Profile completeness is CRITICAL - incomplete profiles must score VERY LOW (15% or less)",
			"A profile with ONLY name/title/one-line summary scores 10-15% MAX regardless of title match",
			"Missing work experience: cap at 20% | Missing work experience AND education: cap at 15%",
			"Missing specific skills: reduce score by 10-15% (not 20%+) if no similar/transferable skills exist | Minimal summaries (<50 words): cap at 25%",
			"EXPERIENCE-BASED EVALUATION: For candidates with 2+ years experience, prioritize work history and practical skills over educational background",
			"EXPERIENCED CANDIDATES (2+ years): Education should be secondary - focus on job performance, achievements, and demonstrated capabilities",
			"ENTRY-LEVEL CANDIDATES (<2 years): Education and certifications carry more weight due to limited work history",
			"To score above 50%, experienced candidates need solid work experience and related/similar skills; entry-level candidates need work experience OR strong education/skills",
			"Previous match history is supplementary only - score based on CURRENT profile content",
			"Be objective and clinical in assessment - no emotional language",
			"Point out gaps and weaknesses directly without softening the message",
			"Don't frame weaknesses as 'opportunities' - call them what they are: deficiencies",
			"If someone is unqualified, say they're unqualified - don't dance around it",
			"Focus on what's missing, not potential - employers hire based on evidence, not hope",
			"SKILL MATCHING: Value similar and transferable skills - exact matches are preferred but not required",
			"SIMILAR SKILLS BONUS: Award modest score increases (3-8 points) for related technologies (e.g., Python/Java, React/Vue, AWS/Azure)",
			"TRANSFERABLE SKILLS: Consider cross-domain skills valuable (e.g., project management across industries, problem-solving abilities)",
			"Don't penalize heavily for missing exact skills if candidate shows strong foundation in similar technologies",
			"Recognize industry-specific qualifications but don't inflate their importance",
			"DATE CONTEXT AWARENESS: Use the current date to assess experience recency and career progression timing",
			"Evaluate career gaps in context of current date when assessing overall profile strength",
		},
		OutputSpec: "Return ONLY a valid JSON object with: matchScore (0-100), strengths (array), weaknesses (array), highlights (array), and feedback (string)",
	}
}

// EnhanceJobMatchPrompt enhances a job matching prompt
func EnhanceJobMatchPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext string, minScore, maxScore int) string {
	template := JobMatchTemplate()

	template.Constraints = append(template.Constraints, JobMatchAntiAIConstraints()...)

	params := map[string]any{
		"useChainOfThought": true,
		"minScore":          minScore,
		"maxScore":          maxScore,
		"currentDate":       time.Now().Format("January 2, 2006"),
	}

	// Add score range to output spec
	template.OutputSpec = fmt.Sprintf("Return ONLY a valid JSON object with: matchScore (%d-%d where %d is no match and %d is perfect match), strengths (array of 3-5 items), weaknesses (array of 2-4 items), highlights (array of 3-5 items), and feedback (2-3 sentences)",
		minScore, maxScore, minScore, maxScore)

	return template.BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, params)
}
