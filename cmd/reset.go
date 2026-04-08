package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the current session",
	Long: `Reset the current session by clearing all context and history.

This command provides a fresh start while keeping the session ID.
It will:
- Clear all conversation history
- Remove all files from context
- Reset session memory
- Preserve the session ID and name

Use --hard to completely delete and recreate the session.`,
	Run: func(cmd *cobra.Command, args []string) {
		hard, _ := cmd.Flags().GetBool("hard")
		force, _ := cmd.Flags().GetBool("force")

		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error getting active session: %v\n", err)
			return
		}
		if session == nil {
			fmt.Println("No active session to reset.")
			return
		}

		if !force {
			fmt.Println("This will reset the current session.")
			if hard {
				fmt.Println("Mode: HARD - Session will be deleted and recreated.")
			} else {
				fmt.Println("Mode: SOFT - History and context will be cleared but session preserved.")
			}
			fmt.Println("Use --force to confirm.")
			return
		}

		if hard {
			// Hard reset: delete session and create new one with same name
			sessionName := session.Name
			sessionID := session.ID

			// Delete the session
			if err := db.DeleteSession(sessionID); err != nil {
				fmt.Printf("Error deleting session: %v\n", err)
				return
			}

			// Create new session with same name
			newSessionID, err := db.CreateSession(sessionName, "")
			if err != nil {
				fmt.Printf("Error creating new session: %v\n", err)
				return
			}

			db.SetActiveSession(newSessionID)
			fmt.Printf("✓ Hard reset complete. New session: %s (name: %s)\n", newSessionID[:8], sessionName)
		} else {
			// Soft reset: clear history and context but keep session
			messages, _ := db.GetHistory(session.ID)
			contextFiles, _ := db.GetContextFiles(session.ID)

			// Clear history
			for _, msg := range messages {
				db.DeleteMessage(msg.ID)
			}

			// Clear context
			for _, file := range contextFiles {
				db.RemoveContextFile(session.ID, file.Path)
			}

			fmt.Printf("✓ Soft reset complete. Cleared %d messages and %d context files.\n", len(messages), len(contextFiles))
			fmt.Printf("Session %s preserved.\n", session.ID[:8])
		}
	},
}

func init() {
	resetCmd.Flags().BoolP("hard", "", false, "Hard reset - delete and recreate session")
	resetCmd.Flags().BoolP("force", "f", false, "Force reset without confirmation")
}
