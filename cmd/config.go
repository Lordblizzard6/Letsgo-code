package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `View and modify Claude Code configuration settings.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Configuration Settings ===\n")
		fmt.Printf("API Key: %s\n", maskString(config.AppConfig.APIKey))
		fmt.Printf("Base URL: %s\n", config.AppConfig.BaseURL)
		fmt.Printf("Model: %s\n", config.AppConfig.Model)
		fmt.Printf("Temperature: %.2f\n", config.AppConfig.Temperature)
		fmt.Printf("Max Tokens: %d\n", config.AppConfig.MaxTokens)
		fmt.Printf("Stream: %v\n", config.AppConfig.Stream)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a specific configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		switch key {
		case "api_key":
			fmt.Printf("API Key: %s\n", maskString(config.AppConfig.APIKey))
		case "base_url":
			fmt.Printf("Base URL: %s\n", config.AppConfig.BaseURL)
		case "model":
			fmt.Printf("Model: %s\n", config.AppConfig.Model)
		case "temperature":
			fmt.Printf("Temperature: %.2f\n", config.AppConfig.Temperature)
		case "max_tokens":
			fmt.Printf("Max Tokens: %d\n", config.AppConfig.MaxTokens)
		case "stream":
			fmt.Printf("Stream: %v\n", config.AppConfig.Stream)
		default:
			fmt.Printf("Unknown setting: %s\n", key)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]
		
		// Load current config
		config.LoadConfig()
		
		switch key {
		case "api_key":
			config.AppConfig.APIKey = value
			fmt.Println("API Key updated.")
		case "base_url":
			config.AppConfig.BaseURL = value
			fmt.Println("Base URL updated.")
		case "model":
			config.AppConfig.Model = value
			fmt.Printf("Model set to: %s\n", value)
		case "temperature":
			// Parse float
			var temp float64
			_, err := fmt.Sscanf(value, "%f", &temp)
			if err != nil {
				fmt.Println("Invalid temperature value")
				return
			}
			config.AppConfig.Temperature = temp
			fmt.Printf("Temperature set to: %.2f\n", temp)
		case "max_tokens":
			var tokens int
			_, err := fmt.Sscanf(value, "%d", &tokens)
			if err != nil {
				fmt.Println("Invalid max_tokens value")
				return
			}
			config.AppConfig.MaxTokens = tokens
			fmt.Printf("Max Tokens set to: %d\n", tokens)
		case "stream":
			config.AppConfig.Stream = value == "true" || value == "1"
			fmt.Printf("Stream set to: %v\n", config.AppConfig.Stream)
		default:
			fmt.Printf("Unknown setting: %s\n", key)
			return
		}
		
		// Save config
		config.SaveConfig()
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Resetting configuration to defaults...")
		config.AppConfig = config.DefaultConfig()
		config.SaveConfig()
		fmt.Println("Configuration reset.")
	},
}

func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
