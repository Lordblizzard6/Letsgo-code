package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export [filename]",
	Short: "Export conversation to file",
	Long:  `Export the current conversation session to a JSON or markdown file.`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := "conversation.json"
		if len(args) > 0 {
			filename = args[0]
		}

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session to export.")
			return
		}

		messages, err := db.GetHistory(session.ID)
		if err != nil {
			fmt.Printf("Error getting history: %v\n", err)
			return
		}

		if len(messages) == 0 {
			fmt.Println("No messages to export.")
			return
		}

		export := map[string]interface{}{
			"session_id":  session.ID,
			"name":        session.Name,
			"created_at":  session.CreatedAt,
			"messages":    messages,
			"exported_at": time.Now(),
		}

		data, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling: %v\n", err)
			return
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			fmt.Printf("Error saving file: %v\n", err)
			return
		}

		fmt.Printf("✓ Exported %d messages to %s\n", len(messages), filename)
	},
}

var importCmd = &cobra.Command{
	Use:   "import [filename]",
	Short: "Import conversation from file",
	Long:  `Import a conversation from a JSON file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		var export map[string]interface{}
		if err := json.Unmarshal(data, &export); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return
		}

		// Create new session
		sessionName := "Imported"
		if name, ok := export["name"].(string); ok && name != "" {
			sessionName = name
		}

		sessionID, err := db.CreateSession(sessionName, "")
		if err != nil {
			fmt.Printf("Error creating session: %v\n", err)
			return
		}

		// Import messages
		if messages, ok := export["messages"].([]interface{}); ok {
			for _, msg := range messages {
				if m, ok := msg.(map[string]interface{}); ok {
					role, _ := m["role"].(string)
					content := m["content"]
					db.SaveMessage(sessionID, role, content)
				}
			}
			fmt.Printf("✓ Imported %d messages to session %s\n", len(messages), sessionID[:8])
		}
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "View or change the current model",
	Long:  `Display current model or switch to a different AI model.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("=== Current Model ===")
			fmt.Printf("Model: %s\n", config.AppConfig.Model)
			fmt.Printf("Base URL: %s\n", config.AppConfig.BaseURL)
			fmt.Println("\nAvailable models:")
			fmt.Println("  anthropic: claude-3-5-sonnet-20240620, claude-3-opus-20240229")
			fmt.Println("  openai: gpt-4o, gpt-4-turbo")
			fmt.Println("  groq: llama-3.3-70b-versatile, llama-3.1-8b-instant")
			fmt.Println("  ollama: local models")
			fmt.Println("  openrouter: various models")
			fmt.Println("\nUsage: claudego model [provider/model]")
			return
		}

		model := args[0]
		config.AppConfig.Model = model

		// Update base URL based on model
		switch model {
		case ".letsGo":
			config.AppConfig.BaseURL = "https://api.anthropic.com/v1"
		case "claude-3-5-sonnet-20240620", "claude-3-opus-20240229":
			config.AppConfig.BaseURL = "https://api.anthropic.com/v1"
		case "gpt-4o", "gpt-4-turbo":
			config.AppConfig.BaseURL = "https://api.openai.com/v1"
		case "llama-3.3-70b-versatile", "llama-3.1-8b-instant":
			config.AppConfig.BaseURL = "https://api.groq.com/openai/v1"
		}

		// Save config
		config.SaveConfig()

		fmt.Printf("✓ Model switched to: %s\n", model)
	},
}
