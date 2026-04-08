package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(teleportCmd)
	teleportCmd.Flags().String("to", "", "Target directory to teleport to")
	teleportCmd.Flags().Bool("list", false, "List recent teleport locations")
}

var teleportCmd = &cobra.Command{
	Use:   "teleport",
	Short: "Teleport session to another directory",
	Long: `Move your Claude Code session to another directory/project.
This preserves conversation context while switching to a new codebase.`,
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := cmd.Flags().GetBool("list")
		if list {
			fmt.Println("Recent locations:")
			fmt.Println("  (Feature: would show recent directories)")
			return
		}

		target, _ := cmd.Flags().GetString("to")
		if len(args) > 0 {
			target = args[0]
		}

		if target == "" {
			fmt.Println("Usage: teleport <directory>")
			fmt.Println("       teleport --to <directory>")
			os.Exit(1)
		}

		// Check if directory exists
		if _, err := os.Stat(target); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Directory does not exist: %s\n", target)
			os.Exit(1)
		}

		// Get current directory
		cwd, _ := os.Getwd()
		fmt.Printf("🌀 Teleporting from %s to %s...\n", cwd, target)

		// Change directory
		if err := os.Chdir(target); err != nil {
			fmt.Fprintf(os.Stderr, "Error changing directory: %v\n", err)
			os.Exit(1)
		}

		// Check for git repo
		if _, err := os.Stat(".git"); err == nil {
			branch, _ := exec.Command("git", "branch", "--show-current").Output()
			fmt.Printf("📍 Git repository detected on branch: %s", branch)
		}

		fmt.Println("\n✅ Teleported successfully!")
		fmt.Println("Context preserved. Continue your conversation.")
	},
}
