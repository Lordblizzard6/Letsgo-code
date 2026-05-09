package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/go-claude-code/internal/db"
)

// GetContextFilesSummary returns a summary of files in the current context
func GetContextFilesSummary() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "❌ Error getting current directory: " + err.Error()
	}

	var files []string
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden directories and common non-code directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		// Only include code files
		ext := filepath.Ext(path)
		if isCodeFile(ext) {
			rel, _ := filepath.Rel(cwd, path)
			files = append(files, rel)
		}
		return nil
	})

	if err != nil {
		return "❌ Error walking directory: " + err.Error()
	}

	if len(files) == 0 {
		return "📁 No code files found in current directory."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📁 Files in context (%d code files):\n\n", len(files)))

	// Group by extension
	extCount := make(map[string]int)
	for _, f := range files {
		ext := filepath.Ext(f)
		extCount[ext]++
	}

	sb.WriteString("By type:\n")
	for ext, count := range extCount {
		sb.WriteString(fmt.Sprintf("  %s: %d files\n", ext, count))
	}

	sb.WriteString("\nRecent files:\n")
	// Show first 15 files
	limit := 15
	if len(files) < limit {
		limit = len(files)
	}
	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("  • %s\n", files[i]))
	}
	if len(files) > limit {
		sb.WriteString(fmt.Sprintf("  ... and %d more files\n", len(files)-limit))
	}

	return sb.String()
}

// GetSessionDiff shows changes made in the current session
func GetSessionDiff() string {
	// Check if git repo
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "⚠️ Not a git repository. No diff available."
	}

	// Get git status
	status, err := executeGitCommand("status", "--porcelain")
	if err != nil {
		return "❌ Error getting git status: " + err.Error()
	}

	if status == "" {
		return "✅ No changes detected in this session."
	}

	// Get diff stats
	stats, err := executeGitCommand("diff", "--stat")
	if err != nil {
		stats = ""
	}

	var sb strings.Builder
	sb.WriteString("📊 Changes in this session:\n\n")
	sb.WriteString("```\n")
	sb.WriteString(status)
	sb.WriteString("\n```\n")

	if stats != "" {
		sb.WriteString("\n📈 Diff stats:\n")
		sb.WriteString("```\n")
		sb.WriteString(stats)
		sb.WriteString("\n```\n")
	}

	return sb.String()
}

// SearchInConversation searches for text in conversation history
func SearchInConversation(messages []interface{}, query string) string {
	return SearchInConversationWithPagination(query, 1, 5)
}

func SearchInConversationWithPagination(query string, page, pageSize int) string {
	if strings.TrimSpace(query) == "" {
		return "❌ Search query cannot be empty."
	}
	activeSession, err := db.GetActiveSession()
	if err != nil {
		return fmt.Sprintf("❌ Could not load active session: %v", err)
	}
	if activeSession == nil {
		return "⚠️ No active session found."
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 5
	}

	offset := (page - 1) * pageSize
	results, total, err := db.SearchMessages(activeSession.ID, query, pageSize, offset)
	if err != nil {
		return fmt.Sprintf("❌ Search failed: %v", err)
	}
	if total == 0 {
		return fmt.Sprintf("🔍 No results for '%s'.", query)
	}

	totalPages := (total + pageSize - 1) / pageSize
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Results for '%s' (page %d/%d, total %d)\n\n", query, page, totalPages, total))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("• #%d [%s] %s\n", r.ID, r.Role, r.Timestamp.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("  %s\n\n", r.Fragment))
	}
	return sb.String()
}

func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".rs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
		".scala": true, ".r": true, ".m": true, ".mm": true, ".cs": true,
		".fs": true, ".hs": true, ".lhs": true, ".clj": true, ".cljs": true,
		".erl": true, ".ex": true, ".exs": true, ".elm": true, ".lua": true,
		".vim": true, ".sh": true, ".bash": true, ".zsh": true, ".fish": true,
		".ps1": true, ".bat": true, ".cmd": true, ".sql": true, ".yaml": true,
		".yml": true, ".json": true, ".xml": true, ".toml": true, ".ini": true,
		".cfg": true, ".conf": true, ".md": true, ".rst": true, ".txt": true,
	}
	return codeExts[ext]
}

func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git error: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return string(output), nil
}
