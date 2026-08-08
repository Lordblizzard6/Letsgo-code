package gui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/config"
)

// autoApproveCategories are the categories the status bar can toggle.
var autoApproveCategories = []string{"file-edit", "bash", "web"}

// WorkStatus is the live state of the active turn, derived from engine events
// (FR-005/006/020). It lives in memory only (data-model.md §3).
type WorkStatus struct {
	Active        bool          // turno en curso
	StreamingText bool          // StreamDelta activo → ocultar franja
	Phase         string        // "working" | "tool" | "agent"
	CurrentStep   string        // herramienta/agente actual
	Elapsed       time.Duration // tiempo transcurrido del turno
	ContextStale  bool          // >15 min sin actualizar (FR-020)
	RateStale     bool
}

// statusBar shows provider/model, usage, the auto-approve toggle and the work
// status strip (contracts/gui-contract.md). Ctrl+Shift+A toggles in app.go.
type statusBar struct {
	providerLabel *widget.Label
	modelLabel    *widget.Label
	usageLabel    *widget.Label
	autoBtn       *widget.Button

	// Configurable status line pieces (FR-018): model, branch, mode, context,
	// rate, version — each can be toggled in Settings and is hidden off here.
	branchLabel  *coloredLabel
	rateLabel    *widget.Label
	versionLbl   *widget.Label
	statusPieces *fyne.Container

	// Work status strip (FR-005): spinner, elapsed, current step.
	strip    *fyne.Container
	stripLbl *coloredLabel
	spinner  *widget.ProgressBarInfinite
	start    time.Time
	status   WorkStatus

	// Plan/Execute/Auto mode indicator (FR-012).
	modeLbl *coloredLabel

	onToggleApprove func(category string)
	root            *fyne.Container
}

// coloredLabel is a monospace text label whose color can be switched to the
// LetsGO accent (FR-024). The Text field mirrors widget.Label so tests keep
// reading .Text; the actual visual uses a themed RichText segment.
type coloredLabel struct {
	Text string

	rt       *widget.RichText
	accented bool
	hidden   bool
}

func newColoredLabel(text string) *coloredLabel {
	l := &coloredLabel{Text: text}
	l.rt = widget.NewRichText(&widget.TextSegment{
		Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}},
		Text:  text,
	})
	l.apply()
	return l
}

// SetText updates the label text and refreshes.
func (l *coloredLabel) SetText(text string) {
	l.Text = text
	l.apply()
}

// SetAccent toggles the LetsGO accent color on the label.
func (l *coloredLabel) SetAccent(on bool) {
	if l.accented == on {
		return
	}
	l.accented = on
	l.apply()
}

func (l *coloredLabel) apply() {
	style := widget.RichTextStyle{TextStyle: fyne.TextStyle{Monospace: true}}
	if l.accented {
		style.ColorName = accentThemeColor
	}
	l.rt.Segments = []widget.RichTextSegment{
		&widget.TextSegment{Style: style, Text: l.Text},
	}
	if l.rt != nil {
		l.rt.Refresh()
	}
}

func (l *coloredLabel) Importance(_ widget.Importance) {}

// Refresh repaints the underlying RichText.
func (l *coloredLabel) Refresh() { l.rt.Refresh() }

// Move forwards to the underlying RichText.
func (l *coloredLabel) Move(pos fyne.Position) { l.rt.Move(pos) }

// Position forwards to the underlying RichText.
func (l *coloredLabel) Position() fyne.Position { return l.rt.Position() }

// Size forwards to the underlying RichText.
func (l *coloredLabel) Size() fyne.Size { return l.rt.Size() }

// MinSize forwards to the underlying RichText.
func (l *coloredLabel) MinSize() fyne.Size { return l.rt.MinSize() }

// Resize forwards to the underlying RichText.
func (l *coloredLabel) Resize(size fyne.Size) { l.rt.Resize(size) }

// Hide forwards to the underlying RichText.
func (l *coloredLabel) Hide() {
	l.hidden = true
	l.rt.Hide()
}

// Show forwards to the underlying RichText.
func (l *coloredLabel) Show() {
	l.hidden = false
	l.rt.Show()
}

// Hidden reports whether the label is hidden.
func (l *coloredLabel) Hidden() bool { return l.hidden }

// Visible reports whether the underlying RichText is visible.
func (l *coloredLabel) Visible() bool { return l.rt.Visible() }

// newStatusBar builds the status bar from the current config.
func newStatusBar() *statusBar {
	s := &statusBar{}
	cfg := config.AppConfig
	provider := config.DetectProviderFromModel(cfg.Model)

	s.providerLabel = widget.NewLabelWithStyle(provider, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	s.modelLabel = widget.NewLabelWithStyle(cfg.Model, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	s.usageLabel = widget.NewLabelWithStyle("usage: —", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	s.autoBtn = widget.NewButtonWithIcon("", theme.ConfirmIcon(), s.toggle)
	s.autoBtn.Importance = widget.MediumImportance

	s.spinner = widget.NewProgressBarInfinite()
	s.spinner.Resize(fyne.NewSize(16, 16))
	s.stripLbl = newColoredLabel("")
	s.strip = container.NewHBox(s.spinner, s.stripLbl.rt)
	s.setStripVisible(false)

	s.modeLbl = newColoredLabel("mode: execute")

	// FR-018: configurable status line pieces.
	s.branchLabel = newColoredLabel("")
	s.refreshBranch()
	s.rateLabel = widget.NewLabelWithStyle("rate: —", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	s.versionLbl = widget.NewLabelWithStyle(appVersion, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	leftPieces := container.NewHBox(s.modeLbl, s.branchLabel, s.modelLabel, s.usageLabel, s.rateLabel, s.versionLbl)
	s.statusPieces = leftPieces
	s.applyFields()

	s.root = container.NewBorder(nil, nil, nil,
		container.NewVBox(
			s.strip,
			container.NewHBox(leftPieces, s.autoBtn),
		),
		container.NewHBox(s.providerLabel),
	)
	s.refreshAuto()
	return s
}

// refreshBranch refreshes the current git branch label for the status line.
func (s *statusBar) refreshBranch() {
	s.branchLabel.SetText("branch: " + gitBranch())
	if appVersion == "" {
		s.versionLbl.SetText("")
	}
}

// applyFields shows/hides status line pieces following statusline.fields
// (FR-018). The mode indicator stays visible regardless (SC-010).
func (s *statusBar) applyFields() {
	cfg := config.AppConfig
	fields := map[string]bool{}
	for _, f := range cfg.Statusline.Fields {
		fields[f] = true
	}
	show := cfg.Statusline.Show
	if !show && len(fields) == 0 {
		// Defaults: mode + model + context when nothing configured yet.
		fields["mode"] = true
		fields["model"] = true
		fields["context"] = true
		show = true
	}

	// Each piece is hidden unless enabled. Mode stays visible (SC-010) by
	// design when show is on; pieces honor the exact field list.
	update := func(obj fyne.CanvasObject, field string) {
		obj.Hide()
		if show && fields[field] {
			obj.Show()
		}
	}
	update(s.modeLbl, "mode")
	update(s.branchLabel, "branch")
	update(s.modelLabel, "model")
	update(s.usageLabel, "context")
	update(s.rateLabel, "rate")
	update(s.versionLbl, "version")
	s.statusPieces.Refresh()
}

// SetMode updates the Plan/Execute/Auto indicator (FR-012).
func (s *statusBar) SetMode(mode string) {
	if mode == "" {
		mode = "execute"
	}
	s.modeLbl.SetText("mode: " + mode)
	if mode == "plan" {
		s.modeLbl.SetAccent(true)
	} else {
		s.modeLbl.SetAccent(false)
	}
	s.modeLbl.Refresh()
}

// toggle flips every category (the onToggleApprove handler inverts each).
func (s *statusBar) toggle() {
	s.setAllApprovals()
}

func (s *statusBar) setAllApprovals() {
	if s.onToggleApprove != nil {
		for _, cat := range autoApproveCategories {
			s.onToggleApprove(cat)
		}
	}
	s.refreshAuto()
}

// refreshAuto updates the button label to reflect current auto-approve state.
func (s *statusBar) refreshAuto() {
	cfg := config.AppConfig
	on := cfg.AutoApprove["file-edit"] && cfg.AutoApprove["bash"] && cfg.AutoApprove["web"]
	if on {
		s.autoBtn.SetText("auto ON")
		s.autoBtn.Importance = widget.HighImportance
	} else {
		s.autoBtn.SetText("auto OFF")
		s.autoBtn.Importance = widget.MediumImportance
	}
}

// SetUsage updates the usage summary label.
func (s *statusBar) SetUsage(in, out int) {
	s.usageLabel.SetText(fmt.Sprintf("usage: %d→%d tok", in, out))
}

// refreshModel updates provider/model labels after a config change.
func (s *statusBar) refreshModel() {
	cfg := config.AppConfig
	provider := config.DetectProviderFromModel(cfg.Model)
	s.providerLabel.SetText(provider)
	s.modelLabel.SetText(cfg.Model)
}

// --- Work status strip (FR-005/006/020) ---

// setStripVisible toggles the working strip.
func (s *statusBar) setStripVisible(vis bool) {
	s.strip.Hidden = !vis
	if vis {
		s.spinner.Start()
	} else {
		s.spinner.Stop()
	}
	s.strip.Refresh()
}

// SetWork updates the work strip from a WorkStatus snapshot (FR-005).
// During text streaming the strip hides; between tool bursts it reappears
// (FR-006). Context/rate stale marks the strip label (FR-020).
func (s *statusBar) SetWork(w WorkStatus) {
	s.status = w
	if !w.Active {
		s.setStripVisible(false)
		return
	}
	if w.StreamingText {
		s.setStripVisible(false)
		return
	}
	s.setStripVisible(true)

	step := w.CurrentStep
	if step == "" {
		step = w.Phase
	}
	if step == "" {
		step = "working"
	}
	stale := ""
	if w.ContextStale || w.RateStale {
		stale = " · desactualizado"
	}
	s.stripLbl.SetText(fmt.Sprintf("Working ● %s · %s · esc to interrupt%s", step, fmtDuration(w.Elapsed), stale))
	// FR-024: the Working strip carries the LetsGO accent while active.
	s.stripLbl.SetAccent(true)
	s.stripLbl.Refresh()
}

// BeginWork starts a turn timer and shows the strip.
func (s *statusBar) BeginWork() {
	s.start = time.Now()
	s.status.Active = true
	s.status.StreamingText = false
	s.status.Elapsed = 0
	s.SetWork(s.status)
}

// MarkStreaming hides the strip during assistant text streaming (FR-006).
func (s *statusBar) MarkStreaming() {
	s.status.StreamingText = true
	s.setStripVisible(false)
}

// MarkTooling brings the strip back between tool bursts (FR-006).
func (s *statusBar) MarkTooling(step string) {
	if !s.status.Active {
		return
	}
	s.status.StreamingText = false
	s.status.Phase = "tool"
	s.status.CurrentStep = step
	s.status.Elapsed = time.Since(s.start)
	s.SetWork(s.status)
}

// MarkStale flags context/rate data older than 15 minutes (FR-020).
func (s *statusBar) MarkStale(contextStale, rateStale bool) {
	s.status.ContextStale = contextStale
	s.status.RateStale = rateStale
	if s.status.Active {
		s.SetWork(s.status)
	}
}

// EndWork hides the strip when the turn finishes.
func (s *statusBar) EndWork() {
	s.status.Active = false
	s.stripLbl.SetAccent(false)
	s.setStripVisible(false)
}

func (s *statusBar) content() fyne.CanvasObject { return s.root }

// gitBranch returns the current git branch, or "" when not in a repo.
func gitBranch() string {
	out, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// fmtDuration renders a duration compactly: 1m 23s / 45s.
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	m := int(d.Minutes())
	sec := int(d.Seconds()) % 60
	if sec == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm %ds", m, sec)
}
