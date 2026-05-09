package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/user/go-claude-code/internal/api"
)

// Task represents a subtask that can be executed
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "in_progress", "completed", "failed"
	Result      string    `json:"result"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AgentID     string    `json:"agent_id,omitempty"`
}

type TaskStore struct {
	Tasks map[string]*Task `json:"tasks"`
	mu    sync.RWMutex
	path  string
}

var (
	taskInstance *TaskStore
	taskOnce     sync.Once
)

func getTaskStore() *TaskStore {
	taskOnce.Do(func() {
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".letsGo", "tasks.json")

		taskInstance = &TaskStore{
			Tasks: make(map[string]*Task),
			path:  path,
		}

		taskInstance.load()
	})

	return taskInstance
}

func (s *TaskStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Tasks)
}

func (s *TaskStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.saveLocked()
}

func (s *TaskStore) saveLocked() error {

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *TaskStore) create(description string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		ID:          generateTaskID(),
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.Tasks[task.ID] = task
	_ = s.saveLocked()

	return task
}

func (s *TaskStore) get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.Tasks[id]
	return task, exists
}

func (s *TaskStore) update(id string, updates map[string]interface{}) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.Tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	if status, ok := updates["status"].(string); ok {
		task.Status = status
	}
	if result, ok := updates["result"].(string); ok {
		task.Result = result
	}

	task.UpdatedAt = time.Now()
	_ = s.saveLocked()

	return task, nil
}

func (s *TaskStore) list() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.Tasks))
	for _, task := range s.Tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *TaskStore) delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Tasks[id]; !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	delete(s.Tasks, id)
	return s.saveLocked()
}

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// TaskCreateTool creates a new task
type TaskCreateTool struct{}

func (t *TaskCreateTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_create",
		Description: "Create a new subtask with a description. Returns the task ID. Use this to delegate work that can be done independently.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Detailed description of what the task should accomplish",
				},
			},
			"required": []string{"description"},
		},
	}
}

func (t *TaskCreateTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	description, _ := m["description"].(string)
	if description == "" {
		return "", fmt.Errorf("description is required")
	}

	store := getTaskStore()
	task := store.create(description)

	return fmt.Sprintf("Created task %s: %s", task.ID[:12], task.Description), nil
}

// TaskGetTool retrieves a task by ID
type TaskGetTool struct{}

func (t *TaskGetTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_get",
		Description: "Get the status and details of a specific task by its ID.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The task ID",
				},
			},
			"required": []string{"id"},
		},
	}
}

func (t *TaskGetTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	id, _ := m["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	store := getTaskStore()

	// Try exact match first
	task, exists := store.get(id)
	if !exists {
		// Try prefix match
		for taskID, t := range store.Tasks {
			if len(taskID) >= len(id) && taskID[:len(id)] == id {
				task = t
				exists = true
				break
			}
		}
	}

	if !exists {
		return "", fmt.Errorf("task not found: %s", id)
	}

	result := fmt.Sprintf("Task: %s\n", task.ID)
	result += fmt.Sprintf("Status: %s\n", task.Status)
	result += fmt.Sprintf("Description: %s\n", task.Description)
	if task.Result != "" {
		result += fmt.Sprintf("Result: %s\n", task.Result)
	}
	result += fmt.Sprintf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04"))
	result += fmt.Sprintf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04"))

	return result, nil
}

// TaskUpdateTool updates a task's status or result
type TaskUpdateTool struct{}

func (t *TaskUpdateTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_update",
		Description: "Update a task's status or result. Use to mark tasks as completed, failed, or in_progress.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The task ID",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "New status: pending, in_progress, completed, or failed",
					"enum":        []string{"pending", "in_progress", "completed", "failed"},
				},
				"result": map[string]interface{}{
					"type":        "string",
					"description": "Result or output of the task",
				},
			},
			"required": []string{"id"},
		},
	}
}

func (t *TaskUpdateTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	id, _ := m["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	store := getTaskStore()

	updates := make(map[string]interface{})
	if status, ok := m["status"].(string); ok {
		updates["status"] = status
	}
	if result, ok := m["result"].(string); ok {
		updates["result"] = result
	}

	// Try prefix match
	var taskID string
	for tid := range store.Tasks {
		if len(tid) >= len(id) && tid[:len(id)] == id {
			taskID = tid
			break
		}
	}

	if taskID == "" {
		return "", fmt.Errorf("task not found: %s", id)
	}

	task, err := store.update(taskID, updates)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated task %s (status: %s)", task.ID[:12], task.Status), nil
}

// TaskListTool lists all tasks
type TaskListTool struct{}

func (t *TaskListTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_list",
		Description: "List all tasks with their status. Optionally filter by status.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by status (optional)",
					"enum":        []string{"pending", "in_progress", "completed", "failed"},
				},
			},
		},
	}
}

func (t *TaskListTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
	}

	statusFilter, _ := m["status"].(string)

	store := getTaskStore()
	tasks := store.list()

	if len(tasks) == 0 {
		return "No tasks found.", nil
	}

	var pending, inProgress, completed, failed int
	var result string

	result = "Tasks:\n\n"

	for _, task := range tasks {
		if statusFilter != "" && task.Status != statusFilter {
			continue
		}

		switch task.Status {
		case "pending":
			pending++
			result += fmt.Sprintf("[ ] %s: %s\n", task.ID[:12], task.Description)
		case "in_progress":
			inProgress++
			result += fmt.Sprintf("[~] %s: %s\n", task.ID[:12], task.Description)
		case "completed":
			completed++
			result += fmt.Sprintf("[x] %s: %s\n", task.ID[:12], task.Description)
		case "failed":
			failed++
			result += fmt.Sprintf("[!] %s: %s\n", task.ID[:12], task.Description)
		}
	}

	result += fmt.Sprintf("\nSummary: %d pending, %d in progress, %d completed, %d failed\n",
		pending, inProgress, completed, failed)

	return result, nil
}

// TaskStopTool stops/abandons a task
type TaskStopTool struct{}

func (t *TaskStopTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_stop",
		Description: "Stop or cancel a task. Marks it as failed.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The task ID to stop",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Reason for stopping the task",
				},
			},
			"required": []string{"id"},
		},
	}
}

func (t *TaskStopTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	id, _ := m["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	reason, _ := m["reason"].(string)
	if reason == "" {
		reason = "Stopped by user"
	}

	store := getTaskStore()

	updates := map[string]interface{}{
		"status": "failed",
		"result": reason,
	}

	// Try prefix match
	var taskID string
	for tid := range store.Tasks {
		if len(tid) >= len(id) && tid[:len(id)] == id {
			taskID = tid
			break
		}
	}

	if taskID == "" {
		return "", fmt.Errorf("task not found: %s", id)
	}

	task, err := store.update(taskID, updates)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Stopped task %s", task.ID[:12]), nil
}

// Exported helper functions for CLI commands

// ListTasks returns all tasks for CLI display
type TaskInfo struct {
	ID         string
	Command    string
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	ExitCode   int
	Output     string
	Completed  bool
}

func ListTasks() []TaskInfo {
	store := getTaskStore()
	tasks := store.list()

	var result []TaskInfo
	for _, t := range tasks {
		info := TaskInfo{
			ID:        t.ID,
			Status:    t.Status,
			StartedAt: t.CreatedAt,
			Completed: t.Status == "completed" || t.Status == "failed",
		}
		if t.Status == "completed" {
			info.ExitCode = 0
		}
		if t.Status == "failed" {
			info.ExitCode = 1
		}
		result = append(result, info)
	}
	return result
}

func CreateTask(command string, args []string) (string, error) {
	store := getTaskStore()
	task := store.create(command)
	task.Status = "in_progress"
	store.save()
	return task.ID, nil
}

func GetTask(id string) (*TaskInfo, error) {
	store := getTaskStore()
	task, exists := store.get(id)
	if !exists {
		// Try prefix match
		for taskID, t := range store.Tasks {
			if len(taskID) >= len(id) && taskID[:len(id)] == id {
				task = t
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	info := &TaskInfo{
		ID:        task.ID,
		Command:   task.Description,
		Status:    task.Status,
		StartedAt: task.CreatedAt,
		Output:    task.Result,
		Completed: task.Status == "completed" || task.Status == "failed",
	}

	if task.Status == "completed" {
		info.ExitCode = 0
	}
	if task.Status == "failed" {
		info.ExitCode = 1
	}

	return info, nil
}

func StopTask(id string) error {
	store := getTaskStore()
	updates := map[string]interface{}{
		"status": "failed",
		"result": "Stopped by user",
	}

	// Try prefix match
	var taskID string
	for tid := range store.Tasks {
		if len(tid) >= len(id) && tid[:len(id)] == id {
			taskID = tid
			break
		}
	}

	if taskID == "" {
		return fmt.Errorf("task not found: %s", id)
	}

	_, err := store.update(taskID, updates)
	return err
}
