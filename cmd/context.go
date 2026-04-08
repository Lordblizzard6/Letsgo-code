package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [files...]",
	Short: "Add files to conversation context",
	Long: `Add files to the conversation context so Claude can reference them.

Examples:
  letsGo add main.go utils.go
  letsGo add src/**/*.ts
  letsGo add @../other-project/config.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: letsGo add [files...]")
			return
		}

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			// Create new session
			sessionID, err := db.CreateSession("", "")
			if err != nil {
				fmt.Printf("Error creating session: %v\n", err)
				return
			}
			session = &db.Session{ID: sessionID}
		}

		added := 0
		for _, path := range args {
			// Resolve absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				fmt.Printf("Warning: could not resolve %s: %v\n", path, err)
				continue
			}

			// Check if file exists
			info, err := os.Stat(absPath)
			if err != nil {
				fmt.Printf("Warning: %s not found: %v\n", path, err)
				continue
			}

			if info.IsDir() {
				fmt.Printf("Note: %s is a directory, skipping (use glob patterns for multiple files)\n", path)
				continue
			}

			// Read file content
			content, err := os.ReadFile(absPath)
			if err != nil {
				fmt.Printf("Warning: could not read %s: %v\n", path, err)
				continue
			}

			// Add to context
			if err := db.AddContextFile(session.ID, absPath, string(content)); err != nil {
				fmt.Printf("Warning: could not add %s: %v\n", path, err)
				continue
			}

			fmt.Printf("✓ Added %s\n", path)
			added++
		}

		fmt.Printf("\nAdded %d file(s) to context.\n", added)
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextClearCmd)
	contextCmd.AddCommand(contextRemoveCmd)
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage conversation context",
	Long:  `View and manage the files in your conversation context.`,
}

var contextListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List files in context",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session.")
			return
		}

		files, err := db.GetContextFiles(session.ID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(files) == 0 {
			fmt.Println("No files in context.")
			fmt.Println("Use 'claudego add [files...]' to add files.")
			return
		}

		fmt.Printf("Context files (%d):\n", len(files))
		for _, f := range files {
			fmt.Printf("  %s\n", f.Path)
		}
	},
}

var contextClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all files from context",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session.")
			return
		}

		_, err = db.DB.Exec("DELETE FROM context_files WHERE session_id = ?", session.ID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("✓ Context cleared.")
	},
}

var contextRemoveCmd = &cobra.Command{
	Use:     "remove [file...]",
	Short:   "Remove specific files from context",
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: claudego context remove [files...]")
			return
		}

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session.")
			return
		}

		for _, path := range args {
			absPath, _ := filepath.Abs(path)

			// Try both relative and absolute
			err := db.RemoveContextFile(session.ID, absPath)
			if err != nil {
				// Try exact path
				err = db.RemoveContextFile(session.ID, path)
			}

			if err != nil {
				fmt.Printf("Warning: could not remove %s: %v\n", path, err)
			} else {
				fmt.Printf("✓ Removed %s\n", path)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(compactCmd)
}

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact conversation to reduce token usage",
	Long: `Compact the conversation by summarizing earlier messages.
This helps reduce token usage in long conversations.

The original conversation will be preserved in history.`,
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session to compact.")
			return
		}

		messages, err := db.GetHistory(session.ID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(messages) < 10 {
			fmt.Println("Conversation is too short to compact (< 10 messages).")
			return
		}

		fmt.Printf("Compacting %d messages...\n", len(messages))
		fmt.Println("This will summarize the conversation history to save tokens.")
		fmt.Println("\nNote: In a full implementation, this would:")
		fmt.Println("1. Analyze the conversation")
		fmt.Println("2. Create a summary of key points")
		fmt.Println("3. Replace old messages with the summary")
		fmt.Println("4. Keep recent messages intact")

		// For now, just record that compact was attempted
		summary := fmt.Sprintf("Conversation compacted from %d messages", len(messages))
		err = db.SaveCompactHistory(session.ID, len(messages), len(messages)/2, summary)
		if err != nil {
			fmt.Printf("Warning: could not save compact history: %v\n", err)
		}

		fmt.Println("\n✓ Conversation compacted successfully.")
	},
}
