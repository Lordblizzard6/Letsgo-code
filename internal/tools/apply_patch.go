package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/user/go-claude-code/internal/api"
)

type ApplyPatchTool struct{}

func (t *ApplyPatchTool) Definition() api.Tool {
	return api.Tool{
		Name:        "apply_patch",
		Description: "Apply a standard unified diff patch to modify existing files or create new files. The patch should follow standard unified diff format (with ---, +++, and @@ hunk headers).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"patch": map[string]interface{}{
					"type":        "string",
					"description": "The unified diff patch to apply.",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Optional explicit file path if not present in the patch headers.",
				},
			},
			"required": []string{"patch"},
		},
	}
}

func (t *ApplyPatchTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	patchText, ok := m["patch"].(string)
	if !ok || strings.TrimSpace(patchText) == "" {
		return "", fmt.Errorf("invalid input: patch string expected")
	}
	explicitPath, _ := m["path"].(string)

	filePath, hunks, err := parseUnifiedDiff(patchText)
	if err != nil {
		return "", fmt.Errorf("failed to parse patch: %w", err)
	}

	if explicitPath != "" {
		filePath = explicitPath
	}

	if filePath == "" {
		return "", fmt.Errorf("no target file found in patch header and no explicit path provided")
	}

	// Clean file path (remove a/ or b/ prefixes if present)
	filePath = filepath.Clean(filePath)

	var originalLines []string
	if content, err := os.ReadFile(filePath); err == nil {
		// Normalize CRLF to LF
		text := strings.ReplaceAll(string(content), "\r\n", "\n")
		originalLines = strings.Split(text, "\n")
	} else if os.IsNotExist(err) {
		originalLines = []string{}
	} else {
		return "", fmt.Errorf("failed to read target file %s: %w", filePath, err)
	}

	updatedLines, err := applyHunks(originalLines, hunks)
	if err != nil {
		return "", fmt.Errorf("failed to apply patch to %s: %w", filePath, err)
	}

	// Ensure directory exists
	if dir := filepath.Dir(filePath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	newContent := strings.Join(updatedLines, "\n")
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write patched file %s: %w", filePath, err)
	}

	return fmt.Sprintf("Successfully applied patch to %s (%d hunks)", filePath, len(hunks)), nil
}

type diffHunk struct {
	oldStart int
	oldCount int
	newStart int
	newCount int
	lines    []string
}

func parseUnifiedDiff(diff string) (string, []diffHunk, error) {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	var targetFile string
	var hunks []diffHunk
	var currentHunk *diffHunk

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- ") {
			// Old file header
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			// New file header
			rawPath := strings.TrimSpace(strings.TrimPrefix(line, "+++ "))
			rawPath = strings.TrimPrefix(rawPath, "b/")
			rawPath = strings.TrimPrefix(rawPath, "a/")
			targetFile = rawPath
			continue
		}
		if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}
			hunk, err := parseHunkHeader(line)
			if err != nil {
				return "", nil, err
			}
			currentHunk = hunk
			continue
		}
		if currentHunk != nil {
			currentHunk.lines = append(currentHunk.lines, line)
		}
	}

	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return targetFile, hunks, scanner.Err()
}

func parseHunkHeader(line string) (*diffHunk, error) {
	// Format: @@ -oldStart,oldCount +newStart,newCount @@
	parts := strings.Split(line, "@@")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid hunk header: %s", line)
	}
	spec := strings.TrimSpace(parts[1])
	ranges := strings.Fields(spec)
	if len(ranges) < 2 {
		return nil, fmt.Errorf("invalid hunk range: %s", spec)
	}

	oldStart, oldCount := 1, 1
	newStart, newCount := 1, 1

	oldPart := strings.TrimPrefix(ranges[0], "-")
	oldSub := strings.Split(oldPart, ",")
	if n, err := strconv.Atoi(oldSub[0]); err == nil {
		oldStart = n
	}
	if len(oldSub) > 1 {
		if n, err := strconv.Atoi(oldSub[1]); err == nil {
			oldCount = n
		}
	}

	newPart := strings.TrimPrefix(ranges[1], "+")
	newSub := strings.Split(newPart, ",")
	if n, err := strconv.Atoi(newSub[0]); err == nil {
		newStart = n
	}
	if len(newSub) > 1 {
		if n, err := strconv.Atoi(newSub[1]); err == nil {
			newCount = n
		}
	}

	return &diffHunk{
		oldStart: oldStart,
		oldCount: oldCount,
		newStart: newStart,
		newCount: newCount,
		lines:    []string{},
	}, nil
}

func applyHunks(original []string, hunks []diffHunk) ([]string, error) {
	if len(hunks) == 0 {
		return original, nil
	}

	var result []string
	origIdx := 0

	for _, hunk := range hunks {
		targetLine := hunk.oldStart - 1
		if targetLine < 0 {
			targetLine = 0
		}

		// Copy untouched lines before this hunk
		for origIdx < targetLine && origIdx < len(original) {
			result = append(result, original[origIdx])
			origIdx++
		}

		for _, hl := range hunk.lines {
			if len(hl) == 0 {
				continue
			}
			prefix := hl[0]
			content := hl[1:]

			switch prefix {
			case ' ':
				// Context line
				if origIdx < len(original) {
					result = append(result, original[origIdx])
					origIdx++
				} else {
					result = append(result, content)
				}
			case '-':
				// Deleted line
				if origIdx < len(original) {
					origIdx++
				}
			case '+':
				// Added line
				result = append(result, content)
			}
		}
	}

	// Append remaining original lines
	for origIdx < len(original) {
		result = append(result, original[origIdx])
		origIdx++
	}

	return result, nil
}
