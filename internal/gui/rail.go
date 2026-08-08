package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// railSlot is a single navigation entry of the vertical rail (data-model.md §1).
type railSlot struct {
	id        string
	label     string
	icon      fyne.ThemeIconName // resolved lazily (theme.Icon) at build time
	paneIndex int                // 0..6 pane inside viewStack; -1 for an action slot
	shortcut  string
	zone      string // "tools" | "system" | "work" (rail zones, 004)
}

// railSlots order defines both layout and the Alt+1..8 mapping
// (contracts/ui-contract.md §1.3/§8): work group first, then the tools zone
// ("HERRAMIENTAS"), the system zone pinned to the bottom, account + collapse
// in the rail footer. Icons are resolved at build time — theme lookups at
// package init would run before the Fyne app exists and log "current app not
// started" errors.
var railSlots = []railSlot{
	{id: "chat", label: "Chat", icon: theme.IconNameMailCompose, paneIndex: 0, shortcut: "Alt+1", zone: "work"},
	{id: "git", label: "Git", icon: theme.IconNameViewRefresh, paneIndex: 3, shortcut: "Alt+2", zone: "tools"},
	{id: "tasks", label: "Tareas", icon: theme.IconNameCheckButton, paneIndex: 6, shortcut: "Alt+3", zone: "tools"},
	{id: "mcp", label: "MCP", icon: theme.IconNameComputer, paneIndex: 4, shortcut: "Alt+4", zone: "tools"},
	{id: "plugins", label: "Plugins", icon: theme.IconNameGrid, paneIndex: 5, shortcut: "Alt+5", zone: "tools"},
	{id: "settings", label: "Configuración", icon: theme.IconNameSettings, paneIndex: 1, shortcut: "Alt+6", zone: "system"},
	{id: "usage", label: "Uso", icon: theme.IconNameStorage, paneIndex: 2, shortcut: "Alt+7", zone: "system"},
	{id: "help", label: "Ayuda", icon: theme.IconNameHelp, paneIndex: -1, shortcut: "Alt+8", zone: "system"},
	{id: "theme", label: "Tema", icon: theme.IconNameColorPalette, paneIndex: -1, shortcut: "", zone: "system"},
	{id: "account", label: "Cuenta", icon: theme.IconNameAccount, paneIndex: -1, shortcut: "", zone: "system"},
}

// slotForPane returns the single slot whose paneIndex equals i. The catalog
// guarantees one destination per pane (no duplicates: sessions was removed in
// 004, data-model.md §1), so this is a pure lookup.
func slotForPane(i int) (railSlot, bool) {
	for _, s := range railSlots {
		if s.paneIndex == i {
			return s, true
		}
	}
	return railSlot{}, false
}

func slotByID(id string) (railSlot, bool) {
	for _, s := range railSlots {
		if s.id == id {
			return s, true
		}
	}
	return railSlot{}, false
}

// railView is the autonomous vertical navigation bar (FR-001, T003). It does
// not depend on viewStack: the Controller wires onSelect/onToggle/onAction so
// the widget stays testable headless.
type railView struct {
	onSelect func(paneIndex int)
	onToggle func()
	onAction func(id string)

	activeID  string
	collapsed bool
	canvas    fyne.Canvas // for tooltips; nil in headless layouts

	buttons map[string]*widget.Button
	tips    map[string]*tooltipButton
	marks   map[string]*canvas.Rectangle
	avatar  *railAvatar
	root    fyne.CanvasObject
}

// minWidth wraps a canvas object and reports a minimum width in px.
type minWidth struct {
	fyne.CanvasObject
	w float32
}

func (m *minWidth) MinSize() fyne.Size {
	s := m.CanvasObject.MinSize()
	s.Width = m.w
	return s
}

// newRailView builds the rail; collapsed forces icon-only mode (≥44px).
func newRailView(collapsed bool) *railView {
	r := &railView{
		collapsed: collapsed,
		buttons:   map[string]*widget.Button{},
		tips:      map[string]*tooltipButton{},
		marks:     map[string]*canvas.Rectangle{},
	}
	r.rebuild()
	return r
}

// setCanvas attaches the window canvas so tooltip popups can anchor (rail is
// built before SetContent; 004 tooltips need the canvas at hover time). It is
// safe to call with nil in headless layouts (toolbars stay dormant).
func (r *railView) setCanvas(c fyne.Canvas) {
	r.canvas = c
	for _, tb := range r.tips {
		tb.canvas = c
	}
}

// content exposes the top-level canvas object.
func (r *railView) content() fyne.CanvasObject { return r.root }

// rebuild (re)creates slot buttons for the current expanded/collapsed state.
// 004 (US1): the rail is split into three zones — work (chat), tools
// ("HERRAMIENTAS" label), system (settings/usage/help/theme, pinned at the
// bottom next to collapse + avatar).
func (r *railView) rebuild() {
	work := container.NewVBox()
	tools := container.NewVBox()
	system := container.NewVBox()
	// Zone header for the middle zone; hidden in collapsed mode.
	toolsHeader := widget.NewLabelWithStyle("HERRAMIENTAS",
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	toolsHeader.Importance = widget.MediumImportance
	if r.collapsed {
		toolsHeader.Hide()
	}
	for _, s := range railSlots {
		slot := s
		label := slot.label
		if r.collapsed {
			label = ""
		}
		tb := newTooltipButton(label, iconFor(slot.id), r.canvas,
			slot.label, slot.shortcut, func() { r.dispatch(slot.id) })
		mark := canvas.NewRectangle(AccentColor)
		mark.Hide()
		r.buttons[slot.id] = &tb.Button
		r.tips[slot.id] = tb
		r.marks[slot.id] = mark
		row := container.NewBorder(nil, nil, &minWidth{CanvasObject: mark, w: 3}, nil, tb)
		switch slot.zone {
		case "work":
			work.Add(row)
		case "tools":
			tools.Add(row)
		default:
			system.Add(row)
		}
	}
	// Account is pinned to the bottom edge (FR-005) with a "GO" avatar and a
	// live status dot (T006); a collapse toggle sits above it.
	footer := r.footer()
	base := container.NewVBox(
		work,
		widget.NewSeparator(),
		toolsHeader, tools,
		widget.NewSeparator(),
		system, widget.NewSeparator(),
		footer,
	)
	r.root = &minWidth{CanvasObject: base, w: r.width()}
	r.applyActive()
}

// footer builds the pinned bottom block: rail collapse toggle, then the
// account avatar ("GO" monogram on an accent circle, status dot on top-right).
func (r *railView) footer() fyne.CanvasObject {
	tipLabel, tipShortcut := "Colapsar rail", ""
	if r.collapsed {
		tipLabel = "Expandir rail"
	}
	collapseBtn := newTooltipButton("", theme.MenuExpandIcon(), r.canvas, tipLabel, tipShortcut, func() {
		if r.onToggle != nil {
			r.onToggle()
		}
	})
	collapseBtn.Importance = widget.LowImportance

	r.avatar = newRailAvatar(func() {
		if r.onAction != nil {
			r.onAction("account")
		}
	})
	row := container.NewHBox(collapseBtn, r.avatar.content())
	return row
}

// setAvatarStatus updates the account status dot (online/idle/error derived
// from hasAnyKey/engine by the Controller; T006).
func (r *railView) setAvatarStatus(c color.NRGBA) {
	if r.avatar != nil {
		r.avatar.setStatus(c)
	}
}

// railAvatar is a tappable widget: accent-stroked circle with the "GO"
// monogram and a status dot anchored to the top-right corner. It draws itself
// so no container layout is needed for the overlap.
type railAvatar struct {
	widget.BaseWidget
	dotColor color.NRGBA
	tap      func()
}

func newRailAvatar(tap func()) *railAvatar {
	a := &railAvatar{
		dotColor: colorGreen, // online default
		tap:      tap,
	}
	a.ExtendBaseWidget(a)
	return a
}

func (a *railAvatar) content() fyne.CanvasObject { return a }

func (a *railAvatar) setStatus(c color.NRGBA) {
	if c == a.dotColor {
		return
	}
	a.dotColor = c
	a.Refresh()
}

func (a *railAvatar) MinSize() fyne.Size {
	return fyne.NewSize(40, 40)
}

func (a *railAvatar) CreateRenderer() fyne.WidgetRenderer {
	circle := canvas.NewCircle(railSurfaceColor())
	circle.StrokeColor = railAccentColor()
	circle.StrokeWidth = 2
	label := canvas.NewText("GO", railAccentColor())
	label.TextSize = TextSizeLarge
	label.TextStyle = fyne.TextStyle{Bold: true}
	dot := canvas.NewCircle(colorGreen)

	r := &railAvatarRenderer{a: a, circle: circle, label: label, dot: dot}
	r.objects = []fyne.CanvasObject{circle, label, dot}
	r.apply()
	return r
}

func (a *railAvatar) Tapped(_ *fyne.PointEvent) {
	if a.tap != nil {
		a.tap()
	}
}

var _ fyne.Tappable = (*railAvatar)(nil)

type railAvatarRenderer struct {
	a         *railAvatar
	circle    *canvas.Circle
	label     *canvas.Text
	dot       *canvas.Circle
	objects   []fyne.CanvasObject
	lastBound fyne.Size
}

func (r *railAvatarRenderer) apply() {
	r.label.Color = railAccentColor()
	r.dot.FillColor = r.a.dotColor
	r.Refresh()
}

func (r *railAvatarRenderer) Layout(size fyne.Size) {
	circle, label, dot := r.circle, r.label, r.dot
	diam := fyne.Min(size.Width, size.Height) - 6
	cw := fyne.NewSize(diam, diam)
	circle.Resize(cw)
	circle.Move(fyne.NewPos((size.Width-cw.Width)/2, (size.Height-cw.Height)/2))
	label.Move(fyne.NewPos(cw.Width/2-label.MinSize().Width/2, cw.Height/2-label.MinSize().Height/2))
	dot.Resize(fyne.NewSize(12, 12))
	dot.Move(fyne.NewPos(circle.Position().X+cw.Width-8, circle.Position().Y-2))
}

func (r *railAvatarRenderer) MinSize() fyne.Size { return fyne.NewSize(40, 40) }

func (r *railAvatarRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *railAvatarRenderer) Refresh() {
	r.label.Color = railAccentColor()
	r.dot.FillColor = r.a.dotColor
	if r.a.Size() != r.lastBound {
		r.Layout(r.a.Size())
		r.lastBound = r.a.Size()
	}
}

func (r *railAvatarRenderer) Destroy() {}

// width forces the rail width: ≥44 collapsed, ~200 expanded (SC-008).
func (r *railView) width() float32 {
	if r.collapsed {
		return 48
	}
	return 200
}

// dispatch routes a slot tap.
func (r *railView) dispatch(id string) {
	s, ok := slotByID(id)
	if !ok {
		return
	}
	if s.paneIndex >= 0 {
		if r.onSelect != nil {
			r.onSelect(s.paneIndex)
		}
		return
	}
	if r.onAction != nil {
		r.onAction(id)
	}
}

// setActive moves the accent mark to the slot for pane i (contracts §1.3).
func (r *railView) setActive(paneIndex int) {
	if s, ok := slotForPane(paneIndex); ok {
		r.setActiveID(s.id)
		return
	}
	r.setActiveID("")
}

// setActiveID highlights the given slot, clearing all others.
func (r *railView) setActiveID(id string) {
	r.activeID = id
	for bid, mark := range r.marks {
		if bid == id {
			mark.Show()
		} else {
			mark.Hide()
		}
	}
}

// aid returns the currently active slot id.
func (r *railView) aid() string { return r.activeID }

// toggle flips collapsed<->expanded and rebuilds (FR-002).
func (r *railView) toggle() {
	r.collapsed = !r.collapsed
	r.rebuild()
}

// setExpanded forces the expanded/collapsed state without flipping.
func (r *railView) setExpanded(expanded bool) {
	if r.collapsed == !expanded {
		return
	}
	r.collapsed = !expanded
	r.rebuild()
}

func (r *railView) applyActive() {
	if r.activeID != "" {
		r.setActiveID(r.activeID)
	}
}
