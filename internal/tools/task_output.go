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

// TaskOutput tracks the output of a background task
type TaskOutput struct {
	TaskID      string    `json:"task_id"`
	Output      string    `json:"output"`
	Error       string    `json:"error,omitempty"`
	ExitCode    int       `json:"exit_code"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// TaskOutputManager manages task outputs
type TaskOutputManager struct {
	outputs map[string]*TaskOutput
	mu      sync.RWMutex
}

var (
	taskOutputInstance *TaskOutputManager
	taskOutputOnce     sync.Once
)

func GetTaskOutputManager() *TaskOutputManager {
	taskOutputOnce.Do(func() {
		taskOutputInstance = &TaskOutputManager{
			outputs: make(map[string]*TaskOutput),
		}
		// Load persisted outputs
		taskOutputInstance.load()
	})
	return taskOutputInstance
}

func (m *TaskOutputManager) load() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".letsGo", "task_outputs.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var outputs []*TaskOutput
	if err := json.Unmarshal(data, &outputs); err != nil {
		return
	}

	for _, o := range outputs {
		m.outputs[o.TaskID] = o
	}
}

func (m *TaskOutputManager) save() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveLocked()
}

// saveLocked persists outputs. Caller must hold mu.
func (m *TaskOutputManager) saveLocked() {
	// Filter to only completed outputs and limit to last 100
	var outputs []*TaskOutput
	for _, o := range m.outputs {
		if o.Completed {
			outputs = append(outputs, o)
		}
	}

	if len(outputs) > 100 {
		outputs = outputs[len(outputs)-100:]
	}

	data, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".letsGo", "task_outputs.json")
	_ = os.WriteFile(path, data, 0644)
}

// CreateOutput creates a new task output entry
func (m *TaskOutputManager) CreateOutput(taskID string) *TaskOutput {
	m.mu.Lock()
	defer m.mu.Unlock()

	output := &TaskOutput{
		TaskID:    taskID,
		Output:    "",
		CreatedAt: time.Now(),
	}
	m.outputs[taskID] = output
	return output
}

// AppendOutput appends output to a task
func (m *TaskOutputManager) AppendOutput(taskID, data string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if o, exists := m.outputs[taskID]; exists {
		o.Output += data
	}
}

// CompleteOutput marks a task as complete
func (m *TaskOutputManager) CompleteOutput(taskID string, exitCode int, err string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if o, exists := m.outputs[taskID]; exists {
		o.Completed = true
		o.ExitCode = exitCode
		o.Error = err
		o.CompletedAt = time.Now()
		m.saveLocked()
	}
}

// GetOutput gets the output for a task
func (m *TaskOutputManager) GetOutput(taskID string) (*TaskOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if o, exists := m.outputs[taskID]; exists {
		return o, nil
	}
	return nil, fmt.Errorf("task output not found: %s", taskID)
}

// ListOutputs lists all task outputs
func (m *TaskOutputManager) ListOutputs() []*TaskOutput {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var outputs []*TaskOutput
	for _, o := range m.outputs {
		outputs = append(outputs, o)
	}
	return outputs
}

// TaskOutputTool retrieves output from background tasks
type TaskOutputTool struct{}

func (t *TaskOutputTool) Name() string {
	return "task_output"
}

func (t *TaskOutputTool) Description() string {
	return "Retrieve the output from a background task. Use this to check on the progress or results of tasks that are running in the background."
}

func (t *TaskOutputTool) Definition() api.Tool {
	return api.Tool{
		Name:        "task_output",
		Description: "Retrieve the output from a background task. Use this to check on the progress or results of tasks that are running in the background.",
		InputSchema: t.InputSchema(),
	}
}

func (t *TaskOutputTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the background task to retrieve output from",
			},
			"wait": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to wait for the task to complete before returning",
				"default":     false,
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum time to wait if wait=true",
				"default":     60,
			},
		},
		"required": []string{"task_id"},
	}
}

func (t *TaskOutputTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	return t.executeInternal(m)
}

func (t *TaskOutputTool) executeInternal(input map[string]interface{}) (string, error) {
	taskID, ok := input["task_id"].(string)
	if !ok {
		return "", fmt.Errorf("task_id is required")
	}

	wait := false
	if w, ok := input["wait"].(bool); ok {
		wait = w
	}

	timeoutSec := 60
	if t, ok := input["timeout_seconds"].(float64); ok {
		timeoutSec = int(t)
	}

	manager := GetTaskOutputManager()

	if wait {
		// Wait for completion
		start := time.Now()
		for time.Since(start) < time.Duration(timeoutSec)*time.Second {
			output, err := manager.GetOutput(taskID)
			if err != nil {
				return "", err
			}
			if output.Completed {
				return formatTaskOutput(output), nil
			}
			time.Sleep(500 * time.Millisecond)
		}
		return "", fmt.Errorf("timeout waiting for task %s", taskID)
	}

	output, err := manager.GetOutput(taskID)
	if err != nil {
		return "", err
	}

	return formatTaskOutput(output), nil
}

func formatTaskOutput(o *TaskOutput) string {
	status := "in_progress"
	if o.Completed {
		if o.ExitCode == 0 && o.Error == "" {
			status = "completed"
		} else {
			status = "failed"
		}
	}

	result := fmt.Sprintf("Task: %s\n", o.TaskID)
	result += fmt.Sprintf("Status: %s\n", status)
	result += fmt.Sprintf("Created: %s\n", o.CreatedAt.Format("15:04:05"))

	if o.Completed {
		result += fmt.Sprintf("Completed: %s\n", o.CompletedAt.Format("15:04:05"))
		result += fmt.Sprintf("Exit Code: %d\n", o.ExitCode)
		if o.Error != "" {
			result += fmt.Sprintf("Error: %s\n", o.Error)
		}
	}

	if o.Output != "" {
		result += fmt.Sprintf("\nOutput:\n%s\n", o.Output)
	}

	return result
}
