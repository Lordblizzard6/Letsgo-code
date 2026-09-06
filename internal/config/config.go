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
	APIKey string `mapstructure:"api_key" json:"api_key"`

	// Provider-specific API keys
	AnthropicAPIKey  string `mapstructure:"anthropic_api_key" json:"anthropic_api_key"`
	OpenAIAPIKey     string `mapstructure:"openai_api_key" json:"openai_api_key"`
	GroqAPIKey       string `mapstructure:"groq_api_key" json:"groq_api_key"`
	OpenRouterAPIKey string `mapstructure:"openrouter_api_key" json:"openrouter_api_key"`
	GeminiAPIKey     string `mapstructure:"gemini_api_key" json:"gemini_api_key"`
	DeepSeekAPIKey   string `mapstructure:"deepseek_api_key" json:"deepseek_api_key"`
	OllamaBaseURL    string `mapstructure:"ollama_base_url" json:"ollama_base_url"`

	Model       string  `mapstructure:"model" json:"model"`
	PlanModel   string  `mapstructure:"plan_model" json:"plan_model"`
	BaseURL     string  `mapstructure:"base_url" json:"base_url"`
	Shell       string  `mapstructure:"shell" json:"shell"`
	Verbose     bool    `mapstructure:"verbose" json:"verbose"`
	Temperature float64 `mapstructure:"temperature" json:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens" json:"max_tokens"`
	Stream      bool    `mapstructure:"stream" json:"stream"`

	// AutoApprove persists per-category tool auto-approval (bash, file-edit, web).
	AutoApprove map[string]bool `mapstructure:"auto_approve" json:"auto_approve"`

	// Statusline persists the configurable status bar fields (FR-018).
	Statusline StatuslineConfig `mapstructure:"statusline" json:"statusline"`

	// Rail persists the shell navigation bar appearance (003).
	Rail RailConfig `mapstructure:"rail" json:"rail"`

	// ThemeVariant is the app theme: "dark" (default) or "light" (FR-007).
	ThemeVariant string `mapstructure:"theme" json:"theme"`
}

// RailConfig controls the vertical navigation rail of the shell (FR-002).
type RailConfig struct {
	Collapsed bool `mapstructure:"collapsed"`
}

// StatuslineConfig controls the configurable status line (FR-018).
type StatuslineConfig struct {
	Show   bool     `mapstructure:"show"`
	Fields []string `mapstructure:"fields"`
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
	viper.SetDefault("statusline.show", true)
	viper.SetDefault("statusline.fields", []string{"model", "mode", "context", "rate", "version"})
	viper.SetDefault("rail.collapsed", false)
	viper.SetDefault("theme", "letsgo")

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

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return err
	}

	// Transparently decrypt stored keys into memory
	AppConfig.APIKey, _ = DecryptSecret(AppConfig.APIKey)
	AppConfig.AnthropicAPIKey, _ = DecryptSecret(AppConfig.AnthropicAPIKey)
	AppConfig.OpenAIAPIKey, _ = DecryptSecret(AppConfig.OpenAIAPIKey)
	AppConfig.GroqAPIKey, _ = DecryptSecret(AppConfig.GroqAPIKey)
	AppConfig.OpenRouterAPIKey, _ = DecryptSecret(AppConfig.OpenRouterAPIKey)
	AppConfig.GeminiAPIKey, _ = DecryptSecret(AppConfig.GeminiAPIKey)
	AppConfig.DeepSeekAPIKey, _ = DecryptSecret(AppConfig.DeepSeekAPIKey)

	return nil
}

func SaveConfig() error {
	// Transparently encrypt secrets before writing to disk
	encAPIKey, _ := EncryptSecret(AppConfig.APIKey)
	encAnthropic, _ := EncryptSecret(AppConfig.AnthropicAPIKey)
	encOpenAI, _ := EncryptSecret(AppConfig.OpenAIAPIKey)
	encGroq, _ := EncryptSecret(AppConfig.GroqAPIKey)
	encOpenRouter, _ := EncryptSecret(AppConfig.OpenRouterAPIKey)
	encGemini, _ := EncryptSecret(AppConfig.GeminiAPIKey)
	encDeepSeek, _ := EncryptSecret(AppConfig.DeepSeekAPIKey)

	// Write all config values including multi-key support and encryption
	viper.Set("api_key", encAPIKey)
	viper.Set("anthropic_api_key", encAnthropic)
	viper.Set("openai_api_key", encOpenAI)
	viper.Set("groq_api_key", encGroq)
	viper.Set("openrouter_api_key", encOpenRouter)
	viper.Set("gemini_api_key", encGemini)
	viper.Set("deepseek_api_key", encDeepSeek)
	viper.Set("ollama_base_url", AppConfig.OllamaBaseURL)
	viper.Set("model", AppConfig.Model)
	viper.Set("plan_model", AppConfig.PlanModel)
	viper.Set("base_url", AppConfig.BaseURL)
	viper.Set("shell", AppConfig.Shell)
	viper.Set("verbose", AppConfig.Verbose)
	viper.Set("temperature", AppConfig.Temperature)
	viper.Set("max_tokens", AppConfig.MaxTokens)
	viper.Set("stream", AppConfig.Stream)
	viper.Set("auto_approve", AppConfig.AutoApprove)
	viper.Set("statusline.show", AppConfig.Statusline.Show)
	viper.Set("statusline.fields", AppConfig.Statusline.Fields)
	viper.Set("rail.collapsed", AppConfig.Rail.Collapsed)
	viper.Set("theme", AppConfig.ThemeVariant)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	// Restrict the config file to the owner only; it contains API keys.
	if err := restrictConfigPermissions(viper.ConfigFileUsed()); err != nil {
		return fmt.Errorf("restrict config permissions: %w", err)
	}
	return nil
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
	case "gemini", "google":
		if AppConfig.GeminiAPIKey != "" {
			return AppConfig.GeminiAPIKey
		}
	case "deepseek":
		if AppConfig.DeepSeekAPIKey != "" {
			return AppConfig.DeepSeekAPIKey
		}
	case "ollama":
		return "ollama"
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
