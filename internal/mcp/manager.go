package mcp

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ServerConfig represents an MCP server configuration
type ServerConfig struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	AutoStart   bool              `json:"auto_start,omitempty"`
	Description string            `json:"description,omitempty"`
}

// Config represents the MCP configuration file
type Config struct {
	Version string                   `json:"version"`
	Servers map[string]*ServerConfig `json:"servers"`
}

// Server represents a running MCP server instance
type Server struct {
	Config    *ServerConfig
	Cmd       *exec.Cmd
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser
	Tools     []Tool
	Resources []Resource
	Prompts   []Prompt
	Mu        sync.RWMutex
	Started   bool
	LastError string
}

// Tool represents an MCP tool
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Manager handles MCP servers
type Manager struct {
	config     *Config
	servers    map[string]*Server
	configPath string
	mu         sync.RWMutex
}

var (
	managerInstance *Manager
	managerOnce     sync.Once
)

// GetManager returns the singleton MCP manager
func GetManager() *Manager {
	managerOnce.Do(func() {
		managerInstance = &Manager{
			config: &Config{
				Version: "1.0",
				Servers: make(map[string]*ServerConfig),
			},
			servers: make(map[string]*Server),
		}
		managerInstance.loadConfig()
	})
	return managerInstance
}

// getConfigPath returns the path to the MCP config file
func (m *Manager) getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "mcp.json")
}

// loadConfig loads MCP configuration from disk
func (m *Manager) loadConfig() {
	configPath := m.getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config file yet, that's ok
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading MCP config: %v\n", err)
		return
	}

	m.config = &config

	// Auto-start servers marked for auto-start
	for name, serverConfig := range config.Servers {
		if serverConfig.AutoStart {
			go m.StartServer(name)
		}
	}
}

// saveConfig saves MCP configuration to disk
func (m *Manager) saveConfig() error {
	configPath := m.getConfigPath()

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// AddServer adds a new MCP server configuration
func (m *Manager) AddServer(config *ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Servers[config.Name]; exists {
		return fmt.Errorf("server '%s' already exists", config.Name)
	}

	m.config.Servers[config.Name] = config
	return m.saveConfig()
}

// RemoveServer removes an MCP server configuration
func (m *Manager) RemoveServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop if running
	if server, exists := m.servers[name]; exists && server.Started {
		m.stopServerLocked(name)
	}

	delete(m.config.Servers, name)
	return m.saveConfig()
}

// GetServer returns a server configuration
func (m *Manager) GetServer(name string) (*ServerConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.config.Servers[name]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", name)
	}

	return config, nil
}

// ListServers returns all configured servers
func (m *Manager) ListServers() map[string]*ServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	result := make(map[string]*ServerConfig)
	for k, v := range m.config.Servers {
		result[k] = v
	}
	return result
}

// StartServer starts an MCP server
func (m *Manager) StartServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.startServerLocked(name)
}

func (m *Manager) startServerLocked(name string) error {
	config, exists := m.config.Servers[name]
	if !exists {
		return fmt.Errorf("server '%s' not found", config.Name)
	}

	// Check if already running
	if server, exists := m.servers[name]; exists && server.Started {
		return fmt.Errorf("server '%s' is already running", name)
	}

	// Create command
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)

	// Set working directory
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

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
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Create server instance
	server := &Server{
		Config:    config,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		Tools:     make([]Tool, 0),
		Resources: make([]Resource, 0),
		Prompts:   make([]Prompt, 0),
		Started:   true,
	}

	m.servers[name] = server

	// Start reading stderr in background
	go m.readStderr(name, stderr)

	// Initialize the connection
	if err := m.initializeServer(name); err != nil {
		m.stopServerLocked(name)
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	return nil
}

// readStderr reads stderr output from a server
func (m *Manager) readStderr(name string, stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// Log or handle stderr output
		fmt.Fprintf(os.Stderr, "[%s] %s\n", name, line)
	}
}

// initializeServer sends initialize request to server
func (m *Manager) initializeServer(name string) error {
	_, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("server not found")
	}

	// Send initialize request
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": true,
			},
			"sampling": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "claude-code-go",
			"version": "0.1.0",
		},
	}

	result, err := m.sendRequest(name, "initialize", initParams)
	if err != nil {
		return err
	}

	// Parse server capabilities
	var initResult struct {
		ProtocolVersion string `json:"protocolVersion"`
		Capabilities    struct {
			Tools     *struct{} `json:"tools,omitempty"`
			Resources *struct{} `json:"resources,omitempty"`
			Prompts   *struct{} `json:"prompts,omitempty"`
		} `json:"capabilities"`
		ServerInfo struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}

	if err := json.Unmarshal(result, &initResult); err != nil {
		return err
	}

	fmt.Printf("✓ MCP server '%s' initialized (%s v%s)\n",
		name, initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// Send initialized notification
	m.sendNotification(name, "notifications/initialized", nil)

	// Fetch tools if supported
	if initResult.Capabilities.Tools != nil {
		if err := m.fetchTools(name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch tools from %s: %v\n", name, err)
		}
	}

	// Fetch resources if supported
	if initResult.Capabilities.Resources != nil {
		if err := m.fetchResources(name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch resources from %s: %v\n", name, err)
		}
	}

	// Fetch prompts if supported
	if initResult.Capabilities.Prompts != nil {
		if err := m.fetchPrompts(name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch prompts from %s: %v\n", name, err)
		}
	}

	return nil
}

// fetchTools fetches available tools from server
func (m *Manager) fetchTools(name string) error {
	result, err := m.sendRequest(name, "tools/list", nil)
	if err != nil {
		return err
	}

	var listResult struct {
		Tools []Tool `json:"tools"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return err
	}

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("server not found")
	}

	server.Mu.Lock()
	server.Tools = listResult.Tools
	server.Mu.Unlock()

	fmt.Printf("  Found %d tools\n", len(listResult.Tools))
	return nil
}

// fetchResources fetches available resources from server
func (m *Manager) fetchResources(name string) error {
	result, err := m.sendRequest(name, "resources/list", nil)
	if err != nil {
		return err
	}

	var listResult struct {
		Resources []Resource `json:"resources"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return err
	}

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("server not found")
	}

	server.Mu.Lock()
	server.Resources = listResult.Resources
	server.Mu.Unlock()

	fmt.Printf("  Found %d resources\n", len(listResult.Resources))
	return nil
}

// fetchPrompts fetches available prompts from server
func (m *Manager) fetchPrompts(name string) error {
	result, err := m.sendRequest(name, "prompts/list", nil)
	if err != nil {
		return err
	}

	var listResult struct {
		Prompts []Prompt `json:"prompts"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return err
	}

	server, exists := m.servers[name]
	if !exists {
		return fmt.Errorf("server not found")
	}

	server.Mu.Lock()
	server.Prompts = listResult.Prompts
	server.Mu.Unlock()

	fmt.Printf("  Found %d prompts\n", len(listResult.Prompts))
	return nil
}

// StopServer stops a running MCP server
func (m *Manager) StopServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stopServerLocked(name)
}

func (m *Manager) stopServerLocked(name string) error {
	server, exists := m.servers[name]
	if !exists || !server.Started {
		return fmt.Errorf("server '%s' is not running", name)
	}

	// Send shutdown request (optional)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Close stdin to signal EOF
	if server.Stdin != nil {
		server.Stdin.Close()
	}

	// Wait for process to exit or timeout
	done := make(chan error, 1)
	go func() {
		done <- server.Cmd.Wait()
	}()

	select {
	case <-done:
		// Process exited cleanly
	case <-ctx.Done():
		// Timeout, kill the process
		server.Cmd.Process.Kill()
	}

	server.Started = false
	delete(m.servers, name)

	fmt.Printf("✓ MCP server '%s' stopped\n", name)
	return nil
}

// sendRequest sends a JSON-RPC request to a server
func (m *Manager) sendRequest(serverName string, method string, params interface{}) (json.RawMessage, error) {
	server, exists := m.servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	if !server.Started {
		return nil, fmt.Errorf("server '%s' is not running", serverName)
	}

	// Generate request ID
	reqID := fmt.Sprintf("%d", time.Now().UnixNano())

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

	// Send request
	if _, err := fmt.Fprintf(server.Stdin, "%s\n", reqData); err != nil {
		return nil, err
	}

	// Read response
	reader := bufio.NewReader(server.Stdout)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue // Skip non-JSON lines
		}

		// Check if this is our response
		if fmt.Sprintf("%v", resp.ID) == reqID {
			if resp.Error != nil {
				return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
			}
			return resp.Result, nil
		}
	}
}

// sendNotification sends a notification (no response expected)
func (m *Manager) sendNotification(serverName string, method string, params interface{}) {
	server, exists := m.servers[serverName]
	if !exists || !server.Started {
		return
	}

	// Create notification (no ID)
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		paramsRaw, err := json.Marshal(params)
		if err != nil {
			return
		}
		req.Params = paramsRaw
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return
	}

	fmt.Fprintf(server.Stdin, "%s\n", reqData)
}

// CallTool calls a tool on an MCP server
func (m *Manager) CallTool(serverName string, toolName string, args map[string]interface{}) (string, error) {
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}

	result, err := m.sendRequest(serverName, "tools/call", params)
	if err != nil {
		return "", err
	}

	// Parse result
	var toolResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError,omitempty"`
	}

	if err := json.Unmarshal(result, &toolResult); err != nil {
		return "", err
	}

	if toolResult.IsError {
		return "", fmt.Errorf("tool execution failed")
	}

	// Combine all text content
	var texts []string
	for _, content := range toolResult.Content {
		if content.Type == "text" {
			texts = append(texts, content.Text)
		}
	}

	return strings.Join(texts, "\n"), nil
}

// ReadResource reads a resource from an MCP server
func (m *Manager) ReadResource(serverName string, uri string) ([]byte, error) {
	params := map[string]interface{}{
		"uri": uri,
	}

	result, err := m.sendRequest(serverName, "resources/read", params)
	if err != nil {
		return nil, err
	}

	// Parse result - could be text or binary
	var resourceResult struct {
		Contents []struct {
			URI      string `json:"uri"`
			MimeType string `json:"mimeType,omitempty"`
			Text     string `json:"text,omitempty"`
			Blob     string `json:"blob,omitempty"`
		} `json:"contents"`
	}

	if err := json.Unmarshal(result, &resourceResult); err != nil {
		return nil, err
	}

	if len(resourceResult.Contents) == 0 {
		return nil, fmt.Errorf("no content returned")
	}

	content := resourceResult.Contents[0]

	// Return text as bytes, or decode blob
	if content.Text != "" {
		return []byte(content.Text), nil
	}

	if content.Blob != "" {
		// Base64 decode
		return base64.StdEncoding.DecodeString(content.Blob)
	}

	return nil, fmt.Errorf("empty content")
}

// GetPrompt gets a prompt from an MCP server
func (m *Manager) GetPrompt(serverName string, promptName string, args map[string]string) (string, error) {
	params := map[string]interface{}{
		"name":      promptName,
		"arguments": args,
	}

	result, err := m.sendRequest(serverName, "prompts/get", params)
	if err != nil {
		return "", err
	}

	// Parse result
	var promptResult struct {
		Description string `json:"description,omitempty"`
		Messages    []struct {
			Role    string `json:"role"`
			Content struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(result, &promptResult); err != nil {
		return "", err
	}

	// Combine all messages
	var texts []string
	for _, msg := range promptResult.Messages {
		texts = append(texts, msg.Content.Text)
	}

	return strings.Join(texts, "\n"), nil
}

// ListServerTools returns all tools from a server
func (m *Manager) ListServerTools(serverName string) ([]Tool, error) {
	server, exists := m.servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	if !server.Started {
		return nil, fmt.Errorf("server '%s' is not running", serverName)
	}

	server.Mu.RLock()
	defer server.Mu.RUnlock()

	// Return a copy
	tools := make([]Tool, len(server.Tools))
	copy(tools, server.Tools)

	return tools, nil
}

// ListServerResources returns all resources from a server
func (m *Manager) ListServerResources(serverName string) ([]Resource, error) {
	server, exists := m.servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	if !server.Started {
		return nil, fmt.Errorf("server '%s' is not running", serverName)
	}

	server.Mu.RLock()
	defer server.Mu.RUnlock()

	resources := make([]Resource, len(server.Resources))
	copy(resources, server.Resources)

	return resources, nil
}

// ListAllTools returns tools from all running servers
func (m *Manager) ListAllTools() map[string][]Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]Tool)
	for name, server := range m.servers {
		if server.Started {
			server.Mu.RLock()
			tools := make([]Tool, len(server.Tools))
			copy(tools, server.Tools)
			server.Mu.RUnlock()
			result[name] = tools
		}
	}

	return result
}

// ListAllResources returns resources from all running servers
func (m *Manager) ListAllResources() map[string][]Resource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]Resource)
	for name, server := range m.servers {
		if server.Started {
			server.Mu.RLock()
			resources := make([]Resource, len(server.Resources))
			copy(resources, server.Resources)
			server.Mu.RUnlock()
			result[name] = resources
		}
	}

	return result
}

// IsServerRunning checks if a server is running
func (m *Manager) IsServerRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[name]
	return exists && server.Started
}
