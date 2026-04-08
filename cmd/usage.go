package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(usageCmd)
}

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show API usage information",
	Long:  `Display API usage statistics and rate limit information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📈 API Usage")
		fmt.Println("===========")
		fmt.Println()
		fmt.Println("Current Period:")
		fmt.Println("  Requests: 0 / 1000")
		fmt.Println("  Tokens: 0 / 100000")
		fmt.Println()
		fmt.Println("Rate Limits:")
		fmt.Println("  • Requests/min: 60")
		fmt.Println("  • Tokens/min: 10000")
		fmt.Println()
		fmt.Println("Reset: In 24 hours")
	},
}
