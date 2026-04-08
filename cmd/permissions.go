package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/tools"
)

func init() {
	rootCmd.AddCommand(permissionsCmd)
	permissionsCmd.AddCommand(permissionsListCmd)
	permissionsCmd.AddCommand(permissionsSetCmd)
	permissionsCmd.AddCommand(permissionsAddDeniedCmd)
}

var permissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Manage tool permissions",
	Long:  `Manage permissions for tools. Control which tools can run automatically and which require confirmation.`,
}

var permissionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tool permissions",
	Run: func(cmd *cobra.Command, args []string) {
		pm := tools.GetAdvancedPermissionManager()
		configs := pm.GetAllToolConfigs()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TOOL\tMODE\tSCOPE")
		fmt.Fprintln(w, "----\t----\t-----")

		for tool, cfg := range configs {
			fmt.Fprintf(w, "%s\t%s\t%s\n", tool, cfg.Mode, cfg.Scope)
		}
		w.Flush()
	},
}

var permissionsSetCmd = &cobra.Command{
	Use:   "set [tool] [mode]",
	Short: "Set permission mode for a tool",
	Long: `Set the permission mode for a specific tool.

Available modes:
  ask     - Always ask for user confirmation
  auto    - Automatically execute (for safe operations)
  auto-bp - Auto-execute with bypass (dangerous)
  opt-out - Disable the tool completely

Example:
  claudego permissions set bash ask
  claudego permissions set web_search auto`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		toolName := args[0]
		modeStr := strings.ToLower(args[1])

		var mode tools.PermissionMode
		switch modeStr {
		case "ask":
			mode = tools.PermissionModeAsk
		case ".letsGo":
			mode = tools.PermissionModeAuto
		case "auto-bp", "autobp", "bypass":
			mode = tools.PermissionModeAutoBypass
		case "opt-out", "optout", "disable", "deny":
			mode = tools.PermissionModeOptOut
		default:
			fmt.Printf("Error: Unknown mode '%s'. Use ask, auto, auto-bp, or opt-out.\n", modeStr)
			return
		}

		pm := tools.GetAdvancedPermissionManager()
		if err := pm.SetToolMode(toolName, mode, tools.ScopeGlobal); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("✓ Set %s permission mode to %s\n", toolName, mode)
	},
}

var permissionsAddDeniedCmd = &cobra.Command{
	Use:   "deny [tool] [pattern]",
	Short: "Add a denied pattern for a tool",
	Long: `Add a denied command or path pattern for a specific tool.

Examples:
  claudego permissions deny bash "rm -rf /"
  claudego permissions deny edit "/etc/passwd"`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		toolName := args[0]
		pattern := args[1]

		pm := tools.GetAdvancedPermissionManager()

		// Determine if it's a command or path denial based on tool
		switch toolName {
		case "bash":
			if err := pm.AddDeniedCommand(toolName, pattern); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("✓ Added denied command pattern for %s: %s\n", toolName, pattern)
		case "edit", "write_file":
			if err := pm.AddDeniedPath(toolName, pattern); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Printf("✓ Added denied path pattern for %s: %s\n", toolName, pattern)
		default:
			fmt.Printf("Error: Tool %s doesn't support denied patterns\n", toolName)
		}
	},
}
