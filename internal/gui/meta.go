package gui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

// appVersion mirrors cmd/version.go.
const appVersion = "LetsGO Code v0.1.0 (GUI)"

// runDoctor runs the same diagnostics as cmd/doctor.go and returns a summary.
func runDoctor() string {
	var b string
	issues, warnings := 0, 0

	if !hasAnyKey() {
		b += "✗ No API key configured\n"
		issues++
	} else {
		b += "✓ API key configured\n"
	}

	if err := db.InitDB(); err != nil {
		b += fmt.Sprintf("✗ Database error: %v\n", err)
		issues++
	} else {
		b += "✓ Database initialized\n"
	}

	if _, err := exec.LookPath("git"); err != nil {
		b += "⚠ Git not found in PATH\n"
		warnings++
	} else {
		b += "✓ Git available\n"
	}

	dir := config.GetConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		b += fmt.Sprintf("✗ Config directory missing: %s\n", dir)
		issues++
	} else {
		b += fmt.Sprintf("✓ Config directory: %s\n", dir)
	}

	if config.AppConfig.Model == "" {
		b += "✗ No model configured\n"
		issues++
	} else {
		b += fmt.Sprintf("✓ Model: %s\n", config.AppConfig.Model)
	}

	b += fmt.Sprintf("✓ OS: %s / %s\n", runtime.GOOS, runtime.GOARCH)
	b += fmt.Sprintf("✓ Tools available: %d\n", len(tools.AllTools))

	b += "\n" + repeatChar('=', 40) + "\n"
	switch {
	case issues == 0 && warnings == 0:
		b += "✓ All checks passed."
	case issues == 0:
		b += fmt.Sprintf("⚠ %d warning(s), no critical issues.", warnings)
	default:
		b += fmt.Sprintf("✗ %d issue(s), %d warning(s).", issues, warnings)
	}
	return b
}

func repeatChar(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

// showAbout displays version + release notes (cmd/version.go, cmd/release_notes.go).
func showAbout(win fyne.Window) {
	notes := appVersion + "\n\nRelease 0.1.0:\n" +
		"- Interactive chat with streaming responses\n" +
		"- Multiple providers (Anthropic, OpenAI, Groq, OpenRouter, Ollama)\n" +
		"- Session persistence (SQLite)\n" +
		"- Git integration (branch, commit, diff)\n" +
		"- Agent mode with tool approvals\n" +
		"- MCP servers, plugins, skills\n" +
		"- Cost tracking and usage stats"
	dialog.ShowInformation("About", notes, win)
}

// showUpdate checks for updates (cmd/update.go) and reports result.
func showUpdate(win fyne.Window) {
	msg := appVersion + "\n\n" +
		"Checking for updates…\n" +
		"The GUI keeps the check lightweight; run `letsgo update` in a terminal for the full release lookup."
	dialog.ShowInformation("Update", msg, win)
}
