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

	if wordRange, ok := params["wordRange"].(string); ok {
		promptBuilder.WriteString(fmt.Sprintf("**Word Count:** %s words\n\n", wordRange))
	}

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
