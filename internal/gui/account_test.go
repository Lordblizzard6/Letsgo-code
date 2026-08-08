package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/config"
)

// TestAccountFlyout verifies the US2 account flyout (contracts §3): the
// popup shows when invoked, Esc closes it, and the Provider/Model lines read
// the current config.
func TestAccountFlyout(t *testing.T) {
	setupTestDB(t)
	win := test.NewWindow(container.NewStack())
	defer win.Close()

	f := newAccountFlyout(win)
	f.anchor = win.Canvas().Content()
	if f.root == nil {
		t.Fatal("flyout root must be built")
	}

	joined := widgetTexts(f.root)
	if len(joined) == 0 {
		t.Fatal("flyout must render provider/model/usage lines")
	}
	if config.AppConfig.Model != "" && !containsText(joined, config.AppConfig.Model) {
		t.Errorf("flyout missing model %q in %v", config.AppConfig.Model, joined)
	}

	f.show()
	if f.pop == nil {
		t.Fatal("show() must create the popup")
	}
	f.esc.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if f.pop != nil {
		t.Fatal("Esc must close the flyout")
	}
}

// TestAccountFlyoutRoutes verifies T019/T020 wiring: the buttons route to
// the settings overlay (keys), the usage KPI pane and the credential clear.
func TestAccountFlyoutRoutes(t *testing.T) {
	ctrl, _ := buildNavController(t)
	ctrl.buildSettingsOverlay()
	ctrl.account = newAccountFlyout(ctrl.win)
	f := ctrl.account
	f.onKeys = ctrl.showSettings
	f.onUsage = ctrl.showUsage
	f.onSignOut = ctrl.clearCredentials

	// Reset keys so the avatar stays deterministic and clearCredentials has
	// work to do.
	config.AppConfig.AnthropicAPIKey = "sk-test-123456"

	keysBtn := findButtonByText(t, f.root, "API keys & proveedores…")
	usageBtn := findButtonByText(t, f.root, "Uso…")
	signBtn := findButtonByText(t, f.root, "Cerrar sesión")
	if keysBtn == nil || usageBtn == nil || signBtn == nil {
		t.Fatal("flyout must expose the three action buttons")
	}

	test.Tap(keysBtn)
	if !ctrl.settingsOverlay.root.Visible() {
		t.Fatal("keys action must open the settings overlay")
	}
	if f.pop != nil {
		t.Fatal("keys action must hide the popup first")
	}

	test.Tap(usageBtn)
	if !visiblePane(ctrl, 2) {
		t.Fatal("usage action must show the KPI pane (index 2)")
	}

	test.Tap(signBtn)
	if config.AppConfig.AnthropicAPIKey != "" || config.AppConfig.APIKey != "" {
		t.Fatal("sign out must clear credentials")
	}
}

func findButtonByText(t *testing.T, root fyne.CanvasObject, text string) *widget.Button {
	t.Helper()
	var found *widget.Button
	walkButtons(root, func(b *widget.Button) bool {
		if b.Text == text {
			found = b
			return false
		}
		return true
	})
	if found == nil {
		t.Fatalf("button %q not found", text)
	}
	return found
}

func walkButtons(o fyne.CanvasObject, fn func(*widget.Button) bool) {
	if b, ok := o.(*widget.Button); ok {
		if !fn(b) {
			return
		}
	}
	switch c := o.(type) {
	case *fyne.Container:
		for _, child := range c.Objects {
			walkButtons(child, fn)
		}
	case *proseClamp:
		walkButtons(c.CanvasObject, fn)
	}
}

func visiblePane(ctrl *Controller, i int) bool {
	if ctrl.viewStack == nil || i >= len(ctrl.viewStack.Objects) {
		return false
	}
	obj := ctrl.viewStack.Objects[i]
	if co, ok := obj.(*fyne.Container); ok {
		return !co.Hidden
	}
	return obj.Visible()
}

func containsText(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}