package security

import (
	"regexp"
	"strings"
)

// PromptSanitizer provides methods to sanitize user input before using in LLM prompts
type PromptSanitizer struct {
	// Patterns that could be used for prompt injection
	injectionPatterns []*regexp.Regexp
}

// NewPromptSanitizer creates a new instance of PromptSanitizer
func NewPromptSanitizer() *PromptSanitizer {
	patterns := []*regexp.Regexp{
		// System instruction attempts
		regexp.MustCompile(`(?i)system\s*[:：]\s*`),
		regexp.MustCompile(`(?i)assistant\s*[:：]\s*`),
		regexp.MustCompile(`(?i)human\s*[:：]\s*`),

		// Role manipulation attempts
		regexp.MustCompile(`(?i)you\s+are\s+now\s+`),
		regexp.MustCompile(`(?i)ignore\s+(previous|all)\s+(instructions?|prompts?)`),
		regexp.MustCompile(`(?i)forget\s+(everything|all)\s+`),

		// Instruction override attempts
		regexp.MustCompile(`(?i)new\s+instructions?\s*[:：]`),
		regexp.MustCompile(`(?i)updated?\s+instructions?\s*[:：]`),
		regexp.MustCompile(`(?i)instead\s+of\s+`),

		// JSON/Code injection attempts
		regexp.MustCompile(`[{}]\s*"[^"]*"\s*[:：]\s*[{}]`),
		regexp.MustCompile(`(?i)return\s+(json|code|script)`),
	}

	return &PromptSanitizer{
		injectionPatterns: patterns,
	}
}

// SanitizeText removes potential prompt injection patterns from user input
func (s *PromptSanitizer) SanitizeText(input string) string {
	if input == "" {
		return input
	}

	sanitized := input

	for _, pattern := range s.injectionPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "[FILTERED]")
	}

	if len(sanitized) > 10000 {
		sanitized = sanitized[:10000] + "... [TRUNCATED]"
	}

	sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeCVText specifically sanitizes CV content
func (s *PromptSanitizer) SanitizeCVText(cvText string) string {
	return s.SanitizeText(cvText)
}

// SanitizeJobDescription specifically sanitizes job description content
func (s *PromptSanitizer) SanitizeJobDescription(jobDesc string) string {
	return s.SanitizeText(jobDesc)
}

// SanitizeInstructions sanitizes user-provided instructions
func (s *PromptSanitizer) SanitizeInstructions(instructions string) string {
	return s.SanitizeText(instructions)
}

// SanitizeExtraContext sanitizes additional context provided by user
func (s *PromptSanitizer) SanitizeExtraContext(context string) string {
	return s.SanitizeText(context)
}
