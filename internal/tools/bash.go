package tools

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/user/go-claude-code/internal/api"
)

type BashTool struct{}

func (t *BashTool) Definition() api.Tool {
	return api.Tool{
		Name:        "bash",
		Description: "Execute bash shell commands on Unix/Linux systems. Use this for command-line operations like listing directories, installing packages, running scripts, checking system info, or any task that requires shell access. The command executes in the current working directory. WARNING: This tool can execute arbitrary commands. Dangerous commands like 'rm -rf /', force pushes, or destructive database operations are blocked. Always check with the user before hard-to-reverse operations.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The bash command to execute. Can include pipes, redirections, and multiple commands separated by semicolons or &&.",
				},
			},
			"required": []string{"command"},
		},
	}
}

func (t *BashTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	command, ok := m["command"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: command string expected")
	}

	// Check for dangerous commands
	isDangerous, warnings, severity := CheckCommandSafety(command)
	if isDangerous {
		if severity == "high" {
			return "", fmt.Errorf("BLOCKED: Command matches high-risk pattern\n%s", FormatSafetyWarning(command, warnings, severity))
		}
		// For medium/low severity, still execute but include warning
		// The UI layer should handle asking for permission
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	result := string(output)

	// Include safety warning in output if dangerous
	if isDangerous && severity != "high" {
		result = FormatSafetyWarning(command, warnings, severity) + "\n\nOutput:\n" + result
	}

	if err != nil {
		return result, err
	}

	return result, nil
}
