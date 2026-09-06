package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(reviewCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show changes made in this session",
	Long:  `Display a diff of all file changes made during the current conversation session.`,
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if session == nil || session.ProjectPath == "" {
			// Try to get current directory
			session = &db.Session{ProjectPath: "."}
		}

		// Check if git repo
		gitCmd := exec.Command("git", "-C", session.ProjectPath, "status", "--porcelain")
		output, err := gitCmd.Output()
		if err != nil {
			fmt.Println("Not a git repository or git not available.")
			return
		}

		if len(output) == 0 {
			fmt.Println("No changes detected.")
			return
		}

		// Show git diff
		diffCmd := exec.Command("git", "-C", session.ProjectPath, "diff")
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		diffCmd.Run()
	},
}

var commitCmd = &cobra.Command{
	Use:   "commit [message]",
	Short: "Commit changes with AI-generated or custom message",
	Long: `Commit all staged and unstaged changes to git.
If no message is provided, Claude will suggest a commit message based on the changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		message := ""
		if len(args) > 0 {
			message = strings.Join(args, " ")
		}

		// Get current directory
		cwd, _ := os.Getwd()

		// Check if git repo
		gitCmd := exec.Command("git", "-C", cwd, "status", "--porcelain")
		output, err := gitCmd.Output()
		if err != nil {
			fmt.Println("Not a git repository or git not available.")
			return
		}

		if len(output) == 0 {
			fmt.Println("No changes to commit.")
			return
		}

		// Stage all changes
		addCmd := exec.Command("git", "-C", cwd, "add", "-A")
		if err := addCmd.Run(); err != nil {
			fmt.Printf("Error staging changes: %v\n", err)
			return
		}

		// If no message, try to generate one
		if message == "" {
			fmt.Println("Analyzing changes to suggest commit message...")
			// In full implementation, this would call the API to analyze diff
			// For now, use a default
			message = "Update files via Claude Code"
		}

		// Commit
		commitGitCmd := exec.Command("git", "-C", cwd, "commit", "-m", message)
		commitGitCmd.Stdout = os.Stdout
		commitGitCmd.Stderr = os.Stderr
		if err := commitGitCmd.Run(); err != nil {
			fmt.Printf("Error committing: %v\n", err)
			return
		}

		fmt.Println("\n✓ Changes committed successfully.")
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review [files...]",
	Short: "Review code for issues and improvements",
	Long: `Have Claude review your code for bugs, style issues, and improvements.
If no files are specified, reviews all changes in the current session.`,
	Run: func(cmd *cobra.Command, args []string) {
		cwd, _ := os.Getwd()

		var filesToReview []string

		if len(args) > 0 {
			filesToReview = args
		} else {
			// Try to get changed files from git
			changedCmd := exec.Command("git", "-C", cwd, "diff", "--name-only")
			output, err := changedCmd.Output()
			if err == nil && len(output) > 0 {
				filesToReview = strings.Split(strings.TrimSpace(string(output)), "\n")
			}
		}

		if len(filesToReview) == 0 {
			fmt.Println("No files specified and no changes detected.")
			fmt.Println("Usage: letsGo review [files...]")
			return
		}

		fmt.Printf("Reviewing %d file(s):\n", len(filesToReview))
		for _, f := range filesToReview {
			fmt.Printf("  - %s\n", f)
		}

		fmt.Println("\nNote: In a full implementation, this would:")
		fmt.Println("1. Read each file")
		fmt.Println("2. Send to Claude for review with a code review prompt")
		fmt.Println("3. Display findings (bugs, style issues, suggestions)")
		fmt.Println("4. Optionally create inline comments or PR review")
	},
}
