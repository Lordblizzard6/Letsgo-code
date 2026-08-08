package gui

import (
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// TestTooltipTextFormat verifies the canonical "Label (Alt+N)" formatting.
func TestTooltipTextFormat(t *testing.T) {
	if got := tooltipText("Git", "Alt+2"); got != "Git (Alt+2)" {
		t.Fatalf("tooltipText(Git, Alt+2) = %q", got)
	}
	if got := tooltipText("Theme", ""); got != "Theme" {
		t.Fatalf("tooltipText(Theme, ) = %q, want bare label", got)
	}
}

// TestTooltipButtonHover verifies the tooltip-button shows a popup on hover and
// hides it on mouse-out, and that tapping still dispatches.
func TestTooltipButtonHover(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	win := test.NewWindow(container.NewStack())
	defer win.Close()

	tapped := 0
	b := newTooltipButton("Git", nil, win.Canvas(), "Git", "Alt+2", func() { tapped++ })
	if b.tip != "Git (Alt+2)" {
		t.Fatalf("tooltipButton tip = %q, want Git (Alt+2)", b.tip)
	}

	b.OnTapped()
	if tapped != 1 {
		t.Fatal("OnTapped did not fire")
	}

	b.MouseIn(nil)
	if b.popup == nil {
		t.Fatal("MouseIn with a canvas must open a popup")
	}
	b.MouseOut()
	if b.popup != nil {
		t.Fatal("MouseOut must close the popup")
	}
}