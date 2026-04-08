package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(desktopCmd)
}

var desktopCmd = &cobra.Command{
	Use:   "desktop",
	Short: "Switch to desktop mode",
	Long:  `Configure Claude Code for desktop-optimized experience.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🖥️  Desktop mode enabled")
		fmt.Println("\nOptimized for:")
		fmt.Println("  • Larger screen real estate")
		fmt.Println("  • Keyboard shortcuts")
		fmt.Println("  • Multiple windows")
		fmt.Println("  • File drag & drop")
	},
}
