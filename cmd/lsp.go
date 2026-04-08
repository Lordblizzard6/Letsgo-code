package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/go-claude-code/internal/lsp"
)

func init() {
	rootCmd.AddCommand(lspCmd)
	lspCmd.AddCommand(lspStartCmd)
	lspCmd.AddCommand(lspStopCmd)
	lspCmd.AddCommand(lspListCmd)
	lspCmd.AddCommand(lspHoverCmd)
	lspCmd.AddCommand(lspDefinitionCmd)
	lspCmd.AddCommand(lspCompleteCmd)
}

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Manage Language Server Protocol (LSP) servers",
	Long:  `Manage LSP servers for code intelligence features like go-to-definition, hover, and completions.`,
}

var lspStartCmd = &cobra.Command{
	Use:   "start [language]",
	Short: "Start an LSP server",
	Long: `Start an LSP server for a programming language.

Supported languages:
  - go (requires gopls)
  - typescript/javascript (requires typescript-language-server)
  - python (requires pylsp or pyright)
  - rust (requires rust-analyzer)
  - c/c++ (requires clangd)

Example:
  claudego lsp start go`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		language := strings.ToLower(args[0])

		// Get project root
		rootPath, _ := os.Getwd()

		// Check if server already running
		manager := lsp.GetManager()
		if manager.IsServerRunning(language) {
			fmt.Printf("LSP server for %s is already running\n", language)
			return
		}

		fmt.Printf("Starting LSP server for %s...\n", language)
		if err := manager.StartServer(language, rootPath); err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nMake sure you have the language server installed:")
			switch language {
			case "go":
				fmt.Println("  go install golang.org/x/tools/gopls@latest")
			case "typescript", "javascript", "ts", "js":
				fmt.Println("  npm install -g typescript-language-server")
			case "python", "py":
				fmt.Println("  pip install python-lsp-server")
			case "rust", "rs":
				fmt.Println("  rustup component add rust-analyzer")
			case "c", "cpp":
				fmt.Println("  # Install clangd from your package manager")
			}
			return
		}

		fmt.Printf("✓ LSP server for %s started\n", language)
	},
}

var lspStopCmd = &cobra.Command{
	Use:   "stop [language]",
	Short: "Stop an LSP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		language := strings.ToLower(args[0])

		manager := lsp.GetManager()
		if err := manager.StopServer(language); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

var lspListCmd = &cobra.Command{
	Use:   "list",
	Short: "List running LSP servers",
	Run: func(cmd *cobra.Command, args []string) {
		manager := lsp.GetManager()
		servers := manager.GetRunningServers()

		if len(servers) == 0 {
			fmt.Println("No LSP servers running.")
			fmt.Println("\nStart a server with:")
			fmt.Println("  claudego lsp start <language>")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "LANGUAGE\tSTATUS")
		fmt.Fprintln(w, "--------\t------")
		for _, lang := range servers {
			fmt.Fprintf(w, "%s\trunning\n", lang)
		}
		w.Flush()
	},
}

var lspHoverCmd = &cobra.Command{
	Use:   "hover [language] [file] [line] [character]",
	Short: "Get hover information at position",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		language := strings.ToLower(args[0])
		filePath := args[1]
		line, _ := strconv.Atoi(args[2])
		character, _ := strconv.Atoi(args[3])

		manager := lsp.GetManager()
		hover, err := manager.GetHover(language, filePath, line, character)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if hover == nil || hover.Contents.Value == "" {
			fmt.Println("No hover information available")
			return
		}

		fmt.Println(hover.Contents.Value)
	},
}

var lspDefinitionCmd = &cobra.Command{
	Use:   "definition [language] [file] [line] [character]",
	Short: "Get definition at position",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		language := strings.ToLower(args[0])
		filePath := args[1]
		line, _ := strconv.Atoi(args[2])
		character, _ := strconv.Atoi(args[3])

		manager := lsp.GetManager()
		locations, err := manager.GetDefinition(language, filePath, line, character)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(locations) == 0 {
			fmt.Println("No definition found")
			return
		}

		fmt.Println("Definitions:")
		for _, loc := range locations {
			path := strings.TrimPrefix(loc.URI, "file://")
			fmt.Printf("  %s:%d:%d\n", path, loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		}
	},
}

var lspCompleteCmd = &cobra.Command{
	Use:   "complete [language] [file] [line] [character]",
	Short: "Get completions at position",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		language := strings.ToLower(args[0])
		filePath := args[1]
		line, _ := strconv.Atoi(args[2])
		character, _ := strconv.Atoi(args[3])

		manager := lsp.GetManager()
		items, err := manager.GetCompletions(language, filePath, line, character)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if len(items) == 0 {
			fmt.Println("No completions available")
			return
		}

		fmt.Printf("Completions (%d found):\n", len(items))
		for i, item := range items {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(items)-10)
				break
			}
			detail := item.Detail
			if detail != "" {
				fmt.Printf("  %s - %s\n", item.Label, detail)
			} else {
				fmt.Printf("  %s\n", item.Label)
			}
		}
	},
}

// detectLanguage detects the programming language from file extension
func detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".letsGo":
		return ".letsGo"
	case ".ts":
		return "typescript"
	case ".js":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	default:
		return ""
	}
}
