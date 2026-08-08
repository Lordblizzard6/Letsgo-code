package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/engine"
)

// planCard is the inline plan proposal card (FR-013): steps, files and
// criteria with Approve / Reject / Edit instructions actions. It mounts in
// the same bottom strip as approval cards so the transcript stays visible.
type planCard struct {
	widget.BaseWidget

	planID string

	onApprove func(planID string)
	onReject  func(planID string)
	onEdit    func(planID, instructions string)

	instructions *widget.Entry
	root         fyne.CanvasObject
}

// newPlanCard builds the card body from a PlanProposed event.
func newPlanCard(proposed engine.PlanProposed) *planCard {
	c := &planCard{planID: proposed.PlanID}

	title := widget.NewLabelWithStyle(
		"Plan — "+proposed.PlanID,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	title.Importance = widget.HighImportance

	var body []fyne.CanvasObject
	body = append(body, title)

	if len(proposed.Steps) > 0 {
		steps := widget.NewRichText()
		for i, s := range proposed.Steps {
			line := strings.Builder{}
			line.WriteString(itoa(i + 1))
			line.WriteString(". ")
			line.WriteString(s.Action)
			if s.File != "" {
				line.WriteString(" — ")
				line.WriteString(s.File)
			}
			steps.Segments = append(steps.Segments, &widget.TextSegment{
				Style: widget.RichTextStyleInline,
				Text:  line.String() + "\n",
			})
		}
		body = append(body, steps)
	}

	if len(proposed.Files) > 0 {
		files := widget.NewLabelWithStyle(
			"Files: "+strings.Join(proposed.Files, ", "),
			fyne.TextAlignLeading,
			fyne.TextStyle{Monospace: true},
		)
		files.Wrapping = fyne.TextWrapWord
		body = append(body, files)
	}

	if proposed.Criteria != "" {
		crit := widget.NewLabelWithStyle(
			"Criteria: "+proposed.Criteria,
			fyne.TextAlignLeading,
			fyne.TextStyle{Monospace: true},
		)
		crit.Wrapping = fyne.TextWrapWord
		body = append(body, crit)
	}

	c.instructions = widget.NewEntry()
	c.instructions.SetPlaceHolder("Edit instructions (optional)…")

	approve := widget.NewButton("Approve", func() {
		if c.onApprove != nil {
			c.onApprove(c.planID)
		}
	})
	approve.Importance = widget.HighImportance
	reject := widget.NewButton("Reject", func() {
		if c.onReject != nil {
			c.onReject(c.planID)
		}
	})
	edit := widget.NewButton("Edit", func() {
		if c.onEdit != nil {
			c.onEdit(c.planID, c.instructions.Text)
		}
	})
	hint := widget.NewLabel("Enter approve · Esc reject")
	hint.Importance = widget.MediumImportance
	hint.TextStyle = fyne.TextStyle{Monospace: true}

	body = append(body,
		c.instructions,
		container.NewHBox(approve, reject, edit),
		hint,
	)

	c.root = container.NewVBox(body...)
	c.ExtendBaseWidget(c)
	return c
}

func (c *planCard) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyEnter, fyne.KeyReturn:
		if c.onApprove != nil {
			c.onApprove(c.planID)
		}
	case fyne.KeyEscape:
		if c.onReject != nil {
			c.onReject(c.planID)
		}
	}
}

func (c *planCard) TypedRune(r rune) {}

// CreateRenderer implements fyne.Widget.
func (c *planCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.root)
}

// requestFocus focuses the instructions entry so Enter/Esc reach TypedKey.
func (c *planCard) requestFocus(win fyne.Window) {
	win.Canvas().Focus(c.instructions)
}
