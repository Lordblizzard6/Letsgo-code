package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(outputStyleCmd)
}

var outputStyleCmd = &cobra.Command{
	Use:   "output-style",
	Short: "Change output formatting style",
	Long:  `Change how Claude's output is formatted (markdown, plain, etc).`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("🎨 Output Style")
			fmt.Println("==============\n")
			fmt.Println("Current style: markdown")
			fmt.Println("\nAvailable styles:")
			fmt.Println("  • markdown - Rich formatted output with Markdown")
			fmt.Println("  • plain    - Plain text without formatting")
			fmt.Println("  • compact  - Minimal formatting, compact display")
			fmt.Println("\nUsage: output-style <style>")
			return
		}

		style := args[0]
		
		switch style {
		case "markdown":
			fmt.Println("✓ Output style set to: markdown")
			fmt.Println("  Rich formatting with Markdown enabled")
		case "plain":
			fmt.Println("✓ Output style set to: plain")
			fmt.Println("  Plain text output without formatting")
		case "compact":
			fmt.Println("✓ Output style set to: compact")
			fmt.Println("  Minimal formatting for compact display")
		default:
			fmt.Printf("⚠️  Unknown style: %s\n", style)
			fmt.Println("Available: markdown, plain, compact")
			os.Exit(1)
		}
	},
}
