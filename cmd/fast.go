package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fastCmd)
	fastCmd.Flags().Bool("on", false, "Enable fast mode")
	fastCmd.Flags().Bool("off", false, "Disable fast mode")
}

var fastCmd = &cobra.Command{
	Use:   "fast",
	Short: "Toggle fast mode for quicker responses",
	Long:  `Enable fast mode for more concise, rapid responses from Claude.`,
	Run: func(cmd *cobra.Command, args []string) {
		on, _ := cmd.Flags().GetBool("on")
		off, _ := cmd.Flags().GetBool("off")

		if on {
			fmt.Println("⚡ Fast mode enabled")
			fmt.Println("  • Concise responses")
			fmt.Println("  • Skip explanations")
			fmt.Println("  • Focus on code/output")
		} else if off {
			fmt.Println("🐢 Fast mode disabled (detailed mode)")
		} else {
			fmt.Println("⚡ Fast mode toggled")
		}
	},
}
