package tools

import (
	"fmt"
	"strings"
)

// PermissionManager handles tool permission checking
type PermissionManager struct {
	// User preferences
	BashPermission       PermitLevel
	FileEditPermission   PermitLevel
	WebSearchPermission  PermitLevel
	WebFetchPermission   PermitLevel

	// Deny lists
	DeniedCommands    []string
	DeniedPaths       []string
	AllowedPaths      []string
}

// Global permission manager instance
var DefaultPermissionManager = &PermissionManager{
	BashPermission:      PermissionAsk,
	FileEditPermission:  PermissionAsk,
	WebSearchPermission: PermissionAuto,
	WebFetchPermission:  PermissionAuto,
	DeniedCommands: []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero of=/dev/sda",
		"> /dev/sda",
		"mkfs.ext4 /dev/sda",
		":(){ :|:& };:", // Fork bomb
	},
	DeniedPaths: []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/ssh",
		"~/.ssh",
	},
}

// CheckPermission checks if a tool execution is allowed
func (pm *PermissionManager) CheckPermission(toolName string, input interface{}) (allowed bool, reason string) {
	switch toolName {
	case "bash":
		return pm.checkBashPermission(input)
	case "edit", "write_file":
		return pm.checkFileEditPermission(input)
	case "web_search":
		return pm.checkWebPermission(pm.WebSearchPermission, "search")
	case "web_fetch":
		return pm.checkWebPermission(pm.WebFetchPermission, "fetch")
	default:
		return true, ""
	}
}

func (pm *PermissionManager) checkBashPermission(input interface{}) (bool, string) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return false, "invalid input"
	}

	command, _ := m["command"].(string)

	// Check deny list
	for _, denied := range pm.DeniedCommands {
		if strings.Contains(command, denied) {
			return false, fmt.Sprintf("Command matches denied pattern: %s", denied)
		}
	}

	// Check if dangerous
	isDangerous, warnings, severity := CheckCommandSafety(command)
	if isDangerous && severity == "high" {
		return false, FormatSafetyWarning(command, warnings, severity)
	}

	// Check permission level
	requiredPerm := GetPermissionForCommand(command, pm.BashPermission)
	if requiredPerm == PermissionDeny {
		return false, "Command requires elevated permissions but is denied by policy"
	}

	if requiredPerm == PermissionAsk {
		return true, "ASK_USER" // Signal that user confirmation is needed
	}

	return true, ""
}

func (pm *PermissionManager) checkFileEditPermission(input interface{}) (bool, string) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return false, "invalid input"
	}

	path, _ := m["path"].(string)

	// Check denied paths
	for _, denied := range pm.DeniedPaths {
		if strings.Contains(path, denied) {
			return false, fmt.Sprintf("Path matches denied pattern: %s", denied)
		}
	}

	// Check permission level
	if pm.FileEditPermission == PermissionDeny {
		return false, "File edits are denied by policy"
	}

	if pm.FileEditPermission == PermissionAsk {
		return true, "ASK_USER"
	}

	return true, ""
}

func (pm *PermissionManager) checkWebPermission(level PermitLevel, action string) (bool, string) {
	switch level {
	case PermissionDeny:
		return false, fmt.Sprintf("Web %s is denied by policy", action)
	case PermissionAsk:
		return true, "ASK_USER"
	default:
		return true, ""
	}
}

// IsPathAllowed checks if a path is allowed for editing
func IsPathAllowed(path string) bool {
	// Check against git ignore
	if ShouldIgnore(path) {
		return false
	}

	// Check denied extensions
	deniedExts := []string{".exe", ".dll", ".so", ".dylib", ".bin"}
	lowerPath := strings.ToLower(path)
	for _, ext := range deniedExts {
		if strings.HasSuffix(lowerPath, ext) {
			return false
		}
	}

	return true
}

// GetPermissionPrompt returns a prompt for user confirmation
func GetPermissionPrompt(toolName string, input interface{}) string {
	switch toolName {
	case "bash":
		m, _ := input.(map[string]interface{})
		command, _ := m["command"].(string)
		return fmt.Sprintf("Allow bash command:\n  %s", command)
	case "edit":
		m, _ := input.(map[string]interface{})
		path, _ := m["path"].(string)
		return fmt.Sprintf("Allow editing file:\n  %s", path)
	case "write_file":
		m, _ := input.(map[string]interface{})
		path, _ := m["path"].(string)
		return fmt.Sprintf("Allow writing to file:\n  %s", path)
	case "web_search":
		return "Allow web search?"
	case "web_fetch":
		m, _ := input.(map[string]interface{})
		url, _ := m["url"].(string)
		return fmt.Sprintf("Allow fetching URL:\n  %s", url)
	default:
		return fmt.Sprintf("Allow %s?", toolName)
	}
}

// ParsePermission parses a permission string
func ParsePermission(s string) PermitLevel {
	switch strings.ToLower(s) {
	case "auto", "yes", "allow":
		return PermissionAuto
	case "ask", "prompt":
		return PermissionAsk
	case "deny", "no", "never":
		return PermissionDeny
	default:
		return PermissionAsk
	}
}
