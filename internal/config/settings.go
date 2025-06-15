package config

import (
	"os"
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

	cookieSecure := getEnv("COOKIE_SECURE", "") == "true"
	isDevelopment := getEnv("IS_DEVELOPMENT", "false") == "true"

	// In development mode, cookies are not secure by default
	if isDevelopment && getEnv("COOKIE_SECURE", "") == "" {
		cookieSecure = false
	}

	return Settings{
		AppName:            "ascentio",
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		DBConnectionString: getEnv("DB_CONNECTION_STRING", "/app/data/ascentio.db?_journal=WAL&_busy_timeout=5000"),
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		IsDevelopment:      isDevelopment,
		TokenSecret:        getEnv("TOKEN_SECRET", ""),
		IsTest:             getEnv("GO_ENV", "") == "test",
		MigrationsDir:      getEnv("MIGRATIONS_DIR", "migrations/sqlite"),

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
	}
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
