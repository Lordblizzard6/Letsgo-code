package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/db"
)

// forkChooser is an inline card listing the message history of the active
// session so the user can pick a fork point (FR-015). Navigating with ↑/↓
// and pressing Enter forks there; Esc closes without forking.
type forkChooser struct {
	widget.BaseWidget

	sessionID string

	list    *widget.List
	items   []db.Message
	current int
	onFork  func(sessionID string, messageID int64)
	onClose func()

	entry *widget.Entry
	root  fyne.CanvasObject
}

// newForkChooser builds the fork-point picker for the given session.
func newForkChooser(sessionID string) *forkChooser {
	f := &forkChooser{sessionID: sessionID}

	title := widget.NewLabelWithStyle(
		"Fork session at message",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	title.Importance = widget.HighImportance

	hint := widget.NewLabel(
		"↑/↓ pick a message · Enter fork · Esc cancel",
	)
	hint.Importance = widget.MediumImportance
	hint.TextStyle = fyne.TextStyle{Monospace: true}

	closeBtn := widget.NewButton("Cancel", func() {
		if f.onClose != nil {
			f.onClose()
		}
	})

	f.list = widget.NewList(
		func() int { return len(f.items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(f.items) {
				return
			}
			item.(*widget.Label).SetText(forkMessageLabel(f.items[id]))
		},
	)
	f.list.OnSelected = func(id widget.ListItemID) {
		if int(id) >= 0 && int(id) < len(f.items) {
			f.current = int(id)
		}
	}

	f.entry = widget.NewEntry()
	f.entry.SetPlaceHolder("")

	top := container.NewHBox(title, closeBtn)
	f.root = container.NewBorder(
		container.NewVBox(hint, top),
		container.NewHBox(
			widget.NewButton("Fork latest", f.forkCurrent),
			f.entry,
		),
		nil, nil,
		f.list,
	)
	f.ExtendBaseWidget(f)
	return f
}

// load fetches the message history of the session.
func (f *forkChooser) load() {
	hist, err := db.GetHistory(f.sessionID)
	if err != nil {
		f.items = nil
	} else {
		f.items = hist
	}
	if len(f.items) == 0 {
		f.items = []db.Message{}
	}
	f.current = 0
	f.list.Refresh()
	if len(f.items) > 0 {
		f.list.Select(0)
	}
}

// move shifts the highlighted message by delta, wrapping within bounds.
func (f *forkChooser) move(delta int) {
	n := len(f.items)
	if n == 0 {
		return
	}
	f.current = (f.current + delta + n) % n
	f.list.Select(f.current)
	f.list.ScrollTo(f.current)
}

// forkCurrent dispatches the chosen message as the fork point.
func (f *forkChooser) forkCurrent() {
	if len(f.items) == 0 {
		if f.onClose != nil {
			f.onClose()
		}
		return
	}
	f.forkAt(f.items[f.current])
}

// forkAt dispatches a specific message as the fork point.
func (f *forkChooser) forkAt(m db.Message) {
	if f.onFork != nil && m.ID > 0 {
		f.onFork(f.sessionID, m.ID)
	}
}

// requestFocus focuses the internal entry so TypedKey receives ↑/↓/Enter/Esc.
func (f *forkChooser) requestFocus(win fyne.Window) {
	f.load()
	win.Canvas().Focus(f.entry)
}

// TypedKey drives the chooser with the keyboard.
func (f *forkChooser) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		f.move(-1)
	case fyne.KeyDown:
		f.move(1)
	case fyne.KeyEnter, fyne.KeyReturn:
		f.forkCurrent()
	case fyne.KeyEscape:
		if f.onClose != nil {
			f.onClose()
		}
	}
}

func (f *forkChooser) TypedRune(r rune) {}

// CreateRenderer implements fyne.Widget.
func (f *forkChooser) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(f.root)
}

// forkMessageLabel renders a readable one-line summary of a history message.
func forkMessageLabel(m db.Message) string {
	role := strings.ToUpper(m.Role)
	text := ""
	switch v := m.Content.(type) {
	case string:
		text = v
	default:
		text = fmt.Sprintf("%v", v)
	}
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	if len(text) > 90 {
		text = text[:90] + "…"
	}
	ts := m.Timestamp.Format("01-02 15:04")
	return fmt.Sprintf("%s %s — %s", role, ts, text)
}
