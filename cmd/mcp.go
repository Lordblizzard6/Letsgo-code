package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/mcp"
)

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpStartCmd)
	mcpCmd.AddCommand(mcpStopCmd)
	mcpCmd.AddCommand(mcpToolsCmd)
	mcpCmd.AddCommand(mcpResourcesCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) servers",
	Long:  `Manage MCP servers for extended functionality. MCP servers provide additional tools, resources, and prompts.`,
}

var mcpAddCmd = &cobra.Command{
	Use:   "add [name] [command] [args...]",
	Short: "Add a new MCP server",
	Long: `Add a new MCP server configuration.

Example:
  letsGo mcp add filesystem npx -y @modelcontextprotocol/server-filesystem /path/to/dir
  letsGo mcp add git npx -y @modelcontextprotocol/server-git`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]
		var argsSlice []string
		if len(args) > 2 {
			argsSlice = args[2:]
		}

		// Parse optional flags
		workingDir, _ := cmd.Flags().GetString("working-dir")
		autoStart, _ := cmd.Flags().GetBool("auto-start")
		description, _ := cmd.Flags().GetString("description")

		config := &mcp.ServerConfig{
			Name:        name,
			Command:     command,
			Args:        argsSlice,
			WorkingDir:  workingDir,
			AutoStart:   autoStart,
			Description: description,
			Env:         make(map[string]string),
		}

		manager := mcp.GetManager()
		if err := manager.AddServer(config); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ MCP server '%s' added\n", name)

		if autoStart {
			fmt.Printf("Starting '%s'...\n", name)
			if err := manager.StartServer(name); err != nil {
				fmt.Printf("Warning: failed to start server: %v\n", err)
			}
		}
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		manager := mcp.GetManager()
		if err := manager.RemoveServer(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ MCP server '%s' removed\n", name)
	},
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		manager := mcp.GetManager()
		servers := manager.ListServers()

		if len(servers) == 0 {
			fmt.Println("No MCP servers configured.")
			fmt.Println("\nAdd a server with:")
			fmt.Println("  claudego mcp add <name> <command> [args...]")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tCOMMAND\tDESCRIPTION")
		fmt.Fprintln(w, "----\t------\t-------\t-----------")

		for name, config := range servers {
			status := "stopped"
			if manager.IsServerRunning(name) {
				status = "running"
			}

			cmd := config.Command
			if len(config.Args) > 0 {
				cmd += " " + strings.Join(config.Args, " ")
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, status, cmd, config.Description)
		}
		w.Flush()
	},
}

var mcpStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		manager := mcp.GetManager()
		if err := manager.StartServer(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

var mcpStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		manager := mcp.GetManager()
		if err := manager.StopServer(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

var mcpToolsCmd = &cobra.Command{
	Use:   "tools [server]",
	Short: "List tools from an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverName := args[0]

		manager := mcp.GetManager()
		tools, err := manager.ListServerTools(serverName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(tools) == 0 {
			fmt.Printf("No tools available from server '%s'\n", serverName)
			return
		}

		fmt.Printf("Tools from '%s':\n\n", serverName)
		for _, tool := range tools {
			fmt.Printf("  %s\n", tool.Name)
			if tool.Description != "" {
				fmt.Printf("    %s\n", tool.Description)
			}
			fmt.Println()
		}
	},
}

var mcpResourcesCmd = &cobra.Command{
	Use:   "resources [server]",
	Short: "List resources from an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverName := args[0]

		manager := mcp.GetManager()
		resources, err := manager.ListServerResources(serverName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(resources) == 0 {
			fmt.Printf("No resources available from server '%s'\n", serverName)
			return
		}

		fmt.Printf("Resources from '%s':\n\n", serverName)
		for _, res := range resources {
			fmt.Printf("  %s\n", res.Name)
			fmt.Printf("    URI: %s\n", res.URI)
			if res.Description != "" {
				fmt.Printf("    %s\n", res.Description)
			}
			if res.MimeType != "" {
				fmt.Printf("    Type: %s\n", res.MimeType)
			}
			fmt.Println()
		}
	},
}

func init() {
	mcpAddCmd.Flags().StringP("working-dir", "w", "", "Working directory for the server")
	mcpAddCmd.Flags().BoolP("auto-start", "a", false, "Auto-start the server when added")
	mcpAddCmd.Flags().StringP("description", "d", "", "Description of the server")
}
