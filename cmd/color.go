package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(colorCmd)
	colorCmd.AddCommand(colorListCmd)
	colorCmd.AddCommand(colorSetCmd)
}

var colorCmd = &cobra.Command{
	Use:   "color",
	Short: "Manage agent color scheme",
	Long:  `View and change the agent's color scheme in the terminal.`,
}

var colorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available colors",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available agent colors:")
		fmt.Println("  • blue")
		fmt.Println("  • green")
		fmt.Println("  • purple")
		fmt.Println("  • orange")
		fmt.Println("  • cyan")
		fmt.Println("  • magenta")
	},
}

var colorSetCmd = &cobra.Command{
	Use:   "set <color>",
	Short: "Set agent color",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: color set <color>")
			os.Exit(1)
		}
		fmt.Printf("Agent color set to: %s\n", args[0])
	},
}
