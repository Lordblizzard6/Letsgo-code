package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionResumeCmd)
	sessionCmd.AddCommand(sessionRenameCmd)
	sessionCmd.AddCommand(sessionInfoCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage chat sessions",
	Long:  `List, resume, and manage your chat sessions.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		sessions, err := db.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tCREATED\tPROJECT")
		fmt.Fprintln(w, "--\t----\t-------\t-------")

		for _, s := range sessions {
			name := s.Name
			if name == "" {
				name = "(unnamed)"
			}
			project := s.ProjectPath
			if len(project) > 30 {
				project = "..." + project[len(project)-27:]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				s.ID[:8], name, s.CreatedAt.Format("2006-01-02 15:04"), project)
		}
		w.Flush()
	},
}

var sessionResumeCmd = &cobra.Command{
	Use:   "resume [session-id]",
	Short: "Resume a previous session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]

		if err := db.ResumeSession(sessionID); err != nil {
			fmt.Printf("Error resuming session: %v\n", err)
			return
		}

		fmt.Printf("✓ Resumed session %s\n", sessionID)
		fmt.Println("Your conversation context has been restored.")
	},
}

var sessionRenameCmd = &cobra.Command{
	Use:   "rename [name]",
	Short: "Rename current session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session found.")
			return
		}

		_, err = db.DB.Exec("UPDATE sessions SET name = ? WHERE id = ?", name, session.ID)
		if err != nil {
			fmt.Printf("Error renaming session: %v\n", err)
			return
		}

		fmt.Printf("✓ Session renamed to '%s'\n", name)
	},
}

var sessionInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current session info",
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

		messages, err := db.GetHistory(session.ID)
		if err != nil {
			fmt.Printf("Error getting history: %v\n", err)
			return
		}

		contextFiles, _ := db.GetContextFiles(session.ID)

		fmt.Printf("Session ID: %s\n", session.ID)
		if session.Name != "" {
			fmt.Printf("Name: %s\n", session.Name)
		}
		fmt.Printf("Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Messages: %d\n", len(messages))
		fmt.Printf("Context Files: %d\n", len(contextFiles))
	},
}
