package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.Flags().Bool("on", false, "Enable plan mode")
	planCmd.Flags().Bool("off", false, "Disable plan mode")
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Toggle plan mode",
	Long:  `Enable or disable plan mode for structured task breakdown.`,
	Run: func(cmd *cobra.Command, args []string) {
		on, _ := cmd.Flags().GetBool("on")
		off, _ := cmd.Flags().GetBool("off")

		if on {
			fmt.Println("📋 Plan mode enabled")
			fmt.Println("\nI'll help you break down tasks into steps:")
			fmt.Println("  1. Define the goal")
			fmt.Println("  2. Break into subtasks")
			fmt.Println("  3. Execute step by step")
			fmt.Println("  4. Review and adjust")
		} else if off {
			fmt.Println("📋 Plan mode disabled")
		} else {
			fmt.Println("📋 Plan mode toggled")
		}
	},
}
