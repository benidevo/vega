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
	MatchDescExcellent = "You are an outstanding fit for the role with minimal gaps."
	MatchDescStrong    = "You have strong qualifications with only minor areas for development."
	MatchDescGood      = "You meet most requirements with some skill gaps that can be addressed."
	MatchDescFair      = "You have potential but may need significant development in key areas."
	MatchDescPartial   = "You have some relevant qualifications but significant gaps exist."
	MatchDescPoor      = "You do not meet the core requirements for this position."

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
