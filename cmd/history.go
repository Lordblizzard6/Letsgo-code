package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historySearchCmd)
	historyCmd.AddCommand(historyClearCmd)
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View and manage conversation history",
	Long:  `Display the conversation history for the current session, or search through past messages.`,
}

var historyShowCmd = &cobra.Command{
	Use:   "show [session-id]",
	Short: "Show conversation history",
	Long:  `Display the conversation history. If no session ID is provided, shows the active session's history.`,
	Run: func(cmd *cobra.Command, args []string) {
		var sessionID string
		var err error

		if len(args) > 0 {
			sessionID = args[0]
		} else {
			session, err := db.GetActiveSession()
			if err != nil {
				fmt.Printf("Error getting active session: %v\n", err)
				return
			}
			if session == nil {
				fmt.Println("No active session. Use 'letsGo history show [session-id]' to view a specific session.")
				return
			}
			sessionID = session.ID
		}

		messages, err := db.GetHistory(sessionID)
		if err != nil {
			fmt.Printf("Error getting history: %v\n", err)
			return
		}

		if len(messages) == 0 {
			fmt.Println("No messages in this session.")
			return
		}

		limit, _ := cmd.Flags().GetInt("limit")
		if limit <= 0 || limit > len(messages) {
			limit = len(messages)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "#\tROLE\tCONTENT")
		fmt.Fprintln(w, "-\t----\t-------")

		// Show last 'limit' messages
		start := len(messages) - limit
		if start < 0 {
			start = 0
		}

		for i, msg := range messages[start:] {
			content := fmt.Sprintf("%v", msg.Content)
			if len(content) > 80 {
				content = content[:77] + "..."
			}
			// Replace newlines with spaces for display
			content = strings.ReplaceAll(content, "\n", " ")
			fmt.Fprintf(w, "%d\t%s\t%s\n", start+i+1, msg.Role, content)
		}

		w.Flush()
		fmt.Printf("\nShowing %d of %d messages\n", limit, len(messages))
	},
}

var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search through conversation history",
	Long:  `Search for specific text in the conversation history.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.ToLower(args[0])

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error getting active session: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session.")
			return
		}

		messages, err := db.GetHistory(session.ID)
		if err != nil {
			fmt.Printf("Error getting history: %v\n", err)
			return
		}

		var matches []db.Message
		for _, msg := range messages {
			content := fmt.Sprintf("%v", msg.Content)
			if strings.Contains(strings.ToLower(content), query) {
				matches = append(matches, msg)
			}
		}

		if len(matches) == 0 {
			fmt.Printf("No messages found containing '%s'\n", args[0])
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ROLE\tCONTENT")
		fmt.Fprintln(w, "----\t-------")

		for _, msg := range matches {
			content := fmt.Sprintf("%v", msg.Content)
			// Show context around match
			idx := strings.Index(strings.ToLower(content), query)
			start := idx - 30
			if start < 0 {
				start = 0
			}
			end := idx + len(query) + 30
			if end > len(content) {
				end = len(content)
			}

			snippet := content[start:end]
			if start > 0 {
				snippet = "..." + snippet
			}
			if end < len(content) {
				snippet = snippet + "..."
			}

			snippet = strings.ReplaceAll(snippet, "\n", " ")
			fmt.Fprintf(w, "%s\t%s\n", msg.Role, snippet)
		}

		w.Flush()
		fmt.Printf("\nFound %d matches for '%s'\n", len(matches), args[0])
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear conversation history for current session",
	Long:  `Remove all messages from the current session's history. This cannot be undone.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error getting active session: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session.")
			return
		}

		if !force {
			fmt.Println("This will clear all messages from the current session.")
			fmt.Println("Use --force to confirm.")
			return
		}

		// Clear history by archiving and creating new session
		messages, err := db.GetHistory(session.ID)
		if err != nil {
			fmt.Printf("Error getting history: %v\n", err)
			return
		}

		// Archive old session
		newName := session.Name + " (archived " + session.CreatedAt.Format("2006-01-02") + ")"
		db.RenameSession(session.ID, newName)

		// Create new session
		newSessionID, err := db.CreateSession(session.Name, "")
		if err != nil {
			fmt.Printf("Error creating new session: %v\n", err)
			return
		}

		db.SetActiveSession(newSessionID)

		fmt.Printf("✓ Cleared %d messages. Created new session: %s\n", len(messages), newSessionID[:8])
		fmt.Printf("Old session archived as: %s\n", newName)
	},
}

func init() {
	historyShowCmd.Flags().IntP("limit", "n", 20, "Number of messages to show")
	historyClearCmd.Flags().BoolP("force", "f", false, "Force clear without confirmation")
}
