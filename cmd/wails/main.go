package main

import (
	"embed"
	"log"

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
// gui.Run through the cobra command — no separate Wails binary (plan.md
// Structure Decision, 005 Phase 3 T012/T013).
func main() {
	if err := gui.Run(assets); err != nil {
		log.Fatal(err)
	}
}