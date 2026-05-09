package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/plugins"
)

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginUninstallCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long: `Install, manage, and configure plugins to extend Claude Code Go functionality.

Currently, only plugins with manifest type "native" are supported.`,
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install [source] [name]",
	Short: "Install a plugin from a git repository or local path",
	Long: `Install a plugin from a git repository or local directory.

Examples:
  # Install from git
  letsgo plugin install https://github.com/user/my-plugin.git my-plugin

  # Install from local directory
  letsgo plugin install /path/to/plugin my-plugin

Note: The plugin directory must contain a plugin.json manifest file.
Only plugins declaring type "native" are supported at this time.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		name := args[1]

		fmt.Printf("Installing plugin '%s' from %s...\n", name, source)

		registry := plugins.GetRegistry()
		if err := registry.InstallPlugin(source, name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Plugin '%s' installed successfully\n", name)
		fmt.Println("\nTo enable the plugin:")
		fmt.Printf("  letsgo plugin enable %s\n", name)
	},
}

var pluginUninstallCmd = &cobra.Command{
	Use:   "uninstall [name]",
	Short: "Remove a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		registry := plugins.GetRegistry()
		if err := registry.UninstallPlugin(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Plugin '%s' uninstalled\n", name)
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	Run: func(cmd *cobra.Command, args []string) {
		registry := plugins.GetRegistry()
		pluginList := registry.ListPlugins()

		if len(pluginList) == 0 {
			fmt.Println("No plugins installed.")
			fmt.Println("\nInstall a plugin with:")
			fmt.Println("  letsgo plugin install <source> <name>")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tTYPE\tDESCRIPTION")
		fmt.Fprintln(w, "----\t-------\t------\t----\t-----------")

		for name, p := range pluginList {
			status := "disabled"
			if p.Enabled {
				if p.Loaded {
					status = "loaded"
				} else {
					status = "enabled"
				}
			}

			desc := p.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, p.Version, status, p.Type, desc)
		}
		w.Flush()
	},
}

var pluginEnableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		registry := plugins.GetRegistry()
		if err := registry.EnablePlugin(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Plugin '%s' enabled\n", name)
	},
}

var pluginDisableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		registry := plugins.GetRegistry()
		if err := registry.DisablePlugin(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Plugin '%s' disabled\n", name)
	},
}

var pluginInfoCmd = &cobra.Command{
	Use:   "info [name]",
	Short: "Show plugin details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		registry := plugins.GetRegistry()
		p, err := registry.GetPlugin(name)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Plugin: %s\n", p.Name)
		fmt.Printf("Version: %s\n", p.Version)
		fmt.Printf("Author: %s\n", p.Author)
		fmt.Printf("Type: %s\n", p.Type)
		fmt.Printf("Status: %s\n", func() string {
			switch p.Path {
			case ".letsGo":
				return ".letsGo"
			}
			if p.Enabled {
				if p.Loaded {
					return "loaded"
				}
				return "enabled (not loaded)"
			}
			return "disabled"
		}())
		fmt.Printf("Path: %s\n", p.Path)
		if p.Description != "" {
			fmt.Printf("\nDescription:\n  %s\n", p.Description)
		}
	},
}
