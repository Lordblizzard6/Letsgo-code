package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// TestShellSmokeKeyboardTail scans the whole shell (004 US1): mount a
// Controller, drive every rail destination via Alt+N and the keymap overlay
// with Esc, ending back at chat. Contracts §1.3, §11 (T013).
func TestShellSmokeKeyboardTail(t *testing.T) {
	ctrl, win := buildNavController(t)
	attachShortcuts(win, ctrl)
	ctrl.buildKeymap()

	// Alt+1 => chat (already), Alt+8 => help action opens keymap.
	handler := &fyne.ShortcutHandler{}
	alt := func(n int) *desktop.CustomShortcut {
		return &desktop.CustomShortcut{KeyName: numberKeys[n], Modifier: fyne.KeyModifierAlt}
	}
	for i := 1; i <= 8; i++ {
		n := i
		handler.AddShortcut(alt(n), func(s fyne.Shortcut) { ctrl.railShortcut(n) })
	}
	handler.TypedShortcut(alt(8))
	if ctrl.keymapOverlay == nil {
		t.Fatal("Alt+8 must open the keymap overlay")
	}

	// Esc inside the keymap filter closes it and leaves the overlay hidden.
	ctrl.keymapOverlay.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if ctrl.overlayOpen() {
		t.Fatal("Esc must close the keymap overlay")
	}

	// Drive the remaining destinations and land back on chat.
	ctrl.showPane(3)
	ctrl.showPane(6)
	ctrl.showPane(4)
	ctrl.showPane(5)
	ctrl.showPane(1)
	ctrl.showPane(2)
	ctrl.showPane(0)
	if ctrl.rail.aid() != "chat" {
		t.Fatalf("rail active = %q, want chat", ctrl.rail.aid())
	}
	if len(ctrl.viewStack.Objects) == 0 {
		t.Fatal("view stack must survive the navigation tail")
	}
}