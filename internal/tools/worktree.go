package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/user/go-claude-code/internal/api"
)

// EnterWorktreeTool enters a git worktree
type EnterWorktreeTool struct{}

func (t *EnterWorktreeTool) Definition() api.Tool {
	return api.Tool{
		Name:        "enter_worktree",
		Description: "Enter a git worktree directory to work on a different branch simultaneously",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the worktree directory",
				},
				"branch": map[string]interface{}{
					"type":        "string",
					"description": "Branch to checkout in the worktree (will create if doesn't exist)",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (t *EnterWorktreeTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: expected object")
	}
	path, ok := m["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required")
	}

	branch := ""
	if b, ok := m["branch"].(string); ok {
		branch = b
	}

	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "", fmt.Errorf("not in a git repository")
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if worktree already exists
	if _, err := os.Stat(absPath); !os.IsNotExist(err) {
		// Worktree exists, just enter it
		return fmt.Sprintf("✓ Entering existing worktree at %s\n\nUse 'cd %s' to switch to it", absPath, absPath), nil
	}

	// Create worktree
	var cmd *exec.Cmd
	if branch != "" {
		// Check if branch exists
		branchCheck := exec.Command("git", "branch", "--list", branch)
		output, err := branchCheck.Output()

		if err != nil || len(output) == 0 {
			// Create new branch
			cmd = exec.Command("git", "worktree", "add", "-b", branch, absPath)
		} else {
			// Use existing branch
			cmd = exec.Command("git", "worktree", "add", absPath, branch)
		}
	} else {
		cmd = exec.Command("git", "worktree", "add", absPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: %w\n%s", err, string(output))
	}

	result := fmt.Sprintf("✓ Created worktree at %s", absPath)
	if branch != "" {
		result += fmt.Sprintf(" (branch: %s)", branch)
	}
	result += "\n\nYou can now:"
	result += fmt.Sprintf("\n  cd %s", absPath)
	result += "\n  work on a different branch simultaneously"
	result += "\n  changes are isolated from the main working tree"

	return result, nil
}

// ExitWorktreeTool exits a git worktree
type ExitWorktreeTool struct{}

func (t *ExitWorktreeTool) Definition() api.Tool {
	return api.Tool{
		Name:        "exit_worktree",
		Description: "Exit the current git worktree and optionally remove it",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"remove": map[string]interface{}{
					"type":        "boolean",
					"description": "Remove the worktree after exiting",
					"default":     false,
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force removal even if there are uncommitted changes",
					"default":     false,
				},
			},
			"required": []string{},
		},
	}
}

func (t *ExitWorktreeTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
	}
	remove := false
	if r, ok := m["remove"].(bool); ok {
		remove = r
	}

	force := false
	if f, ok := m["force"].(bool); ok {
		force = f
	}

	// Check if we're in a worktree
	cmd := exec.Command("git", "rev-parse", "--git-path", "..")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository: %w", err)
	}

	gitPath := string(output)
	if !filepath.IsAbs(gitPath) {
		// This is the main repo, not a worktree
		return "", fmt.Errorf("not in a worktree - this is the main repository")
	}

	// Get current worktree path
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get the main repository path
	mainRepoCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	mainRepoCmd.Dir = filepath.Dir(gitPath)
	mainRepoOutput, err := mainRepoCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find main repository: %w", err)
	}
	mainRepo := string(mainRepoOutput)

	result := fmt.Sprintf("Exiting worktree: %s\n", pwd)
	result += fmt.Sprintf("Main repository: %s\n", mainRepo)

	if remove {
		// Check for uncommitted changes
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusOutput, err := statusCmd.Output()

		if (err != nil || len(statusOutput) > 0) && !force {
			return "", fmt.Errorf("worktree has uncommitted changes. Use force=true to remove anyway")
		}

		// Remove worktree
		removeCmd := exec.Command("git", "worktree", "remove", pwd)
		if force {
			removeCmd = exec.Command("git", "worktree", "remove", "--force", pwd)
		}

		removeOutput, err := removeCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to remove worktree: %w\n%s", err, string(removeOutput))
		}

		result += "✓ Worktree removed\n"
	}

	result += fmt.Sprintf("\n👉 Switch to main repo: cd %s", mainRepo)

	return result, nil
}

// ListWorktreesTool lists all git worktrees
type ListWorktreesTool struct{}

func (t *ListWorktreesTool) Definition() api.Tool {
	return api.Tool{
		Name:        "list_worktrees",
		Description: "List all git worktrees in the current repository",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	}
}

func (t *ListWorktreesTool) Execute(input interface{}) (string, error) {
	// Check if we're in a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		// Try to find .git in parent directories
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("not in a git repository")
		}
	}

	cmd := exec.Command("git", "worktree", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w\n%s", err, string(output))
	}

	result := "📁 Git Worktrees:\n"
	result += "================\n\n"
	result += string(output)
	result += "\n💡 Use 'enter_worktree' to work on a different branch"

	return result, nil
}
