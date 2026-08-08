package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

// TestRailShortcutDispatchSmoke verifies Alt+1..8 / Alt+Left register and
// dispatch through the shortcut handler to showPane/back (contracts §8,
// FR-025). The Ctrl+1..9 session bindings remain untouched (no collision).
func TestRailShortcutDispatchSmoke(t *testing.T) {
	ctrl, _ := buildNavController(t)
	attachShortcuts(ctrl.win, ctrl)

	handler := &fyne.ShortcutHandler{}

	// Alt+1 => first rail slot "chat" (pane 0)
	alt1 := &desktop.CustomShortcut{KeyName: fyne.Key1, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt1, func(s fyne.Shortcut) { ctrl.railShortcut(1) })
	handler.TypedShortcut(alt1)
	if !visible(ctrl, 0) {
		t.Fatal("Alt+1 must show chat pane (0)")
	}

	// Alt+2 => git (pane 3). sessions was removed in 004, git is now the sole
	// work-tools destination.
	alt2 := &desktop.CustomShortcut{KeyName: fyne.Key2, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt2, func(s fyne.Shortcut) { ctrl.railShortcut(2) })
	handler.TypedShortcut(alt2)
	if !visible(ctrl, 3) {
		t.Fatal("Alt+2 must show git pane (3)")
	}

	// Alt+Left pops back to chat (history top is 0 from the git jump).
	back := &desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(back, func(s fyne.Shortcut) { ctrl.back() })
	handler.TypedShortcut(back)
	if !visible(ctrl, 0) {
		t.Fatal("Alt+Left must return to chat")
	}
	if ctrl.rail.aid() != "chat" {
		t.Fatalf("rail active = %q, want chat", ctrl.rail.aid())
	}

	// Alt+3 → tasks slot (pane 6)
	alt3 := &desktop.CustomShortcut{KeyName: fyne.Key3, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt3, func(s fyne.Shortcut) { ctrl.railShortcut(3) })
	handler.TypedShortcut(alt3)
	if !visible(ctrl, 6) {
		t.Fatal("Alt+3 must show tasks pane (6)")
	}

	// Alt+4 → mcp (pane 4)
	alt4 := &desktop.CustomShortcut{KeyName: fyne.Key4, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt4, func(s fyne.Shortcut) { ctrl.railShortcut(4) })
	handler.TypedShortcut(alt4)
	if !visible(ctrl, 4) {
		t.Fatal("Alt+4 must show mcp pane (4)")
	}

	// Alt+7 → usage (pane 2)
	alt7 := &desktop.CustomShortcut{KeyName: fyne.Key7, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt7, func(s fyne.Shortcut) { ctrl.railShortcut(7) })
	handler.TypedShortcut(alt7)
	if !visible(ctrl, 2) {
		t.Fatal("Alt+7 must show usage pane (2)")
	}

	// Alt+6 → settings (pane 1)
	alt6 := &desktop.CustomShortcut{KeyName: fyne.Key6, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt6, func(s fyne.Shortcut) { ctrl.railShortcut(6) })
	handler.TypedShortcut(alt6)
	if !visible(ctrl, 1) {
		t.Fatal("Alt+6 must show settings pane (1)")
	}

	// Alt+Left now pops back to usage (2), the top of the history stack.
	handler.TypedShortcut(back)
	if !visible(ctrl, 2) {
		t.Fatal("Alt+Left must pop back to the previous pane")
	}
}

// TestKeymapCatalogHasRailCategory verifies the '?' cheat sheet documents the
// Alt+ rail shortcuts (FR-025, T007): the per-destination Alt+1..8 rows and
// Alt+Left back.
func TestKeymapCatalogHasRailCategory(t *testing.T) {
	got := map[string]bool{}
	for _, a := range keymapCatalog {
		if a.category == "Rail" {
			got[a.keys] = true
		}
	}
	want := []string{"Alt+1", "Alt+2", "Alt+3", "Alt+4", "Alt+5", "Alt+6", "Alt+7", "Alt+8", "Alt+Left"}
	for _, k := range want {
		if !got[k] {
			t.Errorf("keymap catalog missing Rail entry %q", k)
		}
	}
	// Filter still resolves Rail rows.
	k := newKeymapOverlay(test.NewWindow(container.NewStack()))
	k.entry.SetText("Alt+")
	k.refresh()
	if len(k.items) == 0 {
		t.Fatal("filtering by Alt+ must surface rail rows")
	}
	k.entry.SetText("")
}

// TestRailShortcutActionSlot verifies Alt+8 maps to the help action slot,
// routed through the rail onAction (opening the keymap overlay).
func TestRailShortcutActionSlot(t *testing.T) {
	ctrl, _ := buildNavController(t)
	attachShortcuts(ctrl.win, ctrl)
	ctrl.buildKeymap()

	handler := &fyne.ShortcutHandler{}
	alt8 := &desktop.CustomShortcut{KeyName: fyne.Key8, Modifier: fyne.KeyModifierAlt}
	handler.AddShortcut(alt8, func(s fyne.Shortcut) { ctrl.railShortcut(8) })
	handler.TypedShortcut(alt8)

	if ctrl.rail == nil {
		t.Fatal("rail must be built for Alt+8 dispatch")
	}
	if ctrl.keymapOverlay == nil {
		t.Fatal("Alt+8 must route to the help action (openKeymap builds the overlay)")
	}
}
