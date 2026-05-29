package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port string

	// Authentication
	AuthKey string

	// Cache configuration
	CacheTTL time.Duration

	// Request limits
	MaxDomainsPerRequest int
	RequestTimeout       time.Duration

	// Feature flags
	EnableCache bool
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		AuthKey:              os.Getenv("AUTH_KEY"), // Optional
		CacheTTL:             getDurationEnv("CACHE_TTL_SECONDS", 3600) * time.Second,
		MaxDomainsPerRequest: getIntEnv("MAX_DOMAINS_PER_REQUEST", 10),
		RequestTimeout:       getDurationEnv("REQUEST_TIMEOUT_SECONDS", 5) * time.Second,
		EnableCache:          getBoolEnv("ENABLE_CACHE", true),
	}

	return cfg
}

// IsAuthEnabled returns true if authentication is configured
func (c *Config) IsAuthEnabled() bool {
	return c.AuthKey != ""
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return time.Duration(intVal)
		}
	}
	return time.Duration(defaultValue)
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
