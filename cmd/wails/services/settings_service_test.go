package services

import (
	"testing"
	"time"

	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/engine"
)

// TestSettingsServiceRefresh — T083 (US4): SaveConfig persists in
// internal/config, emits config:changed and refreshes the running engine
// without reinitialising the interface (FR-010/FR-016, C-004).
func TestSettingsServiceRefresh(t *testing.T) {
	hub := NewHub()
	c := &capture{}
	hub.Emit = c.add
	driver := NewToolbarDriver(hub)
	e := engine.New(driver)
	e.SetStreamFunc(scriptSink(script{deltas: []string{"post-refresh"}}))
	hub.SetEngine(e)
	chat := NewChatService(hub)
	e.Start()
	chat.Start()
	t.Cleanup(func() { chat.Stop(); e.Stop() })
	svc := NewSettingsService(hub)

	before := e.SessionID()
	prev := config.AppConfig.Model
	t.Cleanup(func() { config.AppConfig.Model = prev })

	// The GUI SettingsOverlay sends PascalCase keys (bindings field names).
	cfg, err := svc.SaveConfig(map[string]any{"Model": "claude-sonnet-4-refresh", "AutoApprove": map[string]any{"bash": true}})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if cfg.Model != "claude-sonnet-4-refresh" {
		t.Fatalf("config not applied: %+v", cfg)
	}
	waitFor(t, c, func(name string) bool { return name == "config:changed" }, 3*time.Second)
	if payload, ok := c.payload("config:changed"); ok {
		pcfg, _ := payload.(map[string]any)["config"].(config.Config)
		if pcfg.Model != "claude-sonnet-4-refresh" {
			t.Fatal("config:changed must carry the refreshed config")
		}
	}

	// The same engine instance keeps streaming after the refresh — no reinit.
	chat.Send("tras guardar")
	waitFor(t, c, func(name string) bool { return name == "chat:idle" }, 10*time.Second)
	if !c.has("chat:end") {
		t.Fatal("engine must keep streaming after config save")
	}
	if e.SessionID() != before {
		t.Fatal("config refresh must not reset the session")
	}
}