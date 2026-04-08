package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose potential issues with your setup",
	Long:  `Run diagnostics to check your Claude Code installation and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Running diagnostics...\n")

		issues := 0
		warnings := 0

		// Check 1: API Key
		fmt.Println("✓ Checking API configuration...")
		if config.AppConfig.APIKey == "" {
			fmt.Println("  ✗ No API key configured")
			fmt.Println("    Run: claudego login")
			issues++
		} else {
			fmt.Printf("  ✓ API key configured (%s...)\n", config.AppConfig.APIKey[:8])
		}

		// Check 2: Database
		fmt.Println("\n✓ Checking database...")
		if err := db.InitDB(); err != nil {
			fmt.Printf("  ✗ Database error: %v\n", err)
			issues++
		} else {
			fmt.Println("  ✓ Database initialized")
		}

		// Check 3: Git
		fmt.Println("\n✓ Checking git...")
		if _, err := exec.LookPath("git"); err != nil {
			fmt.Println("  ⚠ Git not found in PATH")
			warnings++
		} else {
			fmt.Println("  ✓ Git available")
		}

		// Check 4: Directory permissions
		fmt.Println("\n✓ Checking permissions...")
		home, _ := os.UserHomeDir()
		claudeDir := home + "/.letsGo"
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			fmt.Printf("  ✗ Config directory not found: %s\n", claudeDir)
			issues++
		} else {
			fmt.Printf("  ✓ Config directory exists: %s\n", claudeDir)
		}

		// Check 5: Model availability
		fmt.Println("\n✓ Checking model configuration...")
		if config.AppConfig.Model == "" {
			fmt.Println("  ✗ No model configured")
			issues++
		} else {
			fmt.Printf("  ✓ Model: %s\n", config.AppConfig.Model)
		}

		// Check 6: Platform
		fmt.Println("\n✓ Checking platform...")
		fmt.Printf("  ✓ OS: %s\n", runtime.GOOS)
		fmt.Printf("  ✓ Architecture: %s\n", runtime.GOARCH)

		// Check 7: Tools availability
		fmt.Println("\n✓ Checking tools...")
		toolCount := len(tools.AllTools)
		fmt.Printf("  ✓ %d tools available\n", toolCount)

		// Summary
		fmt.Println("\n" + strings.Repeat("=", 40))
		if issues == 0 && warnings == 0 {
			fmt.Println("✅ All checks passed! Your setup looks good.")
		} else if issues == 0 {
			fmt.Printf("⚠️  Found %d warning(s), but no critical issues.\n", warnings)
		} else {
			fmt.Printf("❌ Found %d issue(s) and %d warning(s).\n", issues, warnings)
			fmt.Println("\nRun 'claudego login' to configure your API key.")
		}
	},
}
