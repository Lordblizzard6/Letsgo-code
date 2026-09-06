package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Use:   "gui [path]",
	Short: "Start the desktop GUI",
	Long:  "Start the Wails desktop GUI for LetsGO Code. If [path] is provided, opens the GUI with that directory as the active project. If omitted, starts without an active project.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initialPath := ""
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			target := strings.TrimSpace(args[0])
			abs, err := filepath.Abs(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving path '%s': %v\n", target, err)
				os.Exit(1)
			}
			info, err := os.Stat(abs)
			if err != nil || !info.IsDir() {
				fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid directory\n", abs)
				os.Exit(1)
			}
			initialPath = abs
		}

		if err := wailsgui.RunWithProject(wailsAssets, initialPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting GUI: %v\n", err)
			os.Exit(1)
		}
	},
}