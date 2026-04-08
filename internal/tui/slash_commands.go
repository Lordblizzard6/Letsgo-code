package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/tools"
)

// SlashCommand represents a slash command
type SlashCommand struct {
	Name        string
	Description string
}

// MentionContext represents what can be mentioned
type MentionContext struct {
	Type        string
	Icon        string
	Description string
}

var (
	// AvailableSlashCommands lista todos los comandos slash disponibles
	AvailableSlashCommands = []SlashCommand{
		{Name: "/help", Description: "Show available commands"},
		{Name: "/quit", Description: "Exit the chat"},
		{Name: "/exit", Description: "Exit the chat"},
		{Name: "/clear", Description: "Clear conversation history"},
		{Name: "/model", Description: "Show/change current model"},
		{Name: "/models", Description: "List all available models"},
		{Name: "/cost", Description: "Show session cost"},
		{Name: "/tokens", Description: "Show token usage"},
		{Name: "/compact", Description: "Compact conversation"},
		{Name: "/save", Description: "Save session"},
		{Name: "/files", Description: "List files in context"},
		{Name: "/diff", Description: "Show changes"},
		{Name: "/settings", Description: "Open settings"},
		{Name: "/context", Description: "Show context info"},
		{Name: "/memory", Description: "Manage memory"},
		{Name: "/task", Description: "Create task"},
		{Name: "/undo", Description: "Undo action"},
		{Name: "/redo", Description: "Redo action"},
		{Name: "/search", Description: "Search conversation"},
	}

	// AvailableMentions lista todos los tipos de menciones
	AvailableMentions = []MentionContext{
		{Type: "file", Icon: "📄", Description: "Reference a file"},
		{Type: "func", Icon: "🔧", Description: "Reference a function"},
		{Type: "class", Icon: "📦", Description: "Reference a class"},
		{Type: "skill", Icon: "🎯", Description: "Use a skill"},
		{Type: "tool", Icon: "🛠️", Description: "Reference a tool"},
	}
)

// ProcessSlashCommand procesa un comando slash y retorna la respuesta
func ProcessSlashCommand(input string) (response string, shouldContinue bool, shouldQuit bool) {
	parts := strings.SplitN(input, " ", 2)
	cmd := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	switch cmd {
	case "/quit", "/exit":
		return "👋 Goodbye!", false, true
	case "/clear":
		return "🧹 Conversation cleared!", true, false
	case "/help":
		return formatHelp(), true, false
	case "/model":
		return handleModel(args), true, false
	case "/models":
		return formatModels(), true, false
	case "/cost":
		return handleCost(args), true, false
	case "/tokens":
		return handleTokens(args), true, false
	case "/compact":
		return "🗜️ Conversation compacted!", true, false
	case "/save":
		return "💾 Session saved!", true, false
	case "/files":
		return tools.GetContextFilesSummary(), true, false
	case "/diff":
		return tools.GetSessionDiff(), true, false
	case "/settings":
		return "⚙️ Press Ctrl+S to open settings", true, false
	case "/context":
		return tools.GetContextFilesSummary(), true, false
	case "/memory":
		return "🧠 Memory: " + args, true, false
	case "/task":
		return "📋 Task: " + args, true, false
	case "/undo":
		return "↩️ Undo available via git", true, false
	case "/redo":
		return "↪️ Redo available via git", true, false
	case "/search":
		return handleSearch(args), true, false
	default:
		return fmt.Sprintf("❓ Unknown command: %s\nType /help for available commands", cmd), true, false
	}
}

// ProcessMentions procesa menciones @ en el texto y expande el contexto
func ProcessMentions(text string) string {
	// Procesar menciones @file
	if strings.Contains(text, "@") {
		// Encontrar todas las menciones
		words := strings.Fields(text)
		var expanded []string
		for _, word := range words {
			if strings.HasPrefix(word, "@") {
				mention := word[1:]
				// Determinar tipo de mención
				if expandedContent := expandMention(mention); expandedContent != "" {
					expanded = append(expanded, expandedContent)
					continue
				}
			}
			expanded = append(expanded, word)
		}
		return strings.Join(expanded, " ")
	}
	return text
}

// expandMention expande una mención al contenido correspondiente
func expandMention(mention string) string {
	// Intentar como archivo primero
	if _, err := os.Stat(mention); err == nil {
		return mentionFile(mention)
	}

	// Intentar con extensión común
	exts := []string{".go", ".py", ".js", ".ts", ".json", ".yaml", ".md", ".txt"}
	for _, ext := range exts {
		if _, err := os.Stat(mention + ext); err == nil {
			return mentionFile(mention + ext)
		}
	}

	// Buscar en directorio actual
	matches, _ := filepath.Glob("*" + mention + "*")
	if len(matches) > 0 {
		return mentionFile(matches[0])
	}

	return ""
}

// ========== Handlers Slash Commands ==========

func handleModel(args string) string {
	if args == "" {
		return fmt.Sprintf("🤖 Current: **%s**\n\nProvider: %s\n\nUse /models for options or Ctrl+S to change", config.AppConfig.Model, config.AppConfig.BaseURL)
	}
	return "🔄 Use Ctrl+S to change models"
}

func handleCost(args string) string {
	tracker := tools.GetCostTracker()
	cost := tracker.GetSessionCost()
	return fmt.Sprintf("💰 Session cost: **$%.4f**", cost)
}

func handleTokens(args string) string {
	return "📊 Token tracking active. Use /cost for breakdown"
}

func handleSearch(args string) string {
	if args == "" {
		return " **Search Usage:**\nType `/search <keyword>` to search through the conversation history.\n\nExample: `/search function`"
	}

	return fmt.Sprintf(" Searching for '%s'... (Feature coming soon - will search through conversation history)", args)
}

// getCurrentSessionID retorna el ID de sesión actual (placeholder)
func getCurrentSessionID() string {
	// Esta función será llamada desde el contexto del modelo
	// Por ahora retornamos vacío, el handler real lo implementará
	return ""
}

// ========== Handlers Mentions ==========

func mentionFile(identifier string) string {
	content, err := os.ReadFile(identifier)
	if err != nil {
		return ""
	}
	// Limit content size
	contentStr := string(content)
	if len(contentStr) > 2000 {
		contentStr = contentStr[:2000] + "\n... [truncated]"
	}
	return fmt.Sprintf("\n📄 File: %s\n```\n%s\n```\n", identifier, contentStr)
}

// ========== Formatters ==========

func formatHelp() string {
	// Styles modernos
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		MarginBottom(1)

	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Width(14)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true)

	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	var b strings.Builder

	// Título
	b.WriteString(titleStyle.Render("📚 Command Reference") + "\n\n")

	// Sección de Navegación
	b.WriteString(sectionStyle.Render("🎮 Navigation") + "\n")
	b.WriteString(fmt.Sprintf("  %s  Scroll conversation up/down\n", keyStyle.Render("↑/↓, PgUp/PgDown")))
	b.WriteString(fmt.Sprintf("  %s     Jump to top/bottom of conversation\n", keyStyle.Render("Home/End")))
	b.WriteString(fmt.Sprintf("  %s         Open settings / change model\n", keyStyle.Render("Ctrl+S")))
	b.WriteString(fmt.Sprintf("  %s              Exit the application\n", keyStyle.Render("Ctrl+C")))
	b.WriteString("\n")

	// Sección de Slash Commands
	b.WriteString(sectionStyle.Render("⚡ Slash Commands") + "\n")

	// Agrupar comandos por categoría
	categories := map[string][]SlashCommand{
		"Core":    {{"/help", "Show this help"}, {"/quit", "Exit chat"}, {"/clear", "Clear history"}},
		"Model":   {{"/model", "Current model"}, {"/models", "List all models"}, {"/settings", "Open settings"}},
		"Session": {{"/cost", "Session cost"}, {"/tokens", "Token usage"}, {"/save", "Save session"}},
		"Context": {{"/files", "Files in context"}, {"/diff", "Show changes"}, {"/context", "Context info"}},
		"Tools":   {{"/task", "Create task"}, {"/memory", "Manage memory"}, {"/search", "Search conversation"}},
	}

	for cat, cmds := range categories {
		catStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
		b.WriteString("  " + catStyle.Render(cat) + "\n")
		for _, cmd := range cmds {
			b.WriteString(fmt.Sprintf("    %s %s\n",
				cmdStyle.Render(cmd.Name),
				descStyle.Render(cmd.Description)))
		}
	}

	// Sección de Mentions
	b.WriteString("\n" + sectionStyle.Render("🔍 Mentions"))
	b.WriteString(tipStyle.Render(" (use @ in your message)") + "\n")
	for _, m := range AvailableMentions {
		b.WriteString(fmt.Sprintf("  %s %s %s\n",
			iconStyle.Render(m.Icon),
			cmdStyle.Render("@"+m.Type),
			descStyle.Render(m.Description)))
	}

	// Tips
	b.WriteString("\n" + sectionStyle.Render("💡 Pro Tips") + "\n")
	b.WriteString(tipStyle.Render("  • Type / then Tab to autocomplete commands\n"))
	b.WriteString(tipStyle.Render("  • Use @filename to quickly include file content\n"))
	b.WriteString(tipStyle.Render("  • Press ↑ to recall previous messages\n"))

	return b.String()
}

func formatModels() string {
	// Styles modernos
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	currentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("86")).
		Foreground(lipgloss.Color("0")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(35)

	selectedCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Background(lipgloss.Color("237")).
		Padding(1).
		Width(35)

	numStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true)

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	providerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("141")).
		Italic(true)

	models := []struct {
		num      string
		name     string
		desc     string
		provider string
		color    string
	}{
		{"1", "Claude 3.5 Sonnet", "Best overall", "Anthropic", "86"},
		{"2", "Claude 3 Opus", "Most powerful", "Anthropic", "141"},
		{"3", "GPT-4o", "OpenAI's best", "OpenAI", "45"},
		{"4", "GPT-4o Mini", "Fast & cheap", "OpenAI", "33"},
		{"5", "Llama 3.3 70B", "Fast via Groq", "Groq", "214"},
		{"6", "Llama 3.1 8B", "Ultra fast", "Groq", "208"},
		{"7", "GPT-OSS 120B", "OpenAI on Groq", "Groq", "220"},
	}

	var b strings.Builder

	// Título
	b.WriteString(titleStyle.Render("🤖 Model Selection") + "\n")

	// Modelo actual
	currentModel := config.AppConfig.Model
	b.WriteString("Current: " + currentStyle.Render(" "+currentModel+" ") + "\n\n")

	// Grid de modelos
	for i, m := range models {
		isCurrent := strings.Contains(currentModel, m.name) ||
			strings.Contains(currentModel, strings.ToLower(strings.ReplaceAll(m.name, " ", "-")))

		style := cardStyle
		if isCurrent {
			style = selectedCardStyle
		}

		providerColor := lipgloss.Color(m.color)
		providerStyled := lipgloss.NewStyle().Foreground(providerColor).Render(m.provider)

		cardContent := fmt.Sprintf("%s %s\n%s\n%s",
			numStyle.Render(m.num+"."),
			nameStyle.Render(m.name),
			descStyle.Render(m.desc),
			providerStyle.Render("via "+providerStyled))

		b.WriteString(style.Render(cardContent) + "\n\n")

		// Cada 2 modelos, agregar espacio (simulación de grid simple)
		if (i+1)%2 == 0 && i < len(models)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + tipStyle.Render("💡 Press the number key in settings (Ctrl+S) to switch models"))

	return b.String()
}

// GetSlashSuggestions retorna sugerencias para autocompletado
func GetSlashSuggestions(prefix string) []SlashCommand {
	var matches []SlashCommand
	for _, cmd := range AvailableSlashCommands {
		if strings.HasPrefix(cmd.Name, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// GetMentionSuggestions retorna sugerencias de menciones
func GetMentionSuggestions(prefix string) []string {
	var matches []string

	// Buscar archivos
	files, _ := filepath.Glob("*" + prefix + "*")
	matches = append(matches, files...)

	// Buscar con extensión
	for _, ext := range []string{".go", ".js", ".ts", ".py", ".md"} {
		pattern := "*" + prefix + "*" + ext
		files, _ := filepath.Glob(pattern)
		for _, f := range files {
			if !contains(matches, f) {
				matches = append(matches, f)
			}
		}
	}

	sort.Strings(matches)
	return matches
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
