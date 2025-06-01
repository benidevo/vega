package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{
			name:       "empty identifier",
			identifier: "",
			want:       "",
		},
		{
			name:       "email identifier",
			identifier: "test@example.com",
			want:       "73d0b1b0ff1911e4", // First 8 bytes of SHA256
		},
		{
			name:       "username identifier",
			identifier: "john_doe",
			want:       "70c4834e062bc4f9", // First 8 bytes of SHA256
		},
		{
			name:       "consistent hashing",
			identifier: "test@example.com",
			want:       "73d0b1b0ff1911e4", // Should be same as above
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashIdentifier(tt.identifier)
			if tt.identifier != "" {
				assert.Equal(t, 16, len(got), "Hash should be 16 characters")
				// Verify consistency
				got2 := HashIdentifier(tt.identifier)
				assert.Equal(t, got, got2, "Hash should be consistent")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRedactEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{
			name:  "empty email",
			email: "",
			want:  "",
		},
		{
			name:  "standard email",
			email: "john.doe@example.com",
			want:  "j***e@example.com",
		},
		{
			name:  "short local part",
			email: "jo@example.com",
			want:  "***@example.com",
		},
		{
			name:  "single char local part",
			email: "j@example.com",
			want:  "***@example.com",
		},
		{
			name:  "invalid email format",
			email: "notanemail",
			want:  "[REDACTED_EMAIL]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactEmail(tt.email)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRedactUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{
			name:     "empty username",
			username: "",
			want:     "",
		},
		{
			name:     "standard username",
			username: "johndoe",
			want:     "j***e",
		},
		{
			name:     "short username",
			username: "jo",
			want:     "***",
		},
		{
			name:     "email as username",
			username: "john@example.com",
			want:     "j***n@example.com",
		},
		{
			name:     "username with underscore",
			username: "john_doe_123",
			want:     "j***3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactUsername(tt.username)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "message without PII",
			message: "User logged in successfully",
			want:    "User logged in successfully",
		},
		{
			name:    "message with email",
			message: "Failed login for user john@example.com from IP 192.168.1.1",
			want:    "Failed login for user [REDACTED_EMAIL] from IP 192.168.1.1",
		},
		{
			name:    "message with multiple emails",
			message: "Sending notification from admin@site.com to user@example.com",
			want:    "Sending notification from [REDACTED_EMAIL] to [REDACTED_EMAIL]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLogMessage(tt.message)
			assert.Equal(t, tt.want, got)
		})
	}
}
