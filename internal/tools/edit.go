package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

type EditTool struct{}

func (t *EditTool) Definition() api.Tool {
	return api.Tool{
		Name:        "edit",
		Description: "Apply a precise surgical change to a file by replacing exactly ONE unique occurrence of old_string with new_string. Use this for targeted modifications like fixing a specific line, updating a variable name, or changing a configuration value. The old_string must appear EXACTLY once in the file - if it appears multiple times, the operation will fail. For multiple changes, make multiple separate edit calls. Use write_file for creating new files or complete rewrites instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to the file to edit",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "The exact literal text to replace",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "The replacement text",
				},
			},
			"required": []string{"path", "old_string", "new_string"},
		},
	}
}

func (t *EditTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	path, ok := m["path"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: path string expected")
	}
	oldStr, ok := m["old_string"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: old_string expected")
	}
	newStr, ok := m["new_string"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: new_string expected")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fullText := string(content)
	count := strings.Count(fullText, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file")
	}
	if count > 1 {
		return "", fmt.Errorf("old_string is ambiguous (%d occurrences found). Please provide more context.", count)
	}

	newContent := strings.Replace(fullText, oldStr, newStr, 1)
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully edited %s", path), nil
}
