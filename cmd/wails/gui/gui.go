// Package gui launches the Wails desktop GUI of LetsGO (005, Phase 3 T012).
//
// The package is reusable: both the CLI entry (`cmd/gui.go`, `letsgo gui`) and
// the standalone Wails entry (`cmd/wails/main.go`, wails3 dev/build) call
// Run. The embedded frontend assets are passed in because go:embed patterns
// cannot traverse parent directories (`..` is forbidden); each caller embeds
// its own legal path (frontend/dist for main.go, wails/frontend/dist for the
// cobra command) and Wails locates index.html recursively.
package gui

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/user/go-claude-code/cmd/wails/services"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/engine"
)

// Run wires the shared core (engine + config) behind the bound services of
// US3 and blocks until the window closes. The engine is the single
// conversation owner (frontend-contract); the Wails GUI is a presentation like
// the TUI. assets must contain the built frontend (index.html at its root,
// or reachable under a subdirectory — Wails resolves it recursively).
//
// Research.md D4: production builds embed the fixed WebView2 runtime
// (`-webview2 embed`); D2: the GUI is launched by the same executable as the
// TUI (`letsgo.exe gui`, see cmd/gui.go) — no separate Wails binary.
func Run(assets fs.FS) error {
	if err := config.LoadConfig(); err != nil {
		log.Printf("config: %v (defaults in memory)", err)
	}

	hub := services.NewHub()
	driver := services.NewToolbarDriver(hub)
	engineInstance := engine.New(driver)
	hub.SetEngine(engineInstance)

	chat := services.NewChatService(hub)

	app := application.New(application.Options{
		Name:        "LetsGO",
		Description: "LetsGO — Claude Code chat client (TUI + Wails GUI)",
		Services: []application.Service{
			application.NewService(chat),
			application.NewService(services.NewSessionsService(hub)),
			application.NewService(services.NewGitService(hub)),
			application.NewService(services.NewTasksService()),
			application.NewService(services.NewMCPService()),
			application.NewService(services.NewPluginsService()),
			application.NewService(services.NewSettingsService(hub)),
			application.NewService(services.NewUsageService()),
			application.NewService(services.NewThemeService(hub)),
			application.NewService(services.NewAccountService(hub)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Route Go→frontend events through the Wails Event bus, batched ~50ms by
	// the chat pump (research.md D3, frontend-contract §2).
	hub.Emit = func(name string, data any) {
		if !app.Event.Emit(name, data) {
			log.Printf("wails emit %s cancelled", name)
		}
	}

	engineInstance.Start()
	chat.Start()

	// Clean shutdown (FR-014): cancel any in-flight stream before stopping the
	// engine loop so no goroutine outlives the window.
	app.OnShutdown(func() {
		chat.Stop()
		engineInstance.Stop()
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "LetsGO",
		Width:  1200,
		Height: 760,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(13, 17, 23),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		return fmt.Errorf("wails gui: %w", err)
	}
	return nil
}