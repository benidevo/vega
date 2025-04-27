package prospector

import "os"

// Config holds the configuration settings for the application,
type Config struct {
	ServerPort         string
	DBConnectionString string
	DBDriver           string
}

// NewConfig initializes and returns a Config struct with default values
// populated from environment variables. If an environment variable is
// not set, a predefined default value is used.
func NewConfig() Config {
	return Config{
		ServerPort:         getEnv("SERVER_PORT", ":8080"),
		DBConnectionString: getEnv("DB_CONNECTION_STRING", "/app/data/prospector.db?_journal=WAL&_busy_timeout=5000"),
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
	}
}

func getEnv(key string, defaultValue string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return
}
