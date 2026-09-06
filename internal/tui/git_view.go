package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TUIGitFile represents a file with changes in git.
type TUIGitFile struct {
	Path      string
	Name      string
	Dir       string
	Status    string // "M", "A", "D", "??"
	Staged    bool
	Additions int
	Deletions int
}

// TUIGitSummary represents the overall git repository state.
type TUIGitSummary struct {
	Branch           string
	Clean            bool
	UncommittedCount int
	CommittedCount   int
	Files            []TUIGitFile
	SelectedFile     string
	DiffContent      string
}

func runGitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// GetTUIGitSummary gathers status, numstat, and branch info.
func GetTUIGitSummary() (TUIGitSummary, error) {
	summary := TUIGitSummary{
		Clean: true,
	}

	branchOut, err := runGitCmd("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return summary, err
	}
	summary.Branch = strings.TrimSpace(branchOut)

	revOut, err := runGitCmd("rev-list", "--count", "HEAD")
	if err == nil {
		if c, err := strconv.Atoi(strings.TrimSpace(revOut)); err == nil {
			summary.CommittedCount = c
		}
	}

	statsMap := make(map[string][2]int)
	parseNumstat := func(output string) {
		for _, line := range strings.Split(output, "\n") {
			parts := strings.Split(strings.TrimSpace(line), "\t")
			if len(parts) >= 3 {
				add, _ := strconv.Atoi(parts[0])
				del, _ := strconv.Atoi(parts[1])
				p := parts[2]
				cur := statsMap[p]
				statsMap[p] = [2]int{cur[0] + add, cur[1] + del}
			}
		}
	}

	if numstatOut, err := runGitCmd("diff", "--numstat"); err == nil {
		parseNumstat(numstatOut)
	}
	if numstatCached, err := runGitCmd("diff", "--cached", "--numstat"); err == nil {
		parseNumstat(numstatCached)
	}

	statusOut, err := runGitCmd("status", "--porcelain")
	if err != nil {
		return summary, err
	}

	lines := strings.Split(statusOut, "\n")
	for _, l := range lines {
		l = strings.TrimRight(l, "\r\n")
		if len(l) < 3 {
			continue
		}
		statusCode := strings.TrimSpace(l[0:2])
		rawPath := strings.TrimSpace(l[3:])
		rawPath = strings.Trim(rawPath, "\"")
		if rawPath == "" {
			continue
		}

		staged := l[0] != ' ' && l[0] != '?'
		stat := statsMap[rawPath]
		cleanDir := filepath.Dir(rawPath)
		if cleanDir == "." {
			cleanDir = ""
		} else {
			cleanDir = filepath.ToSlash(cleanDir)
		}

		summary.Files = append(summary.Files, TUIGitFile{
			Path:      rawPath,
			Name:      filepath.Base(rawPath),
			Dir:       cleanDir,
			Status:    statusCode,
			Staged:    staged,
			Additions: stat[0],
			Deletions: stat[1],
		})
	}

	summary.UncommittedCount = len(summary.Files)
	summary.Clean = len(summary.Files) == 0

	return summary, nil
}

// StageFile stages a file into git index.
func StageFile(filePath string) error {
	_, err := runGitCmd("add", "--", filePath)
	return err
}

// UnstageFile unstages a file from git index.
func UnstageFile(filePath string) error {
	_, err := runGitCmd("restore", "--staged", "--", filePath)
	return err
}

// StageAll stages all changes.
func StageAll() error {
	_, err := runGitCmd("add", "-A")
	return err
}

// UnstageAll unstages all changes.
func UnstageAll() error {
	_, err := runGitCmd("restore", "--staged", ".")
	return err
}

// CommitChanges commits with message.
func CommitChanges(msg string) error {
	_, err := runGitCmd("commit", "-m", msg)
	return err
}

// GetFileDiff returns unified diff for a path.
func GetFileDiff(filePath string) string {
	diff, _ := runGitCmd("diff", "--", filePath)
	if strings.TrimSpace(diff) == "" {
		diff, _ = runGitCmd("diff", "--cached", "--", filePath)
	}
	return diff
}

// RenderGitReview renders the interactive TUI Git review screen.
func RenderGitReview(summary TUIGitSummary, selectedIdx int, inCommitPrompt bool, commitInput string, width, height int) string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")).Background(lipgloss.Color("236")).Padding(0, 1)
	branchStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	title := fmt.Sprintf("🔍 GIT REVIEW & STAGING  |  🌿 Branch: %s  |  Uncommitted: %d",
		branchStyle.Render(summary.Branch), summary.UncommittedCount)

	s.WriteString(headerStyle.Render(title) + "\n\n")

	if summary.Clean {
		cleanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true).Padding(2, 2)
		s.WriteString(cleanStyle.Render("✓ Working tree is clean. No uncommitted changes."))
		s.WriteString("\n\n" + hintStyle.Render("Press [Esc] or [Ctrl+D] to return to chat"))
		return s.String()
	}

	unstagedFiles := []TUIGitFile{}
	stagedFiles := []TUIGitFile{}
	for _, f := range summary.Files {
		if f.Staged {
			stagedFiles = append(stagedFiles, f)
		} else {
			unstagedFiles = append(unstagedFiles, f)
		}
	}

	sectionTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	s.WriteString(sectionTitle.Render(fmt.Sprintf("📂 UNSTAGED CHANGES (%d)", len(unstagedFiles))) + "\n")
	flatIdx := 0
	for _, f := range unstagedFiles {
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if flatIdx == selectedIdx {
			cursor = "▸ "
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208")).Background(lipgloss.Color("237"))
		}
		statBadge := ""
		if f.Additions > 0 || f.Deletions > 0 {
			statBadge = fmt.Sprintf(" %s %s", addStyle.Render(fmt.Sprintf("+%d", f.Additions)), delStyle.Render(fmt.Sprintf("-%d", f.Deletions)))
		}
		pathStr := f.Name
		if f.Dir != "" {
			pathStr = f.Name + " " + dirStyle.Render(f.Dir)
		}
		s.WriteString(fmt.Sprintf("%s%s [%s]%s\n", cursor, itemStyle.Render(pathStr), f.Status, statBadge))
		flatIdx++
	}
	if len(unstagedFiles) == 0 {
		s.WriteString(dimStyle.Render("  (No unstaged changes)") + "\n")
	}

	s.WriteString("\n" + sectionTitle.Render(fmt.Sprintf("📦 STAGED FOR COMMIT (%d)", len(stagedFiles))) + "\n")
	for _, f := range stagedFiles {
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120"))
		if flatIdx == selectedIdx {
			cursor = "▸ "
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Background(lipgloss.Color("237"))
		}
		statBadge := ""
		if f.Additions > 0 || f.Deletions > 0 {
			statBadge = fmt.Sprintf(" %s %s", addStyle.Render(fmt.Sprintf("+%d", f.Additions)), delStyle.Render(fmt.Sprintf("-%d", f.Deletions)))
		}
		pathStr := f.Name
		if f.Dir != "" {
			pathStr = f.Name + " " + dirStyle.Render(f.Dir)
		}
		s.WriteString(fmt.Sprintf("%s%s [%s]%s\n", cursor, itemStyle.Render(pathStr), f.Status, statBadge))
		flatIdx++
	}
	if len(stagedFiles) == 0 {
		s.WriteString(dimStyle.Render("  (No staged changes)") + "\n")
	}

	// Diff Preview Box for selected file
	if summary.SelectedFile != "" && summary.DiffContent != "" {
		s.WriteString("\n" + sectionTitle.Render("📄 DIFF: "+summary.SelectedFile) + "\n")
		diffBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("239")).
			Padding(0, 1).
			Width(width - 4)

		var diffLines []string
		lines := strings.Split(summary.DiffContent, "\n")
		maxLines := 14
		if len(lines) < maxLines {
			maxLines = len(lines)
		}
		for i := 0; i < maxLines; i++ {
			line := lines[i]
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				diffLines = append(diffLines, addStyle.Render(line))
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				diffLines = append(diffLines, delStyle.Render(line))
			} else if strings.HasPrefix(line, "@@") {
				diffLines = append(diffLines, lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render(line))
			} else {
				diffLines = append(diffLines, line)
			}
		}
		if len(lines) > maxLines {
			diffLines = append(diffLines, dimStyle.Render(fmt.Sprintf("... (%d more lines truncated)", len(lines)-maxLines)))
		}
		s.WriteString(diffBox.Render(strings.Join(diffLines, "\n")) + "\n")
	}

	// Commit prompt or key actions
	if inCommitPrompt {
		promptStyle := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("82")).
			Padding(1, 2).
			Width(width - 4)
		promptTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Render("✍️  Enter commit message (Press Enter to commit, Esc to cancel):")
		s.WriteString("\n" + promptStyle.Render(promptTitle+"\n\n> "+commitInput+"█") + "\n")
	} else {
		s.WriteString("\n" + hintStyle.Render("[↑/↓] Navigate  [s/+] Stage file  [u/-] Unstage  [a] Stage all  [c] Commit  [Esc/Ctrl+D] Back to chat"))
	}

	return s.String()
}
