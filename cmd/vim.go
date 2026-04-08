package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(vimCmd)
}

var vimCmd = &cobra.Command{
	Use:   "vim",
	Short: "Toggle vim mode for input",
	Long:  `Toggle vim keybindings for text input in the interactive chat.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🎹 Vim mode toggled")
		fmt.Println("(Vim keybindings would be enabled/disabled for input)")
	},
}
