package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all yoy configuration.
type Config struct {
	// OutputFormat is the default output format (table, json, plain).
	OutputFormat string `yaml:"output_format,omitempty"`

	// ColorMode controls color output (auto, always, never).
	ColorMode string `yaml:"color_mode,omitempty"`

	// DefaultFolder is the default mail folder.
	DefaultFolder string `yaml:"default_folder,omitempty"`

	// MailLimit is the default number of messages to show.
	MailLimit int `yaml:"mail_limit,omitempty"`

	path string `yaml:"-"`
}

// Load reads the config from disk. Returns defaults if file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{
		OutputFormat:  DefaultOutputFormat,
		ColorMode:     DefaultColorMode,
		DefaultFolder: DefaultFolder,
		MailLimit:     DefaultMailLimit,
		path:          FilePath(),
	}

	data, err := os.ReadFile(cfg.path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.applyEnvOverrides()
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.applyEnvOverrides()

	return cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Get returns a config value by key name.
func (c *Config) Get(key string) (string, error) {
	switch strings.ToLower(key) {
	case "output_format":
		return c.OutputFormat, nil
	case "color_mode":
		return c.ColorMode, nil
	case "default_folder":
		return c.DefaultFolder, nil
	case "mail_limit":
		return fmt.Sprintf("%d", c.MailLimit), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set updates a config value by key name.
func (c *Config) Set(key, value string) error {
	switch strings.ToLower(key) {
	case "output_format":
		if value != "table" && value != "json" && value != "plain" {
			return fmt.Errorf("invalid output format %q: must be table, json, or plain", value)
		}
		c.OutputFormat = value
	case "color_mode":
		if value != "auto" && value != "always" && value != "never" {
			return fmt.Errorf("invalid color mode %q: must be auto, always, or never", value)
		}
		c.ColorMode = value
	case "default_folder":
		c.DefaultFolder = value
	case "mail_limit":
		var limit int
		if _, err := fmt.Sscanf(value, "%d", &limit); err != nil || limit < 1 {
			return fmt.Errorf("invalid mail limit %q: must be a positive integer", value)
		}
		c.MailLimit = limit
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// List returns all config key-value pairs.
func (c *Config) List() map[string]string {
	return map[string]string{
		"output_format":  c.OutputFormat,
		"color_mode":     c.ColorMode,
		"default_folder": c.DefaultFolder,
		"mail_limit":     fmt.Sprintf("%d", c.MailLimit),
	}
}

// Path returns the config file path.
func (c *Config) Path() string {
	return c.path
}

// ValidKeys returns all valid config key names.
func ValidKeys() []string {
	return []string{
		"output_format",
		"color_mode",
		"default_folder",
		"mail_limit",
	}
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("YOY_OUTPUT_FORMAT"); v != "" {
		c.OutputFormat = v
	}
	if v := os.Getenv("YOY_COLOR"); v != "" {
		c.ColorMode = v
	}
}
