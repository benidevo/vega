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
	accessTokenMins, err := strconv.Atoi(getEnv("ACCESS_TOKEN_EXPIRY", "60"))
	if err != nil {
		accessTokenMins = 60 // Default to 60 minutes if conversion fails
	}

	refreshTokenHours, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRY", "168"))
	if err != nil {
		refreshTokenHours = 168 // 7 days
	}

	dbMaxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		dbMaxOpenConns = 25
	}

	dbMaxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		dbMaxIdleConns = 5
	}

	dbConnMaxLifetimeMins, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME_MINS", "5"))
	if err != nil {
		dbConnMaxLifetimeMins = 5
	}

	cookieSecure := getEnv("COOKIE_SECURE", "") == "true"
	isDevelopment := getEnv("IS_DEVELOPMENT", "false") == "true"

	// In development mode, cookies are not secure by default
	if isDevelopment && getEnv("COOKIE_SECURE", "") == "" {
		cookieSecure = false
	}

	isTest := getEnv("GO_ENV", "") == "test"

	// Use in-memory database for tests by default
	dbConnectionString := getEnv("DB_CONNECTION_STRING", "/app/data/vega.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_cache_size=10000&_synchronous=NORMAL")
	if isTest && getEnv("DB_CONNECTION_STRING", "") == "" {
		dbConnectionString = ":memory:"
	}

	return Settings{
		AppName:            "vega",
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		DBConnectionString: dbConnectionString,
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		IsDevelopment:      isDevelopment,
		TokenSecret:        getEnv("TOKEN_SECRET", ""),
		IsTest:             isTest,
		MigrationsDir:      getEnv("MIGRATIONS_DIR", "migrations/sqlite"),

		DBMaxOpenConns:    dbMaxOpenConns,
		DBMaxIdleConns:    dbMaxIdleConns,
		DBConnMaxLifetime: time.Duration(dbConnMaxLifetimeMins) * time.Minute,

		AccessTokenExpiry:  time.Duration(accessTokenMins) * time.Minute,
		RefreshTokenExpiry: time.Duration(refreshTokenHours) * time.Hour,
		CookieDomain:       getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:       cookieSecure,
		CookieSameSite:     getEnv("COOKIE_SAME_SITE", "lax"),

		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleClientRedirectURL: getEnv("GOOGLE_CLIENT_REDIRECT_URL", "http://localhost:8000/auth/google/callback"),
		GoogleAuthUserInfoURL:   getEnv("GOOGLE_AUTH_USER_INFO_URL", "https://www.googleapis.com/oauth2/v3/userinfo"),
		GoogleAuthUserInfoScope: getEnv("GOOGLE_AUTH_USER_INFO_SCOPE", "https://www.googleapis.com/auth/userinfo.email"),

		CORSAllowedOrigins:   getCORSOrigins(),
		CORSAllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "false") == "true",

		CreateAdminUser:    getEnv("CREATE_ADMIN_USER", "false") == "true",
		AdminUsername:      getEnv("ADMIN_USERNAME", ""),
		AdminPassword:      getEnv("ADMIN_PASSWORD", ""),
		ResetAdminPassword: getEnv("RESET_ADMIN_PASSWORD", "false") == "true",
		AdminEmail:         getEnv("ADMIN_EMAIL", ""),
		AIProvider:         getEnv("AI_PROVIDER", "gemini"),
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

// getCORSOrigins parses the CORS_ALLOWED_ORIGINS environment variable
// and returns a slice of origins. If not set, returns localhost origins for development.
func getCORSOrigins() []string {
	origins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:8080")
	if origins == "" {
		return []string{"http://localhost:8080"}
	}

	originList := strings.Split(origins, ",")
	for i, origin := range originList {
		originList[i] = strings.TrimSpace(origin)
	}

	return originList
}
