# GDPR-Compliant Logging Guidelines

## Overview

This document outlines the GDPR-compliant logging practices implemented in Vega to protect user privacy and personal data.

## Core Principles

1. **No Direct PII in Logs**: Never log personally identifiable information (PII) directly
2. **Use References**: Use anonymous references (e.g., `user_123`) instead of usernames or emails
3. **Hash Identifiers**: When correlation is needed, use one-way hashes of identifiers
4. **Event-Based Logging**: Log events and outcomes, not personal data

## What NOT to Log

- ❌ Email addresses
- ❌ Usernames
- ❌ Full names
- ❌ IP addresses (unless absolutely necessary and anonymized)
- ❌ Passwords (even hashed ones)
- ❌ OAuth tokens or API keys
- ❌ Any data that can directly identify a person

## What to Log Instead

- ✅ User references: `user_123` instead of usernames
- ✅ Hashed identifiers for correlation
- ✅ Event types: `login_success`, `registration_failed`
- ✅ Anonymous metrics and counts
- ✅ System errors without user context

## Privacy-Aware Logging Utilities

The application provides privacy-aware logging utilities in `internal/common/logger/privacy.go`:

### PrivacyLogger

```go
// Get a privacy-aware logger
log := logger.GetPrivacyLogger("module_name")

// Log authentication events without exposing PII
log.LogAuthEvent("login_success", userID, true)

// Log registration events with hashed identifier
log.LogRegistrationEvent("user_registered", logger.HashIdentifier(email), true)

// Log with user context
log.WithUserContext(userID, correlationID).Info().Msg("Action completed")
```

### Helper Functions

- `HashIdentifier(identifier)` - One-way hash for correlation
- `RedactEmail(email)` - Redacts to `j***e@example.com`
- `RedactUsername(username)` - Redacts to `j***e`
- `SanitizeLogMessage(message)` - Removes PII from messages

## Implementation Examples

### Authentication Logging

```go
// Bad - Logs email directly
log.Info().Str("email", email).Msg("User logged in")

// Good - Uses hashed identifier
log.LogAuthEvent("login_success", userID, true)
```

### Error Logging

```go
// Bad - Includes username in error
log.Error().Err(err).Str("username", username).Msg("Login failed")

// Good - Uses event type and hashed ID
log.Error().Err(err).
    Str("event", "login_failed").
    Str("hashed_id", logger.HashIdentifier(username)).
    Msg("Login attempt failed")
```

### User Actions

```go
// Bad - Logs user email with action
log.Info().Str("email", user.Email).Msg("User updated profile")

// Good - Uses user reference
log.Info().
    Str("event", "profile_updated").
    Str("user_ref", fmt.Sprintf("user_%d", userID)).
    Msg("Profile updated successfully")
```

## Structured Logging Fields

Use consistent field names for better log analysis:

- `event`: The type of event (e.g., "login_success", "registration_failed")
- `user_ref`: Anonymous user reference (e.g., "user_123")
- `hashed_id`: Hashed identifier for correlation
- `correlation_id`: Request correlation ID
- `error_type`: Type of error without PII
- `success`: Boolean for operation outcome

## Compliance Checklist

Before committing code, ensure:

- [ ] No emails, usernames, or names in log statements
- [ ] User IDs are prefixed with "user_" when logged
- [ ] Authentication events use the LogAuthEvent method
- [ ] Error messages don't contain PII
- [ ] Correlation uses hashed identifiers, not raw data
- [ ] Log levels are appropriate (don't log sensitive data at DEBUG level)

## Testing Privacy Compliance

Run the privacy tests to ensure utilities work correctly:

```bash
go test ./internal/common/logger/...
```

Remember: When in doubt, don't log it!
