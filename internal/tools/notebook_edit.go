package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// NotebookCell represents a cell in a Jupyter notebook
type NotebookCell struct {
	CellType string      `json:"cell_type"`
	Source   interface{} `json:"source"` // Can be string or []string
	Metadata interface{} `json:"metadata,omitempty"`
	Outputs  interface{} `json:"outputs,omitempty"`
}

// Notebook represents a Jupyter notebook structure
type Notebook struct {
	Cells    []NotebookCell    `json:"cells"`
	Metadata map[string]interface{} `json:"metadata"`
	NbFormat int               `json:"nbformat"`
	NbFormatMinor int        `json:"nbformat_minor"`
}

// NotebookEditTool allows editing Jupyter notebooks
type NotebookEditTool struct{}

func (t *NotebookEditTool) Definition() api.Tool {
	return api.Tool{
		Name:        "notebook_edit",
		Description: "Edit a Jupyter notebook (.ipynb) file. Can add, modify, or delete cells. The notebook structure is preserved while allowing surgical edits to specific cells.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the .ipynb file",
				},
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform: 'add', 'edit', 'delete', 'read'",
					"enum":        []string{"add", "edit", "delete", "read"},
				},
				"cell_index": map[string]interface{}{
					"type":        "integer",
					"description": "Cell index (0-based) for edit/delete actions",
				},
				"cell_type": map[string]interface{}{
					"type":        "string",
					"description": "Cell type for add action: 'code' or 'markdown'",
					"enum":        []string{"code", "markdown"},
				},
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Cell source content for add/edit actions",
				},
			},
			"required": []string{"path", "action"},
		},
	}
}

func (t *NotebookEditTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	path, _ := m["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	action, _ := m["action"].(string)
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	// Read notebook
	notebook, err := t.readNotebook(path)
	if err != nil {
		if action == "read" {
			return "", err
		}
		// For other actions, create new notebook if doesn't exist
		notebook = &Notebook{
			Cells:         []NotebookCell{},
			Metadata:      make(map[string]interface{}),
			NbFormat:      4,
			NbFormatMinor: 2,
		}
	}

	switch action {
	case "read":
		return t.formatNotebook(notebook), nil
	case "add":
		return t.addCell(notebook, path, m)
	case "edit":
		return t.editCell(notebook, path, m)
	case "delete":
		return t.deleteCell(notebook, path, m)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *NotebookEditTool) readNotebook(path string) (*Notebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook: %w", err)
	}

	var notebook Notebook
	if err := json.Unmarshal(data, &notebook); err != nil {
		return nil, fmt.Errorf("failed to parse notebook: %w", err)
	}

	return &notebook, nil
}

func (t *NotebookEditTool) saveNotebook(path string, notebook *Notebook) error {
	data, err := json.MarshalIndent(notebook, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize notebook: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write notebook: %w", err)
	}

	return nil
}

func (t *NotebookEditTool) addCell(notebook *Notebook, path string, params map[string]interface{}) (string, error) {
	cellType, _ := params["cell_type"].(string)
	if cellType == "" {
		cellType = "code"
	}

	source, _ := params["source"].(string)

	cell := NotebookCell{
		CellType: cellType,
		Source:   source,
		Metadata: make(map[string]interface{}),
	}

	// Handle cell_index for insertion position
	cellIndex := -1 // Append by default
	if idx, ok := params["cell_index"].(float64); ok {
		cellIndex = int(idx)
	}

	if cellIndex >= 0 && cellIndex < len(notebook.Cells) {
		// Insert at position
		notebook.Cells = append(notebook.Cells[:cellIndex], append([]NotebookCell{cell}, notebook.Cells[cellIndex:]...)...)
	} else {
		// Append
		notebook.Cells = append(notebook.Cells, cell)
	}

	if err := t.saveNotebook(path, notebook); err != nil {
		return "", err
	}

	return fmt.Sprintf("Added %s cell at index %d (total: %d cells)", cellType, cellIndex, len(notebook.Cells)), nil
}

func (t *NotebookEditTool) editCell(notebook *Notebook, path string, params map[string]interface{}) (string, error) {
	cellIndex, ok := params["cell_index"].(float64)
	if !ok {
		return "", fmt.Errorf("cell_index is required for edit action")
	}

	idx := int(cellIndex)
	if idx < 0 || idx >= len(notebook.Cells) {
		return "", fmt.Errorf("cell index %d out of range (0-%d)", idx, len(notebook.Cells)-1)
	}

	source, _ := params["source"].(string)
	if source != "" {
		notebook.Cells[idx].Source = source
	}

	if cellType, ok := params["cell_type"].(string); ok {
		notebook.Cells[idx].CellType = cellType
	}

	if err := t.saveNotebook(path, notebook); err != nil {
		return "", err
	}

	return fmt.Sprintf("Edited cell %d (type: %s)", idx, notebook.Cells[idx].CellType), nil
}

func (t *NotebookEditTool) deleteCell(notebook *Notebook, path string, params map[string]interface{}) (string, error) {
	cellIndex, ok := params["cell_index"].(float64)
	if !ok {
		return "", fmt.Errorf("cell_index is required for delete action")
	}

	idx := int(cellIndex)
	if idx < 0 || idx >= len(notebook.Cells) {
		return "", fmt.Errorf("cell index %d out of range (0-%d)", idx, len(notebook.Cells)-1)
	}

	cellType := notebook.Cells[idx].CellType
	notebook.Cells = append(notebook.Cells[:idx], notebook.Cells[idx+1:]...)

	if err := t.saveNotebook(path, notebook); err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted %s cell at index %d (remaining: %d cells)", cellType, idx, len(notebook.Cells)), nil
}

func (t *NotebookEditTool) formatNotebook(notebook *Notebook) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Notebook: %d cells\n\n", len(notebook.Cells)))

	for i, cell := range notebook.Cells {
		result.WriteString(fmt.Sprintf("Cell %d (%s):\n", i, cell.CellType))

		source := t.getCellSource(cell)
		lines := strings.Split(source, "\n")

		// Show first 5 lines
		for j, line := range lines {
			if j >= 5 {
				result.WriteString(fmt.Sprintf("  ... (%d more lines)\n", len(lines)-5))
				break
			}
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			result.WriteString(fmt.Sprintf("  %s\n", line))
		}

		result.WriteString("\n")
	}

	return result.String()
}

func (t *NotebookEditTool) getCellSource(cell NotebookCell) string {
	switch s := cell.Source.(type) {
	case string:
		return s
	case []interface{}:
		var parts []string
		for _, part := range s {
			if str, ok := part.(string); ok {
				parts = append(parts, str)
			}
		}
		return strings.Join(parts, "")
	case []string:
		return strings.Join(s, "")
	default:
		return ""
	}
}
