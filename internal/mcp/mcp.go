package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/tuusuario/mycli/internal/llm"
)

// MCPServer representa una conexión a un servidor MCP
type MCPServer struct {
	Name       string            `json:"name"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
	Timeout    time.Duration
	
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	requestID  int64
	mu         sync.Mutex
	tools      []llm.Tool
	connected  bool
	cancel     context.CancelFunc
}

// MCPRequest estructura de petición JSON-RPC
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse estructura de respuesta JSON-RPC
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError error de MCP
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolListResult resultado de listar herramientas
type ToolListResult struct {
	Tools []MCPTool `json:"tools"`
}

// MCPTool herramienta definida por MCP
type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// CallToolResult resultado de llamar herramienta
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock bloque de contenido MCP
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
}

// NewMCPServer crea una nueva instancia de servidor MCP
func NewMCPServer(name, command string, args []string, env map[string]string, timeout time.Duration) *MCPServer {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &MCPServer{
		Name:    name,
		Command: command,
		Args:    args,
		Env:     env,
		Timeout: timeout,
	}
}

// Connect inicia la conexión con el servidor MCP
func (m *MCPServer) Connect(ctx context.Context) error {
	if m.connected {
		return fmt.Errorf("server already connected")
	}
	
	ctx, m.cancel = context.WithCancel(ctx)
	m.cmd = exec.CommandContext(ctx, m.Command, m.Args...)
	
	// Configurar variables de entorno
	if len(m.Env) > 0 {
		envVars := make([]string, 0, len(m.Env))
		for k, v := range m.Env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		m.cmd.Env = append(m.cmd.Environ(), envVars...)
	}
	
	var err error
	m.stdin, err = m.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	
	m.stdout, err = m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	
	m.stderr, err = m.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	
	// Iniciar proceso
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}
	
	// Leer stderr en background
	go m.readStderr()
	
	// Inicializar protocolo MCP
	if err := m.initialize(ctx); err != nil {
		m.cmd.Process.Kill()
		return fmt.Errorf("initialize: %w", err)
	}
	
	// Listar herramientas disponibles
	if err := m.listTools(ctx); err != nil {
		m.cmd.Process.Kill()
		return fmt.Errorf("list tools: %w", err)
	}
	
	m.connected = true
	return nil
}

func (m *MCPServer) initialize(ctx context.Context) error {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      m.nextRequestID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			"clientInfo": map[string]interface{}{
				"name":    "mycli",
				"version": "1.0.0",
			},
		},
	}
	
	resp, err := m.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	
	// Enviar notificación initialized
	notif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.sendNotification(notif)
	}
}

func (m *MCPServer) listTools(ctx context.Context) error {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      m.nextRequestID(),
		Method:  "tools/list",
	}
	
	resp, err := m.sendRequest(ctx, req)
	if err != nil {
		return err
	}
	
	var result ToolListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("unmarshal tools: %w", err)
	}
	
	// Convertir a herramientas llm.Tool
	m.tools = make([]llm.Tool, 0, len(result.Tools))
	for _, t := range result.Tools {
		toolName := t.Name
		m.tools = append(m.tools, llm.Tool{
			Name:        t.Name,
			Description: t.Description,
			Schema:      t.InputSchema,
			Category:    "mcp",
			Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
				return m.callTool(ctx, toolName, args)
			},
		})
	}
	
	return nil
}

func (m *MCPServer) callTool(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      m.nextRequestID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": json.RawMessage(args),
		},
	}
	
	resp, err := m.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}
	
	var result CallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("unmarshal result: %w", err)
	}
	
	if result.IsError {
		return "", fmt.Errorf("tool execution error")
	}
	
	// Concatenar todo el contenido texto
	var output string
	for _, block := range result.Content {
		if block.Type == "text" {
			output += block.Text
		}
	}
	
	return output, nil
}

func (m *MCPServer) sendRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Serializar y enviar
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	if _, err := m.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}
	
	// Leer respuesta con timeout
	readCtx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	
	reader := bufio.NewReader(m.stdout)
	lineChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	
	go func() {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			errChan <- err
			return
		}
		lineChan <- line
	}()
	
	select {
	case <-readCtx.Done():
		return nil, readCtx.Err()
	case err := <-errChan:
		return nil, fmt.Errorf("read response: %w", err)
	case line := <-lineChan:
		var resp MCPResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
		
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		
		return &resp, nil
	}
}

func (m *MCPServer) sendNotification(req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}
	
	_, err = m.stdin.Write(append(data, '\n'))
	return err
}

func (m *MCPServer) readStderr() {
	reader := bufio.NewReader(m.stderr)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[%s] stderr: %s\n", m.Name, line)
			}
			return
		}
		fmt.Printf("[%s] stderr: %s\n", m.Name, line)
	}
}

func (m *MCPServer) nextRequestID() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestID++
	return m.requestID
}

// GetTools devuelve las herramientas disponibles
func (m *MCPServer) GetTools() []llm.Tool {
	return m.tools
}

// Disconnect cierra la conexión con el servidor MCP
func (m *MCPServer) Disconnect() error {
	if !m.connected {
		return nil
	}
	
	if m.cancel != nil {
		m.cancel()
	}
	
	if m.cmd != nil && m.cmd.Process != nil {
		return m.cmd.Process.Kill()
	}
	
	return nil
}

// IsConnected verifica si está conectado
func (m *MCPServer) IsConnected() bool {
	return m.connected
}

// Manager gestiona múltiples servidores MCP
type Manager struct {
	servers map[string]*MCPServer
	mu      sync.RWMutex
}

// NewManager crea un nuevo manager de MCP
func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*MCPServer),
	}
}

// AddServer añade un servidor MCP
func (mgr *Manager) AddServer(server *MCPServer) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.servers[server.Name] = server
}

// ConnectAll conecta todos los servidores
func (mgr *Manager) ConnectAll(ctx context.Context) error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	
	for _, server := range mgr.servers {
		if err := server.Connect(ctx); err != nil {
			return fmt.Errorf("connect %s: %w", server.Name, err)
		}
	}
	return nil
}

// GetAllTools obtiene todas las herramientas de todos los servidores
func (mgr *Manager) GetAllTools() []llm.Tool {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	
	var allTools []llm.Tool
	for _, server := range mgr.servers {
		allTools = append(allTools, server.GetTools()...)
	}
	return allTools
}

// GetServer obtiene un servidor por nombre
func (mgr *Manager) GetServer(name string) (*MCPServer, bool) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	server, exists := mgr.servers[name]
	return server, exists
}

// DisconnectAll desconecta todos los servidores
func (mgr *Manager) DisconnectAll() error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	
	var lastErr error
	for _, server := range mgr.servers {
		if err := server.Disconnect(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
