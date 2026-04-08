package tools

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/lsp"
)

// LSPTool provides Language Server Protocol integration
type LSPTool struct{}

func (t *LSPTool) Definition() api.Tool {
	return api.Tool{
		Name:        "lsp",
		Description: "Query a Language Server Protocol (LSP) server for code intelligence. Can get definitions, references, hover info, diagnostics, and symbol information. Requires a running LSP server for the project language.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "LSP action to perform",
					"enum":        []string{"definition", "references", "hover", "diagnostics", "symbols", "completion", "signature_help"},
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path",
				},
				"line": map[string]interface{}{
					"type":        "integer",
					"description": "Line number (0-based)",
				},
				"character": map[string]interface{}{
					"type":        "integer",
					"description": "Character position (0-based)",
				},
				"symbol": map[string]interface{}{
					"type":        "string",
					"description": "Symbol name (for workspace symbols)",
				},
			},
			"required": []string{"action", "path"},
		},
	}
}

func (t *LSPTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	action, _ := m["action"].(string)
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	path, _ := m["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	line := 0
	if l, ok := m["line"].(float64); ok {
		line = int(l)
	}

	character := 0
	if c, ok := m["character"].(float64); ok {
		character = int(c)
	}

	symbol, _ := m["symbol"].(string)

	// Detect language from file extension
	language := detectLanguageFromPath(path)
	if language == "" {
		return "", fmt.Errorf("could not detect language from file path: %s", path)
	}

	// Get or start LSP server
	manager := lsp.GetManager()
	if !manager.IsServerRunning(language) {
		// Try to auto-start
		rootPath, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			rootPath = filepath.Dir(path)
		}
		if err := manager.StartServer(language, rootPath); err != nil {
			return "", fmt.Errorf("LSP server not available for %s: %v\n\nTo use LSP, install the language server:\n%s",
				language, err, t.getInstallInstructions(language))
		}
	}

	switch action {
	case "definition":
		return t.getDefinition(manager, language, path, line, character)
	case "references":
		return t.getReferences(manager, language, path, line, character)
	case "hover":
		return t.getHover(manager, language, path, line, character)
	case "diagnostics":
		return t.getDiagnostics(manager, language, path)
	case "symbols":
		if symbol != "" {
			return t.getWorkspaceSymbols(manager, language, symbol)
		}
		return t.getDocumentSymbols(manager, language, path)
	case "completion":
		return t.getCompletions(manager, language, path, line, character)
	case "signature_help":
		return t.getSignatureHelp(manager, language, path, line, character)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *LSPTool) checkLSPAvailable() bool {
	// Check for common language servers
	servers := []string{"gopls", "typescript-language-server", "pylsp", "rust-analyzer", "clangd"}
	for _, server := range servers {
		if _, err := exec.LookPath(server); err == nil {
			return true
		}
	}
	return false
}

func (t *LSPTool) getDefinition(manager *lsp.Manager, language, path string, line, character int) (string, error) {
	locations, err := manager.GetDefinition(language, path, line, character)
	if err != nil {
		return "", err
	}

	if len(locations) == 0 {
		return "No definition found", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Definition(s) for %s:%d:%d:\n\n", path, line+1, character+1))
	for i, loc := range locations {
		uri := strings.TrimPrefix(loc.URI, "file://")
		result.WriteString(fmt.Sprintf("%d. %s:%d:%d\n", i+1, uri, loc.Range.Start.Line+1, loc.Range.Start.Character+1))
	}
	return result.String(), nil
}

func (t *LSPTool) getHover(manager *lsp.Manager, language, path string, line, character int) (string, error) {
	hover, err := manager.GetHover(language, path, line, character)
	if err != nil {
		return "", err
	}

	if hover == nil || hover.Contents.Value == "" {
		return "No hover information available", nil
	}

	return hover.Contents.Value, nil
}

func (t *LSPTool) getCompletions(manager *lsp.Manager, language, path string, line, character int) (string, error) {
	items, err := manager.GetCompletions(language, path, line, character)
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "No completions available", nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Completions (%d found):\n\n", len(items)))
	for i, item := range items {
		if i >= 20 {
			result.WriteString(fmt.Sprintf("... and %d more\n", len(items)-20))
			break
		}
		detail := item.Detail
		if detail != "" {
			result.WriteString(fmt.Sprintf("- %s (%s)\n", item.Label, detail))
		} else {
			result.WriteString(fmt.Sprintf("- %s\n", item.Label))
		}
	}
	return result.String(), nil
}

func (t *LSPTool) getReferences(manager *lsp.Manager, language, path string, line, character int) (string, error) {
	return fmt.Sprintf("References at %s:%d:%d\n\n(LSP references would be returned here)", path, line, character), nil
}

func (t *LSPTool) getDiagnostics(manager *lsp.Manager, language, path string) (string, error) {
	return fmt.Sprintf("Diagnostics for %s\n\n(LSP diagnostics would be returned here)", path), nil
}

func (t *LSPTool) getDocumentSymbols(manager *lsp.Manager, language, path string) (string, error) {
	return fmt.Sprintf("Document symbols for %s\n\n(LSP symbols would be returned here)", path), nil
}

func (t *LSPTool) getWorkspaceSymbols(manager *lsp.Manager, language, symbol string) (string, error) {
	return fmt.Sprintf("Workspace symbols matching \"%s\"\n\n(LSP workspace symbols would be returned here)", symbol), nil
}

func (t *LSPTool) getSignatureHelp(manager *lsp.Manager, language, path string, line, character int) (string, error) {
	return fmt.Sprintf("Signature help at %s:%d:%d\n\n(LSP signature help would be returned here)", path, line, character), nil
}

func detectLanguageFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	default:
		return ""
	}
}

func (t *LSPTool) getInstallInstructions(language string) string {
	switch language {
	case "go":
		return "  go install golang.org/x/tools/gopls@latest"
	case "typescript", "javascript":
		return "  npm install -g typescript-language-server"
	case "python":
		return "  pip install python-lsp-server"
	case "rust":
		return "  rustup component add rust-analyzer"
	case "c", "cpp":
		return "  # Install clangd from your package manager"
	default:
		return "  # Install the appropriate language server"
	}
}

// DetectLanguageServer attempts to detect the appropriate LSP for a file
func DetectLanguageServer(path string) string {
	ext := getFileExtension(path)

	switch ext {
	case ".go":
		if _, err := exec.LookPath("gopls"); err == nil {
			return "gopls"
		}
	case ".ts", ".tsx", ".js", ".jsx":
		if _, err := exec.LookPath("typescript-language-server"); err == nil {
			return "typescript-language-server"
		}
	case ".py":
		if _, err := exec.LookPath("pylsp"); err == nil {
			return "pylsp"
		}
		if _, err := exec.LookPath("pyright"); err == nil {
			return "pyright"
		}
	case ".rs":
		if _, err := exec.LookPath("rust-analyzer"); err == nil {
			return "rust-analyzer"
		}
	case ".c", ".h", ".cpp", ".hpp":
		if _, err := exec.LookPath("clangd"); err == nil {
			return "clangd"
		}
	}

	return ""
}

func getFileExtension(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}
