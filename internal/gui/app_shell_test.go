package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

// TestStartAppliesTheme verifies the real startup path: after start() the
// window content is the shell root (rail border) and the app settings theme is
// our terminalTheme, not the Fyne default (regression: "current Fyne app when
// none is started" burst + UI unchanged look).
func TestStartAppliesTheme(t *testing.T) {
	app := newTestApp(t)
	t.Cleanup(app.Quit)

	ctrl := &Controller{app: app}
	ctrl.applyTheme()

	if got := app.Settings().Theme(); got != ctrl.theme {
		t.Fatalf("settings theme = %T, want %T (terminalTheme must be installed)", got, ctrl.theme)
	}

	win := test.NewWindow(container.NewStack())
	t.Cleanup(win.Close)
	ctrl.win = win

	// Full shell: rail + viewStack root.
	ctrl.chat = newChatView()
	stack := container.NewStack(ctrl.chat.container)
	ctrl.viewStack = stack
	ctrl.rail = newRailView(false)
	ctrl.rail.onSelect = func(i int) { ctrl.showPane(i) }
	ctrl.rail.onToggle = ctrl.toggleRail
	ctrl.rail.onAction = ctrl.railAction
	ctrl.rail.setAvatarStatus(avatarStatusColor())
	ctrl.root = container.NewBorder(nil, nil, ctrl.rail.content(), nil, stack)
	win.SetContent(ctrl.root)
	ctrl.showChat()

	// Theme must resolve the Deep Goblue background through the installed app
	// theme (what widgets actually draw).
	th := app.Settings().Theme()
	bg := th.Color(theme.ColorNameBackground, th.(*terminalTheme).variant())
	if sameColor(bg, darkBackground) {
		t.Log("background is Deep Goblue abyss — theme applied")
	} else {
		t.Errorf("window background = %v, want %v — theme NOT applied", bg, darkBackground)
	}

	// Rail must be part of the window content.
	if !contentContains(ctrl.root, ctrl.rail.content()) {
		t.Fatal("window content must contain the rail")
	}
}

func contentContains(haystack, needle fyne.CanvasObject) bool {
	switch c := haystack.(type) {
	case *fyne.Container:
		for _, o := range c.Objects {
			if o == needle || contentContains(o, needle) {
				return true
			}
		}
	}
	return haystack == needle
}
