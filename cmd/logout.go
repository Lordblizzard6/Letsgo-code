package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear API credentials and logout",
	Long:  `Remove stored API keys and logout from Claude Code.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Print("Are you sure you want to logout? This will remove your API key. [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Logout cancelled")
				return
			}
		}

		// Clear config
		err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		config.AppConfig.APIKey = ""
		if err := config.SaveConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Logged out successfully")
		fmt.Println("  Run 'letsGo login' to authenticate again")
	},
}

func init() {
	logoutCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}
