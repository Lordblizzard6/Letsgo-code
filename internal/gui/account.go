package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/tools"
)

// accountFlyout is the account popup anchored to the rail avatar (US2,
// contracts §3): status dot + provider/model, today's usage summary and
// actions (API keys, Uso, Cerrar sesión). Esc closes it and returns focus.
type accountFlyout struct {
	win fyne.Window

	pop  *widget.PopUp
	root fyne.CanvasObject
	esc  *settingsEsc

	source    usageSource
	sinceFunc func() time.Time

	onKeys    func() // "API keys & proveedores…" → settings
	onUsage   func() // "Uso…" → KPI dashboard
	onSignOut func()
	onClose   func() // always called on hide (focus back to composer)

	// anchor is the widget the popup hangs from (rail avatar in the app;
	// tests may substitute a window object).
	anchor fyne.CanvasObject
}

// newAccountFlyout builds the popup content. win must be the real window;
// anchor defaults to the window content (top-left) in headless tests.
func newAccountFlyout(win fyne.Window) *accountFlyout {
	f := &accountFlyout{
		win:       win,
		source:    costTrackerUsageSource{tracker: tools.GetCostTracker()},
		sinceFunc: startOfToday,
	}
	f.esc = newSettingsEsc(func() { f.hide() })

	dot := canvas.NewCircle(colorGreen)
	name := widget.NewLabelWithStyle("LetsGO", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	status := container.NewHBox(dot, name)

	provider := widget.NewLabel("Proveedor: " + providerNameFor(detectProvider()))
	model := config.AppConfig.Model
	if model == "" {
		model = "(sin configurar)"
	}
	modelLabel := widget.NewLabel("Modelo: " + model)

	var usageText string
	if f.source != nil {
		since := f.sinceFunc()
		h := aggregateUsage(f.source.Stats(since))
		cost := f.source.TotalCost(since)
		if cost > h.cost {
			h.cost = cost
		}
		usageText = fmt.Sprintf("Hoy: $%.2f · %d req", h.cost, h.requests)
	} else {
		usageText = "Hoy: (sin datos)"
	}
	usageRow := widget.NewLabel(usageText)
	usageRow.Importance = widget.MediumImportance

	keys := widget.NewButton("API keys & proveedores…", func() {
		f.hide()
		if f.onKeys != nil {
			f.onKeys()
		}
	})
	usageBtn := widget.NewButton("Uso…", func() {
		f.hide()
		if f.onUsage != nil {
			f.onUsage()
		}
	})
	signOut := widget.NewButton("Cerrar sesión", func() {
		f.hide()
		if f.onSignOut != nil {
			f.onSignOut()
		}
	})
	signOut.Importance = widget.DangerImportance

	content := container.NewVBox(
		status,
		widget.NewSeparator(),
		provider,
		modelLabel,
		usageRow,
		widget.NewSeparator(),
		keys,
		usageBtn,
		signOut,
	)
	f.root = container.NewBorder(
		container.NewHBox(f.esc),
		nil, nil, nil,
		content,
	)
	return f
}

// show opens the popup anchored below the anchor object.
func (f *accountFlyout) show() {
	if f.root == nil || f.win == nil {
		return
	}
	f.pop = widget.NewPopUp(f.root, f.win.Canvas())
	ref := f.anchor
	if ref == nil {
		ref = f.win.Canvas().Content()
	}
	f.pop.ShowAtPosition(ref.Position().Add(fyne.NewPos(0, ref.Size().Height+6)))
	f.win.Canvas().Focus(f.esc)
}

// hide closes the popup and always notifies onClose (focus back to composer).
func (f *accountFlyout) hide() {
	if f.pop != nil {
		f.pop.Hide()
		f.pop = nil
	}
	f.win.Canvas().Unfocus()
	if f.onClose != nil {
		f.onClose()
	}
}

// close is the public close entry used by app wiring.
func (f *accountFlyout) close() { f.hide() }