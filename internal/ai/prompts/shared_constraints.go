package prompts

// SharedConstraints contains reusable prompt constraints and rules
const (
	// BannedAIPhrases contains all corporate buzzwords and AI-sounding language to avoid
	BannedAIPhrases = "'leverage', 'utilize', 'spearheaded', 'orchestrated', 'synergies', 'cutting-edge', 'innovative solutions', 'dynamic', 'passionate', 'results-driven', 'detail-oriented', 'team player', 'go-getter', 'game-changer', 'disruptive', 'seamless', 'robust', 'scalable', 'streamlined', 'optimized', 'enhanced', 'facilitated', 'collaborated with stakeholders', 'deep dive', 'circle back', 'deliverables', 'action items', 'learnings', 'best practices', 'low-hanging fruit', 'value-add'"
)

// AntiAILanguageConstraints returns standardized constraints for eliminating AI-sounding language
func AntiAILanguageConstraints() []string {
	return []string{
		"CRITICAL - WRITE LIKE A HUMAN, NOT AI:",
		"BANNED PHRASES: Never use " + BannedAIPhrases,
		"NATURAL LANGUAGE ONLY: Write like explaining to a colleague or friend",
		"NO CORPORATE BUZZWORDS: Use plain English, not HR jargon or template language",
		"CONVERSATIONAL TONE: Sound like you're talking to someone face-to-face",
		"SPECIFIC > GENERIC: Use concrete details instead of vague statements",
		"HUMAN TEST: If it sounds like AI analysis or template, rewrite completely",
	}
}

// CVAntiAIConstraints returns CV-specific anti-AI language rules
func CVAntiAIConstraints() []string {
	constraints := AntiAILanguageConstraints()
	cvSpecific := []string{
		"TRANSFORM and ENHANCE: Reframe basic responsibilities as achievements",
		"ELEVATE TRUTHFULLY: Enhance language while maintaining complete honesty",
		"Present BEST FOOT FORWARD: Use impactful language without fabricating",
	}
	return append(constraints, cvSpecific...)
}

// CoverLetterAntiAIConstraints returns cover letter-specific anti-AI language rules
func CoverLetterAntiAIConstraints() []string {
	constraints := AntiAILanguageConstraints()
	letterSpecific := []string{
		"Write like having a professional conversation, not submitting a formal document",
		"Show genuine interest and personality - this is from a real person",
		"Skip tired openings like 'I am writing to apply for' - start engaging",
	}
	return append(constraints, letterSpecific...)
}

// JobMatchAntiAIConstraints returns job match analysis-specific anti-AI language rules
func JobMatchAntiAIConstraints() []string {
	constraints := AntiAILanguageConstraints()
	matchSpecific := []string{
		"NATURAL RECRUITER LANGUAGE: Write like an experienced hiring manager giving feedback",
		"CONVERSATIONAL BUT DIRECT: Sound like face-to-face conversation, not formal assessment",
		"Point to concrete examples and gaps, avoid vague statements",
	}
	return append(constraints, matchSpecific...)
}
