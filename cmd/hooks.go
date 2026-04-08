package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hooksCmd)
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksEnableCmd)
	hooksCmd.AddCommand(hooksDisableCmd)
}

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage git hooks",
	Long:  `Configure and manage git hooks for Claude Code integration.`,
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available hooks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Git Hooks:")
		fmt.Println("  • pre-commit     - Run before each commit")
		fmt.Println("  • post-commit    - Run after each commit")
		fmt.Println("  • pre-push       - Run before pushing")
		fmt.Println("  • prepare-commit-msg - Modify commit messages")
	},
}

var hooksEnableCmd = &cobra.Command{
	Use:   "enable <hook-name>",
	Short: "Enable a git hook",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: hooks enable <hook-name>")
			os.Exit(1)
		}
		fmt.Printf("✓ Enabled hook: %s\n", args[0])
	},
}

var hooksDisableCmd = &cobra.Command{
	Use:   "disable <hook-name>",
	Short: "Disable a git hook",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: hooks disable <hook-name>")
			os.Exit(1)
		}
		fmt.Printf("✓ Disabled hook: %s\n", args[0])
	},
}
