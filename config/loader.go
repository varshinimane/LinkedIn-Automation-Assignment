package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Load reads YAML config and overlays environment credentials from .env if present.
// Applies sensible defaults and validates required fields.
func Load(path string) (Config, error) {
	if err := godotenv.Load(); err != nil {
		// Non-fatal; proceed if .env missing.
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	// Override with environment variables if present
	if v := os.Getenv("LINKEDIN_EMAIL"); v != "" {
		cfg.Credentials.Email = v
	}
	if v := os.Getenv("LINKEDIN_PASSWORD"); v != "" {
		cfg.Credentials.Password = v
	}

	// Apply defaults
	applyDefaults(&cfg)

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return cfg, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// applyDefaults sets sensible defaults for optional fields.
func applyDefaults(cfg *Config) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://www.linkedin.com"
	}
	if cfg.Paths.DBPath == "" {
		cfg.Paths.DBPath = "./data/state.db"
	}
	if cfg.Paths.CookieFile == "" {
		cfg.Paths.CookieFile = "./cookies.json"
	}
	if cfg.Humanization.MinDelayMs == 0 {
		cfg.Humanization.MinDelayMs = 500
	}
	if cfg.Humanization.MaxDelayMs == 0 {
		cfg.Humanization.MaxDelayMs = 1800
	}
	if cfg.Humanization.JitterMs == 0 {
		cfg.Humanization.JitterMs = 400
	}
	if cfg.Humanization.ThinkTimeMs == 0 {
		cfg.Humanization.ThinkTimeMs = 900
	}
	if cfg.Humanization.ScrollPauseMs == 0 {
		cfg.Humanization.ScrollPauseMs = 900
	}
	if cfg.Limits.DailyConnections == 0 {
		cfg.Limits.DailyConnections = 10
	}
	if cfg.Limits.DailyMessages == 0 {
		cfg.Limits.DailyMessages = 8
	}
	if cfg.Search.PaginationPages == 0 {
		cfg.Search.PaginationPages = 2
	}
	if cfg.Messaging.FollowupDelayHour == 0 {
		cfg.Messaging.FollowupDelayHour = 24
	}
	if len(cfg.Browser.UserAgents) == 0 {
		cfg.Browser.UserAgents = []string{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		}
	}
	if len(cfg.Browser.Viewports) == 0 {
		cfg.Browser.Viewports = []struct {
			Width  int     `yaml:"width"`
			Height int     `yaml:"height"`
			Scale  float64 `yaml:"scale"`
		}{
			{Width: 1440, Height: 900, Scale: 1.0},
		}
	}
	if cfg.Browser.Timezone == "" {
		cfg.Browser.Timezone = "America/Los_Angeles"
	}
}

// validate checks required fields and returns an error if validation fails.
func validate(cfg *Config) error {
	if cfg.Credentials.Email == "" {
		return fmt.Errorf("credentials.email is required (set in config.yaml or LINKEDIN_EMAIL env var)")
	}
	if cfg.Credentials.Password == "" {
		return fmt.Errorf("credentials.password is required (set in config.yaml or LINKEDIN_PASSWORD env var)")
	}
	return nil
}



