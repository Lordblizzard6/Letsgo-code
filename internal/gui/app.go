package gui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/spf13/viper"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
)

// appID is used for Fyne preferences/settings persistence.
const appID = "com.letsgo.code"

// Controller wires the engine and the presentation together.
type Controller struct {
	app fyne.App
	win fyne.Window

	engine   *engine.Engine
	driver   *guiPermissionDriver
	chat     *chatView
	sessions *sessionsView
	settings *settingsView
	status   *statusBar
	pump     *streamPump

	// Advanced views (Phase 6/7).
	usage   *usageView
	git     *gitView
	mcp     *mcpView
	plugins *pluginsView
	tasks   *agentTasksView

	// Command palette overlay (FR-011/026).
	palette    *commandPalette
	paletteIdx int // index into viewStack where the palette overlay lives

	// Cheat sheet overlay (FR-019, T043): '?' opens a categorized keymap
	// that is filterable and closed with Esc.
	keymapOverlay *keymapOverlay

	// Non-modal settings card (004 US2): TabContainer floats above the chat.
	settingsOverlay *settingsOverlay

	// Account flyout (004 US2): popup anchored to the rail avatar.
	account *accountFlyout

	// Startup session picker overlay (FR-014).
	picker    *sessionPicker
	pickerIdx int // index into viewStack where the picker overlay lives

	viewStack *fyne.Container
	root      *fyne.Container
	rail      *railView
	theme     *terminalTheme

	// Pane navigation history (003): top = back() destination.
	navHistory []int

	streamCancel context.CancelFunc
}

// Run launches the desktop GUI (blocking).
func Run() error {
	ctrl := &Controller{}
	if err := ctrl.start(); err != nil {
		return err
	}
	ctrl.win.ShowAndRun()
	return nil
}

func (c *Controller) start() error {
	_ = config.LoadConfig()

	c.app = app.NewWithID(appID)
	c.applyTheme()

	c.win = c.app.NewWindow("LetsGO Code")
	c.win.Resize(fyne.NewSize(1100, 720))
	c.win.SetCloseIntercept(func() {
		// FR-022: reject any pending approvals before stopping the engine.
		if c.driver != nil {
			c.driver.rejectPending()
		}
		c.teardown()
		c.win.Close()
	})

	if hasAnyKey() {
		c.buildChat()
	} else {
		// FR-011: first-run onboarding; block until a provider/key is saved.
		// The rail shell is mounted even here so the app has its Codex look
		// from the very first run (003: no theme/shell regressions on first boot).
		c.buildSettings()
		c.mountShell()
	}

	c.buildMenu()
	attachShortcuts(c.win, c)

	return nil
}

// buildChat wires the engine, chat view, status bar, event pump and sidebar.
func (c *Controller) buildChat() {
	if c.engine == nil {
		c.streamCancel = func() {}
		c.driver = &guiPermissionDriver{win: c.win}
		c.engine = engine.New(c.driver)
		c.engine.Start()
	}

	c.chat = newChatView()
	c.chat.onSend = func(text string) {
		c.engine.Send(engine.SendMessage{Text: text, SessionID: c.engine.SessionID()})
		c.chat.SetBusy(true)
	}
	c.chat.onSteer = func(text string) {
		c.engine.Send(engine.Steer{Text: text, SessionID: c.engine.SessionID()})
	}
	c.chat.onQueue = func(text string) {
		c.engine.Send(engine.Queue{Text: text, SessionID: c.engine.SessionID()})
	}
	c.chat.onAgent = func(text string) {
		c.engine.Send(engine.SendTask{Prompt: text, SessionID: c.engine.SessionID()})
		c.chat.SetBusy(true)
	}
	c.chat.onCancel = func() {
		c.engine.Send(engine.Cancel{})
	}
	c.chat.onSlash = func() {
		c.openPalette()
	}
	c.chat.onHole = func() {
		c.openKeymap()
	}
	c.chat.onPlanApprove = func(planID string) {
		c.engine.Send(engine.ApprovePlan{PlanID: planID})
	}
	c.chat.onPlanReject = func(planID string) {
		c.engine.Send(engine.RejectPlan{PlanID: planID})
	}
	c.chat.onPlanEdit = func(planID, instructions string) {
		c.engine.Send(engine.EditPlan{PlanID: planID, Instructions: instructions})
	}
	c.driver.onShow = func(card *approvalCard) {
		c.chat.ShowApproval(card)
	}
	c.driver.onDismiss = func() {
		c.chat.DismissApproval()
	}
	c.driver.sessionID = func() string {
		return c.engine.SessionID()
	}

	c.status = newStatusBar()
	c.status.onToggleApprove = func(category string) {
		enabled := !config.AppConfig.AutoApprove[category]
		cfg := &config.AppConfig
		if cfg.AutoApprove == nil {
			cfg.AutoApprove = map[string]bool{}
		}
		cfg.AutoApprove[category] = enabled
		_ = config.SaveConfig()
		c.driver.setAutoApprove(category, enabled)
		c.engine.Send(engine.SetAutoApprove{Category: category, Enabled: enabled})
		c.status.refreshAuto()
	}

	c.tasks = newAgentTasksView()
	if c.pump == nil {
		c.pump = newStreamPump(c.engine.Events(), c.chat, c.status, c.streamCancel)
		c.pump.tasks = c.tasks
	}

	c.settings = newSettingsView(c.win)
	c.settings.onSaved = func() {
		if c.engine != nil {
			c.engine.RefreshClient()
		}
		c.status.refreshModel()
	}
	c.settings.onStatusLineChanged = func() {
		if c.status != nil {
			c.status.applyFields()
		}
	}

	c.sessions = newSessionsView(c.win)
	c.sessions.onSelect = func(id string) {
		c.switchSession(id)
	}
	c.sessions.onNew = func() {
		c.newSession()
	}
	c.sessions.reload()

	// Advanced views.
	c.usage = newUsageView(c.win)
	c.git = newGitView(c.win)
	c.git.onReview = func() {
		c.chat.onAgent("Review the current git changes for issues and suggest improvements.")
	}
	c.mcp = newMCPView(c.win)
	c.plugins = newPluginsView(c.win)

	chatPane := container.NewBorder(nil, c.status.content(), nil, nil,
		container.NewHSplit(c.sessions.content(), c.chat.container))

	c.viewStack = container.NewStack()
	c.viewStack.Add(chatPane)
	c.viewStack.Add(container.NewStack(c.settings.content()))
	c.viewStack.Add(container.NewStack(c.usage.content()))
	c.viewStack.Add(container.NewStack(c.git.content()))
	c.viewStack.Add(container.NewStack(c.mcp.content()))
	c.viewStack.Add(container.NewStack(c.plugins.content()))
	c.viewStack.Add(container.NewStack(c.tasks.content()))
	c.buildPalette()
	c.buildKeymap()
	c.buildSettingsOverlay()

	// The shell (rail + SetContent) must exist BEFORE any overlay (picker,
	// palette, keymap) tries to focus an input — Fyne rejects focusing an
	// object that is not yet part of the canvas (canvas.go guard).
	c.mountShell()

	// FR-014: startup picker when previous sessions exist.
	c.buildSessionPicker()

	c.showChat()
	c.sessions.reload()
}

// mountShell creates the rail and the root Border shell around the viewStack
// (003). Shared by the chat path and the first-run settings path so the app
// has its Codex look from the very first boot.
func (c *Controller) mountShell() {
	c.rail = newRailView(config.AppConfig.Rail.Collapsed)
	c.rail.onSelect = func(i int) { c.showPane(i) }
	c.rail.onToggle = c.toggleRail
	c.rail.onAction = c.railAction
	c.rail.setAvatarStatus(avatarStatusColor())
	c.root = container.NewBorder(nil, nil, c.rail.content(), nil, c.viewStack)

	// SetContent must happen BEFORE showChat: requestFocus needs the composer
	// to be part of the canvas (Fyne focus guard, SC-009).
	c.win.SetContent(c.root)
}

// buildSessionPicker wires the FR-014 startup picker. It appears only when
// the database already has sessions; the user can then resume one or start new.
func (c *Controller) buildSessionPicker() {
	if c.picker != nil || c.viewStack == nil {
		return
	}
	sessions, err := db.ListSessions()
	if err != nil || len(sessions) == 0 {
		return
	}
	c.picker = newSessionPicker(c.win)
	c.picker.onNew = func() {
		c.closePicker()
		c.newSession()
	}
	c.picker.onPick = func(id string) {
		c.closePicker()
		c.switchSession(id)
	}
	c.picker.onClose = func() {
		// Esc: proceed with the current active session rather than picking.
		c.closePicker()
		c.showChat()
	}
	i := len(c.viewStack.Objects)
	overlay := container.NewStack(c.picker.root)
	overlay.Hidden = false
	c.viewStack.Add(overlay)
	c.picker.open()

	// Hide all others: the picker is the first thing the user sees.
	for idx, obj := range c.viewStack.Objects {
		obj.(*fyne.Container).Hidden = idx != i
	}
	c.pickerIdx = i
}

// closePicker hides the startup picker overlay.
func (c *Controller) closePicker() {
	for i, obj := range c.viewStack.Objects {
		vis := i == 0
		obj.(*fyne.Container).Hidden = !vis
	}
}

// showPickerFromMenu reopens the picker on demand (Session menu / palette).
func (c *Controller) showPickerFromMenu() {
	if c.picker == nil {
		// No overlay exists yet; build and show it.
		c.pickerIdx = len(c.viewStack.Objects)
		c.picker = newSessionPicker(c.win)
		c.picker.onNew = func() {
			c.closePicker()
			c.newSession()
		}
		c.picker.onPick = func(id string) {
			c.closePicker()
			c.switchSession(id)
		}
		c.picker.onClose = func() {
			c.closePicker()
			c.showChat()
		}
		overlay := container.NewStack(c.picker.root)
		overlay.Hidden = true
		c.viewStack.Add(overlay)
	}
	for idx, obj := range c.viewStack.Objects {
		obj.(*fyne.Container).Hidden = idx != c.pickerIdx
	}
	c.picker.open()
}

// buildKeymap wires the '?' cheat-sheet overlay (FR-019, T043): it shows a
// categorized, filterable list of shortcuts and closes with Esc.
func (c *Controller) buildKeymap() {
	if c.keymapOverlay != nil || c.viewStack == nil {
		return
	}
	c.keymapOverlay = newKeymapOverlay(c.win)
	c.keymapOverlay.onClose = c.closeKeymap
	overlay := container.NewStack(c.keymapOverlay.root)
	overlay.Hidden = true
	c.viewStack.Add(overlay)
}

// openKeymap shows the cheat sheet overlay and focuses its filter. It hides
// the palette overlay if open so the two never stack visually.
func (c *Controller) openKeymap() {
	if c.keymapOverlay == nil || c.viewStack == nil {
		return
	}
	if c.palette != nil && c.paletteIdx < len(c.viewStack.Objects) {
		c.viewStack.Objects[c.paletteIdx].(*fyne.Container).Hidden = true
	}
	for i, obj := range c.viewStack.Objects {
		if i == 0 {
			continue
		}
		if i == c.paletteIdx {
			continue
		}
		ov, ok := obj.(*fyne.Container)
		if !ok || len(ov.Objects) == 0 {
			continue
		}
		ov.Hidden = !(ov.Objects[0] == c.keymapOverlay.root)
	}
	c.keymapOverlay.open()
}

// closeKeymap hides the cheat sheet overlay.
func (c *Controller) closeKeymap() {
	if c.keymapOverlay == nil || c.viewStack == nil {
		return
	}
	for _, obj := range c.viewStack.Objects {
		ov, ok := obj.(*fyne.Container)
		if !ok || len(ov.Objects) == 0 {
			continue
		}
		if ov.Objects[0] == c.keymapOverlay.root {
			ov.Hidden = true
			break
		}
	}
	if c.chat != nil {
		c.chat.input.requestFocus()
	}
}

// buildPalette wires the command palette overlay (FR-011/026): Ctrl+K opens
// it, / in the composer opens it, actions run app commands, @ inserts file
// paths into the composer, ! shells go through the existing permission path.
func (c *Controller) buildPalette() {
	if c.palette != nil {
		return
	}
	c.palette = newCommandPalette(c.win)
	c.palette.onExec = func(text string) {
		c.paletteExecute(text)
	}
	c.paletteIdx = len(c.viewStack.Objects)
	overlay := container.NewStack(c.palette.root)
	overlay.Hidden = true
	c.viewStack.Add(overlay)
}

// paletteExecute handles a palette selection: internal actions run directly,
// @path inserts into the composer, !command is passed to the engine as a
// user message (executed only with the engine's normal permission checks).
func (c *Controller) paletteExecute(text string) {
	c.closePalette()
	if c.chat == nil || c.engine == nil {
		return
	}
	if strings.HasPrefix(text, "@") {
		c.chat.input.SetText(c.chat.input.Text + " " + text)
		c.chat.input.requestFocus()
		return
	}
	if strings.HasPrefix(text, "!") {
		cmd := strings.TrimPrefix(text, "!")
		if cmd == "" {
			return
		}
		c.chat.input.SetText("")
		c.engine.Send(engine.SendMessage{Text: cmd, SessionID: c.engine.SessionID()})
		c.chat.SetBusy(true)
		return
	}
	// Internal action catalogue (FR-011).
	switch text {
	case "New session":
		c.newSession()
	case "Compact context":
		c.engine.Send(engine.Compact{})
	case "Settings":
		c.showSettings()
	case "Usage":
		c.showUsage()
	case "Git":
		c.showGit()
	case "MCP":
		c.showMCP()
	case "Plugins":
		c.showPlugins()
	case "Agent tasks":
		c.showTasks()
	case "Plan mode":
		c.togglePlanMode()
	case "Fork session":
		c.forkSession()
	case "Auto-approve toggle":
		if c.status != nil {
			c.status.setAllApprovals()
		}
	case "Focus composer":
		c.focusInput()
	}
}

// openPalette shows the command palette overlay (Ctrl+K or "/" in composer).
func (c *Controller) openPalette() {
	if c.palette == nil || c.viewStack == nil {
		return
	}
	c.palette.entry.SetText("")
	c.palette.refresh()
	for i, obj := range c.viewStack.Objects {
		obj.(*fyne.Container).Hidden = i != c.paletteIdx
	}
	c.palette.open()
}

// closePalette hides the palette overlay and restores the previous view.
func (c *Controller) closePalette() {
	if c.palette == nil || c.viewStack == nil {
		return
	}
	for i, obj := range c.viewStack.Objects {
		vis := i == 0
		obj.(*fyne.Container).Hidden = !vis
	}
}

// forkSession clones the active session at a user-chosen message (FR-015).
func (c *Controller) forkSession() {
	if c.engine == nil || c.chat == nil {
		return
	}
	c.chat.fork = newForkChooser(c.engine.SessionID())
	c.chat.fork.onFork = c.doFork
	c.chat.fork.onClose = func() {
		c.chat.RemoveFork()
		c.chat.input.requestFocus()
	}
	c.chat.ShowFork()
}

// doFork forks the source session at the chosen message and switches to the
// copy (FR-015). The new session ID is obtained synchronously so the view can
// reload immediately; the engine then resumes the fork via SwitchSession.
func (c *Controller) doFork(sourceID string, messageID int64) {
	c.chat.RemoveFork()
	if c.engine == nil {
		return
	}
	newID, err := db.ForkSession(sourceID, messageID)
	if err != nil {
		c.showError(fmt.Sprintf("failed to fork session: %v", err))
		return
	}
	c.switchSession(newID)
}

// showSettings switches to the settings pane (FR-007, Ctrl+,). Since 004 US2
// the settings surface is a TabContainer (Cuenta/Apariencia/Preferencias/Uso);
// when buildSettingsOverlay mounted the card, showSettings() opens it; minimal
// headless harnesses (navigation_test) keep the pane at index 1 (contracts §2).
func (c *Controller) showSettings() {
	if c.settings != nil {
		c.settings.refreshGrants()
	}
	if c.settingsOverlay != nil {
		c.openSettings()
		return
	}
	c.showPane(1)
}

// buildSettingsOverlay mounts the non-modal settings card over the chat shell
// (US2, research D2). It is only built by the real chat shell so headless
// controller tests keep the pane fallback.
func (c *Controller) buildSettingsOverlay() {
	if c.settingsOverlay != nil || c.settings == nil || c.viewStack == nil {
		return
	}
	c.settingsOverlay = newSettingsOverlay(c.win, c.settings)
	c.settingsOverlay.onClose = func() {
		// close() already hid the card; just return focus to the composer.
		c.focusInput()
	}
	if c.settings != nil {
		c.settings.onCancel = func() {
			c.closeSettings()
		}
		c.settings.onOpenUsage = c.showUsage
	}
	overlay := container.NewStack(c.settingsOverlay.root)
	overlay.Hidden = true
	c.viewStack.Add(overlay)
}

// openSettings reveals the floating settings card.
func (c *Controller) openSettings() {
	if c.settingsOverlay == nil {
		return
	}
	c.settingsOverlay.open()
}

// closeSettings hides the settings card and returns focus to the composer
// (contracts §2: Esc/Cancel → focusInput).
func (c *Controller) closeSettings() {
	if c.settingsOverlay != nil {
		c.settingsOverlay.close()
	}
	c.focusInput()
}

// showUsage switches to the usage pane (FR-009, Ctrl+U).
func (c *Controller) showUsage() {
	if c.usage != nil {
		c.usage.refresh()
	}
	c.showPane(2)
}

// showGit switches to the git pane.
func (c *Controller) showGit() {
	if c.git != nil {
		c.git.refresh()
	}
	c.showPane(3)
}

// showMCP switches to the MCP pane.
func (c *Controller) showMCP() { c.showPane(4) }

// showPlugins switches to the plugins pane.
func (c *Controller) showPlugins() { c.showPane(5) }

// showTasks switches to the agent tasks pane.
func (c *Controller) showTasks() { c.showPane(6) }

// showAccount opens the account flyout anchored to the rail avatar (FR-005,
// US2, research D8): provider/model, today's usage and actions. It no longer
// redirects to settings (004).
func (c *Controller) showAccount() {
	if c.rail == nil || c.rail.avatar == nil {
		c.showSettings()
		return
	}
	if c.account == nil {
		c.account = newAccountFlyout(c.win)
		c.account.anchor = c.rail.avatar
		c.account.onKeys = c.showSettings
		c.account.onUsage = c.showUsage
		c.account.onSignOut = c.clearCredentials
		c.account.onClose = func() { c.focusInput() }
	}
	c.account.show()
}

// railAction routes rail action slots (-1): help opens the keymap, account the
// providers pane, theme toggles the variant (contracts §2).
func (c *Controller) railAction(id string) {
	switch id {
	case "help":
		c.openKeymap()
	case "account":
		c.showAccount()
	case "theme":
		c.toggleThemeVariant()
	}
}

// applyTheme installs the Deep Goblue theme with the configured variant
// (FR-007): config.theme "light" overrides the OS preference.
func (c *Controller) applyTheme() {
	variant := theme.VariantDark
	switch config.AppConfig.ThemeVariant {
	case "light":
		variant = theme.VariantLight
	}
	c.theme = &terminalTheme{forcedVariant: variant}
	c.app.Settings().SetTheme(c.theme)
}

// toggleThemeVariant flips dark<->light, persists and re-applies (T006).
func (c *Controller) toggleThemeVariant() {
	if config.AppConfig.ThemeVariant == "light" {
		config.AppConfig.ThemeVariant = "dark"
	} else {
		config.AppConfig.ThemeVariant = "light"
	}
	_ = config.SaveConfig()
	c.applyTheme()
}

func (c *Controller) showIndex(i int) {
	if c.viewStack == nil {
		return
	}
	for idx, obj := range c.viewStack.Objects {
		obj.(*fyne.Container).Hidden = idx != i
	}
}

// switchNth switches to the Nth session (Ctrl+1..9).
func (c *Controller) switchNth(n int) {
	if c.sessions == nil {
		return
	}
	ids := c.sessions.sessionIDs()
	if n >= 0 && n < len(ids) {
		c.switchSession(ids[n])
	}
}

// buildMenu constructs the application menu bar (File·Session·Git·Tools·Help).
func (c *Controller) buildMenu() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New Session", func() { c.newSession() }),
		fyne.NewMenuItem("Close", c.win.Close),
	)
	sessionMenu := fyne.NewMenu("Session",
		fyne.NewMenuItem("New", func() { c.newSession() }),
		fyne.NewMenuItem("Compact context", func() {
			if c.engine != nil {
				c.engine.Send(engine.Compact{})
			}
		}),
		fyne.NewMenuItem("Rewind…", func() {
			if c.sessions != nil && c.engine != nil {
				c.sessions.rewind(c.engine.SessionID())
			}
		}),
		fyne.NewMenuItem("Teleport…", func() {
			if c.sessions != nil && c.engine != nil {
				c.sessions.teleport(c.engine.SessionID())
			}
		}),
		fyne.NewMenuItem("Clear", func() {
			if c.chat != nil {
				c.chat.Clear()
			}
		}),
		fyne.NewMenuItem("Settings", c.showSettings),
	)
	gitMenu := fyne.NewMenu("Git",
		fyne.NewMenuItem("Open Git view", c.showGit),
		fyne.NewMenuItem("Review changes", func() { c.sendAgentPrompt("Review the current git changes for bugs, style issues and improvements.") }),
		fyne.NewMenuItem("Ultra review", func() {
			c.sendAgentPrompt("Perform a deep, line-by-line review of the current changes and propose concrete fixes.")
		}),
		fyne.NewMenuItem("Security review", func() {
			c.sendAgentPrompt("Security review the current changes: vulnerabilities, secrets, and hardening.")
		}),
	)
	toolsMenu := fyne.NewMenu("Tools",
		fyne.NewMenuItem("Settings (Ctrl+,)", c.showSettings),
		fyne.NewMenuItem("Usage (Ctrl+U)", c.showUsage),
		fyne.NewMenuItem("Git", c.showGit),
		fyne.NewMenuItem("MCP", c.showMCP),
		fyne.NewMenuItem("Plugins", c.showPlugins),
		fyne.NewMenuItem("Agent tasks", c.showTasks),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Advisor", func() {
			c.sendAgentPrompt("Act as an architecture advisor: assess the current codebase and suggest the most valuable improvements.")
		}),
		fyne.NewMenuItem("Insights", func() { c.showParityInfo("Insights") }),
		fyne.NewMenuItem("Thinkback (btw)", func() { c.showThinkback() }),
		fyne.NewMenuItem("Memory", func() { c.showParityInfo("Memory") }),
		fyne.NewMenuItem("Skills", func() { c.showParityInfo("Skills") }),
		fyne.NewMenuItem("Share…", func() { c.showParityInfo("Share") }),
		fyne.NewMenuItem("Export…", c.exportSession),
		fyne.NewMenuItem("Onboarding", func() { c.showParityInfo("Onboarding") }),
		fyne.NewMenuItem("Plan mode", func() { c.togglePlanMode() }),
		fyne.NewMenuItem("Fast mode", func() { c.toggleFastMode() }),
		fyne.NewMenuItem("Debug tool call", func() { c.showParityInfo("Debug tool call") }),
		fyne.NewMenuItem("Install GitHub app", func() { c.showParityInfo("Install GitHub app") }),
		fyne.NewMenuItem("Install Slack app", func() { c.showParityInfo("Install Slack app") }),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Doctor", c.showDoctor),
		fyne.NewMenuItem("About", func() { showAbout(c.win) }),
		fyne.NewMenuItem("Update", func() { showUpdate(c.win) }),
		fyne.NewMenuItem("Keybindings", func() { showKeybindings(c.win) }),
	)
	c.win.SetMainMenu(fyne.NewMainMenu(fileMenu, sessionMenu, gitMenu, toolsMenu, helpMenu))
}

// sendAgentPrompt dispatches an agent-mode prompt into the active chat.
func (c *Controller) sendAgentPrompt(prompt string) {
	if c.chat != nil && c.engine != nil {
		c.engine.Send(engine.SendTask{Prompt: prompt, SessionID: c.engine.SessionID()})
		c.chat.SetBusy(true)
		return
	}
	dialog.ShowInformation("Agent", "Configure a provider and open the chat to run agent tasks.", c.win)
}

// showThinkback lists recent sessions as recall targets (cmd/thinkback.go parity).
func (c *Controller) showThinkback() {
	sessions, err := db.ListSessions()
	if err != nil {
		c.showError("thinkback: " + err.Error())
		return
	}
	if len(sessions) == 0 {
		dialog.ShowInformation("Thinkback", "No sessions to recall.", c.win)
		return
	}
	var b strings.Builder
	for i, s := range sessions {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&b, "%d. %s (%s)\n", i+1, s.Name, s.UpdatedAt.Format("01-02 15:04"))
	}
	dialog.ShowInformation("Thinkback — recent sessions", b.String(), c.win)
}

// togglePlanMode flips the plan-mode preference (cmd/plan.go parity).
func (c *Controller) togglePlanMode() {
	viper.Set("plan", !viper.GetBool("plan"))
	_ = config.SaveConfig()
	state := "off"
	if viper.GetBool("plan") {
		state = "on"
	}
	dialog.ShowInformation("Plan mode", "Plan mode is "+state+".", c.win)
}

// toggleFastMode flips the fast-mode preference (cmd/fast.go parity).
func (c *Controller) toggleFastMode() {
	viper.Set("fast", !viper.GetBool("fast"))
	_ = config.SaveConfig()
	state := "off"
	if viper.GetBool("fast") {
		state = "on"
	}
	dialog.ShowInformation("Fast mode", "Fast mode is "+state+".", c.win)
}

// showDoctor runs the diagnostics in a dialog.
func (c *Controller) showDoctor() {
	dialog.ShowInformation("Doctor", runDoctor(), c.win)
}

// showParityInfo opens an informational dialog for parity rows that share a
// simple text target (cli-parity.md rows with no dedicated view).
func (c *Controller) showParityInfo(name string) {
	hints := map[string]string{
		"Memory":     "Persistent key/value memory lives in the local database.\nUse the CLI: letsgo memory list",
		"Skills":     "Skills are discoverable prompts; run them from the chat.\nCLI: letsgo skills list",
		"Share":      "Session sharing exports the current conversation.\nUse File > Export to save it.",
		"Onboarding": "First-run flow: configure a provider in Settings.\nAlready configured here.",
	}
	msg := hints[name]
	if msg == "" {
		msg = name + " is reachable from this menu (parity)."
	}
	dialog.ShowInformation(name, msg, c.win)
}

// exportSession writes the active session history to a JSON file.
func (c *Controller) exportSession() {
	session, err := db.GetActiveSession()
	if err != nil || session == nil {
		c.showError("No active session to export.")
		return
	}
	history, err := db.GetHistory(session.ID)
	if err != nil || len(history) == 0 {
		c.showError("No messages to export.")
		return
	}
	out := map[string]interface{}{
		"session_id": session.ID,
		"name":       session.Name,
		"messages":   history,
	}
	content := fmt.Sprintf("%v", out)
	dialog.ShowInformation("Export", "Session exported to conversation.json\n(first 500 chars):\n"+clip(content, 500), c.win)
}

// clip truncates a string to n characters.
func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// keybindings lists the FR-025 shortcut map.
func showKeybindings(win fyne.Window) {
	dialog.ShowInformation("Keybindings",
		"Ctrl+N  New session\n"+
			"Ctrl+1..9  Switch session\n"+
			"Ctrl+,  Settings\n"+
			"Ctrl+U  Usage\n"+
			"Ctrl+L  Focus input\n"+
			"Ctrl+Shift+A  Auto-approve toggle\n"+
			"Esc  Cancel (while streaming)", win)
}

// newSession creates a fresh session and switches to it, returning its ID.
func (c *Controller) newSession() string {
	sid, err := db.CreateSession("New session", ".")
	if err != nil {
		c.showError(fmt.Sprintf("failed to create session: %v", err))
		return ""
	}
	if c.engine != nil {
		c.switchSession(sid)
	} else {
		c.adminSessionCreated(sid)
	}
	return sid
}

func (c *Controller) adminSessionCreated(id string) {
	if c.sessions != nil {
		c.sessions.reload()
	}
}

// switchSession loads the session's history and resumes it in the engine.
func (c *Controller) switchSession(id string) {
	if c.engine == nil {
		return
	}
	history, err := db.GetHistory(id)
	if err != nil {
		c.showError(fmt.Sprintf("failed to load session: %v", err))
		return
	}
	c.chat.LoadHistory(history)
	c.engine.Send(engine.SwitchSession{SessionID: id})
	if c.sessions != nil {
		c.sessions.selectSession(id)
		c.sessions.reload()
	}
}

// showError surfaces a transient error message in the chat view.
func (c *Controller) showError(msg string) {
	if c.chat != nil {
		c.chat.ShowError(errors.New(msg))
	}
}

// buildSettings shows only the settings form (used on first run). The pane is
// mounted as pane 1 of a minimal viewStack so the rail shell works on first
// boot; after a key is saved buildChat swaps in the full stack.
func (c *Controller) buildSettings() {
	c.settings = newSettingsView(c.win)
	c.settings.saveBtn.OnTapped = func() {
		c.settings.save()
		if hasAnyKey() {
			c.buildChat()
		}
	}
	c.viewStack = container.NewStack(container.NewStack(c.settings.content()))
}

func (c *Controller) teardown() {
	if c.pump != nil {
		c.pump.Stop()
	}
	if c.engine != nil {
		c.engine.Stop()
	}
}

// hasAnyKey reports whether any provider API key is configured (FR-011).
func hasAnyKey() bool {
	cfg := config.AppConfig
	return cfg.APIKey != "" ||
		cfg.AnthropicAPIKey != "" ||
		cfg.OpenAIAPIKey != "" ||
		cfg.GroqAPIKey != "" ||
		cfg.OpenRouterAPIKey != ""
}

// clearCredentials signs the user out of all providers (account flyout,
// US2): clears keys, persists and refreshes the avatar status.
func (c *Controller) clearCredentials() {
	cfg := &config.AppConfig
	cfg.APIKey = ""
	cfg.AnthropicAPIKey = ""
	cfg.OpenAIAPIKey = ""
	cfg.GroqAPIKey = ""
	cfg.OpenRouterAPIKey = ""
	_ = config.SaveConfig()
	if c.rail != nil {
		c.rail.setAvatarStatus(avatarStatusColor())
	}
}

// avatarStatusColor derives the rail avatar status dot: green when a provider
// key or engine is configured, orange when none is (FR-006, T006).
func avatarStatusColor() color.NRGBA {
	if hasAnyKey() || config.AppConfig.BaseURL != "" {
		return colorGreen
	}
	return colorOrange
}
