package prompts

// CoverLetterTemplate returns a template for cover letter generation
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

// EnhanceCoverLetterPrompt enhances a cover letter prompt
func EnhanceCoverLetterPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, wordRange string) string {
	template := CoverLetterTemplate()
	params := map[string]any{
		"wordRange": wordRange,
	}
	return template.BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, params)
}
