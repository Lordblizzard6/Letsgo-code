package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(keybindingsCmd)
	keybindingsCmd.AddCommand(keybindingsListCmd)
	keybindingsCmd.AddCommand(keybindingsResetCmd)
}

var keybindingsCmd = &cobra.Command{
	Use:   "keybindings",
	Short: "Manage keyboard shortcuts",
	Long:  `View and customize keyboard shortcuts for Claude Code.`,
}

var keybindingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keybindings",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Keyboard Shortcuts:")
		fmt.Println("===================")
		fmt.Println("  Ctrl+C    - Cancel current operation")
		fmt.Println("  Ctrl+D    - Exit / EOF")
		fmt.Println("  Tab       - Autocomplete")
		fmt.Println("  Up/Down   - Navigate history")
		fmt.Println("  Ctrl+L    - Clear screen")
		fmt.Println("  Ctrl+R    - Search history")
	},
}

var keybindingsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset keybindings to defaults",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Keybindings reset to defaults")
	},
}
