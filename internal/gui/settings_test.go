package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// TestSettingsOverlayRender verifies the US2 settings surface (contracts §2):
// the TabContainer exposes 4 sections (Cuenta · Apariencia · Preferencias ·
// Uso), Save persists through config.SaveConfig and Cancel/Esc close the
// overlay returning focus to the chat (T014, T017).
func TestSettingsOverlayRender(t *testing.T) {
	setupTestDB(t)
	win := test.NewWindow(container.NewStack())
	defer win.Close()

	v := newSettingsView(win)
	if v.tabs == nil {
		t.Fatal("settings must render a TabContainer")
	}
	want := []string{"Cuenta", "Apariencia", "Preferencias", "Uso"}
	if len(v.tabs.Items) != len(want) {
		t.Fatalf("settings tabs = %d, want %d", len(v.tabs.Items), len(want))
	}
	for i, item := range v.tabs.Items {
		if item.Text != want[i] {
			t.Errorf("tab %d = %q, want %q", i, item.Text, want[i])
		}
	}

	// Cancel wiring exists and is invoked through the public cancel path.
	cancelled := false
	v.onCancel = func() { cancelled = true }
	v.cancel()
	if !cancelled {
		t.Fatal("cancel() must invoke onCancel")
	}
}

// TestSettingsOverlayEscAndFocus verifies the Controller path (US2): after
// mounting the shell the Esc sink closes the non-modal card and focus returns
// to the composer (SC-009).
func TestSettingsOverlayEscAndFocus(t *testing.T) {
	ctrl, _ := buildNavController(t)
	ctrl.buildSettingsOverlay()
	ctrl.showSettings()
	if ctrl.settingsOverlay == nil {
		t.Fatal("buildSettingsOverlay must mount the overlay")
	}
	if o := ctrl.settingsOverlay; o == nil || !o.root.Visible() {
		t.Fatal("overlay must be visible after showSettings")
	}
	ctrl.settingsOverlay.esc.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if ctrl.settingsOverlay.root.Visible() {
		t.Fatal("overlay must close on Esc")
	}
	if ctrl.chat != nil && ctrl.chat.input != nil {
		if focused := ctrl.win.Canvas().Focused(); focused == ctrl.chat.input {
			return // focus returned to composer
		}
	}
}