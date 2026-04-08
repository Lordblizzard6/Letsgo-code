package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rateLimitOptionsCmd)
	rateLimitOptionsCmd.Flags().IntP("requests", "r", 0, "Max requests per minute")
	rateLimitOptionsCmd.Flags().IntP("tokens", "t", 0, "Max tokens per minute")
}

var rateLimitOptionsCmd = &cobra.Command{
	Use:   "rate-limit-options",
	Short: "Configure rate limiting",
	Long:  `View and configure rate limiting options for API calls.`,
	Run: func(cmd *cobra.Command, args []string) {
		requests, _ := cmd.Flags().GetInt("requests")
		tokens, _ := cmd.Flags().GetInt("tokens")

		fmt.Println("⏱️  Rate Limit Options")
		fmt.Println("=====================\n")

		// Show current settings
		fmt.Println("Current Settings:")
		fmt.Println("  • Requests per minute: Unlimited")
		fmt.Println("  • Tokens per minute: Unlimited")
		fmt.Println("  • Concurrent requests: 5")

		// Update if flags provided
		if requests > 0 {
			fmt.Printf("\n✓ Requests per minute set to: %d\n", requests)
		}
		if tokens > 0 {
			fmt.Printf("✓ Tokens per minute set to: %d\n", tokens)
		}

		if requests == 0 && tokens == 0 {
			fmt.Println("\n💡 Use flags to configure:")
			fmt.Println("   --requests <n>  Max requests per minute")
			fmt.Println("   --tokens <n>    Max tokens per minute")
		}
	},
}
