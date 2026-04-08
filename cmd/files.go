package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(filesCmd)
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "List files in current context",
	Long:  `Display all files currently tracked in the conversation context.`,
	Run: func(cmd *cobra.Command, args []string) {
		session, _ := db.GetActiveSession()
		var sessionID string
		if session == nil {
			sessionID = "default"
		} else {
			sessionID = session.ID
		}

		files, err := db.GetContextFiles(sessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Println("No files in context. Use 'add' to add files.")
			return
		}

		fmt.Printf("Files in context (%d):\n", len(files))
		for _, f := range files {
			fmt.Printf("  • %s\n", f.Path)
		}
	},
}
