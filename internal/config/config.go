// Package config loads process configuration from environment variables
// with defaults. Keep it boring: no config files, no precedence rules —
// a demo reads env or falls back.
package config

import (
	"os"
	"strconv"
)

// Config is the full runtime configuration.
type Config struct {
	Port               int
	GRPCPort           int
	LogLevel           string
	SupabaseURL        string
	SupabaseServiceKey string
}

// Load returns defaults overlaid with any set environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		Port:     8080,
		GRPCPort: 50051,
		LogLevel: "info",
	}
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Port = n
		}
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.GRPCPort = n
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	cfg.SupabaseURL = os.Getenv("SUPABASE_URL")
	cfg.SupabaseServiceKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	return cfg, nil
}

// UsePostgREST reports whether durable storage is configured. BOTH values
// are required — a URL without the key would only produce 401s at runtime.
func (c *Config) UsePostgREST() bool {
	return c.SupabaseURL != "" && c.SupabaseServiceKey != ""
}
