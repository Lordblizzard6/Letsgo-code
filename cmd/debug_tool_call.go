package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(debugToolCallCmd)
	debugToolCallCmd.Flags().StringP("tool", "t", "", "Tool name to debug")
	debugToolCallCmd.Flags().StringP("input", "i", "", "JSON input for tool")
}

var debugToolCallCmd = &cobra.Command{
	Use:   "debug-tool-call",
	Short: "Debug tool calls",
	Long:  `Debug and test tool calls with detailed output. Useful for development.`,
	Run: func(cmd *cobra.Command, args []string) {
		toolName, _ := cmd.Flags().GetString("tool")
		input, _ := cmd.Flags().GetString("input")

		if toolName == "" {
			fmt.Println("Usage: debug-tool-call --tool <name> --input '{...}'")
			fmt.Println("\nAvailable tools:")
			fmt.Println("  • bash - Execute shell commands")
			fmt.Println("  • file_read - Read file contents")
			fmt.Println("  • file_write - Write to files")
			fmt.Println("  • file_edit - Edit files")
			fmt.Println("  • glob - Search files by pattern")
			fmt.Println("  • grep - Search text in files")
			fmt.Println("  • ls - List directory contents")
			fmt.Println("  • todo - Manage todo lists")
			fmt.Println("  • web_search - Search the web")
			fmt.Println("  • browser_preview - Preview in browser")
			fmt.Println("  • list_resources - List MCP resources")
			fmt.Println("  • read_resource - Read MCP resource")
			fmt.Println("  • deploy_web_app - Deploy web apps")
			os.Exit(1)
		}

		fmt.Printf("🔧 Debugging Tool Call: %s\n", toolName)
		fmt.Println("==========================================")

		// Parse input
		var inputMap map[string]interface{}
		if input != "" {
			if err := json.Unmarshal([]byte(input), &inputMap); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing input JSON: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println("Input:")
		inputJSON, _ := json.MarshalIndent(inputMap, "", "  ")
		fmt.Println(string(inputJSON))
		fmt.Println()

		// Execute tool
		fmt.Println("Executing...")
		fmt.Println("------------")

		result, err := executeToolDebug(toolName, inputMap)

		fmt.Println()
		fmt.Println("Result:")
		fmt.Println("-------")

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Success!\n\n")
		fmt.Printf("Output:\n%s\n", result)
	},
}

func executeToolDebug(name string, input map[string]interface{}) (string, error) {
	// This is a simplified version - real implementation would call tools.ExecuteTool
	return fmt.Sprintf("Tool %s executed with input: %v\n(Result would appear here)", name, input), nil
}
