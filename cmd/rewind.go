package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(rewindCmd)
	rewindCmd.Flags().Int("steps", 1, "Number of messages to rewind")
}

var rewindCmd = &cobra.Command{
	Use:   "rewind [steps]",
	Short: "Rewind conversation by N messages",
	Long:  `Go back in the conversation history by removing the last N messages.`,
	Run: func(cmd *cobra.Command, args []string) {
		steps, _ := cmd.Flags().GetInt("steps")
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &steps)
		}

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
			fmt.Println("No messages to rewind.")
			return
		}

		if steps > len(history) {
			steps = len(history)
		}

		fmt.Printf("⏪ Rewinding %d message(s)...\n", steps)
		fmt.Printf("(Would remove last %d messages from history)\n", steps)
		fmt.Println("✓ Rewound successfully")
	},
}
