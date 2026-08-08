package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// keymapEntry is a single-line filter where Esc closes the cheat sheet
// (FR-019). Filtering happens via the OnChanged handler.
type keymapEntry struct {
	widget.Entry

	esc func()
}

func (e *keymapEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape && e.esc != nil {
		e.esc()
		return
	}
	e.Entry.TypedKey(key)
}

// keymapAction is one shortcut row of the cheat sheet.
type keymapAction struct {
	keys     string // e.g. "Ctrl+K"
	category string // e.g. "Navigation"
	what     string // e.g. "Open command palette"
}

// keymapOverlay is the '?' cheat sheet (FR-019): categorized, filterable,
// closed with Esc. It is keyboard-only like the palette.
type keymapOverlay struct {
	win fyne.Window

	entry *keymapEntry
	list  *widget.List
	all   []keymapAction
	items []string
	root  fyne.CanvasObject

	onClose func()
}

// keymapCatalog is the full set of shortcuts shown by '?'.
var keymapCatalog = []keymapAction{
	{"Ctrl+K", "Navigation", "Open command palette"},
	{"?", "Navigation", "Open this cheat sheet"},
	{"Ctrl+L", "Navigation", "Focus the composer"},
	{"Ctrl+N", "Sessions", "New session"},
	{"Ctrl+1..9", "Sessions", "Switch to session 1..9"},
	{"Alt+1", "Rail", "Chat"},
	{"Alt+2", "Rail", "Git"},
	{"Alt+3", "Rail", "Tareas"},
	{"Alt+4", "Rail", "MCP"},
	{"Alt+5", "Rail", "Plugins"},
	{"Alt+6", "Rail", "Configuración"},
	{"Alt+7", "Rail", "Uso"},
	{"Alt+8", "Rail", "Ayuda (cheat sheet)"},
	{"Alt+Left", "Rail", "Back to previous pane (or chat)"},
	{"Ctrl+,", "Views", "Open settings"},
	{"Ctrl+U", "Views", "Open usage"},
	{"Ctrl+Shift+A", "Views", "Toggle auto-approve"},
	{"Enter", "Composer", "Send (idle) or steer (active turn)"},
	{"Tab", "Composer", "Queue text as next turn (active)"},
	{"Ctrl+Enter", "Composer", "Newline"},
	{"Esc", "Composer", "Interrupt the active stream"},
	{"@", "Composer", "File picker"},
	{"!", "Composer", "Shell passthrough"},
	{"/", "Composer", "Open command palette"},
	{"Enter", "Approval card", "Approve the pending tool"},
	{"Esc", "Approval card", "Reject the pending tool"},
	{"Ctrl+Enter", "Approval card", "Approve and allow for this session"},
	{"Enter", "Plan card", "Approve the plan"},
	{"Esc", "Plan card", "Reject the plan"},
	{"↑/↓", "Palette / picker", "Navigate items"},
	{"Esc", "Palette / picker / overlay", "Close without action"},
}

// newKeymapOverlay builds the cheat sheet overlay (hidden until open).
func newKeymapOverlay(win fyne.Window) *keymapOverlay {
	k := &keymapOverlay{win: win, all: keymapCatalog}

	k.entry = &keymapEntry{}
	k.entry.ExtendBaseWidget(k.entry)
	k.entry.SetPlaceHolder("Filter shortcuts…")
	k.entry.OnChanged = func(string) { k.refresh() }
	k.entry.esc = func() {
		if k.onClose != nil {
			k.onClose()
		}
	}

	k.list = widget.NewList(
		func() int { return len(k.items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(k.items) {
				return
			}
			item.(*widget.Label).SetText(k.items[id])
		},
	)

	k.root = container.NewBorder(k.entry, nil, nil, nil, k.list)
	return k
}

// refresh re-filters the cheat sheet rows by the current query.
func (k *keymapOverlay) refresh() {
	q := strings.ToLower(strings.TrimSpace(k.entry.Text))
	k.items = k.items[:0]
	for _, a := range k.all {
		if q == "" ||
			strings.Contains(strings.ToLower(a.keys), q) ||
			strings.Contains(strings.ToLower(a.category), q) ||
			strings.Contains(strings.ToLower(a.what), q) {
			k.items = append(k.items, formatKeymapRow(a))
		}
	}
	k.list.Refresh()
}

// formatKeymapRow renders "keys — what (category)".
func formatKeymapRow(a keymapAction) string {
	return a.keys + " — " + a.what + " (" + a.category + ")"
}

// open shows the cheat sheet and focuses the filter.
func (k *keymapOverlay) open() {
	k.refresh()
	k.win.Canvas().Focus(k.entry)
}

// close hides the cheat sheet without action.
func (k *keymapOverlay) close() {
	if k.onClose != nil {
		k.onClose()
	}
}
