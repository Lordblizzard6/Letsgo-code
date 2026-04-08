package tools

import (
	"fmt"
	"io"
	"os"

	"github.com/user/go-claude-code/internal/api"
)

type CatTool struct{}

func (t *CatTool) Definition() api.Tool {
	return api.Tool{
		Name:        "cat",
		Description: "Read and display the contents of a text file. Use this tool to view source code, configuration files, documentation, or any text-based file content. Returns the full file content as a string. Best for files under 500KB. For larger files, consider using grep to search for specific content first.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The absolute or relative path to the file to read. Supports relative paths from current directory.",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (t *CatTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	path, ok := m["path"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: path string expected")
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
