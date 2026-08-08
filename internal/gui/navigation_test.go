package gui

import (
	"os"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
	"github.com/user/go-claude-code/internal/config"
)

// buildNavController wires a minimal Controller with a viewStack of panes
// 0..6 plus the rail, mirroring buildChat's shell (contracts §1.3).
func buildNavController(t *testing.T) (*Controller, fyne.Window) {
	t.Helper()
	initSessions(t)
	app := newTestApp(t)
	t.Cleanup(app.Quit)
	win := test.NewWindow(container.NewStack())
	t.Cleanup(win.Close)

	ctrl := &Controller{win: win, app: app}
	ctrl.chat = newChatView()
	ctrl.settings = newSettingsView(win)
	ctrl.usage = newUsageView(win)
	ctrl.git = newGitView(win)
	ctrl.mcp = newMCPView(win)
	ctrl.plugins = newPluginsView(win)
	ctrl.tasks = newAgentTasksView()

	stack := container.NewStack()
	stack.Add(ctrl.chat.container)
	stack.Add(container.NewStack(ctrl.settings.content()))
	stack.Add(container.NewStack(ctrl.usage.content()))
	stack.Add(container.NewStack(ctrl.git.content()))
	stack.Add(container.NewStack(ctrl.mcp.content()))
	stack.Add(container.NewStack(ctrl.plugins.content()))
	stack.Add(container.NewStack(ctrl.tasks.content()))
	ctrl.viewStack = stack

	ctrl.rail = newRailView(false)
	ctrl.rail.onSelect = func(i int) { ctrl.showPane(i) }
	ctrl.rail.onToggle = ctrl.toggleRail
	ctrl.rail.onAction = ctrl.railAction
	root := container.NewBorder(nil, nil, ctrl.rail.content(), nil, stack)
	win.SetContent(root)
	win.SetContent(root)

	ctrl.showChat()
	return ctrl, win
}

// TestPaneBackReturnsToChat verifies showPane pushes history and back() pops
// back to the previous pane, ending at chat (0) with composer focus (SC-009).
func TestPaneBackReturnsToChat(t *testing.T) {
	ctrl, _ := buildNavController(t)

	ctrl.showPane(3) // git
	if !visible(ctrl, 3) {
		t.Fatal("git pane not visible after showPane(3)")
	}
	if ctrl.rail.aid() != "git" {
		t.Fatalf("rail active = %q, want git", ctrl.rail.aid())
	}

	ctrl.back()
	if !visible(ctrl, 0) {
		t.Fatal("back() did not return to chat")
	}
	if ctrl.rail.aid() != "chat" {
		t.Fatalf("rail active = %q, want chat", ctrl.rail.aid())
	}
	if len(ctrl.navHistory) != 0 {
		t.Fatalf("navHistory = %v, want empty", ctrl.navHistory)
	}

	// Multi-step: 2 -> 5 -> back() -> 2 (stack top).
	ctrl.showPane(2)
	ctrl.showPane(5)
	ctrl.back()
	if !visible(ctrl, 2) {
		t.Fatal("back() from 5 must land on 2")
	}
	ctrl.back()
	if !visible(ctrl, 0) {
		t.Fatal("back() from 2 must land on chat")
	}
}

// TestBackGuardsOverlay verifies back() is a no-op while an overlay is open.
func TestBackGuardsOverlay(t *testing.T) {
	ctrl, _ := buildNavController(t)

	ctrl.showPane(4) // mcp
	overlay := container.NewStack(container.NewVBox(widget.NewLabel("overlay")))
	ctrl.viewStack.Add(overlay)
	ctrl.showIndex(len(ctrl.viewStack.Objects) - 1)

	if !ctrl.overlayOpen() {
		t.Fatal("overlayOpen() must report true with the overlay visible")
	}
	ctrl.back()
	if !visible(ctrl, len(ctrl.viewStack.Objects)-1) {
		t.Fatal("back() must not close an overlay")
	}
	if ctrl.rail.aid() != "mcp" {
		t.Fatalf("rail active = %q, want mcp (unchanged)", ctrl.rail.aid())
	}
}

// TestBackChatRootNoop verifies back() at the chat root does nothing.
func TestBackChatRootNoop(t *testing.T) {
	ctrl, _ := buildNavController(t)

	ctrl.back()
	if !visible(ctrl, 0) {
		t.Fatal("back() at root must keep chat visible")
	}
	if len(ctrl.navHistory) != 0 {
		t.Fatalf("navHistory = %v, want empty", ctrl.navHistory)
	}
}

// TestShowPaneSameIndexNoPush verifies navigating to the current pane does
// not grow the history (contracts §1.1).
func TestShowPaneSameIndexNoPush(t *testing.T) {
	ctrl, _ := buildNavController(t)

	ctrl.showPane(1)
	ctrl.showPane(1)
	if len(ctrl.navHistory) != 1 {
		t.Fatalf("navHistory = %v, want [0]", ctrl.navHistory)
	}
	ctrl.back()
	if !visible(ctrl, 0) {
		t.Fatal("back() must return to chat")
	}
}

// TestToggleRailPersists verifies toggleRail flips the rail and writes the
// preference so a fresh LoadConfig round-trips collapsed=true (FR-002,
// contracts §7 TestRailPreferencePersists).
func TestToggleRailPersists(t *testing.T) {
	ctrl, _ := buildNavController(t)
	defer withTestHome()()

	if !ctrl.rail.collapsed {
		ctrl.toggleRail()
	}
	if !ctrl.rail.collapsed {
		t.Fatal("toggleRail must collapse the rail")
	}
	_ = saveAndReload()
}

func visible(ctrl *Controller, i int) bool {
	if i >= len(ctrl.viewStack.Objects) {
		return false
	}
	return !ctrl.viewStack.Objects[i].(*fyne.Container).Hidden
}

// TestToggleThemeVariantPersists verifies the theme action flips the variant
// and round-trips through config (contracts §2 theme action, T006).
func TestToggleThemeVariantPersists(t *testing.T) {
	ctrl, _ := buildNavController(t)
	defer withTestHome()()

	before := config.AppConfig.ThemeVariant
	ctrl.toggleThemeVariant()
	if config.AppConfig.ThemeVariant == before {
		t.Fatal("toggleThemeVariant must flip the theme variant")
	}
	if ctrl.theme == nil {
		t.Fatal("applyTheme must install a theme")
	}
	_ = saveAndReload()
}

// withTestHome redirects HOME/USERPROFILE to a temp dir so config writes and
// reads do not touch the developer machine (mirrors config_test.go).
func withTestHome() func() {
	home := os.Getenv("HOME")
	userProfile := os.Getenv("USERPROFILE")
	tempDir := os.TempDir()
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	return func() {
		os.Setenv("HOME", home)
		os.Setenv("USERPROFILE", userProfile)
	}
}

// saveAndReload persists and re-loads config for the round-trip assertion.
func saveAndReload() error {
	if err := config.SaveConfig(); err != nil {
		return err
	}
	viper.Reset()
	return config.LoadConfig()
}
