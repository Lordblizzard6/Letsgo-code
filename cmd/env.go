package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envGetCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long:  `View and manage environment variables relevant to Claude Code.`,
}

var envListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List Claude Code related environment variables",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		prefixes := []string{"CLAUDE_", "ANTHROPIC_", "OPENAI_", "GROQ_", "OLLAMA_"}

		fmt.Println("Claude Code Environment Variables:")
		fmt.Println("=================================")

		for _, env := range os.Environ() {
			for _, prefix := range prefixes {
				if strings.HasPrefix(env, prefix) {
					parts := strings.SplitN(env, "=", 2)
					key := parts[0]
					value := ""
					if len(parts) > 1 {
						// Mask API keys
						if strings.Contains(strings.ToLower(key), "key") && len(parts[1]) > 10 {
							value = parts[1][:5] + "..." + parts[1][len(parts[1])-4:]
						} else {
							value = parts[1]
						}
					}
					fmt.Printf("  %s=%s\n", key, value)
					break
				}
			}
		}
	},
}

var envGetCmd = &cobra.Command{
	Use:   "get <variable>",
	Short: "Get an environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: env get <variable>")
			os.Exit(1)
		}

		key := args[0]
		value := os.Getenv(key)
		if value == "" {
			fmt.Printf("%s is not set\n", key)
			return
		}

		// Mask if it looks like a key
		if strings.Contains(strings.ToLower(key), "key") && len(value) > 10 {
			fmt.Printf("%s=%s...%s\n", key, value[:5], value[len(value)-4:])
		} else {
			fmt.Printf("%s=%s\n", key, value)
		}
	},
}

var envSetCmd = &cobra.Command{
	Use:   "set <variable> <value>",
	Short: "Set an environment variable (for current session only)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Usage: env set <variable> <value>")
			os.Exit(1)
		}

		key := args[0]
		value := strings.Join(args[1:], " ")

		if err := os.Setenv(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Set %s (current session only)\n", key)
		fmt.Println("  To persist, add to your shell profile or use 'config set'")
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset <variable>",
	Short: "Unset an environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Usage: env unset <variable>")
			os.Exit(1)
		}

		key := args[0]
		if err := os.Unsetenv(key); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Unset %s\n", key)
	},
}

// Helper to get sorted env vars
func getSortedEnvVars() []string {
	vars := os.Environ()
	sort.Strings(vars)
	return vars
}

func init() {
	// Add shell command to show how to persist
	envCmd.AddCommand(&cobra.Command{
		Use:   "shell",
		Short: "Show shell configuration for Claude Code",
		Run: func(cmd *cobra.Command, args []string) {
			shell := os.Getenv("SHELL")
			if strings.Contains(shell, "zsh") {
				fmt.Println("Add to ~/.zshrc:")
				fmt.Println("  export ANTHROPIC_API_KEY=your-key-here")
			} else if strings.Contains(shell, "bash") {
				fmt.Println("Add to ~/.bashrc:")
				fmt.Println("  export ANTHROPIC_API_KEY=your-key-here")
			} else if strings.Contains(shell, "fish") {
				fmt.Println("Add to ~/.config/fish/config.fish:")
				fmt.Println("  set -x ANTHROPIC_API_KEY your-key-here")
			}
		},
	})
}
