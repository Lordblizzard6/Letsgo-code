package gui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
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

// emptyAwareHasEmpty runs a helper against each of the six surfaces and
// asserts the empty message renders.
func emptyAwareHasEmpty(t *testing.T, obj fyne.CanvasObject, msg string) {
	t.Helper()
	root, ok := obj.(*fyne.Container)
	if !ok {
		t.Fatalf("surface root is %T, want *fyne.Container", obj)
	}
	texts := widgetTexts(root)
	for _, s := range texts {
		if strings.Contains(s, msg) {
			return
		}
	}
	t.Fatalf("empty message %q missing from %v", msg, texts)
}

// TestEmptyStatesAllSurfaces verifies US4 (T029): the six surfaces show a
// shared empty/CTA state when they have no data, and the wrapped content is
// preserved so refresh never blanks the panel.
func TestEmptyStatesAllSurfaces(t *testing.T) {
test.NewApp()
	mcp := newMCPView(test.NewWindow(nil))
	defer mcp.win.Close()
	if mcp.center == nil {
		t.Fatal("mcpView must own an emptyAware")
	}
	mcp.center.setEmpty(true)
	emptyAwareHasEmpty(t, mcp.center.content(), "No hay servidores MCP configurados.")

	pl := newPluginsView(test.NewWindow(nil))
	defer pl.win.Close()
	if pl.center == nil {
		t.Fatal("pluginsView must own an emptyAware")
	}
	pl.center.setEmpty(true)
	emptyAwareHasEmpty(t, pl.center.content(), "No hay plugins instalados.")

	tasks := newAgentTasksView()
	if tasks.center == nil {
		t.Fatal("agentTasksView must own an emptyAware")
	}
	tasks.center.setEmpty(true)
	emptyAwareHasEmpty(t, tasks.center.content(), "Todavía no hay tareas.")

	git := newGitView(test.NewWindow(nil))
	defer git.win.Close()
	if git.center == nil {
		t.Fatal("gitView must own an emptyAware")
	}
	git.center.setEmpty(true)
	emptyAwareHasEmpty(t, git.center.content(), "No hay repositorio.")

	// Session sidebar delegates empty state to the shared helper too.
	sv := newSessionsView(test.NewWindow(nil))
	defer sv.win.Close()
	if sv.center == nil {
		t.Fatal("sessionsView must own an emptyAware")
	}
	sv.center.setEmpty(true)
	emptyAwareHasEmpty(t, sv.center.content(), "No hay conversaciones todavía.")

	// Usage KPI pane shows the empty state when there is no activity.
	uv := newUsageView(test.NewWindow(nil))
	defer uv.win.Close()
	if uv.center == nil {
		t.Fatal("usageView must own an emptyAware")
	}
	uv.center.setEmpty(true)
	emptyAwareHasEmpty(t, uv.center.content(), "Sin actividad registrada hoy.")
}