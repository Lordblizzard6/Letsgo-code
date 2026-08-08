package gui

import (
	"fyne.io/fyne/v2"
	"github.com/user/go-claude-code/internal/config"
)

// navigation.go implements the pane navigation model (data-model.md §2,
// contracts/gui-contract.md §1): showPane pushes the previous pane onto the
// navigation history, back() pops it, and the rail accent follows Current.

// current returns the currently visible pane index (0..6+). It is derived
// from the rail's active slot and the viewStack, so it stays in sync even if
// a view is opened through the palette, menu or shortcuts.
func (c *Controller) current() int {
	if c.viewStack == nil {
		return 0
	}
	if c.rail != nil {
		// The rail knows the active pane; fall back to scanning the stack.
		if s, ok := slotByID(c.rail.aid()); ok && s.paneIndex >= 0 {
			return s.paneIndex
		}
	}
	for i, obj := range c.viewStack.Objects {
		if co, ok := obj.(*fyne.Container); ok && !co.Hidden {
			return i
		}
	}
	return 0
}

// overlayOpen reports whether a transient overlay is currently showing
// (palette, keymap, picker). back() must not move behind an open overlay.
func (c *Controller) overlayOpen() bool {
	if c.viewStack == nil {
		return false
	}
	for i, obj := range c.viewStack.Objects {
		if i < 7 {
			continue // panes only; overlays live at index >= 7
		}
		if co, ok := obj.(*fyne.Container); ok && !co.Hidden {
			return true
		}
	}
	return false
}

// showPane shows pane i, pushing the current pane onto the navigation stack
// when i differs, and marks the corresponding rail slot active
// (contracts/gui-contract.md §1.1).
func (c *Controller) showPane(i int) {
	if c.viewStack == nil {
		return
	}
	if i < 0 || i >= len(c.viewStack.Objects) {
		return
	}
	cur := c.current()
	if i != cur && cur < 7 {
		c.navHistory = append(c.navHistory, cur)
	}
	c.showIndex(i)
	if c.rail != nil {
		c.rail.setActive(i)
	}
}

// back pops the navigation stack and shows the previous pane; empty stack or
// chat root is a no-op. Guards: never pop behind an overlay. When returning
// to the chat pane (0) the composer regains focus (SC-009).
func (c *Controller) back() {
	if c.overlayOpen() {
		return
	}
	if c.viewStack == nil {
		return
	}
	cur := c.current()
	if cur == 0 {
		return
	}
	dest := 0
	if n := len(c.navHistory); n > 0 {
		dest = c.navHistory[n-1]
		c.navHistory = c.navHistory[:n-1]
	}
	c.showIndex(dest)
	if c.rail != nil {
		c.rail.setActive(dest)
	}
	if dest == 0 && c.chat != nil {
		c.chat.input.requestFocus()
	}
}

// showChat shows the conversation pane and focuses the composer.
func (c *Controller) showChat() {
	if c.viewStack == nil {
		return
	}
	c.showIndex(0)
	if c.rail != nil {
		c.rail.setActive(0)
	}
	if c.chat != nil {
		c.chat.input.requestFocus()
	}
}

// resetNavigation clears the navigation stack (used on session switches).
func (c *Controller) resetNavigation() {
	c.navHistory = nil
}

// toggleRail flips the rail expanded<->collapsed and persists the preference
// (FR-002, contracts §1.1).
func (c *Controller) toggleRail() {
	if c.rail == nil {
		return
	}
	c.rail.toggle()
	config.AppConfig.Rail.Collapsed = c.rail.collapsed
	_ = config.SaveConfig()
}
