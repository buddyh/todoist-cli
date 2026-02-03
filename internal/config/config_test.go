package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()

	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	// Should end with .todoist-cli
	if filepath.Base(dir) != ".todoist-cli" {
		t.Errorf("ConfigDir() = %q, want to end with .todoist-cli", dir)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()

	if path == "" {
		t.Error("ConfigPath() returned empty string")
	}

	// Should end with config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath() = %q, want to end with config.json", path)
	}
}

func TestLoadFromEnvVar(t *testing.T) {
	// Save original value and restore after test
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)

	testToken := "test-token-from-env"
	os.Setenv("TODOIST_API_TOKEN", testToken)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.APIToken != testToken {
		t.Errorf("Load() returned token %q, want %q", cfg.APIToken, testToken)
	}
}

func TestLoadNoConfig(t *testing.T) {
	// Save original env var and clear it
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)
	os.Unsetenv("TODOIST_API_TOKEN")

	// Use a temp dir that doesn't have a config file
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error when no config exists")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Clear env var to ensure we're testing file-based config
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)
	os.Unsetenv("TODOIST_API_TOKEN")

	// Use a temp dir
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	testToken := "test-token-12345"
	cfg := &Config{APIToken: testToken}

	// Save
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file exists with correct permissions
	configPath := ConfigPath()
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Check file permissions (0600)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Config file has permissions %o, want 0600", perm)
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.APIToken != testToken {
		t.Errorf("Loaded token %q, want %q", loaded.APIToken, testToken)
	}
}

func TestLoadEmptyToken(t *testing.T) {
	// Clear env var
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)
	os.Unsetenv("TODOIST_API_TOKEN")

	// Use a temp dir
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create config with empty token
	configDir := filepath.Join(tempDir, ".todoist-cli")
	os.MkdirAll(configDir, 0700)
	configFile := filepath.Join(configDir, "config.json")
	os.WriteFile(configFile, []byte(`{"api_token": ""}`), 0600)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for empty token")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Clear env var
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)
	os.Unsetenv("TODOIST_API_TOKEN")

	// Use a temp dir
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create config with invalid JSON
	configDir := filepath.Join(tempDir, ".todoist-cli")
	os.MkdirAll(configDir, 0700)
	configFile := filepath.Join(configDir, "config.json")
	os.WriteFile(configFile, []byte(`{invalid json`), 0600)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid JSON")
	}
}

func TestGetToken(t *testing.T) {
	// Save original env var
	original := os.Getenv("TODOIST_API_TOKEN")
	defer os.Setenv("TODOIST_API_TOKEN", original)

	testToken := "test-token-get"
	os.Setenv("TODOIST_API_TOKEN", testToken)

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() returned error: %v", err)
	}

	if token != testToken {
		t.Errorf("GetToken() = %q, want %q", token, testToken)
	}
}

func TestConfigDirPermissions(t *testing.T) {
	// Use a temp dir
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	cfg := &Config{APIToken: "test"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Check directory permissions (0700)
	configDir := ConfigDir()
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config dir not created: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("Config dir has permissions %o, want 0700", perm)
	}
}
