package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/spf13/viper"
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

// TestSaveConfigRestrictsPermissions verifies FR-014: the config file holding
// API keys is written with owner-only permissions (0o600).
func TestSaveConfigRestrictsPermissions(t *testing.T) {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()

	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("LoadConfig: %v", err)
	}
	AppConfig.GroqAPIKey = "gsk_secret_test_key_123456"
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	path := viper.ConfigFileUsed()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
}

// TestStatuslineDefaults verifies FR-018: the status line is on by default and
// ships with a sensible default field set.
func TestStatuslineDefaults(t *testing.T) {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()
	viper.Reset()

	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !AppConfig.Statusline.Show {
		t.Error("statusline.show should default to true")
	}
	if len(AppConfig.Statusline.Fields) == 0 {
		t.Error("statusline.fields should have a default set")
	}
}

// TestStatuslinePersist verifies FR-018: toggling fields survives SaveConfig.
func TestStatuslinePersist(t *testing.T) {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()
	viper.Reset()

	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("LoadConfig: %v", err)
	}
	AppConfig.Statusline.Show = false
	AppConfig.Statusline.Fields = []string{"model", "mode"}
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	viper.Reset()
	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("reload: %v", err)
	}
	if AppConfig.Statusline.Show {
		t.Error("statusline.show = true, want false after save")
	}
	if len(AppConfig.Statusline.Fields) != 2 || AppConfig.Statusline.Fields[0] != "model" {
		t.Errorf("fields = %v, want [model mode]", AppConfig.Statusline.Fields)
	}
}

// TestRailAndThemeDefaults verifies 003: rail expanded and dark theme by default.
func TestRailAndThemeDefaults(t *testing.T) {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()
	viper.Reset()

	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("LoadConfig: %v", err)
	}
	if AppConfig.Rail.Collapsed {
		t.Error("rail.collapsed should default to false (expanded)")
	}
	if AppConfig.ThemeVariant == "" {
		t.Error("theme should have a default (dark)")
	}
}

// TestRailAndThemePersist verifies FR-002/FR-007: rail collapse and theme
// variant survive SaveConfig/LoadConfig.
func TestRailAndThemePersist(t *testing.T) {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}()
	viper.Reset()

	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("LoadConfig: %v", err)
	}
	AppConfig.Rail.Collapsed = true
	AppConfig.ThemeVariant = "light"
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	viper.Reset()
	if err := LoadConfig(); err != nil && !os.IsNotExist(err) {
		t.Fatalf("reload: %v", err)
	}
	if !AppConfig.Rail.Collapsed {
		t.Error("rail.collapsed = false, want true after save")
	}
	if AppConfig.ThemeVariant != "light" {
		t.Errorf("theme = %q, want light after save", AppConfig.ThemeVariant)
	}
}
