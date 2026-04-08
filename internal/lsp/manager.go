package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ServerConfig represents an LSP server configuration
type ServerConfig struct {
	Name       string            `json:"name"`
	Language   string            `json:"language"`
	Command    string            `json:"command"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	RootURI    string            `json:"root_uri,omitempty"`
}

// Server represents a running LSP server instance
type Server struct {
	Config       *ServerConfig
	Cmd          *exec.Cmd
	Stdin        io.WriteCloser
	Stdout       io.ReadCloser
	Stderr       io.ReadCloser
	Mu           sync.RWMutex
	Started      bool
	RequestID    int
	Capabilities ServerCapabilities
	Pending      map[int]chan *JSONRPCResponse
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	TextDocumentSync       int  `json:"textDocumentSync,omitempty"`
	HoverProvider          bool `json:"hoverProvider,omitempty"`
	CompletionProvider     *struct {
		TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	} `json:"completionProvider,omitempty"`
	DefinitionProvider     bool `json:"definitionProvider,omitempty"`
	ReferencesProvider     bool `json:"referencesProvider,omitempty"`
	DocumentSymbolProvider bool `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider     bool `json:"codeActionProvider,omitempty"`
	RenameProvider         bool `json:"renameProvider,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JSONRPCNotification represents a JSON-RPC notification (no ID)
type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Manager handles LSP servers
type Manager struct {
	servers map[string]*Server
	mu      sync.RWMutex
}

var (
	managerInstance *Manager
	managerOnce     sync.Once
)

// GetManager returns the singleton LSP manager
func GetManager() *Manager {
	managerOnce.Do(func() {
		managerInstance = &Manager{
			servers: make(map[string]*Server),
		}
	})
	return managerInstance
}

// StartServer starts an LSP server for a language
func (m *Manager) StartServer(language string, rootPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if server, exists := m.servers[language]; exists && server.Started {
		return fmt.Errorf("LSP server for %s is already running", language)
	}

	// Detect server configuration
	config := m.detectServerConfig(language, rootPath)
	if config == nil {
		return fmt.Errorf("no LSP server available for %s", language)
	}

	// Create command
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Get pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Create server instance
	server := &Server{
		Config:  config,
		Cmd:     cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Started: true,
		Pending: make(map[int]chan *JSONRPCResponse),
	}

	m.servers[language] = server

	// Start reading responses
	go m.readResponses(language)
	go m.readStderr(language, stderr)

	// Initialize the server
	if err := m.initializeServer(language, rootPath); err != nil {
		m.stopServerLocked(language)
		return fmt.Errorf("failed to initialize LSP server: %w", err)
	}

	return nil
}

// detectServerConfig detects LSP server configuration for a language
func (m *Manager) detectServerConfig(language string, rootPath string) *ServerConfig {
	switch language {
	case "go":
		return &ServerConfig{
			Name:     "gopls",
			Language: "go",
			Command:  "gopls",
			Args:     []string{"serve"},
			RootURI:  pathToURI(rootPath),
		}
	case "typescript", "javascript", "ts", "js":
		return &ServerConfig{
			Name:     "typescript-language-server",
			Language: "typescript",
			Command:  "typescript-language-server",
			Args:     []string{"--stdio"},
			RootURI:  pathToURI(rootPath),
		}
	case "python", "py":
		// Try pylsp first, then fallback to pyright
		return &ServerConfig{
			Name:     "pylsp",
			Language: "python",
			Command:  "pylsp",
			RootURI:  pathToURI(rootPath),
		}
	case "rust", "rs":
		return &ServerConfig{
			Name:     "rust-analyzer",
			Language: "rust",
			Command:  "rust-analyzer",
			RootURI:  pathToURI(rootPath),
		}
	case "c", "cpp", "h", "hpp":
		return &ServerConfig{
			Name:     "clangd",
			Language: "cpp",
			Command:  "clangd",
			Args:     []string{"--background-index"},
			RootURI:  pathToURI(rootPath),
		}
	default:
		return nil
	}
}

// initializeServer sends initialize request
func (m *Manager) initializeServer(language string, rootPath string) error {
	server, exists := m.servers[language]
	if !exists {
		return fmt.Errorf("server not found")
	}

	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootPath":  rootPath,
		"rootUri":   pathToURI(rootPath),
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"synchronization": map[string]interface{}{
					"dynamicRegistration": false,
					"willSave":            true,
					"willSaveWaitUntil":   true,
					"didSave":             true,
				},
				"completion": map[string]interface{}{
					"dynamicRegistration": false,
					"completionItem": map[string]interface{}{
						"snippetSupport": true,
					},
				},
				"hover": map[string]interface{}{
					"dynamicRegistration": false,
				},
				"definition": map[string]interface{}{
					"dynamicRegistration": false,
					"linkSupport":         true,
				},
				"documentSymbol": map[string]interface{}{
					"dynamicRegistration": false,
				},
			},
		},
		"workspaceFolders": []map[string]string{
			{
				"uri":  pathToURI(rootPath),
				"name": filepath.Base(rootPath),
			},
		},
	}

	resp, err := m.sendRequest(language, "initialize", params)
	if err != nil {
		return err
	}

	// Parse capabilities
	var result struct {
		Capabilities ServerCapabilities `json:"capabilities"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return err
	}

	server.Capabilities = result.Capabilities

	// Send initialized notification
	m.sendNotification(language, "initialized", map[string]interface{}{})

	fmt.Printf("✓ LSP server for %s initialized\n", language)
	return nil
}

// readResponses reads JSON-RPC responses from the server
func (m *Manager) readResponses(language string) {
	server, exists := m.servers[language]
	if !exists {
		return
	}

	reader := bufio.NewReader(server.Stdout)
	for {
		// Read Content-Length header
		var contentLength int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "Content-Length: ") {
				fmt.Sscanf(line, "Content-Length: %d", &contentLength)
			}
		}

		if contentLength == 0 {
			continue
		}

		// Read the content
		content := make([]byte, contentLength)
		_, err := io.ReadFull(reader, content)
		if err != nil {
			return
		}

		// Parse response
		var resp JSONRPCResponse
		if err := json.Unmarshal(content, &resp); err != nil {
			continue
		}

		// Handle response
		if resp.ID != 0 {
			server.Mu.Lock()
			if ch, exists := server.Pending[resp.ID]; exists {
				ch <- &resp
				delete(server.Pending, resp.ID)
			}
			server.Mu.Unlock()
		}
	}
}

// readStderr reads stderr output
func (m *Manager) readStderr(language string, stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(os.Stderr, "[LSP %s] %s\n", language, line)
	}
}

// sendRequest sends a JSON-RPC request
func (m *Manager) sendRequest(language string, method string, params interface{}) (*JSONRPCResponse, error) {
	server, exists := m.servers[language]
	if !exists {
		return nil, fmt.Errorf("server for %s not found", language)
	}

	if !server.Started {
		return nil, fmt.Errorf("server for %s is not running", language)
	}

	server.Mu.Lock()
	server.RequestID++
	reqID := server.RequestID
	ch := make(chan *JSONRPCResponse, 1)
	server.Pending[reqID] = ch
	server.Mu.Unlock()

	// Marshal params
	var paramsRaw json.RawMessage
	if params != nil {
		var err error
		paramsRaw, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	// Create request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  paramsRaw,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Send with Content-Length header (LSP spec)
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(reqData), reqData)
	if _, err := server.Stdin.Write([]byte(msg)); err != nil {
		return nil, err
	}

	// Wait for response
	resp := <-ch
	return resp, nil
}

// sendNotification sends a notification
func (m *Manager) sendNotification(language string, method string, params interface{}) error {
	server, exists := m.servers[language]
	if !exists || !server.Started {
		return nil
	}

	// Marshal params
	var paramsRaw json.RawMessage
	if params != nil {
		var err error
		paramsRaw, err = json.Marshal(params)
		if err != nil {
			return err
		}
	}

	// Create notification
	notif := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsRaw,
	}

	notifData, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	// Send with Content-Length header
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(notifData), notifData)
	_, err = server.Stdin.Write([]byte(msg))
	return err
}

// StopServer stops an LSP server
func (m *Manager) StopServer(language string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stopServerLocked(language)
}

func (m *Manager) stopServerLocked(language string) error {
	server, exists := m.servers[language]
	if !exists || !server.Started {
		return fmt.Errorf("server for %s is not running", language)
	}

	// Send shutdown request
	m.sendRequest(language, "shutdown", nil)
	m.sendNotification(language, "exit", nil)

	// Close stdin
	if server.Stdin != nil {
		server.Stdin.Close()
	}

	// Kill the process
	if server.Cmd.Process != nil {
		server.Cmd.Process.Kill()
	}

	server.Started = false
	delete(m.servers, language)

	fmt.Printf("✓ LSP server for %s stopped\n", language)
	return nil
}

// GetDefinition gets definition for a symbol
func (m *Manager) GetDefinition(language string, filePath string, line int, character int) ([]Location, error) {
	params := map[string]interface{}{
		"textDocument": map[string]string{
			"uri": pathToURI(filePath),
		},
		"position": map[string]int{
			"line":      line,
			"character": character,
		},
	}

	resp, err := m.sendRequest(language, "textDocument/definition", params)
	if err != nil {
		return nil, err
	}

	var locations []Location
	if err := json.Unmarshal(resp.Result, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// GetHover gets hover information
func (m *Manager) GetHover(language string, filePath string, line int, character int) (*Hover, error) {
	params := map[string]interface{}{
		"textDocument": map[string]string{
			"uri": pathToURI(filePath),
		},
		"position": map[string]int{
			"line":      line,
			"character": character,
		},
	}

	resp, err := m.sendRequest(language, "textDocument/hover", params)
	if err != nil {
		return nil, err
	}

	var hover Hover
	if err := json.Unmarshal(resp.Result, &hover); err != nil {
		return nil, err
	}

	return &hover, nil
}

// GetCompletions gets completion items
func (m *Manager) GetCompletions(language string, filePath string, line int, character int) ([]CompletionItem, error) {
	params := map[string]interface{}{
		"textDocument": map[string]string{
			"uri": pathToURI(filePath),
		},
		"position": map[string]int{
			"line":      line,
			"character": character,
		},
	}

	resp, err := m.sendRequest(language, "textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []CompletionItem `json:"items"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		// Try direct array
		if err := json.Unmarshal(resp.Result, &result.Items); err != nil {
			return nil, err
		}
	}

	return result.Items, nil
}

// DocumentOpened notifies server that document was opened
func (m *Manager) DocumentOpened(language string, filePath string, content string, version int) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        pathToURI(filePath),
			"languageId": language,
			"version":    version,
			"text":       content,
		},
	}

	return m.sendNotification(language, "textDocument/didOpen", params)
}

// DocumentChanged notifies server that document changed
func (m *Manager) DocumentChanged(language string, filePath string, content string, version int) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     pathToURI(filePath),
			"version": version,
		},
		"contentChanges": []map[string]interface{}{
			{
				"text": content,
			},
		},
	}

	return m.sendNotification(language, "textDocument/didChange", params)
}

// DocumentClosed notifies server that document was closed
func (m *Manager) DocumentClosed(language string, filePath string) error {
	params := map[string]interface{}{
		"textDocument": map[string]string{
			"uri": pathToURI(filePath),
		},
	}

	return m.sendNotification(language, "textDocument/didClose", params)
}

// IsServerRunning checks if a server is running
func (m *Manager) IsServerRunning(language string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[language]
	return exists && server.Started
}

// GetRunningServers returns list of running servers
func (m *Manager) GetRunningServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var languages []string
	for lang, server := range m.servers {
		if server.Started {
			languages = append(languages, lang)
		}
	}
	return languages
}

// Location represents a location in a document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Range represents a range in a document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position represents a position in a document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Hover represents hover information
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents markup content
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// pathToURI converts a file path to URI format
func pathToURI(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "file://" + path
	}
	return "file://" + absPath
}

// uriToPath converts URI to file path
func uriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}
