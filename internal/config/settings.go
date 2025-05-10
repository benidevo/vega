package config

import (
	"os"
	"strconv"
	"time"
)

// Settings holds the configuration settings for the application.
type Settings struct {
	ServerPort         string
	DBConnectionString string
	DBDriver           string
	LogLevel           string
	IsDevelopment      bool
	TokenExpiration    time.Duration
	TokenSecret        string
	IsTest             bool
}

// NewSettings initializes and returns a Settings struct with default values
// populated from environment variables. If an environment variable is
// not set, a predefined default value is used.
func NewSettings() Settings {
	tokenExpiration, err := strconv.Atoi(getEnv("TOKEN_EXPIRATION", "60"))
	if err != nil {
		tokenExpiration = 60 // Default to 60 minutes if conversion fails
	}

	return Settings{
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		DBConnectionString: getEnv("DB_CONNECTION_STRING", "/app/data/prospector.db?_journal=WAL&_busy_timeout=5000"),
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		IsDevelopment:      getEnv("IS_DEVELOPMENT", "false") == "true",
		TokenExpiration:    time.Duration(tokenExpiration) * time.Minute,
		TokenSecret:        getEnv("TOKEN_SECRET", "default-secret"),
		IsTest:             getEnv("GO_ENV", "") == "test",
	}
}

func getEnv(key string, defaultValue string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return
}
