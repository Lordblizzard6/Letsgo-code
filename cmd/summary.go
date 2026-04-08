package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().BoolP("save", "s", false, "Save summary to file")
	summaryCmd.Flags().StringP("output", "o", "", "Output file path")
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarize the current conversation",
	Long:  `Generate a summary of the current conversation session.`,
	Run: func(cmd *cobra.Command, args []string) {
		save, _ := cmd.Flags().GetBool("save")
		output, _ := cmd.Flags().GetString("output")

		session, _ := db.GetActiveSession()
		var sessionID string
		if session == nil {
			sessionID = "default"
		} else {
			sessionID = session.ID
		}

		history, err := db.GetHistory(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(history) == 0 {
			fmt.Println("No conversation to summarize.")
			return
		}

		fmt.Println("📋 Conversation Summary")
		fmt.Println("=======================")
		fmt.Printf("\nTotal messages: %d\n", len(history))
		fmt.Printf("Session: %s\n", sessionID)

		// Simple extraction of topics (would be AI-powered in full version)
		topics := extractTopics(history)
		if len(topics) > 0 {
			fmt.Println("\nTopics discussed:")
			for _, t := range topics {
				fmt.Printf("  • %s\n", t)
			}
		}

		summary := generateSimpleSummary(history)
		fmt.Println("\nSummary:")
		fmt.Println(summary)

		if save || output != "" {
			outputFile := output
			if outputFile == "" {
				outputFile = "conversation_summary.md"
			}
			content := fmt.Sprintf("# Conversation Summary\n\nSession: %s\nMessages: %d\n\n## Summary\n\n%s\n",
				sessionID, len(history), summary)
			if err := os.WriteFile(outputFile, []byte(content), 0644); err == nil {
				fmt.Printf("\n✓ Saved to: %s\n", outputFile)
			}
		}
	},
}

func extractTopics(history []db.Message) []string {
	// Simple keyword extraction (placeholder for AI analysis)
	var topics []string
	for _, msg := range history {
		content := fmt.Sprintf("%v", msg.Content)
		lower := strings.ToLower(content)
		if strings.Contains(lower, "error") || strings.Contains(lower, "bug") {
			topics = append(topics, "Debugging/Error handling")
		}
		if strings.Contains(lower, "function") || strings.Contains(lower, "class") {
			topics = append(topics, "Code structure")
		}
		if strings.Contains(lower, "test") {
			topics = append(topics, "Testing")
		}
		if strings.Contains(lower, "git") || strings.Contains(lower, "commit") {
			topics = append(topics, "Version control")
		}
	}
	return uniqueStrings(topics)
}

func uniqueStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func generateSimpleSummary(history []db.Message) string {
	if len(history) == 0 {
		return "No conversation history."
	}

	// Simple summary logic (placeholder for AI summary)
	var parts []string
	parts = append(parts, fmt.Sprintf("This conversation contains %d messages", len(history)))

	userMsgs := 0
	assistantMsgs := 0
	for _, msg := range history {
		if msg.Role == "user" {
			userMsgs++
		} else if msg.Role == "assistant" {
			assistantMsgs++
		}
	}

	parts = append(parts, fmt.Sprintf("(%d from user, %d from assistant)", userMsgs, assistantMsgs))
	parts = append(parts, "\nIn a full implementation, this would use AI to generate a detailed summary.")

	return strings.Join(parts, " ")
}
