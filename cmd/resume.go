package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(resumeCmd)
	resumeCmd.Flags().String("id", "", "Session ID to resume")
}

var resumeCmd = &cobra.Command{
	Use:   "resume [session-id]",
	Short: "Resume a previous session",
	Long:  `Reactivate and continue a previous conversation session.`,
	Run: func(cmd *cobra.Command, args []string) {
		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		} else {
			sessionID, _ = cmd.Flags().GetString("id")
		}

		if sessionID == "" {
			sessions, err := db.ListSessions()
			if err != nil || len(sessions) == 0 {
				fmt.Println("No sessions found. Start a new session with 'chat'.")
				os.Exit(1)
			}

			fmt.Println("Available sessions:")
			for _, s := range sessions {
				fmt.Printf("  • %s - %s\n", s.ID, s.Name)
			}
			fmt.Println("\nUse: resume <session-id>")
			return
		}

		if err := db.ResumeSession(sessionID); err != nil {
			fmt.Fprintf(os.Stderr, "Error resuming session: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Resumed session: %s\n", sessionID)
	},
}
