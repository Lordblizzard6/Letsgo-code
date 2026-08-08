package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/user/go-claude-code/internal/api"
)

// proseWidth caps the transcript column so prose reads at a comfortable
// measure (US3, T024): the chat window centers a ~720px column and code
// blocks wrap at the same limit.
const proseWidth float32 = 720

// proseClamp caps the minimum width a message column can claim so cards wrap
// at proseWidth instead of growing edge-to-edge (US3 legibility).
type proseClamp struct {
	fyne.CanvasObject
	max float32
}

func (p *proseClamp) MinSize() fyne.Size {
	s := p.CanvasObject.MinSize()
	if s.Width > p.max {
		s.Width = p.max
	}
	return s
}

// cardFrame wraps content in the common card frame (US3, T025): 1px border,
// 6px corner radius and 16px padding, all from theme tokens.
func cardFrame(content fyne.CanvasObject) fyne.CanvasObject {
	stroke := canvas.NewRectangle(colorTransparent)
	stroke.StrokeColor = cardBorderColor()
	stroke.StrokeWidth = 1
	stroke.CornerRadius = 6
	padded := container.New(layout.NewCustomPaddedLayout(16, 16, 16, 16), content)
	return container.NewStack(stroke, padded)
}

// messageText extracts the plain text of an api.Message regardless of whether
// Content is a string or a block list.
func messageText(m api.Message) string {
	return messageContentText(m.Content)
}

// messageContentText extracts plain text from a message Content value
// (string, []ContentBlock, or []api.ContentBlock) .
func messageContentText(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []api.ContentBlock:
		var b strings.Builder
		for _, block := range c {
			if block.Text != "" {
				b.WriteString(block.Text)
				b.WriteString("\n")
			}
		}
		return strings.TrimSpace(b.String())
	case []map[string]interface{}:
		var b strings.Builder
		for _, block := range c {
			if t, ok := block["text"].(string); ok {
				b.WriteString(t)
				b.WriteString("\n")
			}
		}
		return strings.TrimSpace(b.String())
	default:
		return fmt.Sprintf("%v", content)
	}
}
