package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Initialize(isDevelopment bool, logLevel string) {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stderr

	if isDevelopment {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	log.Logger = zerolog.New(output).With().
		Timestamp().
		Caller().
		Logger()
}

// GetLogger returns a zerolog.Logger instance with the specified module name
// added as a contextual field.
//
// This allows for structured logging with module-specific context.
func GetLogger(module string) zerolog.Logger {
	return log.Logger.With().Str("module", module).Logger()
}

// GetPrivacyLogger returns a PrivacyLogger instance with GDPR-compliant logging methods
func GetPrivacyLogger(module string) *PrivacyLogger {
	return NewPrivacyLogger(GetLogger(module))
}
