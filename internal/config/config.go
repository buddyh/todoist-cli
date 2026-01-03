package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".todoist-cli"
	configFileName = "config.json"
)

// Config holds the CLI configuration
type Config struct {
	APIToken string `json:"api_token"`
}

// ConfigDir returns the config directory path
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configDirName)
}

// ConfigPath returns the full config file path
func ConfigPath() string {
	return filepath.Join(ConfigDir(), configFileName)
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	// First check environment variable
	if token := os.Getenv("TODOIST_API_TOKEN"); token != "" {
		return &Config{APIToken: token}, nil
	}

	// Then try config file
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not configured. Run 'todoist auth' or set TODOIST_API_TOKEN")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.APIToken == "" {
		return nil, fmt.Errorf("no API token configured. Run 'todoist auth'")
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := ConfigPath()
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetToken returns the API token from environment or config
func GetToken() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.APIToken, nil
}
