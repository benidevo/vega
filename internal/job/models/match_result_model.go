package models

import (
	"time"
)

// MatchResult represents a historical match analysis result for a job.
//
// It stores the result of the analysis performed on a job-user profile combination.
type MatchResult struct {
	ID         int       `json:"id" db:"id" sql:"primary_key;auto_increment"`
	JobID      int       `json:"job_id" db:"job_id" sql:"type:integer;not null;index;references:jobs(id);on_delete:cascade"`
	MatchScore int       `json:"match_score" db:"match_score" sql:"type:integer;not null;check:match_score >= 0 AND match_score <= 100" validate:"required,min=0,max=100"`
	Strengths  []string  `json:"strengths" db:"strengths" sql:"type:text"`   // Stored as JSON
	Weaknesses []string  `json:"weaknesses" db:"weaknesses" sql:"type:text"` // Stored as JSON
	Highlights []string  `json:"highlights" db:"highlights" sql:"type:text"` // Stored as JSON
	Feedback   string    `json:"feedback" db:"feedback" sql:"type:text"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" sql:"type:timestamp;not null;default:current_timestamp"`
}

// MatchSummary represents a condensed version of a match result
// used for providing context in AI prompts.
type MatchSummary struct {
	JobTitle    string `json:"job_title"`
	Company     string `json:"company"`
	MatchScore  int    `json:"match_score"`
	KeyInsights string `json:"key_insights"` // Combined strengths/weaknesses
	DaysAgo     int    `json:"days_ago"`
}
