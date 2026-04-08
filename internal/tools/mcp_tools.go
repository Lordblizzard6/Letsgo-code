package tools

import (
	"fmt"
	"strings"

	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/mcp"
)

// ListMcpResourcesTool lists available resources from MCP servers
type ListMcpResourcesTool struct{}

func (t *ListMcpResourcesTool) Definition() api.Tool {
	return api.Tool{
		Name:        "list_mcp_resources",
		Description: "List available resources from connected MCP (Model Context Protocol) servers. Returns a list of resources that can be read with read_mcp_resource.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Optional: filter by server name",
				},
			},
		},
	}
}

func (t *ListMcpResourcesTool) Execute(input interface{}) (string, error) {
	manager := mcp.GetManager()
	servers := manager.ListServers()

	if len(servers) == 0 {
		return `No MCP servers configured.

To add an MCP server:
  letsgo mcp add <name> <command> [args...]

Example:
  letsgo mcp add filesystem npx -y @modelcontextprotocol/server-filesystem /path/to/dir`, nil
	}

	var result strings.Builder
	result.WriteString("MCP Resources Available:\n\n")

	for name := range servers {
		if !manager.IsServerRunning(name) {
			// Auto-start if not running
			if err := manager.StartServer(name); err != nil {
				result.WriteString(fmt.Sprintf("Server: %s (failed to start: %v)\n", name, err))
				continue
			}
		}

		tools, err := manager.ListServerTools(name)
		if err != nil {
			result.WriteString(fmt.Sprintf("Server: %s (error: %v)\n", name, err))
			continue
		}

		resources, err := manager.ListServerResources(name)
		if err != nil {
			result.WriteString(fmt.Sprintf("Server: %s (error listing resources: %v)\n", name, err))
			continue
		}

		result.WriteString(fmt.Sprintf("Server: %s\n", name))
		result.WriteString(fmt.Sprintf("  Tools: %d\n", len(tools)))
		for _, tool := range tools {
			result.WriteString(fmt.Sprintf("    - %s: %s\n", tool.Name, tool.Description))
		}
		result.WriteString(fmt.Sprintf("  Resources: %d\n", len(resources)))
		for _, res := range resources {
			result.WriteString(fmt.Sprintf("    - %s (%s)\n", res.Name, res.URI))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// ReadMcpResourceTool reads a resource from an MCP server
type ReadMcpResourceTool struct{}

func (t *ReadMcpResourceTool) Definition() api.Tool {
	return api.Tool{
		Name:        "read_mcp_resource",
		Description: "Read a resource from a connected MCP server. The resource URI should be obtained from list_mcp_resources.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"uri": map[string]interface{}{
					"type":        "string",
					"description": "Resource URI to read",
				},
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Server name (optional if URI is unique)",
				},
			},
			"required": []string{"uri"},
		},
	}
}

func (t *ReadMcpResourceTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	uri, _ := m["uri"].(string)
	serverName, _ := m["server"].(string)

	if uri == "" {
		return "", fmt.Errorf("uri is required")
	}

	manager := mcp.GetManager()

	// If server not specified, try to find it from all servers
	if serverName == "" {
		servers := manager.ListServers()
		for name := range servers {
			if strings.Contains(uri, name) || manager.IsServerRunning(name) {
				serverName = name
				break
			}
		}
	}

	if serverName == "" {
		return "", fmt.Errorf("could not determine MCP server for URI: %s. Please specify server name.", uri)
	}

	// Ensure server is running
	if !manager.IsServerRunning(serverName) {
		if err := manager.StartServer(serverName); err != nil {
			return "", fmt.Errorf("failed to start server '%s': %w", serverName, err)
		}
	}

	// Read the resource
	data, err := manager.ReadResource(serverName, uri)
	if err != nil {
		return "", fmt.Errorf("failed to read resource: %w", err)
	}

	return fmt.Sprintf("Resource: %s\nServer: %s\n\n%s", uri, serverName, string(data)), nil
}

// McpCallTool calls an MCP tool directly
type McpCallTool struct{}

func (t *McpCallTool) Definition() api.Tool {
	return api.Tool{
		Name:        "mcp_call",
		Description: "Call a tool on a connected MCP server directly. This allows using any tool exposed by an MCP server.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "MCP server name",
				},
				"tool": map[string]interface{}{
					"type":        "string",
					"description": "Tool name to call",
				},
				"arguments": map[string]interface{}{
					"type":        "object",
					"description": "Tool arguments",
				},
			},
			"required": []string{"server", "tool"},
		},
	}
}

func (t *McpCallTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input")
	}

	serverName, _ := m["server"].(string)
	toolName, _ := m["tool"].(string)

	if serverName == "" {
		return "", fmt.Errorf("server is required")
	}
	if toolName == "" {
		return "", fmt.Errorf("tool is required")
	}

	arguments := make(map[string]interface{})
	if args, ok := m["arguments"].(map[string]interface{}); ok {
		arguments = args
	}

	manager := mcp.GetManager()

	// Ensure server is running
	if !manager.IsServerRunning(serverName) {
		if err := manager.StartServer(serverName); err != nil {
			return "", fmt.Errorf("failed to start server '%s': %w", serverName, err)
		}
	}

	// Call the tool
	result, err := manager.CallTool(serverName, toolName, arguments)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return fmt.Sprintf("MCP Tool Result:\nServer: %s\nTool: %s\n\n%s", serverName, toolName, result), nil
}
