package storage

import (
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
)

// Type aliases for cleaner code
type (
	Company     = jobmodels.Company
	Job         = jobmodels.Job
	MatchResult = jobmodels.MatchResult
	Profile     = settingsmodels.Profile
)

// StorageMetadata contains metadata about the storage
type StorageMetadata struct {
	LastSync time.Time `json:"last_sync"`
	UserID   string    `json:"user_id"`
	IsDirty  bool      `json:"is_dirty"`
}
