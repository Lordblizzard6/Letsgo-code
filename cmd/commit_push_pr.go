package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(commitPushPrCmd)
	commitPushPrCmd.Flags().BoolP("draft", "d", false, "Create as draft PR")
	commitPushPrCmd.Flags().StringP("title", "t", "", "PR title (auto-generated if not provided)")
	commitPushPrCmd.Flags().StringP("base", "b", "main", "Base branch for PR")
}

var commitPushPrCmd = &cobra.Command{
	Use:   "commit-push-pr",
	Short: "Commit, push and create PR in one command",
	Long: `Complete git workflow in one command:
1. Stage all changes
2. Generate commit message (or use provided)
3. Commit
4. Push to remote
5. Create Pull Request`,
	Run: func(cmd *cobra.Command, args []string) {
		draft, _ := cmd.Flags().GetBool("draft")
		title, _ := cmd.Flags().GetString("title")
		base, _ := cmd.Flags().GetString("base")

		// Check if we're in a git repo
		if _, err := exec.LookPath("git"); err != nil {
			fmt.Println("Error: git not found")
			os.Exit(1)
		}

		// Stage all changes
		fmt.Println("📦 Staging changes...")
		if out, err := exec.Command("git", "add", "-A").CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "Error staging: %s\n", out)
			os.Exit(1)
		}

		// Check if there are changes to commit
		status, _ := exec.Command("git", "diff", "--cached", "--name-only").Output()
		if len(strings.TrimSpace(string(status))) == 0 {
			fmt.Println("No changes to commit")
			os.Exit(0)
		}

		// Get commit message
		commitMsg := "Update files via Claude Code"
		if len(args) > 0 {
			commitMsg = strings.Join(args, " ")
		}

		// Commit
		fmt.Printf("📝 Committing: %s\n", commitMsg)
		if out, err := exec.Command("git", "commit", "-m", commitMsg).CombinedOutput(); err != nil {
			// Check if nothing to commit
			if strings.Contains(string(out), "nothing to commit") {
				fmt.Println("Nothing to commit")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error committing: %s\n", out)
			os.Exit(1)
		}

		// Get current branch
		branch, err := exec.Command("git", "branch", "--show-current").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting branch: %v\n", err)
			os.Exit(1)
		}
		currentBranch := strings.TrimSpace(string(branch))

		// Push
		fmt.Printf("🚀 Pushing to %s...\n", currentBranch)
		if out, err := exec.Command("git", "push", "origin", currentBranch).CombinedOutput(); err != nil {
			// Try setting upstream
			if strings.Contains(string(out), "no upstream") {
				out, err = exec.Command("git", "push", "-u", "origin", currentBranch).CombinedOutput()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error pushing: %s\n", out)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error pushing: %s\n", out)
				os.Exit(1)
			}
		}

		// Create PR if gh CLI is available
		if _, err := exec.LookPath("gh"); err == nil {
			fmt.Println("📋 Creating PR...")

			prTitle := title
			if prTitle == "" {
				prTitle = commitMsg
			}

			ghArgs := []string{"pr", "create", "--title", prTitle, "--base", base}
			if draft {
				ghArgs = append(ghArgs, "--draft")
			}

			if out, err := exec.Command("gh", ghArgs...).CombinedOutput(); err != nil {
				// Check if PR already exists
				if strings.Contains(string(out), "already exists") {
					fmt.Println("PR already exists for this branch")
				} else {
					fmt.Printf("⚠️  Could not create PR: %s\n", out)
				}
			} else {
				fmt.Println("✅ PR created successfully!")
				fmt.Println(string(out))
			}
		} else {
			fmt.Println("⚠️  GitHub CLI (gh) not found. Install it to create PRs automatically.")
		}

		fmt.Println("\n✅ Workflow complete!")
	},
}
