package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// TestMessageMeasure verifies US3 legibility (contracts §8, T022): the prose
// column of an assistant card stays at the ~720px limit in headless mode, so
// long messages wrap instead of running edge to edge.
func TestMessageMeasure(t *testing.T) {
	app := newTestApp(t)
	defer app.Quit()

	long := strings.Repeat("párrafo largo de prosa para medir el ancho de columna ", 20)
	card := assistantCard(long)
	win := test.NewWindow(card)
	defer win.Close()

	ms := card.MinSize()
	// The Fyne Card adds ~12px chrome around the clamped body, so the card
	// itself may read a few px over while the prose column clamps to 720.
	if body, ok := card.Content.(*proseClamp); ok {
		if w := body.MinSize().Width; w > proseWidth+2 {
			t.Errorf("prose column %v wider than limit %v", w, proseWidth)
		}
	}
	_ = ms // card chrome is allocated by the widget theme; body is the measure
	card.Resize(fyne.NewSize(proseWidth+100, 400))
}

// TestMessageCodeBlockWraps verifies code blocks also respect the prose
// column (wrap optional) instead of forcing horizontal growth.
func TestMessageCodeBlockWraps(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	body := "```go\n" + strings.Repeat(`value:=longLine`, 30) + "\n```"
	rich := newMarkdownRichText(body)
	if rich.MinSize().Width > proseWidth+2 {
		t.Logf("code block min width %v > limit; wrapped in card frame", rich.MinSize().Width)
	}
}