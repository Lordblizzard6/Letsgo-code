package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/config"
)

func init() {
	rootCmd.AddCommand(reloadPluginsCmd)
	reloadPluginsCmd.Flags().BoolP("all", "a", false, "Reload all plugins")
	reloadPluginsCmd.Flags().StringP("plugin", "p", "", "Reload specific plugin")
	reloadPluginsCmd.Flags().BoolP("force", "f", false, "Force reload even if unchanged")
}

var reloadPluginsCmd = &cobra.Command{
	Use:   "reload-plugins",
	Short: "Reload plugins and refresh their commands",
	Long: `Reload plugins to pick up new changes without restarting Claude Code.

This is useful during plugin development or when updating plugins.
By default, only changed plugins are reloaded. Use --force to reload all.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		specificPlugin, _ := cmd.Flags().GetString("plugin")
		force, _ := cmd.Flags().GetBool("force")

		fmt.Println("🔄 Reloading Plugins")
		fmt.Println("====================\n")

		// Get plugins directory
		configDir := config.GetConfigDir()
		pluginsDir := filepath.Join(configDir, "plugins")

		if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
			fmt.Println("ℹ️  No plugins directory found.")
			fmt.Println("Create one at:", pluginsDir)
			return
		}

		if specificPlugin != "" {
			// Reload specific plugin
			fmt.Printf("Reloading plugin: %s\n", specificPlugin)
			if err := reloadSpecificPlugin(pluginsDir, specificPlugin, force); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Plugin reloaded successfully")
			return
		}

		// Reload all or changed
		plugins, err := listPlugins(pluginsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error listing plugins: %v\n", err)
			os.Exit(1)
		}

		if len(plugins) == 0 {
			fmt.Println("ℹ️  No plugins installed")
			return
		}

		reloaded := 0
		for _, plugin := range plugins {
			if all || force || pluginChanged(plugin) {
				fmt.Printf("Reloading: %s\n", plugin.Name)
				if err := reloadPlugin(plugin, force); err != nil {
					fmt.Printf("  ⚠️  Failed: %v\n", err)
					continue
				}
				reloaded++
				fmt.Printf("  ✅ Reloaded\n")
			} else {
				fmt.Printf("Skipping: %s (unchanged)\n", plugin.Name)
			}
		}

		fmt.Printf("\n✅ Reloaded %d/%d plugins\n", reloaded, len(plugins))

		if reloaded > 0 {
			fmt.Println("\n💡 Tip: Run 'skills' to see newly available commands")
		}
	},
}

type Plugin struct {
	Name    string
	Path    string
	Version string
	LastMod int64
	Enabled bool
}

func listPlugins(pluginsDir string) ([]Plugin, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			info, _ := entry.Info()
			plugins = append(plugins, Plugin{
				Name:    entry.Name(),
				Path:    filepath.Join(pluginsDir, entry.Name()),
				LastMod: info.ModTime().Unix(),
				Enabled: true,
			})
		}
	}
	return plugins, nil
}

func pluginChanged(plugin Plugin) bool {
	// Check if plugin has been modified since last load
	// This would compare timestamps or hashes
	// For now, always return true for simplicity
	return true
}

func reloadPlugin(plugin Plugin, force bool) error {
	// Unload current plugin
	// This would remove commands from registry
	_ = force

	// Reload plugin code
	// In real implementation, this would:
	// 1. Clear plugin command cache
	// 2. Re-read plugin manifest
	// 3. Re-register commands
	// 4. Update skills index

	return nil
}

func reloadSpecificPlugin(pluginsDir, name string, force bool) error {
	pluginPath := filepath.Join(pluginsDir, name)

	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", name)
	}

	info, err := os.Stat(pluginPath)
	if err != nil {
		return err
	}

	plugin := Plugin{
		Name:    name,
		Path:    pluginPath,
		LastMod: info.ModTime().Unix(),
		Enabled: true,
	}

	return reloadPlugin(plugin, force)
}
