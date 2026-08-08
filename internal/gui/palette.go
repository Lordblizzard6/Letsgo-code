package gui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// paletteAction is one navigable app action shown in the command palette
// (FR-011). Only keyboard: ↑↓ navigate, Enter executes, Esc closes.
type paletteAction struct {
	label    string
	category string
	run      func()
}

// commandPalette is a keyboard-driven overlay listing app actions plus the
// `@` file picker and `!` shell passthrough (FR-011/026).
type commandPalette struct {
	entry  *widget.Entry
	list   *widget.List
	win    fyne.Window
	items  []string
	all    []paletteAction
	onExec func(text string)
	root   fyne.CanvasObject
}

// newCommandPalette builds the palette overlay (hidden until open).
func newCommandPalette(win fyne.Window) *commandPalette {
	p := &commandPalette{win: win}
	p.entry = widget.NewEntry()
	p.entry.SetPlaceHolder("Command (@file, !shell, or action)…")
	p.entry.OnChanged = func(string) { p.refresh() }

	p.list = widget.NewList(
		func() int { return len(p.items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(p.items) {
				return
			}
			item.(*widget.Label).SetText(p.items[id])
		},
	)
	p.list.OnSelected = func(id widget.ListItemID) {
		if int(id) >= 0 && int(id) < len(p.items) {
			p.choose(p.items[id])
		}
	}

	p.root = container.NewBorder(p.entry, nil, nil, nil, p.list)
	return p
}

// open shows the palette and focuses the filter input.
func (p *commandPalette) open() {
	p.refresh()
	p.win.Canvas().Focus(p.entry)
}

// choose dispatches a selection to the controller: @files are inserted into
// the composer, !shell goes through the permission driver, actions execute.
func (p *commandPalette) choose(selected string) {
	if p.onExec == nil {
		return
	}
	text := strings.TrimSpace(p.entry.Text)
	label := strings.TrimPrefix(selected, "! ")
	if strings.HasPrefix(text, "!") {
		p.onExec("!" + label)
		return
	}
	if strings.HasPrefix(text, "@") {
		label = strings.TrimPrefix(selected, "file ")
		if label != "" {
			p.onExec("@" + label)
		}
		return
	}
	p.onExec(label)
}

// refresh re-syncs the filtered item list from the current query.
func (p *commandPalette) refresh() {
	q := strings.TrimSpace(p.entry.Text)
	p.items = p.items[:0]
	switch {
	case strings.HasPrefix(q, "!"):
		// Shell passthrough: single preview entry; the permission driver
		// still guards execution (security assumption).
		p.items = append(p.items, "! "+q[1:])
	case strings.HasPrefix(q, "@"):
		// File picker: fuzzy search over the project tree (FR-026).
		for _, f := range fuzzyFiles(q[1:], 25) {
			p.items = append(p.items, "file "+f)
		}
	default:
		for _, a := range p.all {
			if q == "" || strings.Contains(strings.ToLower(a.label), strings.ToLower(q)) {
				p.items = append(p.items, a.label)
			}
		}
	}
	p.list.Refresh()
}

// setActions replaces the app action catalogue (FR-011).
func (p *commandPalette) setActions(actions []paletteAction) {
	p.all = actions
	p.refresh()
}

// fuzzyFile is a ranked file match (prefix scores higher).
type fuzzyFile struct {
	path  string
	score int
}

// fuzzyFiles walks the working directory (respecting hidden dirs and .git)
// returning up to limit paths matching q by substring/prefix ranking (FR-026).
func fuzzyFiles(q string, limit int) []string {
	var matches []fuzzyFile
	lq := strings.ToLower(q)
	if lq == "" {
		return nil
	}
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path != "." && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+".git") {
			return nil
		}
		lower := strings.ToLower(path)
		score := 0
		switch {
		case strings.HasPrefix(lower, lq):
			score = 3
		case strings.Contains(lower, lq):
			score = 1
		default:
			return nil
		}
		matches = append(matches, fuzzyFile{path: path, score: score})
		return nil
	})
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].path < matches[j].path
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m.path)
	}
	return out
}
