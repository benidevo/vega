package config

import (
	"os"
	"strconv"
	"time"
)

// Settings holds the configuration settings for the application.
type Settings struct {
	AppName            string
	ServerPort         string
	DBConnectionString string
	DBDriver           string
	LogLevel           string
	IsDevelopment      bool
	TokenExpiration    time.Duration
	TokenSecret        string
	IsTest             bool
	MigrationsDir      string

	GoogleClientConfigFile  string
	GoogleClientRedirectURL string
	GoogleAuthUserInfoURL   string
	GoogleAuthUserInfoScope string
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
		AppName:                 "Prospector",
		ServerPort:              getEnv("SERVER_PORT", ":8080"),
		DBConnectionString:      getEnv("DB_CONNECTION_STRING", "/app/data/prospector.db?_journal=WAL&_busy_timeout=5000"),
		DBDriver:                getEnv("DB_DRIVER", "sqlite"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		IsDevelopment:           getEnv("IS_DEVELOPMENT", "false") == "true",
		TokenExpiration:         time.Duration(tokenExpiration) * time.Minute,
		TokenSecret:             getEnv("TOKEN_SECRET", "default-secret"),
		IsTest:                  getEnv("GO_ENV", "") == "test",
		MigrationsDir:           getEnv("MIGRATIONS_DIR", "migrations/sqlite"),
		GoogleClientConfigFile:  getEnv("GOOGLE_CLIENT_CONFIG_FILE", "config/google_oauth_credentials.json"),
		GoogleClientRedirectURL: getEnv("GOOGLE_CLIENT_REDIRECT_URL", "http://localhost:8000/auth/google/callback"),
		GoogleAuthUserInfoURL:   getEnv("GOOGLE_AUTH_USER_INFO_URL", "https://www.googleapis.com/oauth2/v3/userinfo"),
		GoogleAuthUserInfoScope: getEnv("GOOGLE_AUTH_USER_INFO_SCOPE", "https://www.googleapis.com/auth/userinfo.email"),
	}
}

func getEnv(key string, defaultValue string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return
}
