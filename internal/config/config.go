package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Provider ProviderConfig `mapstructure:"provider"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Tools    ToolsConfig    `mapstructure:"tools"`
}

type ProviderConfig struct {
	Name        string  `mapstructure:"name"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	BaseURL     string  `mapstructure:"base_url"`
	Temperature float32 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

type AgentConfig struct {
	MaxIterations int    `mapstructure:"max_iterations"`
	SystemPrompt  string `mapstructure:"system_prompt"`
}

type ToolsConfig struct {
	RequireConfirmation bool     `mapstructure:"require_confirmation"`
	AllowedCommands     []string `mapstructure:"allowed_commands"`
	WorkingDirectory    string   `mapstructure:"working_directory"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Buscar en directorio actual y home
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.mycli")

	// Defaults
	viper.SetDefault("provider.name", "anthropic")
	viper.SetDefault("provider.model", "claude-sonnet-4-20250514")
	viper.SetDefault("agent.max_iterations", 10)
	viper.SetDefault("tools.require_confirmation", true)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// API Key de env tiene prioridad
	if cfg.Provider.APIKey == "" {
		cfg.Provider.APIKey = os.Getenv("API_KEY")
	}

	// Working directory
	if cfg.Tools.WorkingDirectory == "" {
		wd, _ := os.Getwd()
		cfg.Tools.WorkingDirectory = wd
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath := filepath.Join(os.Getenv("HOME"), ".mycli", "config.yaml")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	viper.Set("provider", cfg.Provider)
	viper.Set("agent", cfg.Agent)
	viper.Set("tools", cfg.Tools)

	return viper.WriteConfigAs(configPath)
}

func GetSystemPrompt() string {
	return `Eres un asistente de programación experto. Puedes:
- Leer y escribir archivos
- Ejecutar comandos de shell
- Usar git

Siempre:
1. Analiza el código existente antes de modificar
2. Explica qué vas a hacer antes de hacerlo
3. Usa herramientas en lugar de pedir que el usuario ejecute comandos
4. Verifica que los cambios funcionen

Sé conciso pero completo. Usa markdown para código.`
}
