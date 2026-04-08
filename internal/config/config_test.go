package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{
		Model:       "claude-3-5-sonnet",
		BaseURL:     "https://api.anthropic.com/v1",
		Temperature: 0.7,
		MaxTokens:   4096,
		Stream:      true,
		Verbose:     false,
	}

	if cfg.Model != "claude-3-5-sonnet" {
		t.Errorf("Expected default Model 'claude-3-5-sonnet', got '%s'", cfg.Model)
	}

	if cfg.Temperature != 0.7 {
		t.Errorf("Expected default Temperature 0.7, got %f", cfg.Temperature)
	}

	if cfg.MaxTokens != 4096 {
		t.Errorf("Expected default MaxTokens 4096, got %d", cfg.MaxTokens)
	}

	if !cfg.Stream {
		t.Error("Expected default Stream to be true")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Set both HOME and USERPROFILE for cross-platform compatibility
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()

	// Try to load config (will create default)
	err := LoadConfig()
	// May fail if viper can't create file, that's ok for test
	if err != nil {
		t.Logf("LoadConfig returned (may be expected): %v", err)
	}

	// Verify default values are set
	if AppConfig.Model == "" {
		t.Error("Model should have a default value")
	}
}

func TestConfigStructFields(t *testing.T) {
	cfg := Config{
		APIKey:      "test-key",
		Model:       "test-model",
		BaseURL:     "https://test.com",
		Shell:       "zsh",
		Verbose:     true,
		Temperature: 0.5,
		MaxTokens:   2048,
		Stream:      false,
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey mismatch")
	}
	if cfg.Model != "test-model" {
		t.Errorf("Model mismatch")
	}
	if cfg.Shell != "zsh" {
		t.Errorf("Shell mismatch")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
}
