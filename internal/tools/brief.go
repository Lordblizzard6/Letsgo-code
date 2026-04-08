package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

// BriefTool generates a summary/brief of files or context
type BriefTool struct{}

func (t *BriefTool) Definition() api.Tool {
	return api.Tool{
		Name:        "brief",
		Description: "Generate a brief summary of files, directories, or context. Useful for creating overviews of codebases, summarizing multiple files, or providing context before making changes. Can read multiple files and generate a condensed summary.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"paths": map[string]interface{}{
					"type":        "array",
					"description": "Paths to files or directories to summarize",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Summary format: 'compact', 'detailed', or 'outline'",
					"enum":        []string{"compact", "detailed", "outline"},
				},
				"max_lines": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum lines per file summary (default 50)",
				},
			},
			"required": []string{"paths"},
		},
	}
}

func (t *BriefTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	pathsRaw, ok := m["paths"].([]interface{})
	if !ok || len(pathsRaw) == 0 {
		return "", fmt.Errorf("paths array is required")
	}

	format, _ := m["format"].(string)
	if format == "" {
		format = "compact"
	}

	maxLines := 50
	if ml, ok := m["max_lines"].(float64); ok {
		maxLines = int(ml)
	}

	var paths []string
	for _, p := range pathsRaw {
		if str, ok := p.(string); ok {
			paths = append(paths, str)
		}
	}

	var summaries []string
	for _, path := range paths {
		summary, err := t.summarizePath(path, format, maxLines)
		if err != nil {
			summaries = append(summaries, fmt.Sprintf("Error reading %s: %v", path, err))
			continue
		}
		summaries = append(summaries, summary)
	}

	// Generate overall brief
	var result strings.Builder
	result.WriteString(fmt.Sprintf("BRIEF (%s format):\n\n", format))

	for _, summary := range summaries {
		result.WriteString(summary)
		result.WriteString("\n")
	}

	// Add statistics
	result.WriteString(fmt.Sprintf("\n---\nSummarized %d paths\n", len(paths)))

	return result.String(), nil
}

func (t *BriefTool) summarizePath(path string, format string, maxLines int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return t.summarizeDirectory(path, format, maxLines)
	}

	return t.summarizeFile(path, format, maxLines)
}

func (t *BriefTool) summarizeDirectory(path string, format string, maxLines int) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var files, dirs int
	var fileList []string
	var extensions = make(map[string]int)

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if entry.IsDir() {
			dirs++
		} else {
			files++
			ext := filepath.Ext(name)
			if ext != "" {
				extensions[ext]++
			}
			if len(fileList) < 20 {
				fileList = append(fileList, name)
			}
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📁 %s/\n", path))

	switch format {
	case "compact":
		result.WriteString(fmt.Sprintf("   %d files, %d directories\n", files, dirs))
		if len(extensions) > 0 {
			var extList []string
			for ext, count := range extensions {
				extList = append(extList, fmt.Sprintf("%s(%d)", ext, count))
			}
			result.WriteString(fmt.Sprintf("   Types: %s\n", strings.Join(extList, ", ")))
		}

	case "detailed", "outline":
		result.WriteString(fmt.Sprintf("   Files: %d | Directories: %d\n", files, dirs))
		if len(extensions) > 0 {
			result.WriteString("   Extensions:\n")
			for ext, count := range extensions {
				result.WriteString(fmt.Sprintf("     - %s: %d\n", ext, count))
			}
		}
		if len(fileList) > 0 && format == "detailed" {
			result.WriteString("   Contents:\n")
			for _, name := range fileList {
				if len(fileList) >= 20 && name == fileList[len(fileList)-1] {
					result.WriteString("     ... (more files)\n")
					break
				}
				result.WriteString(fmt.Sprintf("     - %s\n", name))
			}
		}
	}

	return result.String(), nil
}

func (t *BriefTool) summarizeFile(path string, format string, maxLines int) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// Generate content preview based on format
	var preview strings.Builder

	switch format {
	case "compact":
		// Just show line count and first few lines
		preview.WriteString(fmt.Sprintf("   %d lines | ", totalLines))

		// Detect file type
		fileType := detectFileType(path, content)
		preview.WriteString(fmt.Sprintf("Type: %s\n", fileType))

		// Show first 3 non-empty lines
		shown := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "/*") {
				if len(line) > 60 {
					line = line[:57] + "..."
				}
				preview.WriteString(fmt.Sprintf("   > %s\n", line))
				shown++
				if shown >= 3 {
					break
				}
			}
		}

	case "detailed":
		preview.WriteString(fmt.Sprintf("   Lines: %d | Size: %d bytes\n", totalLines, len(content)))

		// Show up to maxLines
		showLines := maxLines
		if totalLines < showLines {
			showLines = totalLines
		}

		for i := 0; i < showLines && i < len(lines); i++ {
			line := lines[i]
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			preview.WriteString(fmt.Sprintf("   %4d| %s\n", i+1, line))
		}

		if totalLines > showLines {
			preview.WriteString(fmt.Sprintf("   ... %d more lines ...\n", totalLines-showLines))
		}

	case "outline":
		// Extract structure/outline
		preview.WriteString(fmt.Sprintf("   Lines: %d\n", totalLines))

		// Extract functions, classes, etc.
		outline := extractOutline(path, lines)
		if len(outline) > 0 {
			preview.WriteString("   Structure:\n")
			for _, item := range outline {
				preview.WriteString(fmt.Sprintf("     %s\n", item))
			}
		}
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📄 %s\n", path))
	result.WriteString(preview.String())

	return result.String(), nil
}

func detectFileType(path string, content []byte) string {
	ext := filepath.Ext(path)

	switch ext {
	case ".go":
		return "Go"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".js", ".jsx":
		return "JavaScript"
	case ".py":
		return "Python"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".c", ".h":
		return "C"
	case ".cpp", ".hpp":
		return "C++"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	case ".md":
		return "Markdown"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".html", ".htm":
		return "HTML"
	case ".css":
		return "CSS"
	case ".sql":
		return "SQL"
	case ".sh", ".bash":
		return "Shell"
	case ".ps1":
		return "PowerShell"
	case ".dockerfile":
		return "Dockerfile"
	default:
		// Try to detect by content
		contentStr := string(content[:min(100, len(content))])
		if strings.Contains(contentStr, "package main") || strings.Contains(contentStr, "import (") {
			return "Go (detected)"
		}
		if strings.Contains(contentStr, "#!/usr/bin/env python") || strings.Contains(contentStr, "import ") {
			return "Python (detected)"
		}
		return "Text"
	}
}

func extractOutline(path string, lines []string) []string {
	ext := filepath.Ext(path)
	var outline []string

	switch ext {
	case ".go":
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "func ") {
				name := strings.TrimPrefix(trimmed, "func ")
				if idx := strings.Index(name, "{"); idx > 0 {
					name = name[:idx]
				}
				if idx := strings.Index(name, "("); idx > 0 {
					name = name[:idx]
				}
				outline = append(outline, "func "+strings.TrimSpace(name))
			} else if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, " struct") {
				outline = append(outline, trimmed)
			} else if strings.HasPrefix(trimmed, "type ") && strings.Contains(trimmed, " interface") {
				outline = append(outline, trimmed)
			}
		}

	case ".ts", ".tsx", ".js", ".jsx":
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "async function ") {
				outline = append(outline, trimmed)
			} else if strings.HasPrefix(trimmed, "class ") {
				outline = append(outline, trimmed)
			} else if strings.HasPrefix(trimmed, "const ") && strings.Contains(trimmed, "=") {
				if strings.Contains(trimmed, "= (") || strings.Contains(trimmed, "= function") || strings.Contains(trimmed, "= async") {
					outline = append(outline, trimmed)
				}
			}
		}

	case ".py":
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "async def ") {
				outline = append(outline, trimmed)
			} else if strings.HasPrefix(trimmed, "class ") {
				outline = append(outline, trimmed)
			}
		}
	}

	return outline
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
