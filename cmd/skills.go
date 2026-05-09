package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsInfoCmd)
	skillsCmd.AddCommand(skillsRunCmd)
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Discover and run available skills",
	Long:  `List, search, and execute available skills for common tasks.`,
}

var skillsListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List available skills",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Available Skills ===")

		skills := []struct {
			Name        string
			Category    string
			Description string
		}{
			{"git-commit", "git", "Generate conventional commit messages"},
			{"git-pr", "git", "Create pull request descriptions"},
			{"code-review", "code", "Review code for issues and improvements"},
			{"refactor", "code", "Refactor code with best practices"},
			{"debug", "code", "Debug errors and suggest fixes"},
			{"test", "code", "Generate unit tests for functions"},
			{"doc", "code", "Generate documentation for code"},
			{"explain", "code", "Explain complex code in simple terms"},
			{"optimize", "code", "Optimize code for performance"},
			{"security", "code", "Security review and hardening"},
			{"api-design", "design", "Design API endpoints and schemas"},
			{"db-schema", "design", "Design database schemas"},
			{"docker", "devops", "Create Docker configurations"},
			{"ci-cd", "devops", "Set up CI/CD pipelines"},
			{"deploy", "devops", "Generate deployment scripts"},
			{"onboard", "project", "Project onboarding and setup"},
			{"structure", "project", "Suggest project structure"},
		}

		for _, s := range skills {
			fmt.Printf("  %-15s [%-8s] %s\n", s.Name, s.Category, s.Description)
		}

		fmt.Println("\nUsage: claudego skills run [skill-name]")
	},
}

var skillsInfoCmd = &cobra.Command{
	Use:   "info [skill-name]",
	Short: "Show detailed information about a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]

		skills := map[string]string{
			"git-commit": `Generate conventional commit messages based on git diff.

This skill analyzes your staged changes and suggests appropriate
conventional commit messages following the Conventional Commits spec.

Example usage:
  claudego skills run git-commit`,
			"code-review": `Review code for bugs, style issues, and improvements.

This skill reads your code files and provides detailed feedback on:
- Potential bugs and errors
- Code style and best practices
- Performance optimizations
- Security considerations

Example usage:
  claudego skills run code-review`,
			"refactor": `Refactor code to improve readability and maintainability.

This skill analyzes your code and suggests refactoring improvements
following industry best practices and design patterns.

Example usage:
  claudego skills run refactor`,
			"test": `Generate comprehensive unit tests for your code.

This skill analyzes functions and generates appropriate test cases
covering normal cases, edge cases, and error conditions.

Example usage:
  claudego skills run test`,
		}

		if info, ok := skills[skillName]; ok {
			fmt.Printf("=== Skill: %s ===\n\n", skillName)
			fmt.Println(info)
		} else {
			fmt.Printf("Skill '%s' not found. Run 'claudego skills list' to see available skills.\n", skillName)
		}
	},
}

var skillsRunCmd = &cobra.Command{
	Use:   "run [skill-name]",
	Short: "Run a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]

		// Check if skill exists
		skillsDir := filepath.Join(config.GetConfigDir(), "skills")
		skillPath := filepath.Join(skillsDir, skillName+".md")

		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			// Use built-in skill
			switch skillName {
			case "git-commit":
				runGitCommitSkill()
			case "code-review":
				runCodeReviewSkill()
			default:
				fmt.Printf("Skill '%s' not found.\n", skillName)
				fmt.Println("Run 'claudego skills list' to see available skills.")
			}
			return
		}

		// Read and execute custom skill
		data, err := os.ReadFile(skillPath)
		if err != nil {
			fmt.Printf("Error reading skill: %v\n", err)
			return
		}

		fmt.Printf("Running skill: %s\n\n", skillName)
		fmt.Println(string(data))
	},
}

func runGitCommitSkill() {
	fmt.Println("📝 Git Commit Skill")
	fmt.Println("\nAnalyzing staged changes...")

	// Get git diff
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("No staged changes found.")
		fmt.Println("Stage changes with: git add <files>")
		return
	}

	if len(output) == 0 {
		fmt.Println("No staged changes to analyze.")
		return
	}

	fmt.Printf("\nFound %d bytes of changes.\n", len(output))
	fmt.Println("\nSuggested commit message types:")
	fmt.Println("  feat:     New feature")
	fmt.Println("  fix:      Bug fix")
	fmt.Println("  docs:     Documentation changes")
	fmt.Println("  style:    Code style changes")
	fmt.Println("  refactor: Code refactoring")
	fmt.Println("  test:     Test changes")
	fmt.Println("  chore:    Maintenance tasks")
	fmt.Println("\nRun 'claudego chat' and ask Claude to suggest a commit message.")
}

func runCodeReviewSkill() {
	fmt.Println("🔍 Code Review Skill")
	fmt.Println("\nTo review code:")
	fmt.Println("1. Add files to context: claudego add <files>")
	fmt.Println("2. Or use: claudego review [files...]")
	fmt.Println("3. Start chat and ask for code review")
}
