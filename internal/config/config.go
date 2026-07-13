package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the goshrt application.
// Values are loaded from environment variables with sensible defaults.
type Config struct {
	// Port is the HTTP server listen address (default: ":8080").
	Port string

	// PostgresDSN is the PostgreSQL connection string.
	PostgresDSN string

	// RedisAddr is the Redis server address (host:port).
	RedisAddr string

	// RedisPassword is the Redis AUTH password (empty if none).
	RedisPassword string

	// ClickSyncInterval is how often click counts are synced from Redis to PostgreSQL.
	ClickSyncInterval time.Duration
}

// Load reads configuration from environment variables, applying defaults
// where values are not set. Returns a Config with all fields populated.
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://goshrt:goshrt@localhost:5432/goshrt?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		ClickSyncInterval: getDurationEnv("CLICK_SYNC_INTERVAL", 5*time.Second),
	}
}

// getEnv returns the value of an environment variable or a default if not set.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv returns the duration value of an environment variable or a default if not set.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// getIntEnv returns the integer value of an environment variable or a default if not set.
func getIntEnv(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
