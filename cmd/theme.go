package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(themeCmd)
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeSetCmd)
}

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage terminal theme",
	Long:  `View and change the terminal color theme.`,
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available themes:")
		fmt.Println("  • default")
		fmt.Println("  • dark")
		fmt.Println("  • light")
		fmt.Println("  • high-contrast")
		fmt.Println("  • ocean")
		fmt.Println("  • solarized")
	},
}

var themeSetCmd = &cobra.Command{
	Use:   "set <theme-name>",
	Short: "Set active theme",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: theme set <theme-name>")
			os.Exit(1)
		}
		fmt.Printf("Theme set to: %s\n", args[0])
	},
}
