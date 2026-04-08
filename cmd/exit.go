package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exitCmd)
}

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Exit the interactive session",
	Long:  `Exit the current Claude Code interactive session.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Goodbye!")
	},
}
