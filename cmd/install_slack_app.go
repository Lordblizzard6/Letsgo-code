package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installSlackAppCmd)
	installSlackAppCmd.Flags().StringP("workspace", "w", "", "Slack workspace URL")
	installSlackAppCmd.Flags().StringP("channel", "c", "", "Default channel for notifications")
}

var installSlackAppCmd = &cobra.Command{
	Use:   "install-slack-app",
	Short: "Install Claude Code Slack App",
	Long: `Install the Claude Code Slack App to receive notifications and interact via Slack.

This enables:
  • PR notifications in Slack channels
  • Code review summaries
  • Build status alerts
  • Direct messages for urgent issues

Requires Slack workspace admin approval.`,
	Run: func(cmd *cobra.Command, args []string) {
		workspace, _ := cmd.Flags().GetString("workspace")
		channel, _ := cmd.Flags().GetString("channel")

		fmt.Println("💬 Installing Claude Code Slack App")
		fmt.Println("====================================")

		if workspace == "" {
			fmt.Print("Enter your Slack workspace URL (e.g., mycompany.slack.com): ")
			fmt.Scanln(&workspace)
		}

		if workspace == "" {
			fmt.Println("❌ Workspace URL is required")
			os.Exit(1)
		}

		fmt.Printf("✓ Workspace: %s\n", workspace)

		// Open Slack app installation
		installURL := fmt.Sprintf("https://%s/apps/manage", workspace)

		fmt.Println("\n🌐 Opening Slack app management...")

		var openCmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			openCmd = exec.Command("open", installURL)
		case "windows":
			openCmd = exec.Command("start", installURL)
		default:
			openCmd = exec.Command("xdg-open", installURL)
		}

		if err := openCmd.Start(); err != nil {
			fmt.Printf("\nPlease visit: %s\n", installURL)
		} else {
			fmt.Println("✓ Browser opened")
		}

		fmt.Println("\n📋 Installation steps:")
		fmt.Println("1. Search for 'Claude Code' in Slack App Directory")
		fmt.Println("2. Click 'Add to Slack'")
		fmt.Println("3. Choose permissions (recommend: read messages, post messages)")
		fmt.Println("4. Copy the Bot User OAuth Token")

		fmt.Println("\n⚙️  Configuration:")
		fmt.Println("Set environment variable:")
		fmt.Println("  export SLACK_BOT_TOKEN=xoxb-your-token")

		if channel != "" {
			fmt.Printf("\n  export SLACK_CHANNEL=%s\n", channel)
		}

		fmt.Println("\n✅ Slack App installation initiated!")
		fmt.Println("\nNext steps:")
		fmt.Println("  • Invite @Claude to your channels")
		fmt.Println("  • Use '/claude' command in Slack")
		fmt.Println("  • Configure notification preferences")
	},
}
