package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/user/go-claude-code/internal/api"
)

// AgentState represents the state of an agent execution
type AgentState struct {
	ID         string                 `json:"id"`
	Status     string                 `json:"status"` // "idle", "running", "completed", "failed"
	Goal       string                 `json:"goal"`
	WorkingDir string                 `json:"working_dir"`
	History    []api.Message          `json:"history"`
	ToolsUsed  []string               `json:"tools_used"`
	Result     string                 `json:"result"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type AgentStore struct {
	Agents map[string]*AgentState `json:"agents"`
	mu     sync.RWMutex
	path   string
}

var (
	agentStoreInstance *AgentStore
	agentStoreOnce     sync.Once
)

func getAgentStore() *AgentStore {
	agentStoreOnce.Do(func() {
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".letsGo", "agents.json")

		agentStoreInstance = &AgentStore{
			Agents: make(map[string]*AgentState),
			path:   path,
		}

		agentStoreInstance.load()
	})

	return agentStoreInstance
}

func (s *AgentStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Agents)
}

func (s *AgentStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.Agents, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *AgentStore) create(goal, workingDir string) *AgentState {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent := &AgentState{
		ID:         generateAgentID(),
		Status:     "idle",
		Goal:       goal,
		WorkingDir: workingDir,
		History:    []api.Message{},
		ToolsUsed:  []string{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	s.Agents[agent.ID] = agent
	s.save()

	return agent
}

func (s *AgentStore) get(id string) (*AgentState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, exists := s.Agents[id]
	return agent, exists
}

func (s *AgentStore) update(id string, updates map[string]interface{}) (*AgentState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, exists := s.Agents[id]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	if status, ok := updates["status"].(string); ok {
		agent.Status = status
	}
	if result, ok := updates["result"].(string); ok {
		agent.Result = result
	}
	if history, ok := updates["history"].([]api.Message); ok {
		agent.History = history
	}
	if tools, ok := updates["tools_used"].([]string); ok {
		agent.ToolsUsed = tools
	}

	agent.UpdatedAt = time.Now()
	s.save()

	return agent, nil
}

func generateAgentID() string {
	return fmt.Sprintf("agent_%d", time.Now().UnixNano())
}

// AgentTool allows Claude to spawn sub-agents for parallel task execution
type AgentTool struct{}

func (t *AgentTool) Definition() api.Tool {
	return api.Tool{
		Name:        "agent",
		Description: `Spawn an agent to handle a subtask. The agent has access to bash, file read/edit, and search tools. Use this to delegate work that can be done in parallel or to get a fresh perspective on a problem. The agent works independently and returns its result.`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The task description for the agent. Be specific about what the agent should do.",
				},
				"working_dir": map[string]interface{}{
					"type":        "string",
					"description": "Working directory for the agent (defaults to current directory)",
				},
				"tools": map[string]interface{}{
					"type":        "array",
					"description": "List of tools the agent can use (defaults to standard set)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"required": []string{"prompt"},
		},
	}
}

func (t *AgentTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	prompt, _ := m["prompt"].(string)
	if prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	workingDir, _ := m["working_dir"].(string)
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "."
		}
	}

	store := getAgentStore()
	agent := store.create(prompt, workingDir)

	// Start the agent execution
	// In a full implementation, this would spawn a goroutine that:
	// 1. Sets up a new API client context
	// 2. Sends the prompt with appropriate system instructions
	// 3. Handles tool calls from the sub-agent
	// 4. Returns the final result

	// For now, we simulate the agent execution
	agent.Status = "running"
	store.update(agent.ID, map[string]interface{}{"status": "running"})

	// Execute the agent task
	result, err := executeAgentTask(agent, prompt)

	if err != nil {
		agent.Status = "failed"
		store.update(agent.ID, map[string]interface{}{
			"status": "failed",
			"result": err.Error(),
		})
		return "", fmt.Errorf("agent failed: %w", err)
	}

	agent.Status = "completed"
	store.update(agent.ID, map[string]interface{}{
		"status": "completed",
		"result": result,
	})

	return fmt.Sprintf("Agent %s completed task.\n\nResult:\n%s", agent.ID[:12], result), nil
}

// executeAgentTask runs the agent's task
func executeAgentTask(agent *AgentState, prompt string) (string, error) {
	// This is a simplified implementation
	// In production, this would:
	// - Create a new chat session
	// - Add system prompt about being a sub-agent
	// - Process the task with tool access
	// - Return the accumulated result

	// For now, return a simulated result
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Working on: %s\n", prompt))
	result.WriteString(fmt.Sprintf("Working directory: %s\n\n", agent.WorkingDir))

	result.WriteString("Agent execution would proceed here with:\n")
	result.WriteString("- Access to bash, file read/edit tools\n")
	result.WriteString("- Independent API calls\n")
	result.WriteString("- Result aggregation\n")

	return result.String(), nil
}

// AgentGetTool retrieves agent status
type AgentGetTool struct{}

func (t *AgentGetTool) Definition() api.Tool {
	return api.Tool{
		Name:        "agent_get",
		Description: "Get the status and result of a specific agent by ID.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The agent ID",
				},
			},
			"required": []string{"id"},
		},
	}
}

func (t *AgentGetTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	id, _ := m["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	store := getAgentStore()

	// Try prefix match
	var agent *AgentState
	for aid, a := range store.Agents {
		if len(aid) >= len(id) && aid[:len(id)] == id {
			agent = a
			break
		}
	}

	if agent == nil {
		return "", fmt.Errorf("agent not found: %s", id)
	}

	result := fmt.Sprintf("Agent: %s\n", agent.ID)
	result += fmt.Sprintf("Status: %s\n", agent.Status)
	result += fmt.Sprintf("Goal: %s\n", agent.Goal)
	result += fmt.Sprintf("Working Dir: %s\n", agent.WorkingDir)
	result += fmt.Sprintf("Created: %s\n", agent.CreatedAt.Format("2006-01-02 15:04"))
	result += fmt.Sprintf("Updated: %s\n", agent.UpdatedAt.Format("2006-01-02 15:04"))

	if agent.Result != "" {
		result += fmt.Sprintf("\nResult:\n%s\n", agent.Result)
	}

	return result, nil
}

// AgentListTool lists all agents
type AgentListTool struct{}

func (t *AgentListTool) Definition() api.Tool {
	return api.Tool{
		Name:        "agent_list",
		Description: "List all agents and their status.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}
}

func (t *AgentListTool) Execute(input interface{}) (string, error) {
	store := getAgentStore()

	if len(store.Agents) == 0 {
		return "No agents found.", nil
	}

	var result strings.Builder
	result.WriteString("Agents:\n\n")

	var idle, running, completed, failed int

	for _, agent := range store.Agents {
		switch agent.Status {
		case "idle":
			idle++
			result.WriteString(fmt.Sprintf("[ ] %s: %s\n", agent.ID[:12], agent.Goal))
		case "running":
			running++
			result.WriteString(fmt.Sprintf("[~] %s: %s\n", agent.ID[:12], agent.Goal))
		case "completed":
			completed++
			result.WriteString(fmt.Sprintf("[x] %s: %s\n", agent.ID[:12], agent.Goal))
		case "failed":
			failed++
			result.WriteString(fmt.Sprintf("[!] %s: %s\n", agent.ID[:12], agent.Goal))
		}
	}

	result.WriteString(fmt.Sprintf("\nSummary: %d idle, %d running, %d completed, %d failed\n",
		idle, running, completed, failed))

	return result.String(), nil
}

// Exported helper functions for CLI commands

// AgentInfo represents agent information for CLI
type AgentInfo struct {
	ID        string
	Name      string
	Status    string
	StartedAt time.Time
	Completed bool
	Result    string
}

// ListAgents returns all agents for CLI display
func ListAgents() []AgentInfo {
	store := getAgentStore()
	var result []AgentInfo

	for _, agent := range store.Agents {
		info := AgentInfo{
			ID:        agent.ID,
			Name:      agent.Goal,
			Status:    agent.Status,
			StartedAt: agent.CreatedAt,
			Completed: agent.Status == "completed" || agent.Status == "failed",
			Result:    agent.Result,
		}
		result = append(result, info)
	}
	return result
}

// GetAgent returns a specific agent by ID
func GetAgent(id string) (*AgentInfo, error) {
	store := getAgentStore()

	var agent *AgentState
	for aid, a := range store.Agents {
		if len(aid) >= len(id) && aid[:len(id)] == id {
			agent = a
			break
		}
	}

	if agent == nil {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return &AgentInfo{
		ID:        agent.ID,
		Name:      agent.Goal,
		Status:    agent.Status,
		StartedAt: agent.CreatedAt,
		Completed: agent.Status == "completed" || agent.Status == "failed",
		Result:    agent.Result,
	}, nil
}

// StopAgent stops a running agent
func StopAgent(id string) error {
	store := getAgentStore()

	var agentID string
	for aid := range store.Agents {
		if len(aid) >= len(id) && aid[:len(id)] == id {
			agentID = aid
			break
		}
	}

	if agentID == "" {
		return fmt.Errorf("agent not found: %s", id)
	}

	store.update(agentID, map[string]interface{}{
		"status": "failed",
		"result": "Stopped by user",
	})
	return nil
}
