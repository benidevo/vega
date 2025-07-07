package storage

import (
	"time"

	jobmodels "github.com/benidevo/vega/internal/job/models"
	settingsmodels "github.com/benidevo/vega/internal/settings/models"
)

type (
	Company     = jobmodels.Company
	Job         = jobmodels.Job
	MatchResult = jobmodels.MatchResult
	Profile     = settingsmodels.Profile
)

// UserData represents all data for a single user
type UserData struct {
	Version   string       `json:"version"`
	UpdatedAt time.Time    `json:"updated_at"`
	Checksum  string       `json:"checksum,omitempty"`
	Data      UserDataCore `json:"data"`
}

// UserDataCore contains the actual user data
type UserDataCore struct {
	Profile   *Profile       `json:"profile,omitempty"`
	Companies []*Company     `json:"companies,omitempty"`
	Jobs      []*Job         `json:"jobs,omitempty"`
	Matches   []*MatchResult `json:"matches,omitempty"`
}

// StorageMetadata contains metadata about the storage
type StorageMetadata struct {
	LastSync time.Time `json:"last_sync"`
	UserID   string    `json:"user_id"`
	IsDirty  bool      `json:"is_dirty"`
}
