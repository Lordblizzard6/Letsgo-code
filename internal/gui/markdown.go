package gui

import (
	"net/url"
	"strings"

	"fyne.io/fyne/v2/widget"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// markdownSegments parses markdown with goldmark and builds RichText segments
// (text, bold/italic, headings, hyperlinks, inline code, block code).
func markdownSegments(src string) []widget.RichTextSegment {
	var segs []widget.RichTextSegment
	source := []byte(src)
	root := goldmark.New().Parser().Parse(text.NewReader(source))
	walk(root, source, &segs)
	return segs
}

// newMarkdownRichText renders a markdown string as a widget.RichText.
func newMarkdownRichText(src string) *widget.RichText {
	return widget.NewRichText(markdownSegments(src)...)
}

func walk(n ast.Node, source []byte, segs *[]widget.RichTextSegment) {
	switch node := n.(type) {
	case *ast.Text:
		txt := string(node.Text(source))
		if txt == "" {
			return
		}
		*segs = append(*segs, textSeg(widget.RichTextStyleInline, txt))
		return
	case *ast.String:
		*segs = append(*segs, textSeg(widget.RichTextStyleInline, string(node.Value)))
		return
	case *ast.Emphasis:
		style := widget.RichTextStyleEmphasis
		if node.Level == 2 {
			style = widget.RichTextStyleStrong
		}
		*segs = append(*segs, walkInline(node, source, style)...)
		return
	case *ast.Heading:
		var b strings.Builder
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, &b)
		}
		*segs = append(*segs, textSeg(widget.RichTextStyleHeading, strings.TrimSpace(b.String())))
		return
	case *ast.CodeSpan:
		*segs = append(*segs, textSeg(widget.RichTextStyleCodeInline, string(node.Text(source))))
		return
	case *ast.Link:
		title := string(node.Text(source))
		if title == "" {
			title = string(node.Destination)
		}
		u, err := url.Parse(string(node.Destination))
		if err != nil {
			u = &url.URL{}
		}
		*segs = append(*segs, &widget.HyperlinkSegment{Text: title, URL: u})
		return
	case *ast.FencedCodeBlock:
		*segs = append(*segs, codeBlockSegLines(node.Lines(), source))
		return
	case *ast.CodeBlock:
		*segs = append(*segs, codeBlockSegLines(node.Lines(), source))
		return
	case *ast.List:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			item, ok := child.(*ast.ListItem)
			if !ok {
				continue
			}
			var b strings.Builder
			b.WriteString("• ")
			for c := item.FirstChild(); c != nil; c = c.NextSibling() {
				if p, ok := c.(*ast.Paragraph); ok {
					for t := p.FirstChild(); t != nil; t = t.NextSibling() {
						collectText(t, source, &b)
					}
				} else {
					collectText(c, source, &b)
				}
			}
			*segs = append(*segs, textSeg(widget.RichTextStyleParagraph, strings.TrimRight(b.String(), "\n")))
		}
		return
	case *ast.Paragraph:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			walk(child, source, segs)
		}
		*segs = append(*segs, textSeg(widget.RichTextStyleParagraph, "\n"))
		return
	case *ast.Blockquote:
		var b strings.Builder
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, &b)
			b.WriteString("\n")
		}
		*segs = append(*segs, textSeg(widget.RichTextStyleParagraph, strings.TrimSpace(b.String())))
		return
	default:
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			walk(child, source, segs)
		}
	}
}

// textSeg builds a text segment (pointer, as required by RichTextSegment).
func textSeg(style widget.RichTextStyle, text string) widget.RichTextSegment {
	return &widget.TextSegment{Style: style, Text: text}
}

// codeBlockSegLines joins code block lines into a monospace block segment.
func codeBlockSegLines(lines *text.Segments, source []byte) widget.RichTextSegment {
	return textSeg(widget.RichTextStyleCodeBlock, strings.TrimRight(string(lines.Value(source)), "\n"))
}

// walkInline renders a node's children as inline segments with the given style.
func walkInline(n ast.Node, source []byte, style widget.RichTextStyle) []widget.RichTextSegment {
	var segs []widget.RichTextSegment
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Text:
			txt := string(c.Text(source))
			if txt == "" {
				continue
			}
			segs = append(segs, textSeg(style, txt))
		case *ast.CodeSpan:
			segs = append(segs, textSeg(widget.RichTextStyleCodeInline, string(c.Text(source))))
		case *ast.Link:
			title := string(c.Text(source))
			if title == "" {
				title = string(c.Destination)
			}
			u, err := url.Parse(string(c.Destination))
			if err != nil {
				u = &url.URL{}
			}
			segs = append(segs, &widget.HyperlinkSegment{Text: title, URL: u})
		default:
			for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if t, ok := gc.(*ast.Text); ok {
					segs = append(segs, textSeg(style, string(t.Text(source))))
				}
			}
		}
	}
	return segs
}

// collectText appends a node's text content to the builder (no styling).
func collectText(n ast.Node, source []byte, b *strings.Builder) {
	switch node := n.(type) {
	case *ast.Text:
		b.WriteString(string(node.Text(source)))
	case *ast.CodeSpan:
		b.WriteString(string(node.Text(source)))
	case *ast.String:
		b.WriteString(string(node.Value))
	default:
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, b)
		}
	}
}
