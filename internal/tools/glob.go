package tools

import (
	"fmt"
	"path/filepath"
	"github.com/user/go-claude-code/internal/api"
)

type GlobTool struct{}

func (t *GlobTool) Definition() api.Tool {
	return api.Tool{
		Name:        "glob",
		Description: "Find files matching a specific pattern (e.g. **/*.go)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The glob pattern to match against",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

func (t *GlobTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}
	pattern, _ := m["pattern"].(string)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "No files matched the pattern.", nil
	}

	result := ""
	for _, m := range matches {
		result += m + "\n"
	}
	return result, nil
}
