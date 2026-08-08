package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/gui"
)

func init() {
	rootCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Start the desktop GUI",
	Long:  "Start the Fyne desktop GUI for LetsGO Code",
	Run: func(cmd *cobra.Command, args []string) {
		if err := gui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting GUI: %v\n", err)
			os.Exit(1)
		}
	},
}
