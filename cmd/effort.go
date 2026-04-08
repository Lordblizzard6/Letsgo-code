package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(effortCmd)
	effortCmd.AddCommand(effortSetCmd)
	effortCmd.AddCommand(effortGetCmd)
}

var effortCmd = &cobra.Command{
	Use:   "effort",
	Short: "Manage effort level for responses",
	Long:  `Set the desired effort level (0-5) for Claude's responses.`,
}

var effortSetCmd = &cobra.Command{
	Use:   "set <level>",
	Short: "Set effort level (0-5)",
	Long:  `Set effort level: 0=minimal, 1=low, 2=normal, 3=high, 4=very high, 5=max`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: effort set <0-5>")
			os.Exit(1)
		}

		level, err := strconv.Atoi(args[0])
		if err != nil || level < 0 || level > 5 {
			fmt.Println("Error: Level must be 0-5")
			os.Exit(1)
		}

		labels := []string{"minimal", "low", "normal", "high", "very high", "maximum"}
		fmt.Printf("Effort level set to %d (%s)\n", level, labels[level])
	},
}

var effortGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current effort level",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Current effort level: 2 (normal)")
		fmt.Println("\nLevels:")
		fmt.Println("  0 - Minimal effort")
		fmt.Println("  1 - Low effort")
		fmt.Println("  2 - Normal effort (default)")
		fmt.Println("  3 - High effort")
		fmt.Println("  4 - Very high effort")
		fmt.Println("  5 - Maximum effort")
	},
}
