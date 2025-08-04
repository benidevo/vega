package constants

const (
	// Match categories
	MatchCategoryExcellent = "Excellent Match"
	MatchCategoryStrong    = "Strong Match"
	MatchCategoryGood      = "Good Match"
	MatchCategoryFair      = "Fair Match"
	MatchCategoryPartial   = "Partial Match"
	MatchCategoryPoor      = "Poor Match"

	// Match descriptions
	MatchDescExcellent = "You're an ideal candidate with all key qualifications and relevant experience. Apply with confidence."
	MatchDescStrong    = "You meet 80%+ of requirements with transferable skills covering gaps. Strong application potential."
	MatchDescGood      = "You have solid core qualifications. Address the skill gaps in your cover letter to strengthen your application."
	MatchDescFair      = "You show promise but lack some key requirements. Consider gaining experience in missing areas first."
	MatchDescPartial   = "Your profile shows potential but significant gaps exist. This role may be a stretch at this time."
	MatchDescPoor      = "Your current qualifications don't align with this role. Focus on building relevant skills and experience."

	// Score thresholds
	ScoreThresholdExcellent = 90
	ScoreThresholdStrong    = 80
	ScoreThresholdGood      = 70
	ScoreThresholdFair      = 60
	ScoreThresholdPartial   = 50

	// Service operation names for logging
	OperationJobMatch      = "job_match_analysis"
	OperationCoverLetter   = "cover_letter_generation"
	OperationMatchAnalysis = "match_analysis"

	// Error context types
	ErrorTypeAIServiceUnavailable = "ai_service_unavailable"
	ErrorTypeValidationFailed     = "validation_failed"
	ErrorTypeAIAnalysisFailed     = "ai_analysis_failed"
	ErrorTypeAIGenerationFailed   = "ai_generation_failed"
	ErrorTypeResponseParseFailed  = "response_parse_failed"
)
