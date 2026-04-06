package llm

import (
	"fmt"
	"os"
	"strings"
)

type ProviderConfig struct {
	Name        string  `yaml:"name"`
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url,omitempty"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
	Temperature float32 `yaml:"temperature,omitempty"`
}

func NewProvider(cfg ProviderConfig) (Provider, error) {
	// Priorizar API Key de env si no está en config
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv(strings.ToUpper(cfg.Name) + "_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key no encontrada para proveedor %s", cfg.Name)
	}

	model := cfg.Model
	if model == "" {
		model = getDefaultModel(cfg.Name)
	}

	switch strings.ToLower(cfg.Name) {
	case "openai":
		return NewOpenAI(apiKey, model, cfg.BaseURL)
	case "anthropic":
		return NewAnthropic(apiKey, model)
	case "gemini":
		return NewGemini(apiKey, model)
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", cfg.Name)
	}
}

func getDefaultModel(provider string) string {
	switch provider {
	case "openai":
		return "gpt-4o"
	case "anthropic":
		return "claude-sonnet-4-20250514"
	case "gemini":
		return "gemini-2.5-pro"
	default:
		return "gpt-4o"
	}
}
