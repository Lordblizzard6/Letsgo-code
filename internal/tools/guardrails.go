package tools

import (
	"fmt"
	"strings"
)

// Dangerous command patterns that require confirmation
type DangerousPattern struct {
	Pattern     string
	Description string
	Severity    string // "high", "medium", "low"
}

var dangerousPatterns = []DangerousPattern{
	// High severity - data destruction
	{Pattern: "rm -rf /", Description: "Deleting entire filesystem", Severity: "high"},
	{Pattern: "rm -rf /*", Description: "Deleting entire filesystem", Severity: "high"},
	{Pattern: "rm -rf ~", Description: "Deleting home directory", Severity: "high"},
	{Pattern: "rm -rf $HOME", Description: "Deleting home directory", Severity: "high"},
	{Pattern: "dd if=", Description: "Direct disk writing", Severity: "high"},
	{Pattern: "> /dev/sda", Description: "Writing to disk device", Severity: "high"},
	{Pattern: "mkfs.", Description: "Formatting filesystem", Severity: "high"},
	{Pattern: "format", Description: "Formatting drive", Severity: "high"},

	// Medium severity - destructive operations
	{Pattern: "rm -rf", Description: "Recursive deletion", Severity: "medium"},
	{Pattern: "rm -r", Description: "Recursive deletion", Severity: "medium"},
	{Pattern: "del /f /s /q", Description: "Force recursive deletion (Windows)", Severity: "medium"},
	{Pattern: "rmdir /s /q", Description: "Remove directory tree (Windows)", Severity: "medium"},
	{Pattern: "git reset --hard", Description: "Hard git reset", Severity: "medium"},
	{Pattern: "git clean -fd", Description: "Force remove untracked files", Severity: "medium"},
	{Pattern: "git push --force", Description: "Force push to remote", Severity: "medium"},
	{Pattern: "git push -f", Description: "Force push to remote", Severity: "medium"},
	{Pattern: "> ", Description: "File overwrite", Severity: "medium"},
	{Pattern: ">> ", Description: "File append", Severity: "low"},

	// Git operations that should be careful
	{Pattern: "git checkout -b", Description: "Creating new branch", Severity: "low"},
	{Pattern: "git branch -D", Description: "Force delete branch", Severity: "medium"},
	{Pattern: "git branch -d", Description: "Delete branch", Severity: "low"},
	{Pattern: "git merge", Description: "Merging branches", Severity: "low"},
	{Pattern: "git rebase", Description: "Rebasing", Severity: "low"},
	{Pattern: "git cherry-pick", Description: "Cherry-picking commits", Severity: "low"},
	{Pattern: "git revert", Description: "Reverting commits", Severity: "low"},

	// Network/Security operations
	{Pattern: "curl.*|.*sh", Description: "Piping curl to shell", Severity: "high"},
	{Pattern: "wget.*|.*sh", Description: "Piping wget to shell", Severity: "high"},
	{Pattern: "fetch.*|.*sh", Description: "Piping fetch to shell", Severity: "high"},

	// Permission changes
	{Pattern: "chmod -R 777", Description: "Making everything world-writable", Severity: "medium"},
	{Pattern: "chmod 777", Description: "Making file world-writable", Severity: "medium"},
	{Pattern: "chown -R", Description: "Recursive ownership change", Severity: "medium"},

	// Database operations
	{Pattern: "drop database", Description: "Dropping database", Severity: "high"},
	{Pattern: "drop table", Description: "Dropping table", Severity: "medium"},
	{Pattern: "delete from", Description: "Deleting data", Severity: "medium"},
	{Pattern: "truncate", Description: "Truncating table", Severity: "medium"},

	// Process management
	{Pattern: "kill -9", Description: "Force killing process", Severity: "low"},
	{Pattern: "killall", Description: "Killing all processes by name", Severity: "medium"},
	{Pattern: "pkill", Description: "Killing processes by pattern", Severity: "medium"},
}

// CheckCommandSafety analyzes a command and returns warnings if it's potentially dangerous
func CheckCommandSafety(command string) (isDangerous bool, warnings []string, severity string) {
	lowerCmd := strings.ToLower(command)

	for _, pattern := range dangerousPatterns {
		lowerPattern := strings.ToLower(pattern.Pattern)

		// Check for exact match or substring match
		if strings.Contains(lowerCmd, lowerPattern) {
			isDangerous = true
			warnings = append(warnings, pattern.Description)

			// Track highest severity
			if severity == "" || pattern.Severity == "high" ||
				(severity == "low" && pattern.Severity == "medium") {
				severity = pattern.Severity
			}
		}
	}

	// Additional heuristics
	if strings.Contains(lowerCmd, "rm") && strings.Contains(lowerCmd, "*") {
		if !contains(warnings, "Wildcard deletion") {
			isDangerous = true
			warnings = append(warnings, "Wildcard deletion")
			if severity != "high" {
				severity = "medium"
			}
		}
	}

	// Check for sudo/root operations
	if strings.HasPrefix(lowerCmd, "sudo ") || strings.HasPrefix(lowerCmd, "su ") {
		if !contains(warnings, "Requires elevated privileges") {
			isDangerous = true
			warnings = append(warnings, "Requires elevated privileges")
			if severity != "high" {
				severity = "medium"
			}
		}
	}

	return isDangerous, warnings, severity
}

// IsCommandAllowed checks if a command is in the allowed list
func IsCommandAllowed(command string, allowedCommands []string) bool {
	lowerCmd := strings.ToLower(strings.TrimSpace(command))

	for _, allowed := range allowedCommands {
		if strings.HasPrefix(lowerCmd, strings.ToLower(allowed)) {
			return true
		}
	}

	return false
}

// GetCommandRiskLevel returns a risk assessment for a command
func GetCommandRiskLevel(command string) string {
	isDangerous, _, severity := CheckCommandSafety(command)
	if !isDangerous {
		return "safe"
	}
	return severity
}

// FormatSafetyWarning formats a safety warning for display
func FormatSafetyWarning(command string, warnings []string, severity string) string {
	var emoji string
	switch severity {
	case "high":
		emoji = "🚨"
	case "medium":
		emoji = "⚠️"
	default:
		emoji = "⚡"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s DANGEROUS COMMAND DETECTED %s\n\n", emoji, emoji))
	result.WriteString(fmt.Sprintf("Command: %s\n\n", command))
	result.WriteString("Detected risks:\n")

	for i, warning := range warnings {
		result.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warning))
	}

	result.WriteString("\n")

	if severity == "high" {
		result.WriteString("⚠️  This command could cause IRREVERSIBLE DATA LOSS or SYSTEM DAMAGE.\n")
		result.WriteString("⚠️  Please review carefully before executing.\n")
	} else if severity == "medium" {
		result.WriteString("⚠️  This command could modify or delete important data.\n")
	}

	return result.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Permission levels for bash tool
type PermissionLevel int

const (
	PermissionAuto PermitLevel = iota
	PermissionAsk
	PermissionDeny
)

type PermitLevel int

func (p PermitLevel) String() string {
	switch p {
	case PermissionAuto:
		return "auto"
	case PermissionAsk:
		return "ask"
	case PermissionDeny:
		return "deny"
	default:
		return "unknown"
	}
}

// GetPermissionForCommand determines what permission level a command should use
func GetPermissionForCommand(command string, userPreference PermitLevel) PermitLevel {
	isDangerous, _, severity := CheckCommandSafety(command)

	if !isDangerous {
		return PermissionAuto
	}

	// High severity commands always require explicit permission
	if severity == "high" {
		return PermissionAsk
	}

	// Medium severity respects user preference
	if severity == "medium" {
		if userPreference == PermissionDeny {
			return PermissionDeny
		}
		return PermissionAsk
	}

	// Low severity can be auto if user allows
	if userPreference == PermissionAuto {
		return PermissionAuto
	}

	return PermissionAsk
}
