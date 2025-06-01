package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

// emailRegex matches email addresses
var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

// PrivacyLogger wraps zerolog.Logger with GDPR-compliant logging methods
type PrivacyLogger struct {
	zerolog.Logger
}

// NewPrivacyLogger creates a new PrivacyLogger instance
func NewPrivacyLogger(logger zerolog.Logger) *PrivacyLogger {
	return &PrivacyLogger{Logger: logger}
}

// HashIdentifier creates a one-way hash of an identifier for tracking purposes
// This allows correlation without storing the actual identifier
func HashIdentifier(identifier string) string {
	if identifier == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity
}

// RedactEmail replaces email addresses with a redacted version
func RedactEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "[REDACTED_EMAIL]"
	}

	// Keep first and last char of local part
	local := parts[0]
	if len(local) > 2 {
		local = string(local[0]) + "***" + string(local[len(local)-1])
	} else {
		local = "***"
	}

	return local + "@" + parts[1]
}

// RedactUsername replaces username with a redacted version
func RedactUsername(username string) string {
	if username == "" {
		return ""
	}

	// If it's an email, use email redaction
	if emailRegex.MatchString(username) {
		return RedactEmail(username)
	}

	// For regular usernames, show first and last character
	if len(username) > 2 {
		return string(username[0]) + "***" + string(username[len(username)-1])
	}
	return "***"
}

// UserEvent logs a user-related event in a GDPR-compliant way
func (pl *PrivacyLogger) UserEvent(event string) *zerolog.Event {
	return pl.Info().Str("event", event)
}

// UserError logs a user-related error in a GDPR-compliant way
func (pl *PrivacyLogger) UserError(event string, err error) *zerolog.Event {
	return pl.Error().Err(err).Str("event", event)
}

// WithUserContext adds user context in a privacy-compliant way
func (pl *PrivacyLogger) WithUserContext(userID int, correlationID string) *PrivacyLogger {
	logger := pl.With().
		Str("correlation_id", correlationID).
		Str("user_ref", fmt.Sprintf("user_%d", userID)).
		Logger()
	return &PrivacyLogger{Logger: logger}
}

// WithHashedIdentifier adds a hashed identifier for correlation
func (pl *PrivacyLogger) WithHashedIdentifier(identifier string) *PrivacyLogger {
	logger := pl.With().
		Str("hashed_id", HashIdentifier(identifier)).
		Logger()
	return &PrivacyLogger{Logger: logger}
}

// LogAuthEvent logs authentication events without exposing PII
func (pl *PrivacyLogger) LogAuthEvent(event string, userID int, success bool) {
	pl.Info().
		Str("event", event).
		Str("user_ref", fmt.Sprintf("user_%d", userID)).
		Bool("success", success).
		Msg("Authentication event")
}

// LogRegistrationEvent logs registration events without exposing PII
func (pl *PrivacyLogger) LogRegistrationEvent(event string, hashedIdentifier string, success bool) {
	pl.Info().
		Str("event", event).
		Str("hashed_id", hashedIdentifier).
		Bool("success", success).
		Msg("Registration event")
}

// SanitizeLogMessage removes potential PII from log messages
func SanitizeLogMessage(message string) string {
	sanitized := emailRegex.ReplaceAllString(message, "[REDACTED_EMAIL]")

	return sanitized
}
