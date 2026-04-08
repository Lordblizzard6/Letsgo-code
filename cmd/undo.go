package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(redoCmd)
}

var undoCmd = &cobra.Command{
	Use:   "undo [description]",
	Short: "Undo the last action using git",
	Long:  `Undo the last action by reverting the most recent git commit. Optionally provide a description for tracking.`,
	Run: func(cmd *cobra.Command, args []string) {
		description := "undo last action"
		if len(args) > 0 {
			description = strings.Join(args, " ")
		}

		// Check if we're in a git repo
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			fmt.Println("Error: Not a git repository. Undo requires git to track changes.")
			return
		}

		// Get the last commit hash
		lastCommitCmd := exec.Command("git", "rev-parse", "HEAD")
		lastCommitOut, err := lastCommitCmd.Output()
		if err != nil {
			fmt.Printf("Error getting last commit: %v\n", err)
			return
		}
		lastCommit := strings.TrimSpace(string(lastCommitOut))

		// Check if there are commits to undo
		if lastCommit == "" {
			fmt.Println("No commits to undo.")
			return
		}

		// Get commit message for display
		msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
		msgOut, _ := msgCmd.Output()
		commitMsg := strings.TrimSpace(string(msgOut))

		fmt.Printf("Undoing commit: %s\n", lastCommit[:8])
		if commitMsg != "" {
			fmt.Printf("Message: %s\n", commitMsg)
		}

		// Perform the revert
		revertCmd := exec.Command("git", "revert", "--no-commit", "HEAD")
		revertCmd.Stdout = os.Stdout
		revertCmd.Stderr = os.Stderr
		if err := revertCmd.Run(); err != nil {
			// Try soft reset if revert fails
			fmt.Println("\nAttempting soft reset instead...")
			resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
			resetCmd.Stdout = os.Stdout
			resetCmd.Stderr = os.Stderr
			if err := resetCmd.Run(); err != nil {
				fmt.Printf("Error undoing: %v\n", err)
				return
			}
		}

		// Create a new commit with the undo
		commitCmd := exec.Command("git", "commit", "-m", "Undo: "+description)
		commitCmd.Stdout = os.Stdout
		commitCmd.Stderr = os.Stderr
		if err := commitCmd.Run(); err != nil {
			fmt.Println("Changes reverted. Commit manually if desired.")
		}

		// Save to history
		db.SetMemory("last_undo_commit", lastCommit)
		db.SetMemory("last_undo_time", time.Now().Format(time.RFC3339))

		fmt.Printf("\n✓ Undone: %s\n", description)
	},
}

var redoCmd = &cobra.Command{
	Use:   "redo",
	Short: "Redo the last undone action",
	Long:  `Redo the most recently undone action by reverting the undo commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if we're in a git repo
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			fmt.Println("Error: Not a git repository. Redo requires git to track changes.")
			return
		}

		// Get the last commit message to check if it's an undo
		msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
		msgOut, err := msgCmd.Output()
		if err != nil {
			fmt.Printf("Error checking last commit: %v\n", err)
			return
		}
		commitMsg := strings.TrimSpace(string(msgOut))

		if !strings.HasPrefix(commitMsg, "Undo:") {
			fmt.Println("Last action was not an undo. Nothing to redo.")
			return
		}

		// Get the commit to redo (the undo commit itself)
		undoCommitCmd := exec.Command("git", "rev-parse", "HEAD")
		undoCommitOut, err := undoCommitCmd.Output()
		if err != nil {
			fmt.Printf("Error getting commit: %v\n", err)
			return
		}
		undoCommit := strings.TrimSpace(string(undoCommitOut))

		// Revert the undo (which restores the original)
		fmt.Printf("Redoing: %s\n", commitMsg)
		
		revertCmd := exec.Command("git", "revert", "--no-commit", "HEAD")
		revertCmd.Stdout = os.Stdout
		revertCmd.Stderr = os.Stderr
		if err := revertCmd.Run(); err != nil {
			fmt.Printf("Error redoing: %v\n", err)
			return
		}

		// Commit the redo
		newMsg := strings.Replace(commitMsg, "Undo:", "Redo:", 1)
		commitCmd := exec.Command("git", "commit", "-m", newMsg)
		commitCmd.Stdout = os.Stdout
		commitCmd.Stderr = os.Stderr
		if err := commitCmd.Run(); err != nil {
			fmt.Println("Changes reapplied. Commit manually if desired.")
			return
		}

		// Save to history
		db.SetMemory("last_redo_commit", undoCommit)
		db.SetMemory("last_redo_time", time.Now().Format(time.RFC3339))

		fmt.Printf("\n✓ Redone: %s\n", newMsg)
	},
}
