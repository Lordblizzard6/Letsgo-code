package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/user/go-claude-code/internal/db"

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
	currentSlashSessionID string
	// AvailableSlashCommands lista todos los comandos slash disponibles
	AvailableSlashCommands = []SlashCommand{
		{Name: "/help", Description: "Show available commands"},
		{Name: "/init", Description: "Initialize AGENTS.md rules for this workspace"},
		{Name: "/quit", Description: "Exit the chat"},
		{Name: "/exit", Description: "Exit the chat"},
		{Name: "/clear", Description: "Clear conversation history"},
		{Name: "/model", Description: "Show/change current model"},
		{Name: "/models", Description: "List all available models"},
		{Name: "/cost", Description: "Show session cost"},
		{Name: "/tokens", Description: "Show token usage"},
		{Name: "/compact", Description: "Compact conversation"},
		{Name: "/save", Description: "Save session"},
		{Name: "/rename", Description: "Rename current active session"},
		{Name: "/sessions", Description: "List saved sessions"},
		{Name: "/open", Description: "Resume a session: /open <id>"},
		{Name: "/new", Description: "Create a new session"},
		{Name: "/files", Description: "List files in context"},
		{Name: "/diff", Description: "Show changes and review diff"},
		{Name: "/review", Description: "Review unstaged/staged git changes"},
		{Name: "/tasks", Description: "Inspect executed tools and background tasks"},
		{Name: "/terminal", Description: "Show terminal commands and outputs"},
		{Name: "/skills", Description: "List available agent skills"},
		{Name: "/share", Description: "Share conversation export"},
		{Name: "/rollback", Description: "Revert last message and restore prompt"},
		{Name: "/settings", Description: "Open settings / model picker"},
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
	case "/init":
		cwd, _ := os.Getwd()
		msg, err := initWorkspaceRules(cwd)
		if err != nil {
			return "❌ " + err.Error(), true, false
		}
		return msg, true, false
	case "/clear":
		return handleClear(), true, false
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
	case "/save":
		return handleSave(), true, false
	case "/rename":
		if strings.TrimSpace(args) == "" {
			return "✏️ Uso: /rename <nuevo nombre>", true, false
		}
		targetID := currentSlashSessionID
		if targetID == "" {
			active, _ := db.GetActiveSession()
			if active != nil {
				targetID = active.ID
			}
		}
		if targetID == "" {
			return "❌ No hay sesión activa para renombrar", true, false
		}
		if err := db.RenameSession(targetID, strings.TrimSpace(args)); err != nil {
			return fmt.Sprintf("❌ Error al renombrar: %v", err), true, false
		}
		return fmt.Sprintf("✅ Sesión renombrada a: %s", strings.TrimSpace(args)), true, false
	case "/sessions":
		return handleSessions(), true, false
	case "/open":
		if args == "" {
			return "📂 Uso: /open <id>. Ejecuta /sessions para ver los ids.", true, false
		}
		return "📂 Retomando sesión " + args + "…", true, false
	case "/new":
		return "✨ Creando una nueva sesión…", true, false
	case "/files":
		return tools.GetContextFilesSummary(), true, false
	case "/diff", "/review":
		cmd := exec.Command("git", "diff", "HEAD")
		out, err := cmd.CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return string(out), true, false
		}
		return tools.GetSessionDiff(), true, false
	case "/tasks", "/terminal":
		return "⚡ Abriendo el inspector de actividades y herramientas (Ctrl+B)", true, false
	case "/skills":
		skills := tools.GetBuiltinSkills()
		var names []string
		for _, s := range skills {
			names = append(names, fmt.Sprintf("• %s: %s", s.Name, s.Description))
		}
		if len(names) == 0 {
			return "🎯 Habilidades disponibles: speckit-plan, speckit-implement, speckit-tasks, speckit-specify", true, false
		}
		return "🎯 Habilidades disponibles:\n" + strings.Join(names, "\n"), true, false
	case "/share":
		return "📋 Conversación lista para compartir (exportación en Markdown disponible)", true, false
	case "/rollback":
		targetID := currentSlashSessionID
		if targetID == "" {
			active, _ := db.GetActiveSession()
			if active != nil {
				targetID = active.ID
			}
		}
		if targetID == "" {
			return "❌ No hay sesión activa para rebobinar", true, false
		}
		restored, err := db.RollbackSession(targetID, 0)
		if err != nil {
			return fmt.Sprintf("❌ Error al rebobinar: %v", err), true, false
		}
		return fmt.Sprintf("↩ Mensaje revertido. Prompt restaurado:\n%s", restored), true, false
	case "/settings":
		return "⚙️ Press Ctrl+S to open settings", true, false
	case "/context":
		return tools.GetContextFilesSummary(), true, false
	case "/memory":
		return "🧠 Memory: " + args, true, false
	case "/task":
		return "📋 Task: " + args, true, false
	case "/undo":
		return handleUndoRedo("undo"), true, false
	case "/redo":
		return handleUndoRedo("redo"), true, false
	case "/search":
		return handleSearch(args), true, false
	default:
		return fmt.Sprintf("❓ Unknown command: %s\nType /help for available commands", cmd), true, false
	}
}

func SetSlashCommandSessionID(sessionID string) {
	currentSlashSessionID = sessionID
}

// handleClear returns the presentation copy for /clear. The state change
// itself is engine-owned (contract §2 slash:{clear}, T029): the TUI's enter
// path delegates via engine.Send(SlashCommand{Name: "clear"}) and the engine
// wipes memory + persistence and emits SessionCleared.
func handleClear() string {
	if currentSlashSessionID == "" {
		fmt.Fprintln(os.Stderr, "[slash] /clear failed: empty session id")
		return "❌ /clear no disponible: sesión actual no identificada"
	}
	fmt.Fprintf(os.Stderr, "[slash] /clear requested for session=%s\n", currentSlashSessionID)
	return "🧹 Conversación limpiada en memoria activa y persistencia"
}

// handleSessions lists the persisted sessions via the contract (§3
// ListSessions): the TUI shares the same DB as the GUI (T031, FR-006).
func handleSessions() string {
	sessions, err := db.ListSessions()
	if err != nil {
		return "❌ Error al listar sesiones: " + err.Error()
	}
	if len(sessions) == 0 {
		return "📂 No hay sesiones guardadas. Usa /new para crear una."
	}
	var b strings.Builder
	b.WriteString("📂 Sesiones guardadas:\n\n")
	for i, s := range sessions {
		if i >= 10 {
			b.WriteString(fmt.Sprintf("\n... y %d más (usa /open <id> para retomar)\n", len(sessions)-10))
			break
		}
		short := s.ID
		if len(short) > 14 {
			short = short[:14]
		}
		b.WriteString(fmt.Sprintf("  • `%s` %s (actualizado %s)\n", short, s.Name, s.UpdatedAt.Format("2006-01-02 15:04")))
	}
	b.WriteString("\nUsa /open <id> para retomar una sesión o /new para crear otra.")
	return b.String()
}

func handleSave() string {
	if currentSlashSessionID == "" {
		fmt.Fprintln(os.Stderr, "[slash] /save failed: empty session id")
		return "❌ /save no disponible: sesión actual no identificada"
	}
	if err := db.UpdateSessionTimestamp(currentSlashSessionID); err != nil {
		fmt.Fprintf(os.Stderr, "[slash] /save failed: %v\n", err)
		return fmt.Sprintf("❌ Error al guardar sesión: %v", err)
	}
	fmt.Fprintf(os.Stderr, "[slash] /save executed for session=%s\n", currentSlashSessionID)
	return "💾 Sesión guardada en DB correctamente"
}

func handleUndoRedo(action string) string {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "⚠️ No estás en un repositorio Git. /" + action + " requiere git para rastrear cambios."
	}
	if action == "undo" {
		lastCommitCmd := exec.Command("git", "rev-parse", "HEAD")
		lastCommitOut, err := lastCommitCmd.Output()
		if err != nil {
			return fmt.Sprintf("❌ Error al obtener último commit: %v", err)
		}
		lastCommit := strings.TrimSpace(string(lastCommitOut))
		if lastCommit == "" {
			return "ℹ️ No hay commits para deshacer."
		}
		revertCmd := exec.Command("git", "revert", "--no-commit", "HEAD")
		if err := revertCmd.Run(); err != nil {
			resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
			if err := resetCmd.Run(); err != nil {
				return fmt.Sprintf("❌ Error al deshacer: %v", err)
			}
		}
		_ = exec.Command("git", "commit", "-m", "Undo: last action").Run()
		limit := 8
		if len(lastCommit) < 8 {
			limit = len(lastCommit)
		}
		return fmt.Sprintf("✅ Deshecho commit %s", lastCommit[:limit])
	} else if action == "redo" {
		msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
		msgOut, err := msgCmd.Output()
		if err != nil {
			return fmt.Sprintf("❌ Error al verificar commit: %v", err)
		}
		commitMsg := strings.TrimSpace(string(msgOut))
		if !strings.HasPrefix(commitMsg, "Undo:") {
			return "ℹ️ La última acción no fue un undo. Nada que rehacer."
		}
		revertCmd := exec.Command("git", "revert", "--no-commit", "HEAD")
		if err := revertCmd.Run(); err != nil {
			return fmt.Sprintf("❌ Error al rehacer: %v", err)
		}
		newMsg := strings.Replace(commitMsg, "Undo:", "Redo:", 1)
		_ = exec.Command("git", "commit", "-m", newMsg).Run()
		return "✅ Acción rehecha: " + newMsg
	}
	return "Comando no reconocido"
}

func initWorkspaceRules(cwd string) (string, error) {
	agentsPath := filepath.Join(cwd, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		return fmt.Sprintf("ℹ️ AGENTS.md ya existe en %s", agentsPath), nil
	}

	stack := "Generic"
	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		stack = "Go"
	} else if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		stack = "Node.js / TypeScript"
	} else if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		stack = "Rust"
	} else if _, err := os.Stat(filepath.Join(cwd, "pyproject.toml")); err == nil {
		stack = "Python"
	} else if _, err := os.Stat(filepath.Join(cwd, "pubspec.yaml")); err == nil {
		stack = "Flutter / Dart"
	}

	template := fmt.Sprintf(`# AGENTS.md

Operating instructions and rules for coding agents in this repository.

## 0. Non-negotiables
1. Working code only. Finish the job. Plausibility is not correctness.
2. Direct, concise communication. No flattery, no filler.
3. Surgical changes: Touch only what you must.
4. Goal-driven: Verify your changes with tests before reporting completion.

## 1. Project Context
- Stack: %s

## 2. Conventions & Style
- Follow existing patterns in the codebase.
- Return errors to callers, do not silently swallow them.
`, stack)

	if err := os.WriteFile(agentsPath, []byte(strings.TrimSpace(template)+"\n"), 0644); err != nil {
		return "", fmt.Errorf("falló al escribir AGENTS.md: %w", err)
	}

	return fmt.Sprintf("✅ Creado AGENTS.md para el stack %s", stack), nil
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
	args = strings.TrimSpace(args)
	if args == "" {
		return " **Search Usage:**\nType `/search <keyword>` to search through the conversation history.\n\nExample: `/search function`"
	}
	page := 1
	limit := 5
	query := args
	for _, token := range strings.Fields(args) {
		if strings.HasPrefix(token, "page=") {
			fmt.Sscanf(token, "page=%d", &page)
			query = strings.TrimSpace(strings.Replace(query, token, "", 1))
		}
		if strings.HasPrefix(token, "limit=") {
			fmt.Sscanf(token, "limit=%d", &limit)
			query = strings.TrimSpace(strings.Replace(query, token, "", 1))
		}
	}
	return tools.SearchInConversationWithPagination(query, page, limit)
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
