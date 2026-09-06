package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TUIActivityItem tracks a tool, command, or background task.
type TUIActivityItem struct {
	ID        string
	Type      string // "tool" | "command" | "task"
	Name      string
	Status    string // "running" | "done" | "failed"
	Input     string
	Output    string
	StartTime time.Time
	Duration  time.Duration
	Diff      string
}

// RenderActivityInspector renders the interactive timeline and inspector.
func RenderActivityInspector(items []TUIActivityItem, selectedIdx int, width, height int) string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("236")).Padding(0, 1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	title := fmt.Sprintf("⚡ ACTIVITY INSPECTOR  |  Total Actions: %d", len(items))

	s.WriteString(headerStyle.Render(title) + "\n\n")

	if len(items) == 0 {
		cleanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Padding(2, 2)
		s.WriteString(cleanStyle.Render("No tools or commands executed in this session yet."))
		s.WriteString("\n\n" + hintStyle.Render("Press [Esc] or [Ctrl+B] to return to chat"))
		return s.String()
	}

	// List of activities
	sectionTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	s.WriteString(sectionTitle.Render("📋 EXECUTION TIMELINE (Recent actions)") + "\n")

	// Render recent 8 items or around selectedIdx
	start := 0
	if selectedIdx >= 8 {
		start = selectedIdx - 7
	}
	end := start + 8
	if end > len(items) {
		end = len(items)
	}

	for i := start; i < end; i++ {
		item := items[i]
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if i == selectedIdx {
			cursor = "▸ "
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("237"))
		}

		statusIcon := "✓"
		statusColor := "82"
		if item.Status == "running" {
			statusIcon = "⟳"
			statusColor = "214"
		} else if item.Status == "failed" {
			statusIcon = "✖"
			statusColor = "196"
		}
		statusBadge := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render("[" + statusIcon + "]")

		durStr := ""
		if item.Duration > 0 {
			durStr = fmt.Sprintf(" (%v)", item.Duration.Round(time.Millisecond))
		}

		s.WriteString(fmt.Sprintf("%s%s %s %s%s\n",
			cursor,
			statusBadge,
			itemStyle.Render(item.Name),
			dimStyle.Render(item.Type),
			dimStyle.Render(durStr),
		))
	}

	// Details box for selected item
	if selectedIdx >= 0 && selectedIdx < len(items) {
		selected := items[selectedIdx]
		s.WriteString("\n" + sectionTitle.Render(fmt.Sprintf("🔍 DETAILS: %s", selected.Name)) + "\n")

		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("239")).
			Padding(0, 1).
			Width(width - 4)

		var details strings.Builder
		details.WriteString(fmt.Sprintf("Status: %s  |  Started: %s  |  Duration: %v\n",
			selected.Status, selected.StartTime.Format("15:04:05"), selected.Duration.Round(time.Millisecond)))

		if selected.Input != "" {
			inputPreview := selected.Input
			if len(inputPreview) > 200 {
				inputPreview = inputPreview[:197] + "..."
			}
			details.WriteString("\nParameters:\n  " + dimStyle.Render(inputPreview) + "\n")
		}

		if selected.Output != "" {
			outputPreview := selected.Output
			if len(outputPreview) > 300 {
				outputPreview = outputPreview[:297] + "..."
			}
			details.WriteString("\nResult:\n  " + lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render(outputPreview) + "\n")
		}

		if selected.Diff != "" {
			details.WriteString("\nDiff preview:\n")
			lines := strings.Split(selected.Diff, "\n")
			limit := 6
			if len(lines) < limit {
				limit = len(lines)
			}
			for j := 0; j < limit; j++ {
				l := lines[j]
				if strings.HasPrefix(l, "+") {
					details.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(l) + "\n")
				} else if strings.HasPrefix(l, "-") {
					details.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(l) + "\n")
				} else {
					details.WriteString(l + "\n")
				}
			}
		}

		s.WriteString(boxStyle.Render(details.String()) + "\n")
	}

	s.WriteString("\n" + hintStyle.Render("[↑/↓] Navigate items  [Esc/Ctrl+B] Back to chat"))
	return s.String()
}
