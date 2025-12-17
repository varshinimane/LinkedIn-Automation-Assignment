package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config contains all tunable settings for the bot.
type Config struct {
	Headless bool   `yaml:"headless"`
	Debug    bool   `yaml:"debug"`
	BaseURL  string `yaml:"base_url"`

	Credentials struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"credentials"`

	Profile struct {
		UserAgent string `yaml:"user_agent"`
		Viewport  struct {
			Width  int     `yaml:"width"`
			Height int     `yaml:"height"`
			Scale  float64 `yaml:"scale"`
		} `yaml:"viewport"`
		Timezone string `yaml:"timezone"`
		Proxy    string `yaml:"proxy"`
	} `yaml:"profile"`

	Humanization struct {
		MinDelayMs     int `yaml:"min_delay_ms"`
		MaxDelayMs     int `yaml:"max_delay_ms"`
		JitterMs       int `yaml:"jitter_ms"`
		ScrollPauseMs  int `yaml:"scroll_pause_ms"`
		TypingVariance int `yaml:"typing_variance"`
	} `yaml:"humanization"`

	Persistence struct {
		DBPath string `yaml:"db_path"`
	} `yaml:"persistence"`

	Behavior struct {
		TargetSearchQueries   []string `yaml:"target_search_queries"`
		DailyConnectionLimit  int      `yaml:"daily_connection_limit"`
		DailyMessageLimit     int      `yaml:"daily_message_limit"`
		MessageTemplates      []string `yaml:"message_templates"`
		Blacklist             []string `yaml:"blacklist"`
	} `yaml:"behavior"`
}

// Load reads the YAML config from disk and returns a Config.
func Load(path string) (Config, error) {
	var cfg Config
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

