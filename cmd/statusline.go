package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statuslineCmd)
}

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Toggle status line display",
	Long:  `Show or hide the status line in the interactive interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Status line toggled")
		fmt.Println("(Status bar display would be enabled/disabled)")
	},
}
