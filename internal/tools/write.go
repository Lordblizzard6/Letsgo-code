package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/user/go-claude-code/internal/api"
)

type WriteFileTool struct{}

func (t *WriteFileTool) Definition() api.Tool {
	return api.Tool{
		Name:        "write_file",
		Description: "Create a new file or completely overwrite an existing file with full content. Use this to write source code, configuration files, documentation, or any text content. Automatically creates parent directories if they don't exist. WARNING: This will overwrite existing files completely without confirmation. Use 'edit' for partial modifications instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The full file path to create or overwrite. Supports relative paths from current directory or absolute paths. Parent directories will be created automatically if they don't exist.",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The complete content to write to the file. Must be the entire file content, not just a diff or partial update.",
				},
			},
			"required": []string{"path", "content"},
		},
	}
}

func (t *WriteFileTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	path, _ := m["path"].(string)
	if path == "" {
		if p, ok := m["file_path"].(string); ok {
			path = p
		} else if p, ok := m["filepath"].(string); ok {
			path = p
		} else if p, ok := m["file"].(string); ok {
			path = p
		}
	}
	if path == "" {
		return "", fmt.Errorf("failed to write file: missing path argument")
	}

	content, _ := m["content"].(string)
	if content == "" {
		if c, ok := m["text"].(string); ok {
			content = c
		}
	}

	// Create directories if they don't exist
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directories: %w", err)
		}
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote file to %s", path), nil
}
