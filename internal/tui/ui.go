package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/engine"
	"github.com/user/go-claude-code/internal/tools"
)

// engineEventMsg wraps an engine event for the bubbletea program.
type engineEventMsg struct {
	ev engine.Event
}

// tuiDriver preserves the CLI's historical behavior (tools execute without an
// interactive prompt). It satisfies engine.PermissionDriver byte-for-byte.
type tuiDriver struct{}

func (tuiDriver) Prompt(toolName string, input any) (bool, error) { return true, nil }
func (tuiDriver) AutoApprove(category string) bool               { return true }
func (tuiDriver) SessionGranted(sessionID, category string) bool { return true }

type mode int

const (
	chatMode mode = iota
	settingsMode
)

type model struct {
	engine          *engine.Engine
	sessionID       string
	messages        []api.Message
	input           string
	isStreaming     bool
	currentResponse string
	executingTool   string
	currentMode     mode
	spinner         spinner.Model
	renderer        *glamour.TermRenderer
	width           int
	height          int
	viewport        viewport.Model // Usar viewport para scroll profesional
	totalTokens     int
	totalCost       float64
	inputTokens     int
	outputTokens    int
	// Nuevos campos para mejoras de UI
	inputHistory    []string
	historyIndex    int
	suggestions     []string
	showSuggestions bool
	// Campos para selector visual de modelos
	selectedProvider int  // Índice del proveedor seleccionado (0-4)
	selectedModel    int  // Índice del modelo seleccionado dentro del proveedor
	inModelMenu      bool // true cuando estamos en el submenú de modelos
}

func NewModel() model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	db.InitDB()

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create renderer: %v\n", err)
	}

	// Inicializar viewport para scroll profesional
	vp := viewport.New(80, 20)
	vp.SetContent("")

	m := model{
		messages:        []api.Message{},
		spinner:         s,
		renderer:        r,
		currentMode:     chatMode,
		viewport:        vp,
		inputHistory:    []string{},
		historyIndex:    -1,
		suggestions:     []string{},
		showSuggestions: false,
		selectedProvider: 0,
		selectedModel:    0,
		inModelMenu:      false,
	}

	// El engine es el único dueño del loop de conversación (FR-026).
	m.engine = engine.New(tuiDriver{})

	// PARITY: Initial Git Warning
	if tools.IsGitRepo() {
		status := tools.GetGitStatus()
		if status != "" {
			note := api.Message{
				Role:    "assistant",
				Content: "> ℹ️ **Note:** You have uncommitted changes in this repository. Claude will see these changes when reading files.",
			}
			m.messages = append(m.messages, note)
		}
	}

	// Welcome message con tips útiles
	m.messages = append(m.messages, api.Message{
		Role:    "assistant",
		Content: getWelcomeMessage(),
	})

	m.engine.SeedMessages(m.messages)
	m.sessionID = m.engine.SessionID()

	return m
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case engineEventMsg:
		return m.handleEngineEvent(msg.ev), nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.isStreaming {
				// Cancelación limpia (contrato §1 `cancel`, C-002): el stream
				// se aborta sin persistir el parcial (FR-007, SC-008).
				m.engine.Send(engine.Cancel{})
				return m, nil
			}
			m.engine.Stop()
			return m, tea.Quit
		case "ctrl+s":
			if m.currentMode == chatMode {
				m.currentMode = settingsMode
				m.selectedProvider = 0
				m.selectedModel = 0
				m.inModelMenu = false
			} else {
				m.currentMode = chatMode
			}
			return m, nil
		case "left":
			if m.currentMode == settingsMode && m.inModelMenu {
				// Volver al menú de proveedores
				m.inModelMenu = false
				return m, nil
			}
		case "right":
			if m.currentMode == settingsMode && !m.inModelMenu {
				// Entrar al submenú del proveedor seleccionado
				m.inModelMenu = true
				m.selectedModel = 0
				return m, nil
			}
		case "up":
			if m.currentMode == settingsMode && !m.inModelMenu {
				// Navegar entre proveedores hacia arriba
				if m.selectedProvider > 0 {
					m.selectedProvider--
				}
				return m, nil
			} else if m.currentMode == settingsMode && m.inModelMenu {
				// Navegar hacia arriba en los modelos
				if m.selectedModel > 0 {
					m.selectedModel--
				}
				return m, nil
			} else if m.currentMode == chatMode && m.historyIndex == -1 {
				// Scroll del viewport solo si no estamos en historial de input
				m.viewport.LineUp(1)
				return m, nil
			} else if m.currentMode == chatMode {
				// Navegación de historial de input
				if m.historyIndex < len(m.inputHistory)-1 && len(m.inputHistory) > 0 {
					m.historyIndex++
					m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIndex]
				}
				return m, nil
			}
		case "down":
			if m.currentMode == settingsMode && !m.inModelMenu {
				// Navegar entre proveedores hacia abajo
				if m.selectedProvider < 4 {
					m.selectedProvider++
				}
				return m, nil
			} else if m.currentMode == settingsMode && m.inModelMenu {
				// Navegar hacia abajo en los modelos
				maxModel := getModelCountForProvider(m.selectedProvider)
				if m.selectedModel < maxModel-1 {
					m.selectedModel++
				}
				return m, nil
			} else if m.currentMode == chatMode && m.historyIndex == -1 {
				// Scroll del viewport solo si no estamos en historial de input
				m.viewport.LineDown(1)
				return m, nil
			} else if m.currentMode == chatMode {
				// Navegación de historial de input
				if m.historyIndex > 0 {
					m.historyIndex--
					m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIndex]
				} else if m.historyIndex == 0 {
					m.historyIndex = -1
					m.input = ""
				}
				return m, nil
			}
		case "pgup":
			if m.currentMode == chatMode {
				m.viewport.HalfViewUp()
				return m, nil
			}
		case "pgdown":
			if m.currentMode == chatMode {
				m.viewport.HalfViewDown()
				return m, nil
			}
		case "home":
			if m.currentMode == chatMode {
				m.viewport.GotoTop()
				return m, nil
			}
		case "end":
			if m.currentMode == chatMode {
				m.viewport.GotoBottom()
				return m, nil
			}
		case "tab":
			if m.currentMode == settingsMode && !m.inModelMenu {
				// Cambiar entre proveedores
				m.selectedProvider = (m.selectedProvider + 1) % 5
				return m, nil
			}
		case "shift+tab":
			if m.currentMode == settingsMode && !m.inModelMenu {
				// Cambiar entre proveedores hacia atrás
				m.selectedProvider = (m.selectedProvider - 1 + 5) % 5
				return m, nil
			}
		case "enter":
			if m.currentMode == settingsMode && m.inModelMenu {
				// Seleccionar el modelo actual: persiste vía SaveConfig del
				// contrato (C-004) y refresca el motor in-place.
				modelID := getModelID(m.selectedProvider, m.selectedModel)
				if modelID != "" {
					config.AppConfig.Model = modelID
					if err := config.SaveConfig(); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to persist model: %v\n", err)
					}
					m.engine.RefreshClient()
					m.currentMode = chatMode
					m.inModelMenu = false
				}
				return m, nil
			}
			if m.currentMode == chatMode && m.input != "" && !m.isStreaming {
				input := m.input
				m.input = ""
				m.historyIndex = -1

				// Guardar en historial de input (limitado a 50)
				m.inputHistory = append(m.inputHistory, input)
				if len(m.inputHistory) > 50 {
					m.inputHistory = m.inputHistory[1:]
				}

				// Process slash commands
				if strings.HasPrefix(input, "/") {
					// /compact delega en el comando del contrato (T018,
					// SlashCommand/Compact): el core es el único dueño de la
					// compactación (FR-001, no duplicar lógica).
					if strings.HasPrefix(input, "/compact") {
						history, err := db.GetHistory(m.sessionID)
						response := "🗜️ Historial compactado y persistido"
						if err == nil && len(history) <= 20 {
							response = "🗜️ No se requiere compactación (historial corto)"
						} else if err == nil {
							m.engine.Send(engine.Compact{})
						}
						m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
					// /clear también es del core (T029): el motor limpia
					// memoria + persistencia y emite SessionCleared; la TUI
					// solo muestra la copia y resetea su vista con el evento.
					if strings.HasPrefix(input, "/clear") {
						m.engine.Send(engine.SlashCommand{Name: "clear"})
						m.messages = append(m.messages, api.Message{Role: "assistant", Content: handleClear()})
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
					// Control de sesiones (T031, FR-006, C-003): crear y retomar
					// sobre las funciones de sesión del contrato (db) + el
					// comando SwitchSession del motor; la misma DB que la GUI.
					if strings.HasPrefix(input, "/open ") {
						sid := strings.TrimSpace(strings.TrimPrefix(input, "/open "))
						if err := db.ResumeSession(sid); err != nil {
							response := "❌ Error al retomar sesión: " + err.Error()
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						} else {
							m.engine.Send(engine.SwitchSession{SessionID: sid})
							m.sessionID = sid
							history, err := db.GetHistory(sid)
							if err == nil {
								m.messages = toTUIMessages(history)
							}
							m.totalTokens = 0
							m.totalCost = 0
							m.inputTokens = 0
							m.outputTokens = 0
							response := fmt.Sprintf("📂 Sesión retomada: %s (%d mensajes)", sid, len(history))
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						}
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
					if strings.HasPrefix(input, "/new") {
						sid, err := db.CreateSession("Nueva sesión", "")
						if err != nil {
							response := "❌ Error al crear sesión: " + err.Error()
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						} else {
							m.engine.Send(engine.SwitchSession{SessionID: sid})
							m.sessionID = sid
							m.totalTokens = 0
							m.totalCost = 0
							m.inputTokens = 0
							m.outputTokens = 0
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: "✨ Nueva sesión creada: " + sid})
						}
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
					SetSlashCommandSessionID(m.sessionID)
					response, shouldContinue, shouldQuit := ProcessSlashCommand(input)
					if shouldQuit {
						m.engine.Stop()
						return m, tea.Quit
					}
					if shouldContinue {
						m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
				}

				// Process mentions (@file, @func, etc.)
				input = ProcessMentions(input)

				m.isStreaming = true
				m.currentResponse = ""
				m.engine.Send(engine.SendMessage{Text: input, SessionID: m.sessionID})
				return m, nil
			}
		case "backspace":
			if len(m.input) > 0 && !m.isStreaming {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 && !m.isStreaming {
				m.input += msg.String()

				// Actualizar sugerencias de autocompletado
				m.updateSuggestions()
			}
		}
	case spinner.TickMsg:
		// Solo actualizar el spinner si estamos en streaming o ejecutando tools
		if m.isStreaming || m.executingTool != "" {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Actualizar dimensiones del viewport
		headerHeight := 3
		footerHeight := 4
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		m.updateViewportContent()
		return m, nil
	}

	// Por defecto, pasar mensaje al viewport para su manejo interno
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// toTUIMessages maps the persisted session history (contract §3 GetHistory)
// to the api.Message shape the model renders (T031, FR-006).
func toTUIMessages(history []db.Message) []api.Message {
	out := make([]api.Message, 0, len(history))
	for _, h := range history {
		out = append(out, api.Message{Role: h.Role, Content: h.Content})
	}
	return out
}

// handleEngineEvent translates an engine event into the model's view state.
// The engine already persists every message; the TUI only renders.
func (m model) handleEngineEvent(ev engine.Event) model {	switch e := ev.(type) {
	case engine.StreamStart:
		// El motor anuncia el inicio del stream (contrato §1 `stream:start`).
		// El modelo reacciona al evento en lugar de depender solo del handler
		// de Enter: cualquier camino que inicie un turno (retry, resume)
		// refleja el estado de streaming correctamente.
		m.isStreaming = true
		m.currentResponse = ""
		m.executingTool = ""
	case engine.UserMessageAppended:
		m.messages = append(m.messages, e.Message)
		m.updateViewportContent()
		m.viewport.GotoBottom()
	case engine.StreamDelta:
		m.currentResponse += e.Text
		m.updateViewportContent()
		m.viewport.GotoBottom()
	case engine.StreamDone:
		if m.currentResponse != "" {
			m.messages = append(m.messages, api.Message{Role: "assistant", Content: m.currentResponse})
		}
		m.currentResponse = ""
		m.isStreaming = false
		m.updateViewportContent()
		m.viewport.GotoBottom()
	case engine.StreamCancelled:
		m.isStreaming = false
		m.executingTool = ""
		m.currentResponse = ""
		m.updateViewportContent()
	case engine.ToolRequested:
		m.executingTool = e.Name
		m.updateViewportContent()
	case engine.ToolExecuting:
		m.executingTool = e.ToolCallID
		m.updateViewportContent()
	case engine.ToolResult:
		m.executingTool = ""
		resultBlock := api.ContentBlock{
			Type: "tool_result",
			ToolResult: &api.ToolResult{
				ToolUseID: e.ToolCallID,
				ToolName:  e.Name,
				Content:   e.Content,
				IsError:   e.IsError,
			},
		}
		m.messages = append(m.messages, api.Message{Role: "user", Content: []api.ContentBlock{resultBlock}})
		m.updateViewportContent()
		m.viewport.GotoBottom()
	case engine.ToolRejected:
		m.executingTool = ""
	case engine.ToolTimedOut:
		m.executingTool = ""
	case engine.UsageUpdate:
		m.inputTokens += e.InputTokens
		m.outputTokens += e.OutputTokens
		m.totalTokens += e.InputTokens + e.OutputTokens

		// Track cost
		costTracker := tools.GetCostTracker()
		provider := "anthropic"
		if !strings.Contains(config.AppConfig.BaseURL, "anthropic") {
			if strings.Contains(config.AppConfig.BaseURL, "groq") {
				provider = "groq"
			} else {
				provider = "openai"
			}
		}
		if err := costTracker.RecordUsage(provider, config.AppConfig.Model, e.InputTokens, e.OutputTokens); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to record usage: %v\n", err)
		}
		m.totalCost = costTracker.GetSessionCost()
	case engine.ErrorEvent:
		m.isStreaming = false
		m.executingTool = ""
		m.currentResponse = ""
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		var errorContent strings.Builder
		errorContent.WriteString(errorStyle.Render("⚠️  Error") + "\n\n")
		errorContent.WriteString(fmt.Sprintf("```\n%s\n```", e.Message))
		errorContent.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
			"Tip: You can try:\n"+
				"  • Checking your API key with /settings\n"+
				"  • Verifying your internet connection\n"+
				"  • Trying a different model with Ctrl+S"))
		m.messages = append(m.messages, api.Message{Role: "assistant", Content: errorContent.String()})
		m.updateViewportContent()
	case engine.SessionCleared:
		// El core ya limpió memoria + persistencia (T029); la vista se resetea
		// para reflejar el estado compartido sin tocar la DB.
		m.messages = []api.Message{}
		m.totalTokens = 0
		m.totalCost = 0
		m.inputTokens = 0
		m.outputTokens = 0
		m.updateViewportContent()
	case engine.Idle:
		m.isStreaming = false
		m.executingTool = ""
		m.updateViewportContent()
	}
	return m
}

func (m model) View() string {
	if m.currentMode == settingsMode {
		return m.renderSettingsView()
	}
	return m.renderChatView()
}

// getModelCountForProvider retorna la cantidad de modelos para un proveedor
func getModelCountForProvider(providerIdx int) int {
	counts := []int{2, 2, 4, 3, 2} // Anthropic:2, OpenAI:2, Groq:4, OpenRouter:3, Ollama:2
	if providerIdx >= 0 && providerIdx < len(counts) {
		return counts[providerIdx]
	}
	return 0
}

// getModelID retorna el ID del modelo según proveedor e índice
func getModelID(providerIdx, modelIdx int) string {
	models := [][]string{
		{"claude-sonnet-4-20250514", "claude-3-opus-20240229"}, // Anthropic
		{"gpt-4o", "gpt-4o-mini"},                              // OpenAI
		{"llama-3.3-70b-versatile", "llama-3.1-8b-instant", "openai/gpt-oss-120b", "openai/gpt-oss-20b"}, // Groq
		{"anthropic/claude-sonnet-4", "openai/gpt-4o", "meta-llama/llama-3.3-70b-instruct"},              // OpenRouter
		{"llama3.2", "codellama"}, // Ollama
	}
	if providerIdx >= 0 && providerIdx < len(models) {
		if modelIdx >= 0 && modelIdx < len(models[providerIdx]) {
			return models[providerIdx][modelIdx]
		}
	}
	return ""
}

// getModelName retorna el nombre legible del modelo
func getModelName(providerIdx, modelIdx int) string {
	names := [][]string{
		{"Claude 4 Sonnet", "Claude 3 Opus"},
		{"GPT-4o", "GPT-4o Mini"},
		{"Llama 3.3 70B", "Llama 3.1 8B", "GPT-OSS 120B", "GPT-OSS 20B"},
		{"Claude 4 Sonnet (OR)", "GPT-4o (OR)", "Llama 3.3 70B (OR)"},
		{"Llama 3.2", "CodeLlama"},
	}
	if providerIdx >= 0 && providerIdx < len(names) {
		if modelIdx >= 0 && modelIdx < len(names[providerIdx]) {
			return names[providerIdx][modelIdx]
		}
	}
	return ""
}

// getModelDesc retorna la descripción del modelo
func getModelDesc(providerIdx, modelIdx int) string {
	descs := [][]string{
		{"Latest Claude - Best coding", "Most powerful reasoning"},
		{"Latest multimodal", "Fast & cheap"},
		{"Ultra-fast versatile", "Lightning fast", "OpenAI 120B on Groq", "OpenAI 20B on Groq"},
		{"Via OpenRouter", "Via OpenRouter", "Via OpenRouter"},
		{"Run locally", "Code specialized"},
	}
	if providerIdx >= 0 && providerIdx < len(descs) {
		if modelIdx >= 0 && modelIdx < len(descs[providerIdx]) {
			return descs[providerIdx][modelIdx]
		}
	}
	return ""
}

func (m model) renderSettingsView() string {
	// Estilos
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Padding(1, 0)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("62")).Padding(0, 2)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)

	// Colores por proveedor
	providerColors := []string{"141", "45", "208", "214", "120"} // Morado, Azul, Naranja, Amarillo, Verde
	providerNames := []string{"Anthropic", "OpenAI", "Groq", "OpenRouter", "Ollama"}
	providerIcons := []string{"◉", "◉", "◉", "◉", "◉"}

	var s strings.Builder
	s.WriteString(titleStyle.Render("🤖 Select AI Provider & Model") + "\n")
	s.WriteString(headerStyle.Render(" Current: "+config.AppConfig.Model+" ") + "\n\n")

	if !m.inModelMenu {
		// Menú principal - selección de proveedor
		s.WriteString(sectionStyle.Render("Select Provider (↑↓ to navigate, → to enter)") + "\n\n")

		for i := 0; i < 5; i++ {
			isSelected := i == m.selectedProvider
			color := lipgloss.Color(providerColors[i])

			if isSelected {
				// Proveedor seleccionado - estilo destacado
				boxStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(color).
					Background(lipgloss.Color("237")).
					Padding(1, 2).
					Width(20)

				title := lipgloss.NewStyle().Bold(true).Foreground(color).Render(providerIcons[i] + " " + providerNames[i])
				count := getModelCountForProvider(i)
				subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%d models", count))

				s.WriteString(boxStyle.Render(title+"\n"+subtitle) + "\n\n")
			} else {
				// Proveedor no seleccionado - estilo simple
				title := lipgloss.NewStyle().Foreground(color).Render("  " + providerIcons[i] + " " + providerNames[i])
				s.WriteString(title + "\n")
			}
		}

		s.WriteString("\n" + hintStyle.Render("Press → to enter provider menu") + "\n")
	} else {
		// Submenú - selección de modelo dentro del proveedor
		providerColor := lipgloss.Color(providerColors[m.selectedProvider])
		providerStyle := lipgloss.NewStyle().Bold(true).Foreground(providerColor)

		s.WriteString(sectionStyle.Render("Select Model (↑↓ to navigate, Enter to select)") + "\n\n")
		s.WriteString(providerStyle.Render("▸ "+providerNames[m.selectedProvider]) + "\n\n")

		modelCount := getModelCountForProvider(m.selectedProvider)
		for i := 0; i < modelCount; i++ {
			isSelected := i == m.selectedModel
			modelName := getModelName(m.selectedProvider, i)
			modelDesc := getModelDesc(m.selectedProvider, i)

			if isSelected {
				// Modelo seleccionado
				selectedStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("86")).
					Background(lipgloss.Color("237")).
					Padding(0, 1)
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
				s.WriteString("  " + selectedStyle.Render("▸ "+modelName) + " " + descStyle.Render(modelDesc) + "\n")
			} else {
				// Modelo no seleccionado
				modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
				s.WriteString("    " + modelStyle.Render(modelName) + " " + descStyle.Render(modelDesc) + "\n")
			}
		}

		s.WriteString("\n" + hintStyle.Render("Press ← to go back to providers") + "\n")
	}

	s.WriteString(hintStyle.Render("Press Ctrl+S to return to chat"))
	return s.String()
}

func (m model) renderChatView() string {
	var s strings.Builder

	// Header con estilo mejorado
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	shortID := m.sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	header := titleStyle.Render("LetsGO Code") + " " +
		versionStyle.Render("v0.1.0") + " | " +
		sessionStyle.Render("Session: "+shortID)

	// Línea separadora
	separatorWidth := m.width - 2
	if separatorWidth < 0 {
		separatorWidth = 0
	}
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).
		Render(strings.Repeat("─", separatorWidth))

	s.WriteString(header + "\n" + separator + "\n")

	// Usar viewport para el contenido scrolleable
	s.WriteString(m.viewport.View())

	// Indicador de streaming
	if m.isStreaming || m.executingTool != "" {
		modelName := config.AppConfig.Model
		if len(modelName) > 15 {
			modelName = modelName[:12] + "..."
		}

		thinkingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

		indicator := fmt.Sprintf("%s %s ",
			spinnerStyle.Render(m.spinner.View()),
			thinkingStyle.Render("Let'sGo("+modelName+")"))

		roleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
		s.WriteString("\n" + indicator + roleStyle.Render("thinking...") + "\n")
		if m.currentResponse != "" {
			s.WriteString(roleStyle.Render("Let'sGo("+modelName+"): ") + m.currentResponse + "█\n")
		}
	}

	// Indicador de ejecución de herramientas
	if m.executingTool != "" {
		executingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Italic(true)

		s.WriteString(fmt.Sprintf("%s %s %s %s\n",
			spinnerStyle.Render(m.spinner.View()),
			executingStyle.Render("▶"),
			executingStyle.Render("Executing:"),
			toolStyle.Render(m.executingTool)))
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("62")).Padding(0, 1)
	tokenStr := fmt.Sprintf(" Tokens: %d ", m.totalTokens)
	modelStr := fmt.Sprintf(" Model: %s ", config.AppConfig.Model)
	costStr := fmt.Sprintf(" Cost: $%.4f ", m.totalCost)
	statusLine := statusStyle.Render(modelStr + " | " + tokenStr + " | " + costStr)

	s.WriteString("\n" + statusLine + "\n")

	// Sugerencias de autocompletado
	if m.showSuggestions && len(m.suggestions) > 0 {
		s.WriteString(m.renderSuggestions())
	}

	s.WriteString("> " + m.input + "█")
	return s.String()
}

var program *tea.Program

func RunChat() error {
	m := NewModel()
	// El loop del motor se arranca aquí (T028): sin Start() los comandos del
	// contrato se quedan en cola y el chat nunca procesa mensajes.
	m.engine.Start()
	program = tea.NewProgram(m, tea.WithAltScreen())

	// Pump engine events into the bubbletea program.
	go func() {
		for ev := range m.engine.Events() {
			if program != nil {
				program.Send(engineEventMsg{ev: ev})
			}
		}
	}()

	_, err := program.Run()
	m.engine.Stop()
	program = nil
	return err
}

// updateSuggestions actualiza las sugerencias basadas en el input actual
func (m *model) updateSuggestions() {
	m.suggestions = []string{}
	m.showSuggestions = false

	if strings.HasPrefix(m.input, "/") {
		// Sugerencias para comandos slash
		for _, cmd := range AvailableSlashCommands {
			if strings.HasPrefix(cmd.Name, m.input) {
				m.suggestions = append(m.suggestions, cmd.Name+" - "+cmd.Description)
			}
		}
		m.showSuggestions = len(m.suggestions) > 0 && m.input != "/"
	} else if strings.Contains(m.input, "@") {
		// Extraer la parte después del último @
		parts := strings.Split(m.input, "@")
		if len(parts) > 0 {
			prefix := parts[len(parts)-1]
			if prefix != "" {
				m.suggestions = GetMentionSuggestions(prefix)
				m.showSuggestions = len(m.suggestions) > 0
			}
		}
	}

	// Limitar sugerencias a 5
	if len(m.suggestions) > 5 {
		m.suggestions = m.suggestions[:5]
	}
}

// renderSuggestions renderiza las sugerencias de autocompletado
func (m model) renderSuggestions() string {
	if len(m.suggestions) == 0 {
		return ""
	}

	var s strings.Builder
	suggestionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	s.WriteString("\n")
	for i, sug := range m.suggestions {
		if i == 0 {
			s.WriteString(highlightStyle.Render("  → "+sug) + "\n")
		} else {
			s.WriteString(suggestionStyle.Render("    "+sug) + "\n")
		}
	}

	return s.String()
}

// getWelcomeMessage retorna el mensaje de bienvenida
func getWelcomeMessage() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	b.WriteString(titleStyle.Render("👋 Welcome to LetsGO Code!") + "\n\n")
	b.WriteString(subtitleStyle.Render("Your AI-powered coding assistant") + "\n\n")

	b.WriteString(hintStyle.Render("Quick Tips:") + "\n")
	b.WriteString(hintStyle.Render("  • Type /help for available commands") + "\n")
	b.WriteString(hintStyle.Render("  • Use @filename to reference files") + "\n")
	b.WriteString(hintStyle.Render("  • Press Ctrl+S to change models") + "\n")
	b.WriteString(hintStyle.Render("  • Use ↑/↓ to navigate message history") + "\n\n")

	b.WriteString(hintStyle.Render("Start typing to chat with the AI!") + "\n")

	return b.String()
}

// updateViewportContent actualiza el contenido del viewport con los mensajes
func (m *model) updateViewportContent() {
	var content strings.Builder

	for _, msg := range m.messages {
		role := "User"
		roleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
		if msg.Role == "assistant" {
			// Extraer nombre corto del modelo
			modelName := config.AppConfig.Model
			if len(modelName) > 15 {
				modelName = modelName[:12] + "..."
			}
			role = "Let'sGo(" + modelName + ")"
			roleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
		}

		msgContent := ""
		switch c := msg.Content.(type) {
		case string:
			if msg.Role == "assistant" {
				rendered, err := m.renderer.Render(c)
				if err != nil {
					msgContent = c
				} else {
					msgContent = rendered
				}
			} else {
				msgContent = c
			}
		case []api.ContentBlock:
			for _, b := range c {
				if b.Type == "tool_result" && b.ToolResult != nil {
					toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
					toolID := b.ToolResult.ToolUseID
					if len(toolID) > 4 {
						toolID = toolID[:4]
					}
					statusIcon := "✅"
					statusColor := "42"
					if b.ToolResult.IsError {
						statusIcon = "❌"
						statusColor = "196"
					}
					statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
					msgContent += toolStyle.Render(fmt.Sprintf("%s Tool Result [%s]:", statusStyle.Render(statusIcon), toolID)) + "\n"

					preview := b.ToolResult.Content
					if len(preview) > 500 {
						preview = preview[:500] + "..."
					}
					msgContent += lipgloss.NewStyle().Faint(true).Render(preview) + "\n"
				}
			}
		default:
			msgContent = fmt.Sprintf("%v", c)
		}

		content.WriteString(roleStyle.Render(role+": ") + msgContent + "\n\n")
	}

	// Si hay contenido en streaming, agregarlo al final
	if (m.isStreaming || m.executingTool != "") && m.currentResponse != "" {
		modelName := config.AppConfig.Model
		if len(modelName) > 15 {
			modelName = modelName[:12] + "..."
		}
		role := "Let'sGo(" + modelName + ")"
		roleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

		// Renderizar el contenido parcial si es posible
		rendered, err := m.renderer.Render(m.currentResponse)
		if err != nil {
			rendered = m.currentResponse
		}
		content.WriteString(roleStyle.Render(role+": ") + rendered + "\n")
	}

	m.viewport.SetContent(content.String())
}