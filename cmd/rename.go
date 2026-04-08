package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(renameCmd)
}

var renameCmd = &cobra.Command{
	Use:   "rename <new-name>",
	Short: "Rename the current session",
	Long:  `Give the current conversation session a new name.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: rename <new-name>")
			os.Exit(1)
		}

		newName := strings.Join(args, " ")
		session, _ := db.GetActiveSession()
		if session == nil {
			fmt.Println("No active session. Start one with 'chat' first.")
			os.Exit(1)
		}

		if err := db.RenameSession(session.ID, newName); err != nil {
			fmt.Fprintf(os.Stderr, "Error renaming session: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Session renamed to: %s\n", newName)
	},
}
