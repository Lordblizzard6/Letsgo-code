package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Claude Code CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Claude Code CLI (Go) v0.1.0")
	},
}
