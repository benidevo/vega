package quota

import (
	"time"
)

// QuotaUsage represents a user's quota usage for a specific month
type QuotaUsage struct {
	UserID       int       `db:"user_id"`
	MonthYear    string    `db:"month_year"`
	JobsAnalyzed int       `db:"jobs_analyzed"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// QuotaStatus represents the current quota status for a user
type QuotaStatus struct {
	Used      int       `json:"used"`
	Limit     int       `json:"limit"`
	ResetDate time.Time `json:"reset_date"`
}

// QuotaCheckResult represents the result of a quota check
type QuotaCheckResult struct {
	Allowed bool        `json:"allowed"`
	Reason  string      `json:"reason"`
	Status  QuotaStatus `json:"status"`
}

// Job represents a minimal job structure for quota checking
type Job struct {
	ID              int        `db:"id"`
	FirstAnalyzedAt *time.Time `db:"first_analyzed_at"`
}

const (
	// FreeUserMonthlyLimit is the number of new job analyses allowed per month for free users
	FreeUserMonthlyLimit = 5

	// QuotaReasonOK indicates the operation is allowed
	QuotaReasonOK = "ok"

	// QuotaReasonReanalysis indicates this is a re-analysis (always allowed)
	QuotaReasonReanalysis = "re-analysis allowed"

	// QuotaReasonLimitReached indicates the monthly limit has been reached
	QuotaReasonLimitReached = "Monthly limit of 5 job analyses reached"
)
