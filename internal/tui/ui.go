package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/user/go-claude-code/internal/api"
	"github.com/user/go-claude-code/internal/config"
	"github.com/user/go-claude-code/internal/db"
	"github.com/user/go-claude-code/internal/tools"
)

type deltaMsg string
type toolUseMsg api.ToolUse
type toolResultMsg api.ToolResult
type toolInputDeltaMsg struct {
	id    string
	delta string
}
type usageMsg struct {
	input  int
	output int
}
type errorMsg error
type finishMsg struct{}

type mode int

const (
	chatMode mode = iota
	settingsMode
)

type model struct {
	sessionID       string
	messages        []api.Message
	input           string
	isStreaming     bool
	currentResponse string
	toolInputs      map[string]string
	pendingToolUses []api.ToolUse
	executingTool   string
	currentMode     mode
	spinner         spinner.Model
	renderer        *glamour.TermRenderer
	width           int
	height          int
	viewport        viewport.Model // Usar viewport para scroll profesional
	client          *api.Client
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

	sessionID := uuid.New().String()
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
		sessionID:        sessionID,
		messages:         []api.Message{},
		client:           createClientForCurrentModel(),
		toolInputs:       make(map[string]string),
		spinner:          s,
		renderer:         r,
		currentMode:      chatMode,
		viewport:         vp,
		inputHistory:     []string{},
		historyIndex:     -1,
		suggestions:      []string{},
		showSuggestions:  false,
		selectedProvider: 0,
		selectedModel:    0,
		inModelMenu:      false,
	}

	// PARITY: Initial Git Warning
	if tools.IsGitRepo() {
		status := tools.GetGitStatus()
		if status != "" {
			m.messages = append(m.messages, api.Message{
				Role:    "assistant",
				Content: "> ℹ️ **Note:** You have uncommitted changes in this repository. Claude will see these changes when reading files.",
			})
		}
	}

	// Welcome message con tips útiles
	m.messages = append(m.messages, api.Message{
		Role:    "assistant",
		Content: getWelcomeMessage(),
	})

	return m
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case usageMsg:
		m.totalTokens += msg.input + msg.output
		m.inputTokens += msg.input
		m.outputTokens += msg.output

		// Debug logging
		fmt.Fprintf(os.Stderr, "[TUI Debug] usageMsg received: input=%d, output=%d, isStreaming=%v\n",
			msg.input, msg.output, m.isStreaming)

		// Si input=0 y output=0, es una señal de finalización del stream
		if msg.input == 0 && msg.output == 0 && m.isStreaming {
			fmt.Fprintf(os.Stderr, "[TUI Debug] Resetting isStreaming to false\n")
			m.isStreaming = false
			if m.currentResponse != "" {
				m.messages = append(m.messages, api.Message{Role: "assistant", Content: m.currentResponse})
				db.SaveMessage(m.sessionID, "assistant", m.currentResponse)
				m.currentResponse = ""
			}
		}

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
		if err := costTracker.RecordUsage(provider, config.AppConfig.Model, msg.input, msg.output); err != nil {
			// Log error but don't interrupt user experience
			fmt.Fprintf(os.Stderr, "Warning: failed to record usage: %v\n", err)
		}
		m.totalCost = costTracker.GetSessionCost()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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
				// Seleccionar el modelo actual
				modelID := getModelID(m.selectedProvider, m.selectedModel)
				if modelID != "" {
					config.AppConfig.Model = modelID
					m.client = createClientForCurrentModel()
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
					SetSlashCommandSessionID(m.sessionID)
					response, shouldContinue, shouldQuit := ProcessSlashCommand(input)
					if shouldQuit {
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

				m.messages = append(m.messages, api.Message{Role: "user", Content: input})
				db.SaveMessage(m.sessionID, "user", input)
				m.isStreaming = true
				m.currentResponse = ""
				m.updateViewportContent()
				m.viewport.GotoBottom()
				return m, m.sendMessageCmd()
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
	case deltaMsg:
		m.currentResponse += string(msg)
		// Actualizar viewport durante streaming para mostrar contenido en tiempo real
		m.updateViewportContent()
		m.viewport.GotoBottom()
	case toolInputDeltaMsg:
		m.toolInputs[msg.id] += msg.delta
	case toolUseMsg:
		tu := api.ToolUse(msg)
		m.pendingToolUses = append(m.pendingToolUses, tu)
		// Almacenar el nombre de la herramienta asociado al ID para recuperarlo en toolResultMsg
		if m.toolInputs == nil {
			m.toolInputs = make(map[string]string)
		}
		// Usamos un prefijo especial o un mapa dedicado para nombres de herramientas
		m.toolInputs["_name_"+tu.ID] = tu.Name
	case toolResultMsg:
		res := api.ToolResult(msg)
		m.executingTool = ""

		// Recuperar el nombre de la herramienta usando el ID
		if name, ok := m.toolInputs["_name_"+res.ToolUseID]; ok {
			res.ToolName = name
			delete(m.toolInputs, "_name_"+res.ToolUseID)
		}

		content := []api.ContentBlock{{Type: "tool_result", ToolResult: &res}}
		m.messages = append(m.messages, api.Message{Role: "user", Content: content})
		db.SaveMessage(m.sessionID, "user", content)

		if len(m.pendingToolUses) == 0 {
			m.isStreaming = true
			return m, m.sendMessageCmd()
		}
		return m, nil
	case errorMsg:
		m.isStreaming = false
		m.executingTool = ""
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		var errorContent strings.Builder
		errorContent.WriteString(errorStyle.Render("⚠️  Error") + "\n\n")
		errorContent.WriteString(fmt.Sprintf("```\n%s\n```", msg.Error()))
		errorContent.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
			"Tip: You can try:\n"+
				"  • Checking your API key with /settings\n"+
				"  • Verifying your internet connection\n"+
				"  • Trying a different model with Ctrl+S"))
		m.messages = append(m.messages, api.Message{Role: "assistant", Content: errorContent.String()})
	case finishMsg:
		m.isStreaming = false
		if m.currentResponse != "" {
			m.messages = append(m.messages, api.Message{Role: "assistant", Content: m.currentResponse})
			db.SaveMessage(m.sessionID, "assistant", m.currentResponse)
			m.currentResponse = ""
		}

		if len(m.pendingToolUses) > 0 {
			var cmds []tea.Cmd
			m.executingTool = fmt.Sprintf("%d tools", len(m.pendingToolUses))
			for _, tu := range m.pendingToolUses {
				cmds = append(cmds, m.executeToolCmd(tu))
			}
			m.pendingToolUses = []api.ToolUse{}
			return m, tea.Batch(cmds...)
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
	if m.isStreaming {
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
		s.WriteString(roleStyle.Render("Let'sGo("+modelName+"): ") + m.currentResponse + "█\n")
	}

	// Indicador de ejecución de herramientas
	if m.executingTool != "" {
		executingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true)

		spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Italic(true)

		s.WriteString(fmt.Sprintf("\n%s %s %s %s\n",
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

func (m model) executeToolCmd(tu api.ToolUse) tea.Cmd {
	return func() tea.Msg {
		var input map[string]interface{}
		jsonStr := m.toolInputs[tu.ID]
		err := json.Unmarshal([]byte(jsonStr), &input)
		if err != nil {
			return toolResultMsg{ToolUseID: tu.ID, Content: fmt.Sprintf("Error parsing tool input: %v", err), IsError: true}
		}
		result, err := tools.ExecuteTool(tu.Name, input)
		isError := false
		if err != nil {
			result = err.Error()
			isError = true
		}
		return toolResultMsg{ToolUseID: tu.ID, Content: result, IsError: isError}
	}
}

func (m model) sendMessageCmd() tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		systemPrompt := api.GetSystemPrompt(config.AppConfig.Model, cwd)
		prunedMessages := db.PruneContext(m.messages, 10)
		req := api.Request{
			Model:     config.AppConfig.Model,
			System:    systemPrompt,
			MaxTokens: 4096,
			Stream:    true,
			Messages:  prunedMessages,
			Tools:     tools.GetToolDefinitions(),
		}
		err = m.client.StreamRequest(req, func(delta string) {
			if program != nil {
				program.Send(deltaMsg(delta))
			}
		}, func(tu api.ToolUse) {
			if program != nil {
				program.Send(toolUseMsg(tu))
			}
		}, func(id string, delta string) {
			if program != nil {
				program.Send(toolInputDeltaMsg{id: id, delta: delta})
			}
		}, func(in, out int) {
			if program != nil {
				program.Send(usageMsg{input: in, output: out})
			}
		})
		if err != nil {
			return errorMsg(err)
		}
		return finishMsg{}
	}
}

func RunChat() error {
	m := NewModel()
	program = tea.NewProgram(m, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// createClientForCurrentModel crea un cliente API con la API key y URL correctas según el proveedor del modelo actual
func createClientForCurrentModel() *api.Client {
	provider := config.DetectProviderFromModel(config.AppConfig.Model)
	apiKey := config.GetAPIKeyForProvider(provider)

	// Seleccionar la URL base correcta según el proveedor
	var baseURL string
	switch provider {
	case "anthropic":
		baseURL = "https://api.anthropic.com/v1/messages"
	case "openai":
		baseURL = "https://api.openai.com/v1/chat/completions"
	case "groq":
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
	case "openrouter":
		baseURL = "https://openrouter.ai/api/v1/chat/completions"
	case "ollama":
		baseURL = "http://localhost:11434/api/chat"
	default:
		baseURL = config.AppConfig.BaseURL
	}

	return api.NewClient(apiKey, baseURL)
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
	if m.isStreaming && m.currentResponse != "" {
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
