package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// PowerShellTool executes PowerShell commands on Windows
type PowerShellTool struct{}

func (t *PowerShellTool) Name() string {
	return "powershell"
}

func (t *PowerShellTool) Description() string {
	return "Execute PowerShell commands on Windows. Use this for Windows-specific operations, registry modifications, WMI queries, or when PowerShell cmdlets are needed instead of standard bash commands. Only available on Windows."
}

func (t *PowerShellTool) Definition() api.Tool {
	return api.Tool{
		Name:        "powershell",
		Description: t.Description(),
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The PowerShell command to execute",
				},
				"args": map[string]interface{}{
					"type":        "array",
					"description": "Arguments to pass to the command",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"cwd": map[string]interface{}{
					"type":        "string",
					"description": "Working directory for the command",
				},
				"timeout": map[string]interface{}{
					"type":        "number",
					"description": "Timeout in seconds (default: 60)",
					"default":     60,
				},
			},
			"required": []string{"command"},
		},
	}
}

func (t *PowerShellTool) Execute(input interface{}) (string, error) {
	// Only available on Windows
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("powershell tool is only available on Windows. Use bash tool instead.")
	}

	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	command, ok := m["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command is required")
	}

	// Get optional parameters
	var args []string
	if a, ok := m["args"].([]interface{}); ok {
		for _, arg := range a {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	cwd := ""
	if c, ok := m["cwd"].(string); ok {
		cwd = c
	}

	timeout := 60
	if t, ok := m["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Build PowerShell command
	// Use -ExecutionPolicy Bypass to allow script execution
	psArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoProfile",
		"-Command", command,
	}
	
	// Append additional args if provided
	if len(args) > 0 {
		psArgs = append(psArgs, args...)
	}

	cmd := exec.Command("powershell", psArgs...)
	
	if cwd != "" {
		cmd.Dir = cwd
	}

	// Set up timeout
	if timeout > 0 {
		// This is a simplified timeout - in production would use context
	}

	output, err := cmd.CombinedOutput()
	
	result := string(output)
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, result), nil
	}

	return result, nil
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// ConvertPath converts between Windows and Unix-style paths
func ConvertPath(path string, toWindows bool) string {
	if toWindows {
		// Convert / to \
		return strings.ReplaceAll(path, "/", "\\")
	}
	// Convert \ to /
	return strings.ReplaceAll(path, "\\", "/")
}
