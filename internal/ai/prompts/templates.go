package prompts

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
	return EnhanceCoverLetterPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, wordRange)
}

// EnhanceJobMatchPrompt enhances a job matching prompt
func (pe *PromptEnhancer) EnhanceJobMatchPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext string, minScore, maxScore int) string {
	return EnhanceJobMatchPrompt(systemInstruction, applicantName, jobDescription, applicantProfile, extraContext, minScore, maxScore)
}

// EnhanceCVGenerationPrompt enhances a CV generation prompt
func (pe *PromptEnhancer) EnhanceCVGenerationPrompt(systemInstruction, cvText, jobDescription, extraContext string) string {
	return EnhanceCVGenerationPrompt(systemInstruction, cvText, jobDescription, extraContext)
}
