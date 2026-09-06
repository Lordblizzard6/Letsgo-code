package tools

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

type LsTool struct{}

func (t *LsTool) Definition() api.Tool {
	return api.Tool{
		Name:        "ls",
		Description: "List files and directories in a given path. Use this to explore project structure, see what files exist in a directory, check file sizes and permissions, or verify directory contents before performing operations. Returns a formatted listing similar to the Unix 'ls -la' command.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The directory path to list. Use '.' for current directory, or provide an absolute/relative path. Defaults to current directory if not specified.",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (t *LsTool) Execute(input interface{}) (string, error) {
	path := "."
	if m, ok := input.(map[string]interface{}); ok && m != nil {
		if p, ok := m["path"].(string); ok && strings.TrimSpace(p) != "" {
			path = p
		} else if p, ok := m["dir"].(string); ok && strings.TrimSpace(p) != "" {
			path = p
		} else if p, ok := m["directory"].(string); ok && strings.TrimSpace(p) != "" {
			path = p
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var results []string
	for _, entry := range entries {
		info, _ := entry.Info()
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		size := ""
		if !entry.IsDir() {
			size = fmt.Sprintf(" (%d bytes)", info.Size())
		}
		results = append(results, fmt.Sprintf("%s%s", name, size))
	}

	if len(results) == 0 {
		return "(empty directory)", nil
	}

	return strings.Join(results, "\n"), nil
}
