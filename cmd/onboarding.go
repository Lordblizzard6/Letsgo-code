package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(onboardingCmd)
	onboardingCmd.Flags().BoolP("reset", "r", false, "Restart onboarding")
	onboardingCmd.Flags().BoolP("skip", "s", false, "Skip onboarding")
}

var onboardingCmd = &cobra.Command{
	Use:   "onboarding",
	Short: "Interactive first-time setup wizard",
	Long: `Run the Claude Code onboarding wizard to configure your environment.

This wizard will help you:
  • Configure your AI provider (Anthropic, OpenAI, Groq)
  • Set up git integration
  • Choose your preferred theme and settings
  • Learn the basic commands`,
	Run: func(cmd *cobra.Command, args []string) {
		reset, _ := cmd.Flags().GetBool("reset")
		skip, _ := cmd.Flags().GetBool("skip")

		if skip {
			fmt.Println("⏭️  Onboarding skipped")
			markOnboardingComplete()
			return
		}

		fmt.Println("🚀 Welcome to Claude Code!")
		fmt.Println("==========================\n")

		if reset {
			fmt.Println("♻️  Restarting onboarding...\n")
		} else if isOnboardingComplete() {
			fmt.Println("✓ Onboarding already completed")
			fmt.Println("Use --reset to run again\n")
			return
		}

		fmt.Println("Let's get you set up with Claude Code.")
		fmt.Println("This will only take a few minutes.\n")

		// Step 1: API Configuration
		fmt.Println("📋 Step 1: AI Provider Configuration")
		fmt.Println("-------------------------------------")
		
		if err := stepAPIConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
			os.Exit(1)
		}

		// Step 2: Git Setup
		fmt.Println("\n📋 Step 2: Git Integration")
		fmt.Println("--------------------------")
		
		if err := stepGitSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠️  Git setup warning: %v\n", err)
		}

		// Step 3: Preferences
		fmt.Println("\n📋 Step 3: Preferences")
		fmt.Println("----------------------")
		
		if err := stepPreferences(); err != nil {
			fmt.Fprintf(os.Stderr, "\n⚠️  Preferences warning: %v\n", err)
		}

		// Step 4: Quick Tutorial
		fmt.Println("\n📋 Step 4: Quick Start Guide")
		fmt.Println("----------------------------")
		showQuickStart()

		// Mark complete
		markOnboardingComplete()

		fmt.Println("\n🎉 Onboarding Complete!")
		fmt.Println("======================")
		fmt.Println("\nYou're ready to use Claude Code!")
		fmt.Println("\nNext steps:")
		fmt.Println("  • Run 'claudego chat' to start a conversation")
		fmt.Println("  • Run 'claudego help' to see all commands")
		fmt.Println("  • Run 'claudego doctor' to check your setup")
	},
}

func stepAPIConfig() error {
	// Check current config
	config.LoadConfig()
	
	if config.AppConfig.APIKey != "" {
		masked := config.AppConfig.APIKey[:5] + "..." + config.AppConfig.APIKey[len(config.AppConfig.APIKey)-4:]
		fmt.Printf("✓ API Key already configured: %s\n", masked)
		
		fmt.Print("\nWould you like to reconfigure? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "y" {
			fmt.Printf("✓ Using existing configuration (%s)\n", config.AppConfig.Model)
			return nil
		}
	}

	// Provider selection
	fmt.Println("\nChoose your AI provider:")
	fmt.Println("  1. Anthropic (Claude) - Recommended")
	fmt.Println("  2. OpenAI (GPT-4)")
	fmt.Println("  3. Groq (Fast, cost-effective)")
	fmt.Println("  4. Other (custom endpoint)")
	
	fmt.Print("\nSelect [1-4]: ")
	var choice string
	fmt.Scanln(&choice)
	
	switch choice {
	case "1":
		fmt.Println("\n📝 Anthropic Configuration")
		fmt.Println("Get your API key from: https://console.anthropic.com/")
		fmt.Print("Enter API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		
		if apiKey != "" {
			config.AppConfig.APIKey = apiKey
			config.AppConfig.BaseURL = "https://api.anthropic.com/v1"
			config.AppConfig.Model = "claude-3-5-sonnet-20240620"
			config.SaveConfig()
			fmt.Println("✓ Configuration saved")
		}
		
	case "2":
		fmt.Println("\n📝 OpenAI Configuration")
		fmt.Println("Get your API key from: https://platform.openai.com/")
		fmt.Print("Enter API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		
		if apiKey != "" {
			config.AppConfig.APIKey = apiKey
			config.AppConfig.BaseURL = "https://api.openai.com/v1"
			config.AppConfig.Model = "gpt-4o"
			config.SaveConfig()
			fmt.Println("✓ Configuration saved")
		}
		
	case "3":
		fmt.Println("\n📝 Groq Configuration")
		fmt.Println("Get your API key from: https://console.groq.com/")
		fmt.Print("Enter API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		
		if apiKey != "" {
			config.AppConfig.APIKey = apiKey
			config.AppConfig.BaseURL = "https://api.groq.com/openai/v1"
			config.AppConfig.Model = "llama-3.3-70b-versatile"
			config.SaveConfig()
			fmt.Println("✓ Configuration saved")
		}
		
	case "4":
		fmt.Println("\n📝 Custom Configuration")
		fmt.Print("Enter API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)
		fmt.Print("Enter Base URL: ")
		var baseURL string
		fmt.Scanln(&baseURL)
		fmt.Print("Enter Model Name: ")
		var model string
		fmt.Scanln(&model)
		
		if apiKey != "" && baseURL != "" && model != "" {
			config.AppConfig.APIKey = apiKey
			config.AppConfig.BaseURL = baseURL
			config.AppConfig.Model = model
			config.SaveConfig()
			fmt.Println("✓ Configuration saved")
		}
		
	default:
		fmt.Println("Skipping API configuration")
	}
	
	return nil
}

func stepGitSetup() error {
	// Check if git is installed
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Println("⚠️  Git not found. Please install git to use version control features.")
		return nil
	}

	// Check if in a git repo
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("📝 Not in a git repository")
		fmt.Print("Would you like to initialize git here? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) == "y" {
			if err := exec.Command("git", "init").Run(); err != nil {
				return fmt.Errorf("failed to initialize git: %w", err)
			}
			fmt.Println("✓ Git repository initialized")
		}
	} else {
		// Show git status
		out, _ := exec.Command("git", "remote", "-v").Output()
		if len(out) > 0 {
			fmt.Println("✓ Git repository detected with remotes")
		} else {
			fmt.Println("✓ Git repository detected (no remotes configured)")
		}
	}
	
	return nil
}

func stepPreferences() error {
	fmt.Println("Choose your default theme:")
	fmt.Println("  1. Dark (default)")
	fmt.Println("  2. Light")
	fmt.Println("  3. Auto (follows system)")
	
	fmt.Print("\nSelect [1-3, default=1]: ")
	var choice string
	fmt.Scanln(&choice)
	
	switch choice {
	case "2":
		fmt.Println("✓ Light theme selected")
	case "3":
		fmt.Println("✓ Auto theme selected")
	default:
		fmt.Println("✓ Dark theme selected (default)")
	}
	
	fmt.Println("\nChoose your vim mode preference:")
	fmt.Println("  1. Disabled (default)")
	fmt.Println("  2. Enabled")
	
	fmt.Print("\nSelect [1-2, default=1]: ")
	fmt.Scanln(&choice)
	
	if choice == "2" {
		fmt.Println("✓ Vim mode enabled")
	} else {
		fmt.Println("✓ Vim mode disabled (default)")
	}
	
	return nil
}

func showQuickStart() {
	fmt.Println("📚 Essential Commands:")
	fmt.Println("----------------------")
	fmt.Println()
	fmt.Println("  claudego chat              Start interactive chat")
	fmt.Println("  claudego chat <message>    Send one-off message")
	fmt.Println("  claudego add <files>       Add files to context")
	fmt.Println("  claudego commit            Commit with AI message")
	fmt.Println("  claudego review            Review code changes")
	fmt.Println("  claudego branch            Manage git branches")
	fmt.Println("  claudego status            Check system status")
	fmt.Println()
	fmt.Println("💡 Pro Tips:")
	fmt.Println("  • Use Tab for command completion")
	fmt.Println("  • Press Ctrl+C to cancel current operation")
	fmt.Println("  • Use /help in chat for available commands")
}

func isOnboardingComplete() bool {
	configDir := config.GetConfigDir()
	markerFile := filepath.Join(configDir, ".onboarding_complete")
	_, err := os.Stat(markerFile)
	return err == nil
}

func markOnboardingComplete() {
	configDir := config.GetConfigDir()
	os.MkdirAll(configDir, 0755)
	markerFile := filepath.Join(configDir, ".onboarding_complete")
	os.WriteFile(markerFile, []byte("completed"), 0644)
}
