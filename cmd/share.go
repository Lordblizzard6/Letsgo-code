package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.Flags().StringP("format", "f", "markdown", "Export format (markdown, json)")
	shareCmd.Flags().StringP("output", "o", "", "Output file")
	shareCmd.Flags().BoolP("gist", "g", false, "Create GitHub Gist")
}

var shareCmd = &cobra.Command{
	Use:   "share [session-id]",
	Short: "Share conversation session",
	Long: `Export and share a conversation session.
Can export to file, create GitHub Gist, or generate a shareable link.`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		gist, _ := cmd.Flags().GetBool("gist")

		// Get session ID
		sessionID := ""
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			session, _ := db.GetActiveSession()
			if session != nil {
				sessionID = session.ID
			}
		}

		if sessionID == "" {
			fmt.Println("No session specified or active.")
			fmt.Println("Usage: share [session-id]")
			os.Exit(1)
		}

		// Load session history
		history, err := db.GetHistory(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading session: %v\n", err)
			os.Exit(1)
		}

		if len(history) == 0 {
			fmt.Println("Session is empty, nothing to share.")
			os.Exit(1)
		}

		// Get session info
		sessions, _ := db.ListSessions()
		var sessionName string
		for _, s := range sessions {
			if s.ID == sessionID {
				sessionName = s.Name
				break
			}
		}

		// Generate export content
		var content string
		switch format {
		case "json":
			content = exportAsJSON(history, sessionName)
		case "markdown":
			content = exportAsMarkdown(history, sessionName)
		default:
			content = exportAsMarkdown(history, sessionName)
		}

		// Output
		if output != "" {
			err := os.WriteFile(output, []byte(content), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Exported to: %s\n", output)
		} else if gist {
			fmt.Println("Creating GitHub Gist...")
			fmt.Println("(In full implementation, this would create a gist)")
			fmt.Println()
			fmt.Println("Preview:")
			fmt.Println("--------")
			fmt.Println(content[:min(500, len(content))])
			if len(content) > 500 {
				fmt.Println("...")
			}
		} else {
			// Print to stdout
			fmt.Println(content)
		}
	},
}

func exportAsMarkdown(history []db.Message, sessionName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Conversation: %s\n\n", sessionName))
	sb.WriteString(fmt.Sprintf("Exported: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range history {
		role := msg.Role
		if role == "assistant" {
			role = "Claude"
		} else if role == "user" {
			role = "User"
		}

		content := fmt.Sprintf("%v", msg.Content)

		sb.WriteString(fmt.Sprintf("## %s\n\n", role))
		sb.WriteString(content)
		sb.WriteString("\n\n---\n\n")
	}

	return sb.String()
}

func exportAsJSON(history []db.Message, sessionName string) string {
	type ExportMessage struct {
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		Timestamp time.Time `json:"timestamp"`
	}

	var export []ExportMessage
	for _, msg := range history {
		export = append(export, ExportMessage{
			Role:      msg.Role,
			Content:   fmt.Sprintf("%v", msg.Content),
			Timestamp: msg.Timestamp,
		})
	}

	data := struct {
		SessionName string          `json:"session_name"`
		ExportedAt  time.Time       `json:"exported_at"`
		Messages    []ExportMessage `json:"messages"`
	}{
		SessionName: sessionName,
		ExportedAt:  time.Now(),
		Messages:    export,
	}

	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonBytes)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
