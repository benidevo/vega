package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAccountActivity(t *testing.T) {
	lastLogin := time.Now().Add(-24 * time.Hour)
	createdAt := time.Now().Add(-30 * 24 * time.Hour)

	activity := NewAccountActivity(lastLogin, createdAt)

	assert.NotNil(t, activity)
	assert.Equal(t, lastLogin, activity.LastLogin)
	assert.Equal(t, createdAt, activity.CreatedAt)
}

func TestNewSecuritySettings(t *testing.T) {
	lastLogin := time.Now().Add(-24 * time.Hour)
	createdAt := time.Now().Add(-30 * 24 * time.Hour)
	activity := NewAccountActivity(lastLogin, createdAt)

	settings := NewSecuritySettings(activity)

	assert.NotNil(t, settings)
	assert.Equal(t, activity, settings.Activity)
}

func TestNewSecuritySettingsWithNilActivity(t *testing.T) {
	settings := NewSecuritySettings(nil)

	assert.NotNil(t, settings)
	assert.Nil(t, settings.Activity)
}
