package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// tooltipText renders the canonical tooltip string for a rail slot or action:
// "Label (Alt+N)" when a shortcut exists, or just the label otherwise
// (contracts/ui-contract.md §8, D4). The tooltip is mandatory in collapsed
// mode so the icon never loses its meaning.
func tooltipText(label, shortcut string) string {
	if shortcut == "" {
		return label
	}
	return label + " (" + shortcut + ")"
}

// tooltipButton is a button that shows a floating tooltip on hover (Fyne
// v2.8.0 offers no built-in tooltip widget). The tip is a small popup shown
// above the button while the pointer stays over it; keyboard focus still works
// normally. canvas may be nil in headless layouts — the tip then stays dormant,
// which keeps tests deterministic.
type tooltipButton struct {
	widget.Button
	tip    string
	canvas fyne.Canvas
	popup  *widget.PopUp
}

// newTooltipButton builds a button with the canonical "Label (Alt+N)" hover tip.
func newTooltipButton(label string, icon fyne.Resource, canvas fyne.Canvas, tipLabel, tipShortcut string, tapped func()) *tooltipButton {
	b := &tooltipButton{
		tip:    tooltipText(tipLabel, tipShortcut),
		canvas: canvas,
	}
	b.ExtendBaseWidget(b)
	// Reuse Button's own OnTapped wiring.
	if icon != nil {
		b.SetIcon(icon)
	}
	b.SetText(label)
	b.OnTapped = tapped
	return b
}

// setTooltipText updates the popup body (icon-only mode keeps the label).
func (b *tooltipButton) setTooltipText(s string) {
	b.tip = s
}

// MouseIn shows the tip popup near the button when a canvas is available.
func (b *tooltipButton) MouseIn(_ *desktop.MouseEvent) {
	if b.canvas == nil || b.tip == "" {
		return
	}
	lbl := widget.NewLabel(b.tip)
	lbl.Importance = widget.MediumImportance
	b.popup = widget.NewPopUp(lbl, b.canvas)
	b.popup.ShowAtPosition(b.Position().Add(fyne.NewPos(0, b.Size().Height+4)))
}

// MouseMoved keeps the popup anchored to the button.
func (b *tooltipButton) MouseMoved(*desktop.MouseEvent) {
	if b.popup != nil {
		b.popup.Move(b.Position().Add(fyne.NewPos(0, b.Size().Height+4)))
	}
}

// MouseOut closes the popup.
func (b *tooltipButton) MouseOut() {
	if b.popup != nil {
		b.popup.Hide()
		b.popup = nil
	}
}

var _ desktop.Hoverable = (*tooltipButton)(nil)