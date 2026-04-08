package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// WorkflowTool executes workflow scripts
type WorkflowTool struct{}

func (t *WorkflowTool) Name() string {
	return "workflow"
}

func (t *WorkflowTool) Description() string {
	return "Execute a predefined workflow or script. Use this to run common automation tasks, build processes, or custom scripts defined in the .letsGo/workflows directory."
}

func (t *WorkflowTool) Definition() api.Tool {
	return api.Tool{
		Name:        "workflow",
		Description: "Execute a predefined workflow or script",
		InputSchema: t.InputSchema(),
	}
}

func (t *WorkflowTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the workflow to execute",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Arguments to pass to the workflow",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"cwd": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the workflow",
			},
		},
		"required": []string{"name"},
	}
}

func (t *WorkflowTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	return t.executeInternal(m)
}

func (t *WorkflowTool) executeInternal(input map[string]interface{}) (string, error) {
	name, ok := input["name"].(string)
	if !ok {
		return "", fmt.Errorf("workflow name is required")
	}

	// Find workflow
	workflowsDir := t.getWorkflowsDir()
	workflowPath := filepath.Join(workflowsDir, name+".json")

	// Check for shell script version
	scriptPath := filepath.Join(workflowsDir, name+".sh")
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(workflowsDir, name+".ps1")
	}

	// Try JSON workflow first
	if _, err := os.Stat(workflowPath); err == nil {
		return t.executeJSONWorkflow(workflowPath, input)
	}

	// Try shell script
	if _, err := os.Stat(scriptPath); err == nil {
		return t.executeScriptWorkflow(scriptPath, input)
	}

	// Try built-in workflows
	return t.executeBuiltinWorkflow(name, input)
}

func (t *WorkflowTool) getWorkflowsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "workflows")
}

func (t *WorkflowTool) executeJSONWorkflow(path string, input map[string]interface{}) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var workflow struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Steps       []string          `json:"steps"`
		Env         map[string]string `json:"env"`
	}

	if err := json.Unmarshal(data, &workflow); err != nil {
		return "", err
	}

	// Get cwd from input or use default
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Execute steps
	var outputs []string
	for i, step := range workflow.Steps {
		// Replace placeholders
		step = strings.ReplaceAll(step, "{{cwd}}", cwd)

		// Execute step
		parts := strings.Fields(step)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = cwd

		// Set env vars
		for k, v := range workflow.Env {
			cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("step %d failed: %v\n%s", i+1, err, string(output))
		}

		outputs = append(outputs, string(output))
	}

	return fmt.Sprintf("Workflow '%s' completed successfully.\n\nOutput:\n%s",
		workflow.Name, strings.Join(outputs, "\n---\n")), nil
}

func (t *WorkflowTool) executeScriptWorkflow(path string, input map[string]interface{}) (string, error) {
	// Get args
	var args []string
	if a, ok := input["args"].([]interface{}); ok {
		for _, arg := range a {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	// Get cwd
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Execute script
	var cmd *exec.Cmd
	if strings.HasSuffix(path, ".ps1") {
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", path)
		cmd.Args = append(cmd.Args, args...)
	} else {
		cmd = exec.Command("bash", path)
		cmd.Args = append(cmd.Args, args...)
	}

	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("script failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

func (t *WorkflowTool) executeBuiltinWorkflow(name string, input map[string]interface{}) (string, error) {
	// Built-in workflows
	workflows := map[string]func(map[string]interface{}) (string, error){
		"build":  t.buildWorkflow,
		"test":   t.testWorkflow,
		"lint":   t.lintWorkflow,
		"format": t.formatWorkflow,
		"clean":  t.cleanWorkflow,
	}

	if fn, ok := workflows[name]; ok {
		return fn(input)
	}

	return "", fmt.Errorf("workflow '%s' not found. Create it in ~/.letsGo/workflows/", name)
}

func (t *WorkflowTool) buildWorkflow(input map[string]interface{}) (string, error) {
	// Detect project type and build
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Check for common build files
	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		cmd := exec.Command("go", "build", "./...")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("build failed: %v\n%s", err, string(output))
		}
		return "Go build completed successfully.\n" + string(output), nil
	}

	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		cmd := exec.Command("npm", "run", "build")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("build failed: %v\n%s", err, string(output))
		}
		return "npm build completed successfully.\n" + string(output), nil
	}

	if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		cmd := exec.Command("cargo", "build")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("build failed: %v\n%s", err, string(output))
		}
		return "Cargo build completed successfully.\n" + string(output), nil
	}

	if _, err := os.Stat(filepath.Join(cwd, "Makefile")); err == nil {
		cmd := exec.Command("make")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("build failed: %v\n%s", err, string(output))
		}
		return "Make build completed successfully.\n" + string(output), nil
	}

	return "", fmt.Errorf("could not detect project type for build workflow")
}

func (t *WorkflowTool) testWorkflow(input map[string]interface{}) (string, error) {
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		cmd := exec.Command("go", "test", "./...")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		cmd := exec.Command("npm", "test")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		cmd := exec.Command("cargo", "test")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	hasPytestIni := false
	if _, err := os.Stat(filepath.Join(cwd, "pytest.ini")); err == nil {
		hasPytestIni = true
	}

	hasSetupPy := false
	if _, err := os.Stat(filepath.Join(cwd, "setup.py")); err == nil {
		hasSetupPy = true
	}

	if hasPytestIni || hasSetupPy {
		cmd := exec.Command("python", "-m", "pytest")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	return "", fmt.Errorf("could not detect project type for test workflow")
}

func (t *WorkflowTool) lintWorkflow(input map[string]interface{}) (string, error) {
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		// Try golangci-lint first, then go vet
		cmd := exec.Command("golangci-lint", "run")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		if err != nil && len(output) == 0 {
			// Fall back to go vet
			cmd = exec.Command("go", "vet", "./...")
			cmd.Dir = cwd
			output, err = cmd.CombinedOutput()
		}
		return string(output), err
	}

	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		cmd := exec.Command("npm", "run", "lint")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	return "", fmt.Errorf("could not detect project type for lint workflow")
}

func (t *WorkflowTool) formatWorkflow(input map[string]interface{}) (string, error) {
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		cmd := exec.Command("go", "fmt", "./...")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		cmd := exec.Command("npx", "prettier", "--write", ".")
		cmd.Dir = cwd
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	return "", fmt.Errorf("could not detect project type for format workflow")
}

func (t *WorkflowTool) cleanWorkflow(input map[string]interface{}) (string, error) {
	cwd, _ := input["cwd"].(string)
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	patterns := []string{
		"*.tmp",
		"*.log",
		".DS_Store",
		"Thumbs.db",
		"node_modules/.cache",
		"__pycache__",
		"*.pyc",
		"*.class",
		"target/debug",
		"target/release",
		"dist",
		"build",
	}

	var removed []string
	var errs []error
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(cwd, pattern))
		if err != nil {
			errs = append(errs, fmt.Errorf("pattern %s: %w", pattern, err))
			continue
		}
		for _, match := range matches {
			if err := os.RemoveAll(match); err != nil {
				errs = append(errs, fmt.Errorf("remove %s: %w", match, err))
				continue
			}
			removed = append(removed, match)
		}
	}

	if len(errs) > 0 {
		return fmt.Sprintf("Cleaned %d files/directories with %d errors", len(removed), len(errs)), fmt.Errorf("errors during cleanup: %v", errs)
	}

	return fmt.Sprintf("Cleaned %d files/directories", len(removed)), nil
}
