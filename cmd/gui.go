package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	wailsgui "github.com/user/go-claude-code/cmd/wails/gui"
)

// Wails GUI assets are embedded here (relative to cmd/): `letsgo gui` runs the
// exact same GUI that `wails3 build` produces, inside the same executable that
// serves the TUI — Fyne is gone (005 Phase 3, FR-018, gui-contract §7).
//
//go:embed all:wails/frontend/dist
var wailsAssets embed.FS

func init() {
	rootCmd.AddCommand(guiCmd)
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Start the desktop GUI",
	Long:  "Start the Wails desktop GUI for LetsGO Code",
	Run: func(cmd *cobra.Command, args []string) {
		if err := wailsgui.Run(wailsAssets); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting GUI: %v\n", err)
			os.Exit(1)
		}
	},
}