package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// toolPanel renders a single tool invocation inline in the conversation:
// name + status header, input summary, diff preview for edits, and command
// output (contracts/gui-contract.md FR-024 — never plain text).
type toolPanel struct {
	widget.BaseWidget

	name    string
	status  string
	input   string
	output  string
	isError bool

	header   *widget.Label
	statusLb *widget.Label
	body     *fyne.Container
}

// newToolPanel creates a panel for the named tool.
func newToolPanel(name string) *toolPanel {
	p := &toolPanel{name: name}
	p.header = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	p.statusLb = widget.NewLabel("")
	p.body = container.NewVBox(p.header, p.statusLb)
	p.ExtendBaseWidget(p)
	return p
}

// SetRequested updates the panel with pending-approval input.
func (p *toolPanel) SetRequested(input string) {
	p.status = "requested"
	p.input = input
	p.refresh()
}

// SetExecuting marks the tool as approved and running.
func (p *toolPanel) SetExecuting() {
	p.status = "executing"
	p.refresh()
}

// SetResult finalizes the panel with tool output.
func (p *toolPanel) SetResult(output string, isError bool) {
	p.status = "done"
	p.output = output
	p.isError = isError
	p.refresh()
}

// SetRejected marks the tool as rejected (never executed).
func (p *toolPanel) SetRejected() {
	p.status = "rejected"
	p.refresh()
}

// SetTimedOut marks the tool as rejected due to approval timeout.
func (p *toolPanel) SetTimedOut() {
	p.status = "timed out"
	p.refresh()
}

func (p *toolPanel) refresh() {
	status := p.status
	if p.isError {
		status = "error"
	}
	p.header.SetText(p.name)
	p.statusLb.SetText(status)
	p.statusLb.Importance = importanceFor(p.status, p.isError)

	var parts []fyne.CanvasObject
	parts = append(parts, p.header, p.statusLb)

	if p.input != "" {
		parts = append(parts, widget.NewRichText(&widget.TextSegment{
			Style: widget.RichTextStyleCodeInline,
			Text:  summarizeInput(p.input),
		}))
	}
	if p.output != "" {
		out := p.output
		if len(out) > 600 {
			out = out[:600] + "…"
		}
		style := widget.RichTextStyleCodeBlock
		if p.isError {
			style = widget.RichTextStyleCodeBlock
		}
		parts = append(parts, widget.NewRichText(&widget.TextSegment{Style: style, Text: out}))
	}

	p.body.Objects = parts
	p.body.Refresh()
}

func importanceFor(status string, isError bool) widget.Importance {
	if isError || status == "rejected" || status == "timed out" {
		return widget.DangerImportance
	}
	if status == "executing" {
		return widget.WarningImportance
	}
	return widget.MediumImportance
}

// summarizeInput renders the tool input as a compact single-line summary.
func summarizeInput(input string) string {
	s := strings.TrimSpace(input)
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}

// CreateRenderer implements fyne.Widget.
func (p *toolPanel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.body)
}
