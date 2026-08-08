package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// emptyState renders a friendly empty surface: a short message and an optional
// CTA button (contracts/ui-contract.md §5/§10). Surfaces use it so a loaded,
// empty list never renders as a bare blank panel.
func emptyState(msg string, ctaText string, onCTA func()) fyne.CanvasObject {
	body := []fyne.CanvasObject{widget.NewLabelWithStyle(msg, fyne.TextAlignCenter, fyne.TextStyle{})}
	if ctaText != "" {
		body = append(body, widget.NewButton(ctaText, func() {
			if onCTA != nil {
				onCTA()
			}
		}))
	}
	return container.NewCenter(container.NewVBox(body...))
}

// loadingSurface wraps an existing surface with a non-blocking "Refreshing…"
// indicator so async refreshes never blank the previous content
// (contracts/ui-contract §5, §10; "conservative refresh"). It is a widget so
// the rail/testers can show/hide/replace it like any CanvasObject.
type loadingSurface struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	refreshing *widget.Label
	root       *fyne.Container
}

// loadingState wraps content, optionally showing the refresh indicator.
func loadingState(content fyne.CanvasObject, refreshing bool) fyne.CanvasObject {
	l := &loadingSurface{
		content:    content,
		refreshing: widget.NewLabelWithStyle("Refreshing…", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true}),
	}
	l.ExtendBaseWidget(l)
	if refreshing {
		l.refreshing.Show()
	} else {
		l.refreshing.Hide()
	}
	l.root = container.NewStack(content, container.NewCenter(l.refreshing))
	return l
}

// CreateRenderer renders the wrapped surface plus the refresh overlay.
func (l *loadingSurface) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.root)
}

// setRefreshing toggles the refresh indicator; content never disappears.
func (l *loadingSurface) setRefreshing(b bool) {
	if b {
		l.refreshing.Show()
	} else {
		l.refreshing.Hide()
	}
	l.root.Refresh()
}

// base returns the wrapped surface object.
func (l *loadingSurface) base() fyne.CanvasObject { return l.content }

// asLoadingSurface type-asserts a canvas object to the refresh wrapper.
func asLoadingSurface(obj fyne.CanvasObject) (*loadingSurface, bool) {
	l, ok := obj.(*loadingSurface)
	return l, ok
}

// emptyAware stacks the surface content with a hidden empty-state layer.
// Surfaces call setEmpty(true) from their refresh when there is no data, so
// an empty list never renders as a bare blank panel; the content underneath
// is preserved for the next refresh (contracts/ui-contract §5, §10).
type emptyAware struct {
	base    fyne.CanvasObject
	empty   fyne.CanvasObject
	cta     *widget.Button
	root    *fyne.Container
}

// newEmptyAware wraps content with an optional empty state + CTA button.
func newEmptyAware(content fyne.CanvasObject, msg, ctaText string, onCTA func()) *emptyAware {
	body := []fyne.CanvasObject{
		widget.NewLabelWithStyle(msg, fyne.TextAlignCenter, fyne.TextStyle{}),
	}
	var btn *widget.Button
	if ctaText != "" {
		btn = widget.NewButton(ctaText, func() {
			if onCTA != nil {
				onCTA()
			}
		})
		body = append(body, btn)
	}
	empty := container.NewCenter(container.NewVBox(body...))
	e := &emptyAware{
		base:  content,
		empty: empty,
		cta:   btn,
		root:  container.NewStack(content, empty),
	}
	e.empty.Hide()
	return e
}

// setEmpty toggles the empty overlay; content underneath never disappears.
func (e *emptyAware) setEmpty(b bool) {
	if b {
		e.empty.Show()
	} else {
		e.empty.Hide()
	}
	e.root.Refresh()
}

// content returns the stacked surface root.
func (e *emptyAware) content() fyne.CanvasObject { return e.root }