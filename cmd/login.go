package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/user/go-claude-code/internal/config"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure your API Key or local provider",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Select your provider:")
		fmt.Println("1. Anthropic (Claude) - API Key sk-ant-*")
		fmt.Println("2. OpenAI - API Key sk-*")
		fmt.Println("3. Groq - API Key gsk_*")
		fmt.Println("4. OpenRouter - API Key sk-or-*")
		fmt.Println("5. Ollama (Local) - No API key needed")
		fmt.Print("\nChoice (1-5), or paste your API key directly: ")

		var input string
		fmt.Scanln(&input)

		var baseURL, model, apiKey string

		switch input {
		case "1", "sk-ant":
			fmt.Print("Enter Anthropic API Key (sk-ant-...): ")
			byteKey, _ := term.ReadPassword(int(os.Stdin.Fd()))
			apiKey = string(byteKey)
			baseURL = "https://api.anthropic.com/v1/messages"
			model = "claude-3-5-sonnet-20240620"
			viper.Set("anthropic_api_key", apiKey)
			viper.Set("api_key", apiKey)
			viper.Set("base_url", baseURL)
			viper.Set("model", model)
			config.AppConfig.AnthropicAPIKey = apiKey
			fmt.Println("\n✔ Configured for Anthropic")

		case "2", "sk-":
			fmt.Print("Enter OpenAI API Key (sk-...): ")
			byteKey, _ := term.ReadPassword(int(os.Stdin.Fd()))
			apiKey = string(byteKey)
			baseURL = "https://api.openai.com/v1/chat/completions"
			model = "gpt-4o"
			viper.Set("openai_api_key", apiKey)
			viper.Set("api_key", apiKey)
			viper.Set("base_url", baseURL)
			viper.Set("model", model)
			config.AppConfig.OpenAIAPIKey = apiKey
			fmt.Println("\n✔ Configured for OpenAI")

		case "3", "gsk_":
			fmt.Print("Enter Groq API Key (gsk_...): ")
			byteKey, _ := term.ReadPassword(int(os.Stdin.Fd()))
			apiKey = string(byteKey)
			baseURL = "https://api.groq.com/openai/v1/chat/completions"
			model = "llama-3.3-70b-versatile"
			viper.Set("groq_api_key", apiKey)
			viper.Set("api_key", apiKey)
			viper.Set("base_url", baseURL)
			viper.Set("model", model)
			config.AppConfig.GroqAPIKey = apiKey
			fmt.Println("\n✔ Configured for Groq")

		case "4":
			fmt.Print("Enter OpenRouter API Key (sk-or-v1-...): ")
			byteKey, _ := term.ReadPassword(int(os.Stdin.Fd()))
			apiKey = string(byteKey)
			baseURL = "https://openrouter.ai/api/v1/chat/completions"
			model = "anthropic/claude-3.5-sonnet"
			viper.Set("openrouter_api_key", apiKey)
			viper.Set("api_key", apiKey)
			viper.Set("base_url", baseURL)
			viper.Set("model", model)
			config.AppConfig.OpenRouterAPIKey = apiKey
			fmt.Println("\n✔ Configured for OpenRouter")

		case "5":
			fmt.Print("Enter Ollama URL [http://localhost:11434]: ")
			var ollamaURL string
			fmt.Scanln(&ollamaURL)
			if ollamaURL == "" {
				ollamaURL = "http://localhost:11434"
			}
			apiKey = "ollama-local"
			baseURL = ollamaURL
			model = "llama3.1"
			viper.Set("api_key", apiKey)
			viper.Set("base_url", baseURL)
			viper.Set("model", model)
			fmt.Println("\n✔ Configured for Ollama (local)")

		default:
			// Try to auto-detect from key format
			if len(input) > 6 && input[:6] == "sk-ant" {
				apiKey = input
				baseURL = "https://api.anthropic.com/v1/messages"
				model = "claude-3-5-sonnet-20240620"
				fmt.Println("✔ Detected Anthropic Provider")
			} else if len(input) > 4 && input[:4] == "gsk_" {
				apiKey = input
				baseURL = "https://api.groq.com/openai/v1/chat/completions"
				model = "llama-3.3-70b-versatile"
				fmt.Println("✔ Detected Groq Provider")
			} else if len(input) > 7 && input[:7] == "sk-or-v1" {
				apiKey = input
				baseURL = "https://openrouter.ai/api/v1/chat/completions"
				model = "anthropic/claude-3.5-sonnet"
				fmt.Println("✔ Detected OpenRouter Provider")
			} else {
				// Assume OpenAI format
				apiKey = input
				baseURL = "https://api.openai.com/v1/chat/completions"
				model = "gpt-4o"
				fmt.Println("✔ Using OpenAI-compatible provider")
			}
		}

		viper.Set("api_key", apiKey)
		viper.Set("base_url", baseURL)
		viper.Set("model", model)

		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}
		fmt.Printf("\nConfig saved! Using %s with %s\n", model, baseURL)
	},
}
