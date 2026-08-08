package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// composerEntry is a multi-line input where plain Enter submits (repose) or
// steers the active turn (FR-009), Tab queues the text as the next turn during
// an active turn (FR-010), Ctrl+Enter inserts a newline, and Esc cancels the
// running stream. The input is NEVER disabled during an active turn (FR-008).
type composerEntry struct {
	widget.Entry

	submit  func(text string)
	queue   func(text string)
	cancel  func()
	onSlash func()
	onHole  func()

	// busy reports whether a turn is active; overridable in tests.
	busy func() bool

	// mods reports the current keyboard modifiers; overridable in tests.
	mods func() fyne.KeyModifier

	// onFocus reports focus transitions so the host can restyle the ring.
	onFocus func(focused bool)
}

func (e *composerEntry) FocusGained() {
	e.Entry.FocusGained()
	if e.onFocus != nil {
		e.onFocus(true)
	}
}

func (e *composerEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onFocus != nil {
		e.onFocus(false)
	}
}

func (e *composerEntry) TypedKey(key *fyne.KeyEvent) {
	mods := e.mods
	if mods == nil {
		mods = currentModifiers
	}
	active := e.busy
	if active == nil {
		active = func() bool { return false }
	}
	switch key.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		if mods()&fyne.KeyModifierControl != 0 {
			e.Entry.TypedKey(key)
			return
		}
		if e.submit != nil {
			e.submit(e.Text)
		}
		return
	case fyne.KeyTab:
		// FR-010: Tab with text during an active turn queues the next turn.
		if active() && e.queue != nil {
			e.queue(e.Text)
			return
		}
		e.Entry.TypedKey(key)
		return
	case fyne.KeyEscape:
		if e.cancel != nil {
			e.cancel()
		}
		return
	}
	e.Entry.TypedKey(key)
}

func (e *composerEntry) TypedRune(r rune) {
	if r == '/' && e.Text == "" && e.onSlash != nil {
		e.onSlash()
		return
	}
	if r == '?' && e.Text == "" && e.onHole != nil {
		e.onHole()
		return
	}
	e.Entry.TypedRune(r)
}

// requestFocus moves keyboard focus to the composer input (FR-025).
func (e *composerEntry) requestFocus() {
	e.Entry.FocusGained()
}

// currentModifiers returns the currently active keyboard modifiers from the
// desktop driver (fyne.KeyEvent carries no modifier in this version).
func currentModifiers() fyne.KeyModifier {
	if d, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		return d.CurrentKeyModifiers()
	}
	return 0
}

// composerBox wraps the composer input with a focus ring (T008): a 1px border
// that turns accent when the input has keyboard focus and returns to the
// surface border otherwise. It draws the ring in its own renderer over the
// input rather than relying on the theme InputBorder.
type composerBox struct {
	widget.BaseWidget

	input   *composerEntry
	focused bool

	ring *canvas.Rectangle
	root *fyne.Container
}

// composerFocusColor resolves the 2px accent ring token (US3, T027).
func composerFocusColor() color.Color {
	if th := fyne.CurrentApp().Settings().Theme(); th != nil {
		return th.Color(theme.ColorNameFocus, fyne.CurrentApp().Settings().ThemeVariant())
	}
	return AccentColor
}

func newComposerBox(input *composerEntry) *composerBox {
	b := &composerBox{input: input}
	b.ring = canvas.NewRectangle(colorTransparent)
	b.ring.StrokeColor = b.idleColor()
	b.ring.StrokeWidth = 2
	b.root = container.NewStack(b.ring, input)
	b.ExtendBaseWidget(b)
	return b
}

// idleColor resolves the rest-state ring color through the theme tokens
// (US3, T027: no hardcoded border in the composer shell).
func (b *composerBox) idleColor() color.Color {
	if th := fyne.CurrentApp().Settings().Theme(); th != nil {
		return th.Color(theme.ColorNameInputBorder, fyne.CurrentApp().Settings().ThemeVariant())
	}
	return darkBorder
}

// setFocused flips the ring to the theme focus color (2px accent, US3) and
// back to the input border token when the input loses focus.
func (b *composerBox) setFocused(f bool) {
	if b.focused == f {
		return
	}
	b.focused = f
	var color color.Color
	if f {
		color = composerFocusColor()
	} else {
		color = b.idleColor()
	}
	b.ring.StrokeColor = color
	b.ring.Refresh()
}

func (b *composerBox) FocusGained() {
	b.setFocused(true)
}

func (b *composerBox) FocusLost() {
	b.setFocused(false)
}

func (b *composerBox) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.root)
}
