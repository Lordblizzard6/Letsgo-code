package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AGENTS.md rules and project configuration for LetsGO Code",
	Long:  `Analyzes the current workspace tech stack and creates an AGENTS.md file with conventions, test commands, and non-negotiables for the agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			return
		}
		msg, err := InitializeProjectRules(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Println(msg)
	},
}

// InitializeProjectRules creates an AGENTS.md if none exists, detecting the project stack.
func InitializeProjectRules(cwd string) (string, error) {
	agentsPath := filepath.Join(cwd, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		return fmt.Sprintf("ℹ️ AGENTS.md already exists at %s", agentsPath), nil
	}

	// Detect stack
	stack := "Generic / Unknown"
	installCmd := "TODO"
	buildCmd := "TODO"
	testCmd := "TODO"
	lintCmd := "TODO"

	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		stack = "Go"
		installCmd = "go mod download"
		buildCmd = "go build ./..."
		testCmd = "go test ./..."
		lintCmd = "golangci-lint run"
	} else if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		stack = "Node.js / TypeScript / JavaScript"
		installCmd = "npm install"
		buildCmd = "npm run build"
		testCmd = "npm test"
		lintCmd = "npm run lint"
	} else if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		stack = "Rust"
		installCmd = "cargo fetch"
		buildCmd = "cargo build"
		testCmd = "cargo test"
		lintCmd = "cargo clippy"
	} else if _, err := os.Stat(filepath.Join(cwd, "pyproject.toml")); err == nil {
		stack = "Python"
		installCmd = "pip install -e ."
		buildCmd = "python -m build"
		testCmd = "pytest"
		lintCmd = "flake8"
	} else if _, err := os.Stat(filepath.Join(cwd, "pubspec.yaml")); err == nil {
		stack = "Flutter / Dart"
		installCmd = "flutter pub get"
		buildCmd = "flutter build"
		testCmd = "flutter test"
		lintCmd = "flutter analyze"
	}

	template := fmt.Sprintf(`# AGENTS.md

Operating instructions and project rules for coding agents in this repository.

## 0. Non-negotiables
1. Working code only. Finish the job. Plausibility is not correctness.
2. Direct, concise communication. No flattery, no filler.
3. Surgical changes: Touch only what you must. No drive-by refactoring or unnecessary format changes.
4. Goal-driven: Verify your changes with tests or build commands before reporting completion.

## 1. Project Context
- **Stack**: %s
- **Install**: %s
- **Build**: %s
- **Test**: %s
- **Lint**: %s

## 2. Conventions & Style
- Follow existing patterns in the codebase.
- Error handling: return errors to callers, do not silently swallow them.
- Simplicity first: write the minimum code required to solve the task.
`, stack, installCmd, buildCmd, testCmd, lintCmd)

	if err := os.WriteFile(agentsPath, []byte(strings.TrimSpace(template)+"\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	return fmt.Sprintf("✅ Created %s for stack %s", agentsPath, stack), nil
}
