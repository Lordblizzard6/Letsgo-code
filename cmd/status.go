package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system and session status",
	Long:  `Display comprehensive status of Claude Code, including session info, git status, and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Claude Code Status")
		fmt.Println("====================")
		fmt.Println()

		// Session info
		session, _ := db.GetActiveSession()
		if session != nil {
			fmt.Printf("📝 Active Session: %s\n", session.Name)
			fmt.Printf("   ID: %s\n", session.ID[:8])
			fmt.Printf("   Project: %s\n", session.ProjectPath)
		} else {
			fmt.Println("📝 No active session")
		}

		// Git status
		fmt.Println()
		if _, err := os.Stat(".git"); err == nil {
			fmt.Println("🌿 Git Repository:")

			// Branch
			if branch, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
				fmt.Printf("   Branch: %s", branch)
			}

			// Status
			if status, err := exec.Command("git", "status", "--short").Output(); err == nil {
				if len(status) > 0 {
					fmt.Printf("   Changes: %d files\n", len(status))
				} else {
					fmt.Println("   Working tree: clean")
				}
			}

			// Last commit
			if lastCommit, err := exec.Command("git", "log", "-1", "--oneline").Output(); err == nil {
				fmt.Printf("   Last commit: %s", lastCommit)
			}
		} else {
			fmt.Println("🌿 Not in a git repository")
		}

		// Configuration
		fmt.Println()
		fmt.Println("⚙️  Configuration:")
		fmt.Printf("   Model: %s\n", config.AppConfig.Model)

		// API Key status
		err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "   Error loading config: %v\n", err)
		} else if config.AppConfig.APIKey != "" {
			fmt.Printf("   API Key: %s...%s\n", config.AppConfig.APIKey[:5], config.AppConfig.APIKey[len(config.AppConfig.APIKey)-4:])
		} else {
			fmt.Println("   API Key: Not configured")
		}

		// System info
		fmt.Println()
		fmt.Println("🖥️  System:")
		fmt.Printf("   OS: %s\n", runtime.GOOS)
		fmt.Printf("   Arch: %s\n", runtime.GOARCH)

		cwd, _ := os.Getwd()
		fmt.Printf("   Working directory: %s\n", cwd)
	},
}
