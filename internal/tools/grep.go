package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"github.com/user/go-claude-code/internal/api"
)

type GrepTool struct{}

func (t *GrepTool) Definition() api.Tool {
	return api.Tool{
		Name:        "grep",
		Description: "Search for text patterns in files recursively. Use this to find specific code, function definitions, variable usage, error messages, or any text pattern across multiple files. Supports regex patterns. Automatically adapts to use PowerShell on Windows or standard grep on Unix/Linux systems. Returns matching lines with file paths and line numbers.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "The regex pattern to search for. Can include regex syntax like .* for wildcards, [abc] for character classes, or ^/$ for line anchors. Example: 'func.*Foo' to find function definitions.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory or file to search in. Use '.' for current directory or specify a subdirectory to narrow search scope.",
				},
			},
			"required": []string{"pattern"},
		},
	}
}

func (t *GrepTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}
	pattern, _ := m["pattern"].(string)
	path, ok := m["path"].(string)
	if !ok {
		path = "."
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Use Select-String in PowerShell for Windows
		cmd = exec.Command("powershell", "-Command", fmt.Sprintf("Get-ChildItem -Recurse -Path %s | Select-String -Pattern '%s'", path, pattern))
	} else {
		// Use standard grep for Unix
		cmd = exec.Command("grep", "-rnE", pattern, path)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matches found.", nil
		}
		return "", err
	}

	return string(out), nil
}
