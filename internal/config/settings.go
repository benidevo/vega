package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	IsTest             bool
	MigrationsDir      string

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	TokenSecret        string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	CookieDomain       string
	CookieSecure       bool
	CookieSameSite     string

	GoogleClientID          string
	GoogleClientSecret      string
	GoogleClientRedirectURL string
	GoogleAuthUserInfoURL   string
	GoogleAuthUserInfoScope string

	CORSAllowedOrigins   []string
	CORSAllowCredentials bool

	CreateAdminUser    bool
	AdminUsername      string
	AdminPassword      string
	AdminEmail         string
	ResetAdminPassword bool

	AIProvider   string
	GeminiAPIKey string
	GeminiModel  string
}

// NewSettings initializes and returns a Settings struct with default values
// populated from environment variables. If an environment variable is
// not set, a predefined default value is used.
func NewSettings() Settings {
	isDevelopment := getEnv("IS_DEVELOPMENT", "false") == "true"
	isTest := getEnv("GO_ENV", "") == "test"

	// Production-optimized defaults
	accessTokenExpiry := 60 * time.Minute // 1 hour
	refreshTokenExpiry := 168 * time.Hour // 7 days
	dbMaxOpenConns := 25
	dbMaxIdleConns := 5
	dbConnMaxLifetime := 5 * time.Minute

	if envVal := getEnv("ACCESS_TOKEN_EXPIRY", ""); envVal != "" {
		if mins, err := strconv.Atoi(envVal); err == nil {
			accessTokenExpiry = time.Duration(mins) * time.Minute
		}
	}

	if envVal := getEnv("REFRESH_TOKEN_EXPIRY", ""); envVal != "" {
		if hours, err := strconv.Atoi(envVal); err == nil {
			refreshTokenExpiry = time.Duration(hours) * time.Hour
		}
	}

	// Database connection string with sensible production default
	dbConnectionString := "/app/data/vega.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_cache_size=10000&_synchronous=NORMAL"
	if envDB := getEnv("DB_CONNECTION_STRING", ""); envDB != "" {
		dbConnectionString = envDB
	}
	if isTest && getEnv("DB_CONNECTION_STRING", "") == "" {
		dbConnectionString = ":memory:"
	}

	cookieSecure := !isDevelopment
	if envSecure := getEnv("COOKIE_SECURE", ""); envSecure != "" {
		cookieSecure = envSecure == "true"
	}

	corsOrigins := []string{"http://localhost:8080"}
	if isDevelopment {
		corsOrigins = []string{"http://localhost:8080", "http://localhost:8000"}
	}
	if envCORS := getEnv("CORS_ALLOWED_ORIGINS", ""); envCORS != "" {
		corsOrigins = strings.Split(envCORS, ",")
		for i, origin := range corsOrigins {
			corsOrigins[i] = strings.TrimSpace(origin)
		}
	}

	return Settings{
		AppName:            "vega",
		ServerPort:         ":8080",
		DBConnectionString: dbConnectionString,
		DBDriver:           "sqlite",
		LogLevel:           getEnv("LOG_LEVEL", getDefaultLogLevel(isDevelopment)),
		IsDevelopment:      isDevelopment,
		TokenSecret:        getEnv("TOKEN_SECRET", ""),
		IsTest:             isTest,
		MigrationsDir:      "migrations/sqlite",

		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: dbConnMaxLifetime,

		AccessTokenExpiry:  accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
		CookieDomain:       getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:       cookieSecure,
		CookieSameSite:     "lax",

		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleClientRedirectURL: getEnv("GOOGLE_CLIENT_REDIRECT_URL", "http://localhost:8000/auth/google/callback"),
		GoogleAuthUserInfoURL:   "https://www.googleapis.com/oauth2/v3/userinfo",
		GoogleAuthUserInfoScope: "https://www.googleapis.com/auth/userinfo.email",

		CORSAllowedOrigins:   corsOrigins,
		CORSAllowCredentials: false,

		CreateAdminUser:    getEnv("CREATE_ADMIN_USER", "false") == "true",
		AdminUsername:      getEnv("ADMIN_USERNAME", ""),
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		ResetAdminPassword: getEnv("RESET_ADMIN_PASSWORD", "false") == "true",
		AdminEmail:         getEnv("ADMIN_EMAIL", ""),
		AIProvider:         "gemini",
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		GeminiModel:        "gemini-2.5-flash",
	}
}

// NewTestSettings creates settings optimized for testing with in-memory database
func NewTestSettings() Settings {
	settings := NewSettings()
	settings.IsTest = true
	settings.DBConnectionString = ":memory:"
	settings.ServerPort = ":0" // Use random available port
	return settings
}

// NewTestSettingsWithTempDB creates settings for testing with a temporary file database
// This is useful for tests that need persistence or migration testing
// The caller is responsible for cleaning up the returned temp file path
func NewTestSettingsWithTempDB() (Settings, string) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("vega_test_%d.db", time.Now().UnixNano()))

	settings := NewSettings()
	settings.IsTest = true
	settings.DBConnectionString = tempFile + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"
	settings.ServerPort = ":0"

	return settings, tempFile
}

func getEnv(key string, defaultValue string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return
}

// getDefaultLogLevel returns appropriate log level based on environment
func getDefaultLogLevel(isDevelopment bool) string {
	if isDevelopment {
		return "debug"
	}
	return "info"
}
