package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(modelsCmd)
}

var modelsCmd = &cobra.Command{
	Use:   "models [provider]",
	Short: "Discover and list available AI models dynamically",
	Long: `Query provider endpoints in real-time or from local catalogs to discover available models.
Supported providers: ollama, openrouter, groq, openai, gemini, anthropic, deepseek.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider := ""
		if len(args) > 0 {
			provider = strings.ToLower(args[0])
		}

		providers := []string{"anthropic", "openrouter", "openai", "groq", "ollama", "gemini", "deepseek"}
		if provider != "" {
			providers = []string{provider}
		}

		for _, p := range providers {
			apiKey := config.GetAPIKeyForProvider(p)
			baseURL := config.AppConfig.BaseURL
			if p == "ollama" && config.AppConfig.OllamaBaseURL != "" {
				baseURL = config.AppConfig.OllamaBaseURL
			}

			models, err := api.DiscoverModels(p, apiKey, baseURL)
			if err != nil {
				fmt.Printf("⚠️  [%s] Error: %v\n\n", strings.ToUpper(p), err)
				continue
			}

			fmt.Printf("📦 Provider: %s (%d models)\n", strings.ToUpper(p), len(models))
			fmt.Println(strings.Repeat("-", 60))
			for _, m := range models {
				desc := m.Description
				if desc != "" {
					fmt.Printf("  • %-35s %s\n", m.ID, desc)
				} else {
					fmt.Printf("  • %s\n", m.ID)
				}
			}
			fmt.Println()
		}
	},
}
