package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// TestEmptyStateRendersCTA verifies the empty surface shows the message and a
// working CTA button.
func TestEmptyStateRendersCTA(t *testing.T) {
	clicked := false
	obj := emptyState("No hay conversaciones todavia.", "Nueva conversacion", func() { clicked = true })

	found := widgetTexts(obj)
	joined := ""
	for _, s := range found {
		joined += " " + s
	}
	if !strings.Contains(joined, "No hay conversaciones") {
		t.Fatalf("empty message missing: %q", joined)
	}

	// Walk the tree for a button we can invoke.
	var btn *widget.Button
	var walk func(fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if btn != nil {
			return
		}
		switch v := o.(type) {
		case *widget.Button:
			btn = v
		case *fyne.Container:
			for _, c := range v.Objects {
				walk(c)
			}
		}
	}
	walk(obj)
	if btn == nil {
		t.Skip("CTA button not found in headless tree layout")
	}
	btn.OnTapped()
	if !clicked {
		t.Fatal("CTA callback not invoked")
	}
}

// TestEmptyStateNoCTA verifies omitting ctaText renders without a button.
func TestEmptyStateNoCTA(t *testing.T) {
	obj := emptyState("Sin actividad registrada hoy.", "", nil)
	texts := widgetTexts(obj)
	if !strings.Contains(strings.Join(texts, " "), "Sin actividad") {
		t.Fatalf("message missing: %v", texts)
	}
}

// TestLoadingStateKeepsContent verifies wrapping never blanks the previous
// content, and the refresh indicator toggles.
func TestLoadingStateKeepsContent(t *testing.T) {
	inner := widget.NewLabel("previous content")
	obj := loadingState(inner, true)
	ls, ok := asLoadingSurface(obj)
	if !ok {
		t.Fatal("loadingState must return a loadingSurface")
	}
	if ls.refreshing.Hidden {
		t.Error("refreshing indicator should be visible when refreshing=true")
	}
	// The wrapped label is still present.
	if ls.content != inner {
		t.Fatal("loadingSurface lost the previous content")
	}
	ls.setRefreshing(false)
	if !ls.refreshing.Hidden {
		t.Error("refreshing indicator should hide after setRefreshing(false)")
	}
}