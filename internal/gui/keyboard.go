package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// focusInput moves keyboard focus back to the composer so a keyboard-only
// user never has to touch the mouse (SC-009). Safe when the chat is not built.
func (c *Controller) focusInput() {
	if c.chat != nil && c.viewStack != nil {
		for i, obj := range c.viewStack.Objects {
			obj.(*fyne.Container).Hidden = i != 0
		}
		c.chat.input.requestFocus()
	}
}

// attachShortcuts registers the global keyboard map
// (contracts/gui-contract.md, FR-025). Handlers are provided by the controller
// so navigation reaches sessions, settings, usage and the input box.
func attachShortcuts(win fyne.Window, c *Controller) {
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}, func() {
		c.newSession()
		c.focusInput()
	})
	for i := 0; i < 9; i++ {
		idx := i
		addShortcut(win, newNumberShortcut(idx), func() {
			c.switchNth(idx)
			c.focusInput()
		})
	}
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierControl}, func() {
		c.showSettings()
	})
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyU, Modifier: fyne.KeyModifierControl}, func() {
		c.showUsage()
	})
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyL, Modifier: fyne.KeyModifierControl}, func() {
		if c.chat != nil {
			c.chat.input.requestFocus()
		}
	})
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl}, func() {
		c.openPalette()
	})
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyA, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}, func() {
		if c.status != nil {
			c.status.setAllApprovals()
		}
	})
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeySlash, Modifier: fyne.KeyModifierShift}, func() {
		c.openKeymap()
	})
	// FR-003/FR-025: rail navigation (003/004). Alt+1..8 mirror railSlots focus
	// order (chat, git, tasks, mcp, plugins, settings, usage, help); Alt+Left
	// pops the pane history. The mapping reads railSlots at handler time so the
	// catalog stays the single source of truth.
	for i := 1; i <= 8; i++ {
		n := i
		addShortcut(win, &desktop.CustomShortcut{KeyName: numberKeys[n], Modifier: fyne.KeyModifierAlt}, func() {
			c.railShortcut(n)
		})
	}
	addShortcut(win, &desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierAlt}, func() {
		c.back()
	})
}

// numberKeys maps 0..9 to fyne key names (shared by Ctrl+n and Alt+n).
var numberKeys = map[int]fyne.KeyName{
	0: fyne.Key0, 1: fyne.Key1, 2: fyne.Key2, 3: fyne.Key3, 4: fyne.Key4,
	5: fyne.Key5, 6: fyne.Key6, 7: fyne.Key7, 8: fyne.Key8, 9: fyne.Key9,
}

// newNumberShortcut returns the Ctrl+n shortcut for session switching (002).
func newNumberShortcut(n int) *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: numberKeys[n], Modifier: fyne.KeyModifierControl}
}

// railShortcut dispatches Alt+n to the nth rail slot (contracts §3). The
// mapping follows railSlots order; action slots route through onAction.
func (c *Controller) railShortcut(n int) {
	if n < 1 || n > len(railSlots) {
		return
	}
	slot := railSlots[n-1]
	if slot.paneIndex >= 0 {
		c.showPane(slot.paneIndex)
		return
	}
	if c.rail != nil && c.rail.onAction != nil {
		c.rail.onAction(slot.id)
	}
}

func addShortcut(win fyne.Window, s *desktop.CustomShortcut, fn func()) {
	win.Canvas().AddShortcut(s, func(shortcut fyne.Shortcut) { fn() })
}
