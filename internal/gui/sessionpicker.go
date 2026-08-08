package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/db"
)

// sessionPicker is the startup overlay offered when previous sessions exist
// (FR-014): New / Resume last / Choose, filterable by cwd and date. It is
// keyboard-driven: ↑↓ navigate, Enter picks, Esc closes without action.
type sessionPicker struct {
	win fyne.Window

	entry   *pickerEntry
	list    *widget.List
	items   []db.Session
	current int
	onPick  func(id string)
	onNew   func()
	onClose func()
	root    fyne.CanvasObject
}

// pickerEntry is a single-line filter where ↑/↓ move the highlighted session,
// Enter picks it, and Esc closes the picker (FR-014).
type pickerEntry struct {
	widget.Entry

	up    func()
	down  func()
	enter func()
	esc   func()
}

func (e *pickerEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		if e.up != nil {
			e.up()
		}
		return
	case fyne.KeyDown:
		if e.down != nil {
			e.down()
		}
		return
	case fyne.KeyEnter, fyne.KeyReturn:
		if e.enter != nil {
			e.enter()
		}
		return
	case fyne.KeyEscape:
		if e.esc != nil {
			e.esc()
		}
		return
	}
	e.Entry.TypedKey(key)
}

// newSessionPicker builds the picker overlay (shown by the Controller).
func newSessionPicker(win fyne.Window) *sessionPicker {
	p := &sessionPicker{win: win}

	p.entry = &pickerEntry{}
	p.entry.ExtendBaseWidget(p.entry)
	p.entry.SetPlaceHolder("Filter by name, cwd or date (↑↓ / Enter / Esc)…")
	p.entry.OnChanged = func(string) {
		p.refresh()
	}

	p.list = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(p.items) {
				return
			}
			item.(*widget.Label).SetText(sessionDisplayName(p.items[id]))
		},
	)
	p.list.OnSelected = func(id widget.ListItemID) {
		if int(id) >= 0 && int(id) < len(p.items) {
			p.current = int(id)
		}
	}
	p.entry.up = func() { p.move(-1) }
	p.entry.down = func() { p.move(1) }
	p.entry.enter = p.chooseCurrent
	p.entry.esc = p.close

	p.root = container.NewBorder(p.entry, nil, nil, nil, p.list)
	return p
}

// move shifts the highlighted row by delta, wrapping within bounds.
func (p *sessionPicker) move(delta int) {
	n := len(p.items)
	if n == 0 {
		return
	}
	p.current = (p.current + delta + n) % n
	p.list.Select(p.current)
	p.list.ScrollTo(p.current)
	if delta == 0 {
		p.list.Select(p.current)
	}
}

// chooseCurrent dispatches the highlighted session (FR-014).
func (p *sessionPicker) chooseCurrent() {
	if len(p.items) == 0 {
		return
	}
	if p.current < 0 || p.current >= len(p.items) {
		p.move(0)
		return
	}
	if p.onPick != nil {
		p.onPick(p.items[p.current].ID)
	}
}

// refresh re-syncs the picker list from the session database, honoring the
// filter over name, cwd, date and id.
func (p *sessionPicker) refresh() {
	sessions, err := db.ListSessions()
	if err != nil {
		p.items = nil
		p.list.Refresh()
		return
	}
	query := strings.ToLower(strings.TrimSpace(p.entry.Text))
	p.items = p.items[:0]
	for _, s := range sessions {
		label := strings.ToLower(sessionDisplayName(s))
		nameCwd := strings.ToLower(s.Name + " " + s.ProjectPath)
		if query == "" ||
			strings.Contains(label, query) ||
			strings.Contains(nameCwd, query) ||
			strings.Contains(strings.ToLower(s.ID), query) {
			p.items = append(p.items, s)
		}
	}
	if p.current >= len(p.items) {
		p.current = 0
	}
	p.list.Refresh()
	if len(p.items) > 0 {
		p.list.Select(p.current)
	}
}

// selectID highlights the session with the given ID, if present.
func (p *sessionPicker) selectID(id string) {
	for i, s := range p.items {
		if s.ID == id {
			p.current = i
			p.list.Select(i)
			p.list.ScrollTo(i)
			return
		}
	}
}

// close hides the picker without acting (Esc / Cancel).
func (p *sessionPicker) close() {
	if p.onClose != nil {
		p.onClose()
	}
}

// open shows the picker and focuses the filter.
func (p *sessionPicker) open() {
	p.refresh()
	p.win.Canvas().Focus(p.entry)
}
