package gui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/user/go-claude-code/internal/mcp"
)

// mcpView lists configured MCP servers, their connection state and discovered
// tools (FR-019).
type mcpView struct {
	win  fyne.Window
	root *fyne.Container

	manager    *mcp.Manager
	listNames  []string
	serverList *widget.List
	center     *emptyAware
	stateLbl   *widget.Label
	detailLbl  *widget.Label
	toolLbl    *widget.Label
	selected   string
}

// newMCPView builds the MCP server management panel.
func newMCPView(win fyne.Window) *mcpView {
	v := &mcpView{
		win:     win,
		manager: mcp.GetManager(),
	}

	v.stateLbl = widget.NewLabel("")
	v.stateLbl.Wrapping = fyne.TextWrapWord

	v.detailLbl = widget.NewLabel("")
	v.detailLbl.Wrapping = fyne.TextWrapWord
	v.detailLbl.TextStyle = fyne.TextStyle{Monospace: true}

	v.toolLbl = widget.NewLabel("")
	v.toolLbl.Wrapping = fyne.TextWrapWord
	v.toolLbl.TextStyle = fyne.TextStyle{Monospace: true}

	v.serverList = widget.NewList(
		func() int { return len(v.listNames) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(v.listNames) {
				return
			}
			name := v.listNames[id]
			lb := item.(*widget.Label)
			if v.manager.IsServerRunning(name) {
				lb.SetText("● " + name)
			} else {
				lb.SetText("○ " + name)
			}
		},
	)
	v.serverList.OnSelected = func(id widget.ListItemID) {
		if int(id) >= 0 && int(id) < len(v.listNames) {
			v.selected = v.listNames[id]
			v.renderDetail(v.listNames[id])
		}
	}
	v.serverList.OnUnselected = func(widget.ListItemID) {}

	actionsRow := container.NewHBox(
		widget.NewButton("Start", v.startSelected),
		widget.NewButton("Stop", v.stopSelected),
		widget.NewButton("Add…", v.addServer),
		widget.NewButton("Remove", v.removeSelected),
		widget.NewButton("Refresh", v.refresh),
	)

	left := container.NewBorder(nil, nil, nil, nil, v.serverList)
	v.center = newEmptyAware(v.serverList, "No hay servidores MCP configurados.", "Añadir servidor", v.addServer)
	left = container.NewBorder(nil, nil, nil, nil, v.center.content())

	right := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("State", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			v.stateLbl,
			v.detailLbl,
		),
		nil, nil, nil,
		container.NewBorder(
			widget.NewLabelWithStyle("Tools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			container.NewScroll(v.toolLbl),
		),
	)

	v.root = container.NewBorder(
		actionsRow, nil, nil, nil,
		container.NewHSplit(left, right),
	)
	v.refresh()
	return v
}

// content returns the panel for embedding in the main window.
func (v *mcpView) content() fyne.CanvasObject { return v.root }

// refresh reloads the server list and resets the detail panes.
func (v *mcpView) refresh() {
	servers := v.manager.ListServers()
	v.listNames = nil
	for name := range servers {
		v.listNames = append(v.listNames, name)
	}
	sort.Strings(v.listNames)
	v.serverList.Refresh()
	if v.center != nil {
		v.center.setEmpty(len(v.listNames) == 0)
	}
	v.stateLbl.SetText("")
	v.detailLbl.SetText("")
	v.toolLbl.SetText("")
	if v.selected != "" {
		v.renderDetail(v.selected)
	}
}

// renderDetail shows connection state and tools for the named server.
func (v *mcpView) renderDetail(name string) {
	if cfg, err := v.manager.GetServer(name); err == nil && cfg != nil {
		desc := cfg.Description
		if desc != "" {
			v.detailLbl.SetText(desc)
		} else {
			v.detailLbl.SetText("")
		}
	}
	if !v.manager.IsServerRunning(name) {
		v.stateLbl.SetText("stopped")
		v.toolLbl.SetText("Start the server to discover tools.")
		return
	}
	tools, err := v.manager.ListServerTools(name)
	if err != nil {
		v.stateLbl.SetText("running (tools unavailable: " + err.Error() + ")")
		v.toolLbl.SetText("")
		return
	}
	v.stateLbl.SetText("connected")
	var tnames []string
	for _, t := range tools {
		tnames = append(tnames, t.Name)
	}
	sort.Strings(tnames)
	if len(tnames) == 0 {
		v.toolLbl.SetText("(no tools exposed)")
		return
	}
	v.toolLbl.SetText(strings.Join(tnames, "\n"))
}

// startSelected starts the currently selected server.
func (v *mcpView) startSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("MCP", "Select a server first.", v.win)
		return
	}
	if err := v.manager.StartServer(name); err != nil {
		dialog.ShowError(fmt.Errorf("start %s: %v", name, err), v.win)
		return
	}
	dialog.ShowInformation("MCP", "Started "+name, v.win)
	v.refresh()
}

// stopSelected stops the currently selected server.
func (v *mcpView) stopSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("MCP", "Select a server first.", v.win)
		return
	}
	if err := v.manager.StopServer(name); err != nil {
		dialog.ShowError(fmt.Errorf("stop %s: %v", name, err), v.win)
		return
	}
	dialog.ShowInformation("MCP", "Stopped "+name, v.win)
	v.refresh()
}

// addServer prompts for a config (name, command, args) and registers it.
func (v *mcpView) addServer() {
	nameE := widget.NewEntry()
	nameE.SetPlaceHolder("Name")
	cmdE := widget.NewEntry()
	cmdE.SetPlaceHolder("Command (e.g. npx)")
	argsE := widget.NewEntry()
	argsE.SetPlaceHolder("Args (space separated)")
	dlg := dialog.NewForm("Add MCP server", "Add", "Cancel",
		[]*widget.FormItem{
			{Text: "Name", Widget: nameE},
			{Text: "Command", Widget: cmdE},
			{Text: "Args", Widget: argsE},
		},
		func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(nameE.Text)
			if name == "" {
				return
			}
			cfg := &mcp.ServerConfig{
				Name:    name,
				Command: strings.TrimSpace(cmdE.Text),
				Args:    strings.Fields(argsE.Text),
			}
			if err := v.manager.AddServer(cfg); err != nil {
				dialog.ShowError(fmt.Errorf("add server: %v", err), v.win)
				return
			}
			v.refresh()
			dialog.ShowInformation("MCP", "Added "+name, v.win)
		}, v.win)
	dlg.Show()
}

// removeSelected deletes the selected server config.
func (v *mcpView) removeSelected() {
	name := v.selected
	if name == "" {
		dialog.ShowInformation("MCP", "Select a server first.", v.win)
		return
	}
	if err := v.manager.RemoveServer(name); err != nil {
		dialog.ShowError(fmt.Errorf("remove %s: %v", name, err), v.win)
		return
	}
	v.selected = ""
	v.refresh()
}
