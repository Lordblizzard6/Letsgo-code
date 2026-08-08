package gui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/plugins"
)

// pluginsView lists installed plugins with install/remove and load status
// (FR-020).
type pluginsView struct {
	win  fyne.Window
	root *fyne.Container

	registry   *plugins.Registry
	listNames  []string
	pluginList *widget.List
	center     *emptyAware
	selected   string
}

// newPluginsView builds the plugin management panel.
func newPluginsView(win fyne.Window) *pluginsView {
	v := &pluginsView{
		win:      win,
		registry: plugins.GetRegistry(),
	}

	v.pluginList = widget.NewList(
		func() int { return len(v.listNames) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(v.listNames) {
				return
			}
			name := v.listNames[id]
			p, err := v.registry.GetPlugin(name)
			lb := item.(*widget.Label)
			if err != nil || p == nil {
				lb.SetText(name + "  (unknown)")
				return
			}
			state := "installed"
			if p.Enabled {
				state = "enabled"
			}
			if p.Loaded {
				state += "/loaded"
			}
			lb.SetText(fmt.Sprintf("%s  [%s]", name, state))
		},
	)
	v.pluginList.OnSelected = func(id widget.ListItemID) {
		if int(id) >= 0 && int(id) < len(v.listNames) {
			v.selected = v.listNames[id]
		}
	}

	actions := container.NewHBox(
		widget.NewButton("Reload", v.reloadSelected),
		widget.NewButton("Enable", v.enableSelected),
		widget.NewButton("Disable", v.disableSelected),
		widget.NewButton("Install…", v.installPrompt),
		widget.NewButton("Remove", v.removeSelected),
		widget.NewButton("Refresh", v.refresh),
	)

	v.center = newEmptyAware(v.pluginList, "No hay plugins instalados.", "Instalar…", v.installPrompt)
	v.root = container.NewBorder(
		actions,
		nil, nil, nil,
		v.center.content(),
	)
	v.refresh()
	return v
}

// content returns the panel for embedding in the main window.
func (v *pluginsView) content() fyne.CanvasObject { return v.root }

// refresh re-lists configured plugins.
func (v *pluginsView) refresh() {
	plugins := v.registry.ListPlugins()
	v.listNames = nil
	for name := range plugins {
		v.listNames = append(v.listNames, name)
	}
	sort.Strings(v.listNames)
	v.pluginList.Refresh()
	if v.center != nil {
		v.center.setEmpty(len(v.listNames) == 0)
	}
}

// reloadSelected reloads the selected plugin.
func (v *pluginsView) reloadSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("Plugins", "Select a plugin first.", v.win)
		return
	}
	if err := v.registry.LoadPlugin(name); err != nil {
		dialog.ShowError(fmt.Errorf("load %s: %v", name, err), v.win)
		return
	}
	v.refresh()
}

// enableSelected enables the selected plugin.
func (v *pluginsView) enableSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("Plugins", "Select a plugin first.", v.win)
		return
	}
	if err := v.registry.EnablePlugin(name); err != nil {
		dialog.ShowError(fmt.Errorf("enable %s: %v", name, err), v.win)
		return
	}
	v.refresh()
}

// disableSelected disables the selected plugin.
func (v *pluginsView) disableSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("Plugins", "Select a plugin first.", v.win)
		return
	}
	if err := v.registry.DisablePlugin(name); err != nil {
		dialog.ShowError(fmt.Errorf("disable %s: %v", name, err), v.win)
		return
	}
	v.refresh()
}

// installPrompt asks for a plugin source (git URL / path) and installs it.
func (v *pluginsView) installPrompt() {
	srcE := widget.NewEntry()
	srcE.SetPlaceHolder("git URL or local path")
	nameE := widget.NewEntry()
	nameE.SetPlaceHolder("plugin name (optional)")
	dlg := dialog.NewForm("Install plugin", "Install", "Cancel",
		[]*widget.FormItem{
			{Text: "Source", Widget: srcE},
			{Text: "Name", Widget: nameE},
		},
		func(ok bool) {
			if !ok {
				return
			}
			src := strings.TrimSpace(srcE.Text)
			if src == "" {
				return
			}
			if err := v.registry.InstallPlugin(src, strings.TrimSpace(nameE.Text)); err != nil {
				dialog.ShowError(fmt.Errorf("install: %v", err), v.win)
				return
			}
			v.refresh()
		}, v.win)
	dlg.Show()
}

// removeSelected uninstalls the selected plugin.
func (v *pluginsView) removeSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("Plugins", "Select a plugin first.", v.win)
		return
	}
	if err := v.registry.UninstallPlugin(name); err != nil {
		dialog.ShowError(fmt.Errorf("remove %s: %v", name, err), v.win)
		return
	}
	v.selected = ""
	v.refresh()
}
