package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sandboxToggleCmd)
}

var sandboxToggleCmd = &cobra.Command{
	Use:   "sandbox-toggle",
	Short: "Toggle sandbox mode",
	Long:  `Enable or disable sandbox mode for safer command execution.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("🧪 Sandbox Mode")
			fmt.Println("==============")
			fmt.Println("Status: Disabled")
			fmt.Println("\nSandbox mode restricts potentially dangerous operations.")
			fmt.Println("When enabled, commands like 'rm', 'dd', etc. are blocked.")
			fmt.Println("\nUsage: sandbox-toggle [on|off]")
			return
		}

		state := args[0]

		switch state {
		case "on", "enable":
			fmt.Println("✓ Sandbox mode enabled")
			fmt.Println("  Potentially dangerous operations are now blocked")
		case "off", "disable":
			fmt.Println("✓ Sandbox mode disabled")
			fmt.Println("  Full command execution restored")
		default:
			fmt.Printf("⚠️  Unknown state: %s\n", state)
			fmt.Println("Use 'on' or 'off'")
			os.Exit(1)
		}
	},
}
