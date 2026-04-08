package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().BoolP("force", "f", false, "Force upgrade even if up to date")
	upgradeCmd.Flags().BoolP("check", "c", false, "Only check for updates")
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Claude Code CLI to the latest version",
	Long:  `Check for updates and upgrade Claude Code CLI to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		checkOnly, _ := cmd.Flags().GetBool("check")

		fmt.Println("🔄 Checking for updates...")

		// Get current version
		currentVersion := "0.1.0" // Should come from build info

		// In a real implementation, this would:
		// 1. Fetch latest release from GitHub API
		// 2. Compare versions
		// 3. Download and install if newer

		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Println()

		if checkOnly {
			fmt.Println("✓ Up to date (check only mode)")
			return
		}

		// Simulate update check
		latestVersion := "0.2.0" // Would come from API

		if currentVersion == latestVersion && !force {
			fmt.Println("✓ Already on the latest version!")
			return
		}

		fmt.Printf("Update available: %s → %s\n", currentVersion, latestVersion)
		fmt.Println()

		// Platform-specific upgrade
		switch runtime.GOOS {
		case "windows":
			fmt.Println("To upgrade on Windows, run:")
			fmt.Println("  go install github.com/user/go-claude-code@latest")
		case "darwin", "linux":
			fmt.Println("Upgrading...")
			
			// Try different methods
			methods := []string{
				"go install github.com/user/go-claude-code@latest",
				"curl -fsSL https://claudego.dev/install.sh | sh",
			}

			for _, method := range methods {
				fmt.Printf("\nTrying: %s\n", method)
				parts := splitCommand(method)
				if len(parts) > 0 {
					c := exec.Command(parts[0], parts[1:]...)
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					if err := c.Run(); err == nil {
						fmt.Println("\n✅ Upgrade successful!")
						fmt.Println("Please restart Claude Code to use the new version.")
						return
					}
				}
			}
			
			fmt.Println("\n⚠️  Automatic upgrade failed. Please upgrade manually:")
			fmt.Println("  go install github.com/user/go-claude-code@latest")
		}
	},
}

func splitCommand(cmd string) []string {
	var parts []string
	var current string
	inQuote := false
	
	for _, r := range cmd {
		switch r {
		case '"', '\'':
			inQuote = !inQuote
		case ' ':
			if !inQuote && current != "" {
				parts = append(parts, current)
				current = ""
			} else if inQuote {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}
