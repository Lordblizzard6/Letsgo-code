package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// approvalCard is the inline approval card shown above the composer
// (FR-001/002/003/021): tool name, summary with diff highlights, and
// Enter=approve / Esc=reject / Ctrl+Enter=approve+grant session hints.
// It never blocks the transcript: the card lives in the bottom strip.
type approvalCard struct {
	widget.BaseWidget

	toolName string
	summary  string

	onEnter func()
	onEsc   func()
	onGrant func()

	entry    *widget.Entry
	grantBtn *widget.Button
	root     fyne.CanvasObject
}

// newApprovalCard builds the inline card body.
func newApprovalCard(toolName, summary string) *approvalCard {
	c := &approvalCard{toolName: toolName, summary: summary}

	title := widget.NewLabelWithStyle(
		"Approve \""+toolName+"\"?",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	title.Importance = widget.HighImportance

	body := diffRichText(summary)

	hint := widget.NewLabel(
		"Enter approve · Esc reject · Ctrl+Enter approve for this session",
	)
	hint.Importance = widget.MediumImportance
	hint.TextStyle = fyne.TextStyle{Monospace: true}

	c.entry = widget.NewEntry()
	c.entry.SetPlaceHolder("")

	// FR-016: a visible grant button (Enter/Ctrl+Enter also work); clicking it
	// approves AND persists a per-session grant for this category.
	grantBtn := widget.NewButton("Allow for this session", func() {
		if c.onGrant != nil {
			c.onGrant()
		}
	})
	grantBtn.Importance = widget.HighImportance
	c.grantBtn = grantBtn

	c.root = container.NewVBox(title, body, hint, container.NewHBox(grantBtn, c.entry))

	// Accent-left bar (T008): a thin accent strip marks the card as a live
	// decision point, echoing the rail active mark (contracts §1.4).
	bar := canvas.NewRectangle(AccentColor)
	c.root = container.NewBorder(nil, nil, &minWidth{CanvasObject: bar, w: 4}, nil, c.root)
	c.ExtendBaseWidget(c)
	return c
}

// diffRichText renders the tool input with diff-style highlighting
// (lines starting with +, -, @@ are colored via theme color names).
func diffRichText(summary string) *widget.RichText {
	lines := strings.Split(summary, "\n")
	segments := make([]widget.RichTextSegment, 0, len(lines))
	for _, line := range lines {
		style := widget.RichTextStyleInline
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "+++"), strings.HasPrefix(trimmed, "---"):
			style.TextStyle.Bold = true
			style.ColorName = theme.ColorNamePrimary
		case strings.HasPrefix(trimmed, "+"):
			style.ColorName = theme.ColorNameSuccess
		case strings.HasPrefix(trimmed, "-"):
			style.ColorName = theme.ColorNameError
		case strings.HasPrefix(trimmed, "@@"):
			style.ColorName = theme.ColorNamePrimary
		}
		segments = append(segments, &widget.TextSegment{Style: style, Text: line + "\n"})
	}
	return widget.NewRichText(segments...)
}

// requestFocus focuses the entry so TypedKey receives Enter/Esc.
func (c *approvalCard) requestFocus(win fyne.Window) {
	win.Canvas().Focus(c.entry)
}

func (c *approvalCard) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		if c.onEnter != nil {
			c.onEnter()
		}
	case fyne.KeyEscape:
		if c.onEsc != nil {
			c.onEsc()
		}
	}
}

func (c *approvalCard) TypedRune(r rune) {}

// CreateRenderer implements fyne.Widget.
func (c *approvalCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.root)
}

// pendingIndicator shows the "N pending" marker above the approval zone
// (FR-021): a compact line that only appears when more than one card waits.
type pendingIndicator struct {
	widget.BaseWidget

	label *widget.Label
	count int
	root  fyne.CanvasObject
}

func newPendingIndicator() *pendingIndicator {
	p := &pendingIndicator{
		label: widget.NewLabel(""),
	}
	p.label.Importance = widget.WarningImportance
	p.label.TextStyle = fyne.TextStyle{Monospace: true}
	p.label.Hidden = true
	p.root = container.NewHBox(widget.NewIcon(iconFor("approval")), p.label)
	p.ExtendBaseWidget(p)
	return p
}

// set updates the counter; the indicator hides at 0 and 1.
func (p *pendingIndicator) set(n int) {
	p.count = n
	if n > 1 {
		p.label.SetText("approvals: " + itoa(n) + " pending")
		p.label.Show()
	} else {
		p.label.Hide()
	}
}

func (p *pendingIndicator) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.root)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
