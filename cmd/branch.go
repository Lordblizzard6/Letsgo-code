package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.AddCommand(branchListCmd)
	branchCmd.AddCommand(branchCreateCmd)
	branchCmd.AddCommand(branchSwitchCmd)
	branchCmd.AddCommand(branchDeleteCmd)
	branchCmd.AddCommand(branchMergeCmd)
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Manage git branches",
	Long:  `List, create, switch, delete, and merge git branches.`,
}

var branchListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all branches",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		out, err := exec.Command("git", "branch", "-a").CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	},
}

var branchCreateCmd = &cobra.Command{
	Use:     "create <branch-name>",
	Short:   "Create a new branch",
	Aliases: []string{"new", "c"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: branch create <branch-name>")
			os.Exit(1)
		}

		branchName := args[0]
		out, err := exec.Command("git", "checkout", "-b", branchName).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating branch: %s\n", string(out))
			os.Exit(1)
		}
		fmt.Printf("✓ Created and switched to branch: %s\n", branchName)
	},
}

var branchSwitchCmd = &cobra.Command{
	Use:     "switch <branch-name>",
	Short:   "Switch to an existing branch",
	Aliases: []string{"checkout", "sw"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: branch switch <branch-name>")
			os.Exit(1)
		}

		branchName := args[0]
		out, err := exec.Command("git", "checkout", branchName).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error switching branch: %s\n", string(out))
			os.Exit(1)
		}
		fmt.Printf("✓ Switched to branch: %s\n", branchName)
	},
}

var branchDeleteCmd = &cobra.Command{
	Use:     "delete <branch-name>",
	Short:   "Delete a branch",
	Aliases: []string{"del", "rm", "d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: branch delete <branch-name>")
			os.Exit(1)
		}

		force, _ := cmd.Flags().GetBool("force")
		branchName := args[0]

		var out []byte
		var err error
		if force {
			out, err = exec.Command("git", "branch", "-D", branchName).CombinedOutput()
		} else {
			out, err = exec.Command("git", "branch", "-d", branchName).CombinedOutput()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting branch: %s\n", string(out))
			os.Exit(1)
		}
		fmt.Printf("✓ Deleted branch: %s\n", branchName)
	},
}

var branchMergeCmd = &cobra.Command{
	Use:   "merge <branch-name>",
	Short: "Merge a branch into current branch",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: branch merge <branch-name>")
			os.Exit(1)
		}

		branchName := args[0]
		out, err := exec.Command("git", "merge", branchName).CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error merging branch: %s\n", string(out))
			os.Exit(1)
		}
		fmt.Printf("✓ Merged branch: %s\n", branchName)
		fmt.Println(string(out))
	},
}

func init() {
	branchDeleteCmd.Flags().BoolP("force", "f", false, "Force delete unmerged branch")
}
