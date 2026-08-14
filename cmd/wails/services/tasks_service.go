package services

import (
	"github.com/user/go-claude-code/internal/tools"
)

// TasksService lists sub-agents with their state for the rail Tareas pane
// (T046, gui-contract §4). The store lives in the shared core, so the GUI
// and the engine see the same agents.
type TasksService struct{}

func NewTasksService() *TasksService { return &TasksService{} }

// Agent is the JSON-safe row for the tasks pane.
type Agent struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Goal   string `json:"goal"`
	Result string `json:"result"`
}

// List returns all sub-agents ordered by creation (oldest first).
func (s *TasksService) List() []Agent {
	store := tools.GetAgentStore()
	agents := store.ListAgents()
	out := make([]Agent, 0, len(agents))
	for _, a := range agents {
		out = append(out, Agent{
			ID:     a.ID,
			Status: a.Status,
			Goal:   a.Goal,
			Result: a.Result,
		})
	}
	return out
}