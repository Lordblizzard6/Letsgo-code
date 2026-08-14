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

type TodoItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "in_progress", "done"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TodoStore struct {
	Items []TodoItem `json:"items"`
	mu    sync.RWMutex
	path  string
}

var (
	todoInstance *TodoStore
	todoOnce     sync.Once
)

func getTodoStore() *TodoStore {
	todoOnce.Do(func() {
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, ".letsGo", "todos.json")

		todoInstance = &TodoStore{
			path: path,
		}

		// Load existing todos
		todoInstance.load()
	})

	return todoInstance
}

func (s *TodoStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		s.Items = []TodoItem{}
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		s.Items = []TodoItem{}
		return err
	}

	return json.Unmarshal(data, &s.Items)
}

func (s *TodoStore) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// saveLocked persists entries. Caller must hold mu.
func (s *TodoStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.Items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *TodoStore) add(description string) *TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := TodoItem{
		ID:          generateTodoID(),
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.Items = append(s.Items, item)
	s.saveLocked()

	return &item
}

func (s *TodoStore) update(id, status string) (*TodoItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Items {
		if s.Items[i].ID == id {
			s.Items[i].Status = status
			s.Items[i].UpdatedAt = time.Now()
			s.saveLocked()
			return &s.Items[i], nil
		}
	}

	return nil, fmt.Errorf("todo item not found: %s", id)
}

func (s *TodoStore) remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Items {
		if s.Items[i].ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			s.saveLocked()
			return nil
		}
	}

	return fmt.Errorf("todo item not found: %s", id)
}

func (s *TodoStore) list() []TodoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]TodoItem, len(s.Items))
	copy(result, s.Items)
	return result
}

func (s *TodoStore) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Items = []TodoItem{}
	s.saveLocked()
}

func generateTodoID() string {
	return fmt.Sprintf("todo_%d", time.Now().UnixNano())
}

// TodoWriteTool implements the todo management tool
type TodoWriteTool struct{}

func (t *TodoWriteTool) Definition() api.Tool {
	return api.Tool{
		Name:        "todo_write",
		Description: "Manage a list of todos. Can add new todos, update existing ones, or clear the list. Useful for tracking tasks and maintaining context across long conversations. The todo list is persisted across sessions.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"todos": map[string]interface{}{
					"type":        "array",
					"description": "The complete list of todos to set. Replaces any existing todos.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"description": "Unique identifier for the todo. Use existing IDs to update, or omit to create new.",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "The todo description",
							},
							"status": map[string]interface{}{
								"type":        "string",
								"description": "Status: 'pending', 'in_progress', or 'done'",
								"enum":        []string{"pending", "in_progress", "done"},
							},
						},
						"required": []string{"description", "status"},
					},
				},
			},
			"required": []string{"todos"},
		},
	}
}

func (t *TodoWriteTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	store := getTodoStore()

	todosRaw, ok := m["todos"].([]interface{})
	if !ok {
		return "", fmt.Errorf("todos array is required")
	}

	// Clear existing todos
	store.clear()

	// Add new todos
	var added []TodoItem
	for _, todoRaw := range todosRaw {
		todoMap, ok := todoRaw.(map[string]interface{})
		if !ok {
			continue
		}

		desc, _ := todoMap["description"].(string)
		if desc == "" {
			continue
		}

		item := store.add(desc)

		// Update status if specified
		if status, ok := todoMap["status"].(string); ok {
			store.update(item.ID, status)
			item.Status = status
		}

		// Update ID if specified (preserve from existing)
		if id, ok := todoMap["id"].(string); ok && id != "" {
			item.ID = id
		}

		added = append(added, *item)
	}

	// Format result
	var result string
	result = fmt.Sprintf("Updated todo list (%d items):\n\n", len(added))

	pending := 0
	inProgress := 0
	done := 0

	for _, item := range added {
		switch item.Status {
		case "pending":
			pending++
			result += fmt.Sprintf("[ ] %s\n", item.Description)
		case "in_progress":
			inProgress++
			result += fmt.Sprintf("[~] %s\n", item.Description)
		case "done":
			done++
			result += fmt.Sprintf("[x] %s\n", item.Description)
		}
	}

	result += fmt.Sprintf("\nSummary: %d pending, %d in progress, %d done\n", pending, inProgress, done)

	return result, nil
}
