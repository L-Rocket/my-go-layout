// Package config loads process configuration: defaults < configs/config.yaml
// < environment variables. Env always wins so deployments can override a
// checked-in file without editing it.
package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config is the full runtime configuration.
type Config struct {
	Port               int    `yaml:"port"`
	GRPCPort           int    `yaml:"grpc_port"`
	LogLevel           string `yaml:"log_level"`
	SupabaseURL        string `yaml:"supabase_url"`
	SupabaseServiceKey string `yaml:"supabase_service_key"`
}

const configPath = "configs/config.yaml"

// Load returns defaults, overlaid with configs/config.yaml when it
// exists, overlaid with any set environment variables. A malformed yaml
// file is fatal-by-error: silently ignoring it would hide typos.
func Load() (*Config, error) {
	cfg := &Config{
		Port:     8080,
		GRPCPort: 50051,
		LogLevel: "info",
	}

	data, err := os.ReadFile(configPath)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	case !os.IsNotExist(err):
		return nil, err
	}

	// env wins over everything
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
	if v := os.Getenv("SUPABASE_URL"); v != "" {
		cfg.SupabaseURL = v
	}
	if v := os.Getenv("SUPABASE_SERVICE_ROLE_KEY"); v != "" {
		cfg.SupabaseServiceKey = v
	}
	return cfg, nil
}

// UsePostgREST reports whether durable storage is configured. BOTH values
// are required — a URL without the key would only produce 401s at runtime.
func (c *Config) UsePostgREST() bool {
	return c.SupabaseURL != "" && c.SupabaseServiceKey != ""
}
