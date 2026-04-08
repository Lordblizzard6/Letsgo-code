package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// PermissionMode represents the permission mode for tools
type PermissionMode string

const (
	// PermissionModeAsk - Always ask user for confirmation
	PermissionModeAsk PermissionMode = "ask"
	// PermissionModeAuto - Automatically execute safe operations
	PermissionModeAuto PermissionMode = "auto"
	// PermissionModeAutoBypass - Auto-execute with bypass (enterprise)
	PermissionModeAutoBypass PermissionMode = "auto-bp"
	// PermissionModeOptOut - Disable the tool
	PermissionModeOptOut PermissionMode = "opt-out"
)

// PermissionScope represents the scope of a permission
type PermissionScope string

const (
	ScopeGlobal  PermissionScope = "global"
	ScopeProject PermissionScope = "project"
	ScopeSession PermissionScope = "session"
)

// ToolPermission configuration for a specific tool
type ToolPermission struct {
	Mode            PermissionMode  `json:"mode"`
	Scope           PermissionScope `json:"scope"`
	AllowedPaths    []string        `json:"allowed_paths,omitempty"`
	DeniedPaths     []string        `json:"denied_paths,omitempty"`
	AllowedCommands []string        `json:"allowed_commands,omitempty"`
	DeniedCommands  []string        `json:"denied_commands,omitempty"`
}

// PermissionConfig holds all permission settings
type PermissionConfig struct {
	Version     string                     `json:"version"`
	GlobalMode  PermissionMode             `json:"global_mode"`
	Tools       map[string]*ToolPermission `json:"tools"`
	LastUpdated string                     `json:"last_updated"`
}

// AdvancedPermissionManager handles all permission logic
type AdvancedPermissionManager struct {
	config     *PermissionConfig
	configPath string
	mu         sync.RWMutex
	// Session overrides
	sessionOverrides map[string]PermissionMode
}

var (
	advancedPMInstance *AdvancedPermissionManager
	advancedPMOnce     sync.Once
)

// GetAdvancedPermissionManager returns the singleton instance
func GetAdvancedPermissionManager() *AdvancedPermissionManager {
	advancedPMOnce.Do(func() {
		advancedPMInstance = &AdvancedPermissionManager{
			config: &PermissionConfig{
				Version:    "1.0",
				GlobalMode: PermissionModeAsk,
				Tools:      make(map[string]*ToolPermission),
			},
			sessionOverrides: make(map[string]PermissionMode),
		}
		advancedPMInstance.loadConfig()
	})
	return advancedPMInstance
}

// getConfigPath returns the path to the permissions config file
func (pm *AdvancedPermissionManager) getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", "permissions.json")
}

// loadConfig loads permissions from disk
func (pm *AdvancedPermissionManager) loadConfig() {
	configPath := pm.getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Use defaults if file doesn't exist
		pm.setDefaults()
		return
	}

	var config PermissionConfig
	if err := json.Unmarshal(data, &config); err != nil {
		pm.setDefaults()
		return
	}

	pm.config = &config

	// Ensure all tools have entries
	pm.ensureToolPermissions()
}

// saveConfig saves permissions to disk
func (pm *AdvancedPermissionManager) saveConfig() error {
	configPath := pm.getConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// setDefaults sets default permissions
func (pm *AdvancedPermissionManager) setDefaults() {
	defaults := map[string]PermissionMode{
		"bash":        PermissionModeAsk,
		"edit":        PermissionModeAsk,
		"write_file":  PermissionModeAsk,
		"web_search":  PermissionModeAuto,
		"web_fetch":   PermissionModeAuto,
		"agent":       PermissionModeAsk,
		"task_create": PermissionModeAuto,
		"todo_write":  PermissionModeAuto,
	}

	for tool, mode := range defaults {
		pm.config.Tools[tool] = &ToolPermission{
			Mode:  mode,
			Scope: ScopeGlobal,
		}
	}

	// Set dangerous defaults
	pm.config.Tools["bash"].DeniedCommands = []string{
		"rm -rf /", "rm -rf /*", "dd if=/dev/zero of=/dev/sda",
		"> /dev/sda", "mkfs.", ":(){ :|:& };:",
	}
	pm.config.Tools["edit"].DeniedPaths = []string{
		"/etc/passwd", "/etc/shadow", "/etc/ssh", "~/.ssh/id_*",
	}
}

// ensureToolPermissions ensures all known tools have permission entries
func (pm *AdvancedPermissionManager) ensureToolPermissions() {
	knownTools := []string{
		"bash", "edit", "write_file", "read_file", "ls", "cat", "glob", "grep",
		"web_search", "web_fetch", "agent", "agent_get", "agent_list",
		"task_create", "task_get", "task_update", "task_list", "task_stop",
		"todo_write", "brief", "notebook_edit",
		"enter_plan_mode", "exit_plan_mode", "plan_step_add", "plan_show",
		"skill", "list_mcp_resources", "read_mcp_resource", "mcp_call",
		"lsp", "ask_user",
	}

	for _, tool := range knownTools {
		if _, exists := pm.config.Tools[tool]; !exists {
			pm.config.Tools[tool] = &ToolPermission{
				Mode:  pm.config.GlobalMode,
				Scope: ScopeGlobal,
			}
		}
	}
}

// CheckPermission checks if a tool can be executed
func (pm *AdvancedPermissionManager) CheckPermission(toolName string, input interface{}) PermissionResult {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check session override first
	if mode, exists := pm.sessionOverrides[toolName]; exists {
		if mode == PermissionModeOptOut {
			return PermissionResult{
				Allowed: false,
				Reason:  fmt.Sprintf("Tool '%s' is disabled for this session", toolName),
				Mode:    mode,
			}
		}
		if mode == PermissionModeAuto || mode == PermissionModeAutoBypass {
			return PermissionResult{Allowed: true, Mode: mode}
		}
	}

	// Get tool permission config
	tp, exists := pm.config.Tools[toolName]
	if !exists {
		// Unknown tool - use global default
		tp = &ToolPermission{Mode: pm.config.GlobalMode, Scope: ScopeGlobal}
	}

	// Check if tool is opted out
	if tp.Mode == PermissionModeOptOut {
		return PermissionResult{
			Allowed: false,
			Reason:  fmt.Sprintf("Tool '%s' is disabled", toolName),
			Mode:    tp.Mode,
		}
	}

	// Special handling for specific tools
	switch toolName {
	case "bash":
		return pm.checkBashPermission(tp, input)
	case "edit", "write_file":
		return pm.checkFilePermission(tp, input)
	case "web_search", "web_fetch":
		return pm.checkWebPermission(tp, toolName, input)
	default:
		// For other tools, just check the mode
		if tp.Mode == PermissionModeAsk {
			return PermissionResult{
				Allowed:               true,
				Mode:                  tp.Mode,
				NeedsUserConfirmation: true,
				Prompt:                pm.buildPrompt(toolName, input),
			}
		}
		return PermissionResult{Allowed: true, Mode: tp.Mode}
	}
}

// PermissionResult contains the result of a permission check
type PermissionResult struct {
	Allowed               bool
	Reason                string
	Mode                  PermissionMode
	NeedsUserConfirmation bool
	Prompt                string
	RiskLevel             string // "safe", "low", "medium", "high"
}

func (pm *AdvancedPermissionManager) checkBashPermission(tp *ToolPermission, input interface{}) PermissionResult {
	m, ok := input.(map[string]interface{})
	if !ok {
		return PermissionResult{Allowed: false, Reason: "Invalid input", Mode: tp.Mode}
	}

	command, _ := m["command"].(string)

	// Check denied commands
	for _, denied := range tp.DeniedCommands {
		if strings.Contains(command, denied) {
			return PermissionResult{
				Allowed:   false,
				Reason:    fmt.Sprintf("Command matches denied pattern: %s", denied),
				Mode:      tp.Mode,
				RiskLevel: "high",
			}
		}
	}

	// Check global denied commands
	for _, denied := range getGlobalDeniedCommands() {
		if strings.Contains(command, denied) {
			return PermissionResult{
				Allowed:   false,
				Reason:    fmt.Sprintf("Command is globally denied: %s", denied),
				Mode:      tp.Mode,
				RiskLevel: "high",
			}
		}
	}

	// Check command safety
	isDangerous, warnings, severity := CheckCommandSafety(command)
	if isDangerous && severity == "high" {
		return PermissionResult{
			Allowed:   false,
			Reason:    FormatSafetyWarning(command, warnings, severity),
			Mode:      tp.Mode,
			RiskLevel: severity,
		}
	}

	// Check if command is in allowed list (if specified)
	if len(tp.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range tp.AllowedCommands {
			if strings.HasPrefix(command, allowedCmd) {
				allowed = true
				break
			}
		}
		if !allowed {
			return PermissionResult{
				Allowed: false,
				Reason:  "Command not in allowed list",
				Mode:    tp.Mode,
			}
		}
	}

	// Determine if we need to ask
	if tp.Mode == PermissionModeAsk || (isDangerous && tp.Mode == PermissionModeAuto) {
		return PermissionResult{
			Allowed:               true,
			Mode:                  tp.Mode,
			NeedsUserConfirmation: true,
			Prompt:                fmt.Sprintf("Execute bash command:\n  %s", command),
			RiskLevel:             severity,
		}
	}

	return PermissionResult{Allowed: true, Mode: tp.Mode, RiskLevel: severity}
}

func (pm *AdvancedPermissionManager) checkFilePermission(tp *ToolPermission, input interface{}) PermissionResult {
	m, ok := input.(map[string]interface{})
	if !ok {
		return PermissionResult{Allowed: false, Reason: "Invalid input", Mode: tp.Mode}
	}

	path, _ := m["path"].(string)

	// Check denied paths
	for _, denied := range tp.DeniedPaths {
		if strings.Contains(path, denied) {
			return PermissionResult{
				Allowed:   false,
				Reason:    fmt.Sprintf("Path matches denied pattern: %s", denied),
				Mode:      tp.Mode,
				RiskLevel: "high",
			}
		}
	}

	// Check global denied paths
	for _, denied := range getGlobalDeniedPaths() {
		if strings.Contains(path, denied) {
			return PermissionResult{
				Allowed:   false,
				Reason:    fmt.Sprintf("Path is globally denied: %s", denied),
				Mode:      tp.Mode,
				RiskLevel: "high",
			}
		}
	}

	// Check if path is in allowed list (if specified)
	if len(tp.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range tp.AllowedPaths {
			if strings.HasPrefix(path, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return PermissionResult{
				Allowed: false,
				Reason:  "Path not in allowed list",
				Mode:    tp.Mode,
			}
		}
	}

	if tp.Mode == PermissionModeAsk {
		return PermissionResult{
			Allowed:               true,
			Mode:                  tp.Mode,
			NeedsUserConfirmation: true,
			Prompt:                fmt.Sprintf("Edit file:\n  %s", path),
			RiskLevel:             "medium",
		}
	}

	return PermissionResult{Allowed: true, Mode: tp.Mode, RiskLevel: "low"}
}

func (pm *AdvancedPermissionManager) checkWebPermission(tp *ToolPermission, toolName string, input interface{}) PermissionResult {
	if tp.Mode == PermissionModeAsk {
		m, _ := input.(map[string]interface{})
		url, _ := m["url"].(string)
		if url == "" {
			url, _ = m["query"].(string)
		}

		return PermissionResult{
			Allowed:               true,
			Mode:                  tp.Mode,
			NeedsUserConfirmation: true,
			Prompt:                fmt.Sprintf("Allow %s:\n  %s", toolName, url),
			RiskLevel:             "low",
		}
	}

	return PermissionResult{Allowed: true, Mode: tp.Mode, RiskLevel: "low"}
}

func (pm *AdvancedPermissionManager) buildPrompt(toolName string, input interface{}) string {
	switch toolName {
	case "agent":
		return "Allow spawning an autonomous agent?"
	case "task_create":
		return "Allow creating a subtask?"
	default:
		return fmt.Sprintf("Allow executing %s?", toolName)
	}
}

// SetToolMode sets the permission mode for a tool
func (pm *AdvancedPermissionManager) SetToolMode(toolName string, mode PermissionMode, scope PermissionScope) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.config.Tools[toolName] = &ToolPermission{
		Mode:  mode,
		Scope: scope,
	}

	return pm.saveConfig()
}

// SetGlobalMode sets the global default permission mode
func (pm *AdvancedPermissionManager) SetGlobalMode(mode PermissionMode) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.config.GlobalMode = mode
	return pm.saveConfig()
}

// SetSessionOverride sets a temporary permission override for the session
func (pm *AdvancedPermissionManager) SetSessionOverride(toolName string, mode PermissionMode) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.sessionOverrides[toolName] = mode
}

// ClearSessionOverride clears a session override
func (pm *AdvancedPermissionManager) ClearSessionOverride(toolName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.sessionOverrides, toolName)
}

// GetToolConfig returns the permission config for a tool
func (pm *AdvancedPermissionManager) GetToolConfig(toolName string) *ToolPermission {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if tp, exists := pm.config.Tools[toolName]; exists {
		return tp
	}
	return &ToolPermission{Mode: pm.config.GlobalMode, Scope: ScopeGlobal}
}

// GetAllToolConfigs returns all tool permission configs
func (pm *AdvancedPermissionManager) GetAllToolConfigs() map[string]*ToolPermission {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy
	result := make(map[string]*ToolPermission)
	for k, v := range pm.config.Tools {
		result[k] = v
	}
	return result
}

// AddDeniedCommand adds a denied command pattern for a tool
func (pm *AdvancedPermissionManager) AddDeniedCommand(toolName string, pattern string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	tp, exists := pm.config.Tools[toolName]
	if !exists {
		return fmt.Errorf("tool %s not found", toolName)
	}

	for _, existing := range tp.DeniedCommands {
		if existing == pattern {
			return nil // Already exists
		}
	}

	tp.DeniedCommands = append(tp.DeniedCommands, pattern)
	return pm.saveConfig()
}

// AddDeniedPath adds a denied path pattern for a tool
func (pm *AdvancedPermissionManager) AddDeniedPath(toolName string, pattern string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	tp, exists := pm.config.Tools[toolName]
	if !exists {
		return fmt.Errorf("tool %s not found", toolName)
	}

	for _, existing := range tp.DeniedPaths {
		if existing == pattern {
			return nil // Already exists
		}
	}

	tp.DeniedPaths = append(tp.DeniedPaths, pattern)
	return pm.saveConfig()
}

// Global defaults
func getGlobalDeniedCommands() []string {
	return []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero of=/dev/sda",
		"> /dev/sda",
		"mkfs.ext4 /dev/sda",
		":(){ :|:& };:", // Fork bomb
	}
}

func getGlobalDeniedPaths() []string {
	return []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/ssh",
		"~/.ssh/id_rsa",
		"~/.ssh/id_ed25519",
	}
}
