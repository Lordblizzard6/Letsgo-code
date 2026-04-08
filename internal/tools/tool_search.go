package tools

import (
	"fmt"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// ToolSearch searches available tools
type ToolSearch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// ToolSearchTool allows searching for available tools
type ToolSearchTool struct{}

func (t *ToolSearchTool) Name() string {
	return "tool_search"
}

func (t *ToolSearchTool) Description() string {
	return "Search for available tools by name, description, or functionality. Use this when you need to find the right tool for a task but aren't sure which one to use."
}

func (t *ToolSearchTool) Definition() api.Tool {
	return api.Tool{
		Name:        "tool_search",
		Description: "Search for available tools by name, description, or functionality. Use this when you need to find the right tool for a task but aren't sure which one to use.",
		InputSchema: t.InputSchema(),
	}
}

func (t *ToolSearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query to find tools - can be keywords, partial names, or functionality descriptions",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Optional category filter: file, shell, web, task, agent, context, plan, mcp, lsp",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return",
				"default":     10,
			},
		},
		"required": []string{"query"},
	}
}

func (t *ToolSearchTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	return t.executeInternal(m)
}

func (t *ToolSearchTool) executeInternal(input map[string]interface{}) (string, error) {
	query, ok := input["query"].(string)
	if !ok {
		return "", fmt.Errorf("query is required")
	}
	
	category := ""
	if c, ok := input["category"].(string); ok {
		category = c
	}
	
	limit := 10
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}
	
	// Build tool catalog
	tools := buildToolCatalog()
	
	// Filter by query
	var results []ToolSearch
	queryLower := strings.ToLower(query)
	
	for _, tool := range tools {
		// Category filter
		if category != "" && tool.Category != category {
			continue
		}
		
		// Search in name and description
		if strings.Contains(strings.ToLower(tool.Name), queryLower) ||
		   strings.Contains(strings.ToLower(tool.Description), queryLower) {
			results = append(results, tool)
		}
	}
	
	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}
	
	if len(results) == 0 {
		return fmt.Sprintf("No tools found matching '%s'", query), nil
	}
	
	// Format results
	output := fmt.Sprintf("Found %d tool(s) matching '%s':\n\n", len(results), query)
	for _, tool := range results {
		output += fmt.Sprintf("  • %s (%s)\n    %s\n\n", tool.Name, tool.Category, tool.Description)
	}
	
	return output, nil
}

func buildToolCatalog() []ToolSearch {
	return []ToolSearch{
		{Name: "ls", Description: "List files in a directory", Category: "file"},
		{Name: "cat", Description: "Read and display file contents", Category: "file"},
		{Name: "read_file", Description: "Read file with optional offset and limit", Category: "file"},
		{Name: "write_file", Description: "Create or overwrite a file", Category: "file"},
		{Name: "edit", Description: "Edit a file by replacing text", Category: "file"},
		{Name: "glob", Description: "Find files matching a pattern", Category: "file"},
		{Name: "grep", Description: "Search for text in files", Category: "file"},
		{Name: "notebook_edit", Description: "Edit Jupyter notebook cells", Category: "file"},
		{Name: "bash", Description: "Execute bash shell commands", Category: "shell"},
		{Name: "web_search", Description: "Search the web for information", Category: "web"},
		{Name: "web_fetch", Description: "Fetch and extract content from URLs", Category: "web"},
		{Name: "todo_write", Description: "Create and manage todo lists", Category: "task"},
		{Name: "task_create", Description: "Create a background task", Category: "task"},
		{Name: "task_get", Description: "Get task status and output", Category: "task"},
		{Name: "task_list", Description: "List all tasks", Category: "task"},
		{Name: "task_stop", Description: "Stop a running task", Category: "task"},
		{Name: "task_output", Description: "Retrieve output from background task", Category: "task"},
		{Name: "agent", Description: "Create and run an autonomous agent", Category: "agent"},
		{Name: "agent_get", Description: "Get agent status and output", Category: "agent"},
		{Name: "agent_list", Description: "List all agents", Category: "agent"},
		{Name: "brief", Description: "Create a summary of context", Category: "context"},
		{Name: "enter_plan_mode", Description: "Enter plan mode for complex tasks", Category: "plan"},
		{Name: "exit_plan_mode", Description: "Exit plan mode", Category: "plan"},
		{Name: "plan_step_add", Description: "Add a step to the current plan", Category: "plan"},
		{Name: "plan_show", Description: "Show the current plan", Category: "plan"},
		{Name: "plan_approve", Description: "Approve a plan before execution", Category: "plan"},
		{Name: "plan_step_complete", Description: "Mark a plan step as complete", Category: "plan"},
		{Name: "skill", Description: "Use a predefined skill or workflow", Category: "plan"},
		{Name: "list_mcp_resources", Description: "List available MCP resources", Category: "mcp"},
		{Name: "read_mcp_resource", Description: "Read an MCP resource", Category: "mcp"},
		{Name: "mcp_call", Description: "Call an MCP tool", Category: "mcp"},
		{Name: "lsp", Description: "Use LSP for code intelligence", Category: "lsp"},
		{Name: "ask_user", Description: "Ask the user a question", Category: "user"},
		{Name: "tool_search", Description: "Search for available tools", Category: "user"},
	}
}
