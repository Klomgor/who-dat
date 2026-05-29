// Package config loads runtime configuration from environment variables with sane
// defaults, so the same binary works locally, in Docker, and on Vercel.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port          string        // HTTP listen port (binary only)
	AuthKey       string        // optional bearer token; auth disabled when empty
	CacheTTL      time.Duration // how long lookups stay cached
	EnableCache   bool          // toggles the in-memory cache
	LookupTimeout time.Duration // per-request upstream timeout
	MaxDomains    int           // max domains per /multi request
	RatePerMinute int           // per-IP request budget; <= 0 disables rate limiting
	RateBurst     int           // per-IP burst allowance
}

// Load reads configuration from the environment.
func Load() *Config {
	return &Config{
		Port:          env("PORT", "8080"),
		AuthKey:       os.Getenv("AUTH_KEY"),
		CacheTTL:      time.Duration(intEnv("CACHE_TTL_SECONDS", 3600)) * time.Second,
		EnableCache:   boolEnv("ENABLE_CACHE", true),
		LookupTimeout: time.Duration(intEnv("LOOKUP_TIMEOUT_SECONDS", 10)) * time.Second,
		MaxDomains:    intEnv("MAX_DOMAINS", 10),
		RatePerMinute: intEnv("RATE_PER_MINUTE", 60),
		RateBurst:     intEnv("RATE_BURST", 20),
	}
}

// AuthEnabled reports whether bearer auth is configured.
func (c *Config) AuthEnabled() bool { return c.AuthKey != "" }

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func boolEnv(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
