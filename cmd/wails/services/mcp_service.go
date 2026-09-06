package services

import (
	"github.com/user/go-claude-code/internal/mcp"
)

// MCPService exposes MCP servers, their state and tools for the rail MCP pane
// (T047, gui-contract §4). Operations delegate to the shared mcp.Manager.
type MCPService struct{}

func NewMCPService() *MCPService { return &MCPService{} }

// Server is the JSON-safe row for the MCP pane.
type Server struct {
	Name    string           `json:"name"`
	Command string           `json:"command"`
	Running bool             `json:"running"`
	Tools   []mcp.Tool       `json:"tools,omitempty"`
	Config  mcp.ServerConfig `json:"config"`
}

// List returns every configured server with its runtime state.
func (s *MCPService) List() []Server {
	m := mcp.GetManager()
	configs := m.ListServers()
	var out []Server
	for name, cfg := range configs {
		tools, err := m.ListServerTools(name)
		out = append(out, Server{
			Name:    name,
			Command: cfg.Command,
			Running: m.IsServerRunning(name),
			Tools:   toolsOrEmpty(tools, err),
			Config:  *cfg,
		})
	}
	return out
}

// Start launches a configured server and re-lists its tools.
func (s *MCPService) Start(name string) error { return mcp.GetManager().StartServer(name) }

// Stop halts a running server.
func (s *MCPService) Stop(name string) error { return mcp.GetManager().StopServer(name) }

// Add registers a new server (the "Añadir servidor" CTA).
func (s *MCPService) Add(cfg *mcp.ServerConfig) error {
	return mcp.GetManager().AddServer(cfg)
}

func toolsOrEmpty(tools []mcp.Tool, err error) []mcp.Tool {
	if err != nil {
		return nil
	}
	return tools
}