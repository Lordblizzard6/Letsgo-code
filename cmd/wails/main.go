package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/go-claude-code/cmd/wails/gui"
)

// Wails embeds the built frontend (cmd/wails/frontend/dist) into the binary.
// Build with `wails3 build -clean -platform windows/amd64 -webview2 embed`
// (research.md D4): the fixed WebView2 runtime is embedded and no system
// WebView2 dependency remains at runtime.
//
//go:embed all:frontend/dist
var assets embed.FS

// main is the thin Wails entry used by `wails3 dev`/`wails3 build`. The
// production path is the CLI `letsgo gui` (cmd/gui.go), which calls the same
// gui.RunWithProject through the cobra command — no separate Wails binary (plan.md
// Structure Decision, 005 Phase 3 T012/T013).
func main() {
	initialPath := ""
	if len(os.Args) > 1 && strings.TrimSpace(os.Args[1]) != "" {
		target := strings.TrimSpace(os.Args[1])
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

	if err := gui.RunWithProject(assets, initialPath); err != nil {
		log.Fatal(err)
	}
}