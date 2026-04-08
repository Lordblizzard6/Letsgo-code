package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installGitHubAppCmd)
	installGitHubAppCmd.Flags().BoolP("local", "l", false, "Install for local development")
	installGitHubAppCmd.Flags().StringP("org", "o", "", "Organization to install (default: personal)")
}

var installGitHubAppCmd = &cobra.Command{
	Use:   "install-github-app",
	Short: "Install Claude Code GitHub App",
	Long: `Install the Claude Code GitHub App to enable advanced GitHub integrations.

This allows Claude to:
  • Comment on pull requests automatically
  • Review code changes
  • Monitor CI/CD status
  • Subscribe to repository events

Installation requires GitHub authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		local, _ := cmd.Flags().GetBool("local")
		org, _ := cmd.Flags().GetString("org")

		fmt.Println("🔗 Installing Claude Code GitHub App")
		fmt.Println("====================================\n")

		// Check for gh CLI
		if _, err := exec.LookPath("gh"); err != nil {
			fmt.Println("⚠️  GitHub CLI (gh) not found.")
			fmt.Println("Please install it first: https://cli.github.com/")
			os.Exit(1)
		}

		// Check authentication
		if err := exec.Command("gh", "auth", "status").Run(); err != nil {
			fmt.Println("🔐 Not authenticated with GitHub.")
			fmt.Println("\nPlease run: gh auth login")
			os.Exit(1)
		}

		fmt.Println("✓ GitHub CLI authenticated")

		if local {
			fmt.Println("\n📦 Local Development Mode")
			fmt.Println("This will configure the app for local testing.")
			fmt.Println("\nSteps:")
			fmt.Println("1. Go to: https://github.com/apps/claude-code")
			fmt.Println("2. Click 'Install'")
			if org != "" {
				fmt.Printf("3. Select organization: %s\n", org)
			} else {
				fmt.Println("3. Select your personal account")
			}
			fmt.Println("4. Choose repository access")
			fmt.Println("5. Copy the installation token")

			fmt.Println("\n⚙️  Configuration:")
			fmt.Println("export GITHUB_APP_TOKEN=your-token-here")

		} else {
			fmt.Println("\n🌐 Opening GitHub App installation page...")

			appURL := "https://github.com/apps/claude-code/installations/new"
			if org != "" {
				appURL = fmt.Sprintf("https://github.com/organizations/%s/settings/installations", org)
			}

			var openCmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				openCmd = exec.Command("open", appURL)
			case ".letsGo":
				openCmd = exec.Command("open", appURL)
			case "windows":
				openCmd = exec.Command("start", appURL)
			default:
				openCmd = exec.Command("xdg-open", appURL)
			}

			if err := openCmd.Start(); err != nil {
				fmt.Printf("\nPlease visit: %s\n", appURL)
			} else {
				fmt.Println("✓ Browser opened")
			}

			fmt.Println("\n📋 After installation:")
			fmt.Println("1. Select repositories to give Claude access to")
			fmt.Println("2. The app will be installed and ready")
			fmt.Println("3. Use 'pr-comments' command to see Claude's activity")
		}

		fmt.Println("\n✅ GitHub App installation initiated!")
		fmt.Println("\nNext steps:")
		fmt.Println("  • Use 'claudego pr-comments' to view activity")
		fmt.Println("  • Claude will automatically review PRs")
		fmt.Println("  • Enable notifications for real-time updates")
	},
}
