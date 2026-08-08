package gui

import (
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// agentTask tracks a single parallel sub-agent (FR-018).
type agentTask struct {
	ID      string
	Status  string
	Summary string
}

// agentTasksView renders AgentTaskUpdate events as a progress list
// (contracts/engine-contract.md, FR-018).
type agentTasksView struct {
	root   *fyne.Container
	list   *widget.List
	tasks  []agentTask
	center *emptyAware
}

// newAgentTasksView builds the parallel sub-agent panel.
func newAgentTasksView() *agentTasksView {
	v := &agentTasksView{}

	v.list = widget.NewList(
		func() int { return len(v.tasks) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel(""), widget.NewLabel(""))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if int(id) < 0 || int(id) >= len(v.tasks) {
				return
			}
			t := v.tasks[id]
			status := item.(*fyne.Container).Objects[0].(*widget.Label)
			name := item.(*fyne.Container).Objects[1].(*widget.Label)
			status.SetText(t.Status)
			name.SetText(t.ID + ": " + t.Summary)
		},
	)

	// US4 (T032): reuse the shared empty-state helper instead of a hand-rolled
	// label so the surface meets contracts §5/§10 (empty + CTA).
	v.center = newEmptyAware(container.NewScroll(v.list), "Todavía no hay tareas.", "", nil)
	v.root = container.NewBorder(nil, nil, nil, nil, v.center.content())
	return v
}

// content returns the panel for embedding in the main window.
func (v *agentTasksView) content() fyne.CanvasObject { return v.root }

// update applies an AgentTaskUpdate event from the engine (thread-safe via fyne.Do).
func (v *agentTasksView) update(id, status, summary string) {
	for i, t := range v.tasks {
		if t.ID == id {
			v.tasks[i].Status = status
			if summary != "" {
				v.tasks[i].Summary = summary
			}
			v.refresh()
			return
		}
	}
	v.tasks = append(v.tasks, agentTask{ID: id, Status: status, Summary: summary})
	v.refresh()
}

// summary renders a one-line status of all tasks for the chat view.
func (v *agentTasksView) summary() string {
	if len(v.tasks) == 0 {
		return ""
	}
	running, done := 0, 0
	for _, t := range v.tasks {
		switch t.Status {
		case "running", "scheduled", "pending":
			running++
		case "done", "completed", "complete":
			done++
		}
	}
	return fmt.Sprintf("sub-agents: %d running, %d complete", running, done)
}

// refresh repaints the list and toggles the empty-state overlay (US4).
func (v *agentTasksView) refresh() {
	sort.SliceStable(v.tasks, func(i, j int) bool { return v.tasks[i].ID < v.tasks[j].ID })
	v.list.Refresh()
	v.center.setEmpty(len(v.tasks) == 0)
}
