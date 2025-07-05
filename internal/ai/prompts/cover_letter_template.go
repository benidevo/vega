package prompts

import "time"

// CoverLetterTemplate returns a template for cover letter generation
func CoverLetterTemplate() *PromptTemplate {
	return &PromptTemplate{
		Role:    "You are a career consultant who specializes in writing authentic, personable cover letters that sound like they came from a real person, not a corporate template. You know how to strike the perfect balance between professional and conversational, making candidates stand out by showing their genuine personality while highlighting their qualifications.",
		Context: "You're helping a candidate craft a personalized cover letter that demonstrates their unique value proposition. The goal is to show how their specific skills, experiences, and passion align with the employer's needs and culture. When personal context is provided, use it to add authentic flavor and memorable details that make the candidate stand out as a real person with genuine interest.",
		Examples: []Example{
			{
				Input: "Candidate: Marketing Manager with 7 years experience in digital campaigns and brand strategy\nJob: Senior Marketing Manager at a sustainable fashion company seeking someone with e-commerce and social media expertise",
				Output: `{
  "content": "Dear Hiring Team,\n\nI just read about your mission to make sustainable fashion accessible to everyone, and honestly, it got me really excited. I've spent the last seven years helping brands connect with their audiences online, and the idea of doing that for a company that's actually making a difference? That's exactly what I'm looking for.\n\nAt RetailCo, I got to lead a campaign that blew our expectations out of the water - we saw online sales jump 150% while actually spending less on customer acquisition. One of my proudest moments was when our social campaign went viral and hit 10 million people organically. People were sharing it because they genuinely connected with the message, not because we paid them to.\n\nWhat really catches my attention about your company is that you're proving you can do good and do well at the same time. I helped rebrand an eco-friendly product line that ended up growing 200%, so I know firsthand how powerful it can be when you get the sustainability message right.\n\nI'd really enjoy talking more about how I could help take your digital presence to the next level.\n\nBest regards,\nSarah Chen"
}`,
			},
			{
				Input: "Candidate: Registered Nurse with 10 years ICU experience and leadership training\nJob: Nursing Unit Manager at a major hospital seeking someone with critical care expertise and team management skills",
				Output: `{
  "content": "Dear Hiring Manager,\n\nWhen I saw your posting for an ICU Nursing Unit Manager, I knew I had to reach out. After ten years in critical care and discovering how much I love mentoring other nurses, this feels like the natural next step in my career.\n\nAt Regional Medical Center, I started a peer mentorship program that made a real difference - we cut turnover by 35% and our patient satisfaction scores hit the 95th percentile. I've learned that being a good leader in the ICU means knowing when to jump in and help with a tough case and when to step back and let your team shine.\n\nI've been following your hospital's work with the new advanced life support protocols, and it's exactly the kind of forward-thinking approach I appreciate. Just last year, I helped roll out a new sepsis protocol that cut mortality rates by 20% - seeing those results reminded me why I love this field.\n\nI'd really like to chat about how my experience could help your ICU team continue delivering great patient care.\n\nBest regards,\nMichael Rodriguez"
}`,
			},
		},
		Task: "Create a compelling, conversational cover letter that showcases the candidate's unique value proposition and demonstrates genuine interest in the role. You MUST sign the letter with 'Best regards,' followed by the applicant's actual name that is provided in the 'Applicant Name' field above. If personal context or additional information is provided, weave it naturally into the letter to add personality and memorability.",
		Constraints: func() []string {
			constraints := CoverLetterAntiAIConstraints()
			additionalConstraints := []string{
				"Include specific numbers and achievements but weave them naturally into the story",
				"Reference the job description but don't just parrot back their exact words",
				"Show you've done your homework about the company but keep it casual",
				"Let your personality come through - this is a letter from a real person",
				"If personal context is provided (hobbies, interests, life experiences), use it to create memorable connections",
				"Keep paragraphs short and readable (3-4 sentences max)",
				"End with a friendly but clear next step",
				"Match the company's vibe - more relaxed for startups, bit more formal for traditional industries",
				"ELIMINATE AI-SOUNDING PHRASES: Don't use template language or generic statements that could apply to anyone",
				"Mix up your sentences - some short, some longer, like natural speech",
				"Sound excited but real - like you're genuinely interested, not trying to impress",
				"Always sign with the applicant's actual name that was provided, never a placeholder",
				"Always end the letter with 'Best regards,' followed by the applicant's name",
				"Do not use em dashes (—) in your writing, use commas or rewrite the sentence instead",
			}
			return append(constraints, additionalConstraints...)
		}(),
		OutputSpec: "Return ONLY a valid JSON object with a 'content' field containing the cover letter text with proper formatting (\\n for line breaks)",
	}
}

// EnhanceCoverLetterPrompt enhances a cover letter prompt
func EnhanceCoverLetterPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, wordRange string) string {
	template := CoverLetterTemplate()
	params := map[string]any{
		"wordRange":   wordRange,
		"currentDate": time.Now().Format("January 2, 2006"),
	}
	return template.BuildPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, params)
}
