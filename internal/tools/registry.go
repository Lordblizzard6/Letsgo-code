package tools

import (
	"fmt"

	"github.com/user/go-claude-code/internal/api"
)

var AllTools = []Tool{
	// File operations
	&LsTool{},
	&CatTool{},
	&GlobTool{},
	&GrepTool{},
	&EditTool{},
	&ApplyPatchTool{},
	&WriteFileTool{},
	&NotebookEditTool{},

	// Shell execution
	&BashTool{},
	&PowerShellTool{},

	// Web operations
	&WebSearchTool{},
	&WebFetchTool{},
	&WebBrowserTool{},

	// Task management
	&TodoWriteTool{},
	&TaskCreateTool{},
	&TaskGetTool{},
	&TaskUpdateTool{},
	&TaskListTool{},
	&TaskStopTool{},
	&TaskOutputTool{},

	// Agent system
	&AgentTool{},
	&AgentGetTool{},
	&AgentListTool{},

	// Context and summaries
	&BriefTool{},

	// Plan mode
	&EnterPlanModeTool{},
	&ExitPlanModeTool{},
	&PlanStepAddTool{},
	&PlanShowTool{},
	&PlanApproveTool{},
	&PlanStepCompleteTool{},

	// Skills
	&SkillTool{},

	// MCP integration
	&ListMcpResourcesTool{},
	&ReadMcpResourceTool{},
	&McpCallTool{},

	// LSP
	&LSPTool{},

	// User interaction
	&AskUserTool{},

	// Utility
	&SleepTool{},

	// Tool search
	&ToolSearchTool{},

	// Workflow
	&WorkflowTool{},

	// Worktree
	&EnterWorktreeTool{},
	&ExitWorktreeTool{},
	&ListWorktreesTool{},
}

func GetToolDefinitions() []api.Tool {
	defs := make([]api.Tool, len(AllTools))
	for i, t := range AllTools {
		defs[i] = t.Definition()
	}
	return defs
}

// GetPlanToolDefinitions returns only read-only/inspection tools allowed during Plan mode.
func GetPlanToolDefinitions() []api.Tool {
	var defs []api.Tool
	for _, t := range AllTools {
		name := t.Definition().Name
		switch name {
		case "ls", "cat", "glob", "grep", "lsp", "web_search", "web_fetch",
			"brief", "ask_user", "sleep", "tool_search",
			"list_mcp_resources", "read_mcp_resource",
			"plan_show", "plan_step_add", "plan_step_complete", "plan_approve":
			defs = append(defs, t.Definition())
		}
	}
	return defs
}

func ExecuteTool(name string, input interface{}) (string, error) {
	for _, t := range AllTools {
		if t.Definition().Name == name {
			return t.Execute(input)
		}
	}
	return "", nil // Tool not found
}

// ExecuteToolWithPermission checks permissions before executing a tool
func ExecuteToolWithPermission(name string, input interface{}) (string, error) {
	// Check permissions first
	allowed, reason := DefaultPermissionManager.CheckPermission(name, input)
	if !allowed {
		return "", fmt.Errorf("permission denied: %s", reason)
	}

	// If user confirmation is needed, we would handle that at a higher level
	// For now, just execute
	return ExecuteTool(name, input)
}
