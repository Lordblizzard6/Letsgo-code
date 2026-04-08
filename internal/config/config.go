package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	// Legacy single API key (backward compatibility)
	APIKey string `mapstructure:"api_key"`

	// Provider-specific API keys
	AnthropicAPIKey  string `mapstructure:"anthropic_api_key"`
	OpenAIAPIKey     string `mapstructure:"openai_api_key"`
	GroqAPIKey       string `mapstructure:"groq_api_key"`
	OpenRouterAPIKey string `mapstructure:"openrouter_api_key"`

	Model       string  `mapstructure:"model"`
	BaseURL     string  `mapstructure:"base_url"`
	Shell       string  `mapstructure:"shell"`
	Verbose     bool    `mapstructure:"verbose"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Stream      bool    `mapstructure:"stream"`
}

var AppConfig Config

func LoadConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	configDir := filepath.Join(home, ".letsGo")
	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetDefault("model", "llama-3.3-70b-versatile")
	viper.SetDefault("base_url", "https://api.groq.com/openai/v1/chat/completions")
	viper.SetDefault("shell", "bash")
	viper.SetDefault("temperature", 0.7)
	viper.SetDefault("max_tokens", 4096)
	viper.SetDefault("stream", true)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file automatically
		fmt.Printf("Creating default config file at %s\n", configPath)
		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to create default config file: %w", err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return viper.Unmarshal(&AppConfig)
}

func SaveConfig() error {
	// Write all config values including multi-key support
	viper.Set("api_key", AppConfig.APIKey)
	viper.Set("anthropic_api_key", AppConfig.AnthropicAPIKey)
	viper.Set("openai_api_key", AppConfig.OpenAIAPIKey)
	viper.Set("groq_api_key", AppConfig.GroqAPIKey)
	viper.Set("openrouter_api_key", AppConfig.OpenRouterAPIKey)
	viper.Set("model", AppConfig.Model)
	viper.Set("base_url", AppConfig.BaseURL)
	viper.Set("shell", AppConfig.Shell)
	viper.Set("verbose", AppConfig.Verbose)
	viper.Set("temperature", AppConfig.Temperature)
	viper.Set("max_tokens", AppConfig.MaxTokens)
	viper.Set("stream", AppConfig.Stream)

	return viper.WriteConfig()
}

// GetAPIKeyForProvider returns the appropriate API key for the given provider
func GetAPIKeyForProvider(provider string) string {
	switch provider {
	case "anthropic":
		if AppConfig.AnthropicAPIKey != "" {
			return AppConfig.AnthropicAPIKey
		}
	case "openai":
		if AppConfig.OpenAIAPIKey != "" {
			return AppConfig.OpenAIAPIKey
		}
	case "groq":
		if AppConfig.GroqAPIKey != "" {
			return AppConfig.GroqAPIKey
		}
	case "openrouter":
		if AppConfig.OpenRouterAPIKey != "" {
			return AppConfig.OpenRouterAPIKey
		}
	}

	// Fallback to legacy API key
	return AppConfig.APIKey
}

// DetectProviderFromModel returns the provider based on the model name
func DetectProviderFromModel(model string) string {
	model = strings.ToLower(model)
	switch {
	// Verificar modelos Groq que tienen formato con / primero
	case strings.HasPrefix(model, "openai/gpt-oss"):
		return "groq"
	case strings.HasPrefix(model, "meta-llama/") && strings.Contains(model, "versatile"):
		return "groq"
	case strings.Contains(model, "claude"):
		return "anthropic"
	case strings.Contains(model, "gpt") && !strings.Contains(model, "/"):
		return "openai"
	case strings.Contains(model, "o1") || strings.Contains(model, "o3"):
		return "openai"
	// Verificar formato OpenRouter (proveedor/modelo) antes que llama genérico
	case strings.Contains(model, "/"):
		return "openrouter"
	case strings.Contains(model, "llama") || strings.Contains(model, "mixtral") || strings.Contains(model, "gemma"):
		return "groq"
	default:
		return "openai" // default fallback
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo")
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Model:       "llama-3.3-70b-versatile",
		BaseURL:     "https://api.groq.com/openai/v1/chat/completions",
		Shell:       "bash",
		Temperature: 0.7,
		MaxTokens:   4096,
		Stream:      true,
	}
}
