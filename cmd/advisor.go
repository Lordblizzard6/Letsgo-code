package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(advisorCmd)
}

var advisorCmd = &cobra.Command{
	Use:   "advisor",
	Short: "Get coding advice and suggestions",
	Long:  `Get personalized coding advice, best practices, and improvement suggestions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("💡 Claude Advisor")
		fmt.Println("================")
		fmt.Println("\nAsk me about:")
		fmt.Println("  • Best practices for your code")
		fmt.Println("  • Architecture suggestions")
		fmt.Println("  • Performance optimizations")
		fmt.Println("  • Refactoring recommendations")
		fmt.Println("\n(Interactive advisor mode would start here)")
	},
}
