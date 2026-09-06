package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
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

type pendingApproval struct {
	toolName string
	input    map[string]interface{}
	inputStr string
	respCh   chan bool
}

type tuiApprovalMsg struct {
	req pendingApproval
}

// tuiDriver implements engine.PermissionDriver for interactive confirmation
type tuiDriver struct {
	approvalCh chan pendingApproval
}

func (d tuiDriver) Prompt(toolName string, input any) (bool, error) {
	cat := engine.ToolCategory(toolName)
	if config.AppConfig.AutoApprove[cat] || d.approvalCh == nil {
		return true, nil
	}
	respCh := make(chan bool, 1)
	mInput, _ := input.(map[string]interface{})
	d.approvalCh <- pendingApproval{
		toolName: toolName,
		input:    mInput,
		inputStr: fmt.Sprintf("%v", input),
		respCh:   respCh,
	}
	select {
	case allow := <-respCh:
		return allow, nil
	case <-time.After(engine.ApprovalTimeout):
		return false, engine.ErrApprovalTimedOut
	}
}

func (d tuiDriver) AutoApprove(category string) bool {
	return config.AppConfig.AutoApprove[category]
}

func (d tuiDriver) SessionGranted(sessionID, category string) bool {
	allowed, err := db.IsGranted(sessionID, category)
	return err == nil && allowed
}

type mode int

const (
	chatMode mode = iota
	settingsMode
	gitMode
	activityMode
	sessionsMode
)

// AutocompleteItem represents a suggested command or mention in TUI.
type AutocompleteItem struct {
	Type        string // "cmd" | "mention"
	Icon        string
	Title       string
	Description string
	Value       string
}

// AutocompleteState tracks popup suggestions.
type AutocompleteState struct {
	Active        bool
	Trigger       string // "/" or "@"
	Query         string
	SelectedIndex int
	Items         []AutocompleteItem
}

type model struct {
	engine           *engine.Engine
	driver           tuiDriver
	sessionID        string
	messages         []api.Message
	input            string
	textarea         textarea.Model
	agentMode        string // "build" (default) or "plan"
	pendingApproval  *pendingApproval
	showDiff         bool
	isStreaming      bool
	currentResponse  string
	executingTool    string
	currentMode      mode
	spinner          spinner.Model
	renderer         *glamour.TermRenderer
	width            int
	height           int
	viewport         viewport.Model
	totalTokens      int
	totalCost        float64
	inputTokens      int
	outputTokens     int
	inputHistory     []string
	historyIndex     int
	suggestions      []string
	showSuggestions  bool
	autocomplete     AutocompleteState
	activityHistory  []TUIActivityItem
	activitySelected int
	gitSummary       TUIGitSummary
	gitSelectedIdx   int
	gitCommitPrompt  bool
	gitCommitInput   string
	savedSessions    []db.Session
	sessionSelected  int
	selectedProvider int
	selectedModel    int
	inModelMenu      bool
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

	vp := viewport.New(80, 20)
	vp.SetContent("")

	ta := textarea.New()
	ta.Placeholder = "Ask a question, enter @file, or /command... (Alt+Enter for newline, Tab to toggle mode)"
	ta.Focus()
	ta.Prompt = "│ "
	ta.CharLimit = 8000
	ta.SetWidth(80)
	ta.SetHeight(2)
	ta.ShowLineNumbers = false

	apprCh := make(chan pendingApproval, 8)
	driver := tuiDriver{approvalCh: apprCh}

	gitSummary := TUIGitSummary{Clean: true}
	if tools.IsGitRepo() {
		if sum, err := GetTUIGitSummary(); err == nil {
			gitSummary = sum
		}
	}

	m := model{
		messages:         []api.Message{},
		spinner:          s,
		renderer:         r,
		currentMode:      chatMode,
		agentMode:        "build",
		textarea:         ta,
		viewport:         vp,
		driver:           driver,
		inputHistory:     []string{},
		historyIndex:     -1,
		suggestions:      []string{},
		showSuggestions:  false,
		activityHistory:  []TUIActivityItem{},
		gitSummary:       gitSummary,
		selectedProvider: 0,
		selectedModel:    0,
		inModelMenu:      false,
	}

	m.engine = engine.New(driver)

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
	case tuiApprovalMsg:
		m.pendingApproval = &msg.req
		m.showDiff = false
		return m, nil
	case engineEventMsg:
		return m.handleEngineEvent(msg.ev), nil
	case tea.KeyMsg:
		// Intercept keys if a tool permission is pending approval
		if m.pendingApproval != nil {
			switch msg.String() {
			case "y", "Y":
				m.pendingApproval.respCh <- true
				m.pendingApproval = nil
				m.showDiff = false
				return m, nil
			case "n", "N", "esc":
				m.pendingApproval.respCh <- false
				m.pendingApproval = nil
				m.showDiff = false
				return m, nil
			case "c", "C", "a", "A":
				cat := engine.ToolCategory(m.pendingApproval.toolName)
				m.engine.Send(engine.GrantSession{SessionID: m.sessionID, Category: cat, Allow: true})
				m.pendingApproval.respCh <- true
				m.pendingApproval = nil
				m.showDiff = false
				return m, nil
			case "p", "P":
				cat := engine.ToolCategory(m.pendingApproval.toolName)
				cwd, _ := os.Getwd()
				_ = engine.GrantProject(cwd, cat)
				m.engine.Send(engine.GrantSession{SessionID: m.sessionID, Category: cat, Allow: true})
				m.pendingApproval.respCh <- true
				m.pendingApproval = nil
				m.showDiff = false
				return m, nil
			case "d", "D":
				m.showDiff = !m.showDiff
				return m, nil
			}
			return m, nil
		}

		if m.currentMode == gitMode {
			switch msg.String() {
			case "ctrl+c":
				m.currentMode = chatMode
				return m, nil
			case "esc", "ctrl+d":
				if m.gitCommitPrompt {
					m.gitCommitPrompt = false
					m.gitCommitInput = ""
				} else {
					m.currentMode = chatMode
				}
				return m, nil
			case "up":
				if !m.gitCommitPrompt && m.gitSelectedIdx > 0 {
					m.gitSelectedIdx--
					if m.gitSelectedIdx < len(m.gitSummary.Files) {
						m.gitSummary.SelectedFile = m.gitSummary.Files[m.gitSelectedIdx].Path
						m.gitSummary.DiffContent = GetFileDiff(m.gitSummary.SelectedFile)
					}
				}
				return m, nil
			case "down":
				if !m.gitCommitPrompt && m.gitSelectedIdx < len(m.gitSummary.Files)-1 {
					m.gitSelectedIdx++
					if m.gitSelectedIdx < len(m.gitSummary.Files) {
						m.gitSummary.SelectedFile = m.gitSummary.Files[m.gitSelectedIdx].Path
						m.gitSummary.DiffContent = GetFileDiff(m.gitSummary.SelectedFile)
					}
				}
				return m, nil
			case "s", "+":
				if !m.gitCommitPrompt && m.gitSelectedIdx >= 0 && m.gitSelectedIdx < len(m.gitSummary.Files) {
					_ = StageFile(m.gitSummary.Files[m.gitSelectedIdx].Path)
					m.gitSummary, _ = GetTUIGitSummary()
				}
				return m, nil
			case "u", "-":
				if !m.gitCommitPrompt && m.gitSelectedIdx >= 0 && m.gitSelectedIdx < len(m.gitSummary.Files) {
					_ = UnstageFile(m.gitSummary.Files[m.gitSelectedIdx].Path)
					m.gitSummary, _ = GetTUIGitSummary()
				}
				return m, nil
			case "a":
				if !m.gitCommitPrompt {
					_ = StageAll()
					m.gitSummary, _ = GetTUIGitSummary()
				}
				return m, nil
			case "c":
				if !m.gitCommitPrompt {
					m.gitCommitPrompt = true
					m.gitCommitInput = ""
				}
				return m, nil
			case "enter":
				if m.gitCommitPrompt {
					if strings.TrimSpace(m.gitCommitInput) != "" {
						_ = CommitChanges(strings.TrimSpace(m.gitCommitInput))
						m.messages = append(m.messages, api.Message{
							Role:    "assistant",
							Content: fmt.Sprintf("✓ Cambios confirmados con commit: %q", m.gitCommitInput),
						})
						m.updateViewportContent()
						m.gitSummary, _ = GetTUIGitSummary()
					}
					m.gitCommitPrompt = false
					m.gitCommitInput = ""
				}
				return m, nil
			case "backspace":
				if m.gitCommitPrompt && len(m.gitCommitInput) > 0 {
					m.gitCommitInput = m.gitCommitInput[:len(m.gitCommitInput)-1]
				}
				return m, nil
			default:
				if m.gitCommitPrompt && len(msg.Runes) > 0 {
					m.gitCommitInput += string(msg.Runes)
				}
				return m, nil
			}
		}

		if m.currentMode == activityMode {
			switch msg.String() {
			case "esc", "ctrl+b", "ctrl+c":
				m.currentMode = chatMode
				return m, nil
			case "up":
				if m.activitySelected > 0 {
					m.activitySelected--
				}
				return m, nil
			case "down":
				if m.activitySelected < len(m.activityHistory)-1 {
					m.activitySelected++
				}
				return m, nil
			}
			return m, nil
		}

		if m.currentMode == sessionsMode {
			switch msg.String() {
			case "esc", "ctrl+c":
				m.currentMode = chatMode
				return m, nil
			case "up":
				if m.sessionSelected > 0 {
					m.sessionSelected--
				}
				return m, nil
			case "down":
				if m.sessionSelected < len(m.savedSessions)-1 {
					m.sessionSelected++
				}
				return m, nil
			case "enter":
				if m.sessionSelected >= 0 && m.sessionSelected < len(m.savedSessions) {
					sid := m.savedSessions[m.sessionSelected].ID
					_ = db.ResumeSession(sid)
					m.sessionID = sid
					m.engine.SwitchSession(sid)
					history, _ := db.GetHistory(sid)
					m.messages = toTUIMessages(history)
					m.updateViewportContent()
					m.currentMode = chatMode
				}
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			if m.isStreaming {
				m.engine.Send(engine.Cancel{})
				return m, nil
			}
			if m.textarea.Value() != "" {
				m.textarea.Reset()
				m.input = ""
				m.autocomplete.Active = false
				m.autocomplete.Items = nil
				return m, nil
			}
			m.engine.Stop()
			return m, tea.Quit
		case "esc":
			if m.autocomplete.Active {
				m.autocomplete.Active = false
				m.autocomplete.Items = nil
				return m, nil
			}
			if m.currentMode != chatMode {
				m.currentMode = chatMode
				return m, nil
			}
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
		case "ctrl+d":
			if m.currentMode == gitMode {
				m.currentMode = chatMode
			} else {
				m.currentMode = gitMode
				if sum, err := GetTUIGitSummary(); err == nil {
					m.gitSummary = sum
				}
				m.gitSelectedIdx = 0
				m.gitCommitPrompt = false
			}
			return m, nil
		case "ctrl+b":
			if m.currentMode == activityMode {
				m.currentMode = chatMode
			} else {
				m.currentMode = activityMode
				if len(m.activityHistory) > 0 {
					m.activitySelected = len(m.activityHistory) - 1
				}
			}
			return m, nil
		case "left":
			if m.currentMode == settingsMode && m.inModelMenu {
				m.inModelMenu = false
				return m, nil
			}
		case "right":
			if m.currentMode == settingsMode && !m.inModelMenu {
				m.inModelMenu = true
				m.selectedModel = 0
				return m, nil
			}
		case "up":
			if m.autocomplete.Active {
				if m.autocomplete.SelectedIndex > 0 {
					m.autocomplete.SelectedIndex--
				}
				return m, nil
			}
			if m.currentMode == settingsMode && !m.inModelMenu {
				if m.selectedProvider > 0 {
					m.selectedProvider--
				}
				return m, nil
			} else if m.currentMode == settingsMode && m.inModelMenu {
				if m.selectedModel > 0 {
					m.selectedModel--
				}
				return m, nil
			} else if m.currentMode == chatMode && m.historyIndex == -1 && m.textarea.Value() == "" {
				m.viewport.LineUp(1)
				return m, nil
			} else if m.currentMode == chatMode && m.textarea.Value() == "" {
				if m.historyIndex < len(m.inputHistory)-1 && len(m.inputHistory) > 0 {
					m.historyIndex++
					m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIndex]
					m.textarea.SetValue(m.input)
				}
				return m, nil
			}
		case "down":
			if m.autocomplete.Active {
				if m.autocomplete.SelectedIndex < len(m.autocomplete.Items)-1 {
					m.autocomplete.SelectedIndex++
				}
				return m, nil
			}
			if m.currentMode == settingsMode && !m.inModelMenu {
				if m.selectedProvider < 4 {
					m.selectedProvider++
				}
				return m, nil
			} else if m.currentMode == settingsMode && m.inModelMenu {
				maxModel := getModelCountForProvider(m.selectedProvider)
				if m.selectedModel < maxModel-1 {
					m.selectedModel++
				}
				return m, nil
			} else if m.currentMode == chatMode && m.historyIndex == -1 && m.textarea.Value() == "" {
				m.viewport.LineDown(1)
				return m, nil
			} else if m.currentMode == chatMode && m.textarea.Value() == "" {
				if m.historyIndex > 0 {
					m.historyIndex--
					m.input = m.inputHistory[len(m.inputHistory)-1-m.historyIndex]
					m.textarea.SetValue(m.input)
				} else if m.historyIndex == 0 {
					m.historyIndex = -1
					m.input = ""
					m.textarea.Reset()
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
			if m.currentMode == chatMode && m.textarea.Value() == "" {
				m.viewport.GotoTop()
				return m, nil
			}
		case "end":
			if m.currentMode == chatMode && m.textarea.Value() == "" {
				m.viewport.GotoBottom()
				return m, nil
			}
		case "tab":
			if m.autocomplete.Active && len(m.autocomplete.Items) > 0 && m.autocomplete.SelectedIndex < len(m.autocomplete.Items) {
				selected := m.autocomplete.Items[m.autocomplete.SelectedIndex]
				m.textarea.SetValue(selected.Value)
				m.input = m.textarea.Value()
				m.autocomplete.Active = false
				m.autocomplete.Items = nil
				return m, nil
			}
			if m.currentMode == settingsMode && !m.inModelMenu {
				m.selectedProvider = (m.selectedProvider + 1) % 5
				return m, nil
			}
			if m.showSuggestions && len(m.suggestions) > 0 {
				parts := strings.Split(m.suggestions[0], " ")
				if len(parts) > 0 {
					m.textarea.SetValue(parts[0] + " ")
					m.input = m.textarea.Value()
					m.showSuggestions = false
				}
				return m, nil
			}
			// OpenCode style: Tab toggles Build vs Plan mode when input is empty
			if m.textarea.Value() == "" {
				if m.agentMode == "plan" {
					m.agentMode = "build"
					m.engine.Send(engine.SetPlanMode{Mode: "auto"})
				} else {
					m.agentMode = "plan"
					m.engine.Send(engine.SetPlanMode{Mode: "plan"})
				}
				return m, nil
			}
		case "shift+tab":
			if m.currentMode == settingsMode && !m.inModelMenu {
				m.selectedProvider = (m.selectedProvider - 1 + 5) % 5
				return m, nil
			}
		case "alt+enter", "ctrl+j":
			if m.currentMode == chatMode && !m.isStreaming {
				m.textarea.InsertString("\n")
				m.input = m.textarea.Value()
				return m, nil
			}
		case "ctrl+z":
			if m.currentMode == chatMode && !m.isStreaming {
				restored, err := db.RollbackSession(m.sessionID, 0)
				if err != nil {
					m.messages = append(m.messages, api.Message{Role: "assistant", Content: "❌ Error al rebobinar: " + err.Error()})
				} else {
					m.engine.SwitchSession(m.sessionID)
					history, err := db.GetHistory(m.sessionID)
					if err == nil {
						m.messages = toTUIMessages(history)
					}
					m.textarea.SetValue(restored)
					m.input = restored
					m.messages = append(m.messages, api.Message{Role: "assistant", Content: "↩ Mensaje revertido. Prompt restaurado en el editor."})
				}
				m.updateViewportContent()
				m.viewport.GotoBottom()
				return m, nil
			}
		case "enter":
			if m.autocomplete.Active && len(m.autocomplete.Items) > 0 && m.autocomplete.SelectedIndex < len(m.autocomplete.Items) {
				selected := m.autocomplete.Items[m.autocomplete.SelectedIndex]
				m.textarea.SetValue(selected.Value)
				m.input = m.textarea.Value()
				m.autocomplete.Active = false
				m.autocomplete.Items = nil
				return m, nil
			}
			if m.currentMode == settingsMode && m.inModelMenu {
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
			if m.currentMode == chatMode && !m.isStreaming {
				input := strings.TrimSpace(m.textarea.Value())
				if input == "" {
					input = strings.TrimSpace(m.input)
				}
				if input == "" {
					return m, nil
				}
				m.autocomplete.Active = false
				m.autocomplete.Items = nil
				m.textarea.Reset()
				m.input = ""
				m.historyIndex = -1

				m.inputHistory = append(m.inputHistory, input)
				if len(m.inputHistory) > 50 {
					m.inputHistory = m.inputHistory[1:]
				}

				if strings.HasPrefix(input, "/") {
					trimmed := strings.TrimSpace(input)
					if trimmed == "/diff" || trimmed == "/review" {
						m.currentMode = gitMode
						if sum, err := GetTUIGitSummary(); err == nil {
							m.gitSummary = sum
						}
						m.gitSelectedIdx = 0
						m.gitCommitPrompt = false
						return m, nil
					}
					if trimmed == "/tasks" || trimmed == "/terminal" {
						m.currentMode = activityMode
						if len(m.activityHistory) > 0 {
							m.activitySelected = len(m.activityHistory) - 1
						}
						return m, nil
					}
					if trimmed == "/sessions" {
						if sessions, err := db.ListSessions(); err == nil && len(sessions) > 0 {
							m.savedSessions = sessions
							m.sessionSelected = 0
							m.currentMode = sessionsMode
							return m, nil
						}
					}
					if trimmed == "/model" || trimmed == "/models" || trimmed == "/settings" {
						m.currentMode = settingsMode
						m.selectedProvider = 0
						m.selectedModel = 0
						m.inModelMenu = false
						return m, nil
					}
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
					if strings.HasPrefix(input, "/clear") {
						m.engine.Send(engine.SlashCommand{Name: "clear"})
						m.messages = append(m.messages, api.Message{Role: "assistant", Content: handleClear()})
						m.updateViewportContent()
						m.viewport.GotoBottom()
						return m, nil
					}
					if strings.HasPrefix(input, "/open ") {
						sid := strings.TrimSpace(strings.TrimPrefix(input, "/open "))
						if err := db.ResumeSession(sid); err != nil {
							response := "❌ Error al retomar sesión: " + err.Error()
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: response})
						} else {
							m.engine.SwitchSession(sid)
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
							m.engine.SwitchSession(sid)
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
					if strings.HasPrefix(input, "/rollback") {
						restored, err := db.RollbackSession(m.sessionID, 0)
						if err != nil {
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: "❌ Error al rebobinar: " + err.Error()})
						} else {
							m.engine.SwitchSession(m.sessionID)
							history, err := db.GetHistory(m.sessionID)
							if err == nil {
								m.messages = toTUIMessages(history)
							}
							m.textarea.SetValue(restored)
							m.input = restored
							m.messages = append(m.messages, api.Message{Role: "assistant", Content: "↩ Mensaje revertido. Prompt restaurado en el editor."})
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

				input = ProcessMentions(input)
				m.isStreaming = true
				m.currentResponse = ""
				m.engine.Send(engine.SendMessage{Text: input, SessionID: m.sessionID})
				return m, nil
			}
		default:
			if !m.isStreaming {
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				m.input = m.textarea.Value()
				m.updateSuggestions()
				m.updateAutocomplete()
				return m, cmd
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
		headerHeight := 3
		footerHeight := 8
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		if m.viewport.Height < 5 {
			m.viewport.Height = 5
		}
		if m.width > 4 {
			m.textarea.SetWidth(m.width - 4)
		}
		if r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(msg.Width-4),
		); err == nil {
			m.renderer = r
		}
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
		m.activityHistory = append(m.activityHistory, TUIActivityItem{
			ID:        e.ToolCallID,
			Type:      "tool",
			Name:      e.Name,
			Status:    "running",
			Input:     fmt.Sprintf("%v", e.Input),
			StartTime: time.Now(),
		})
		m.updateViewportContent()
	case engine.ToolExecuting:
		m.executingTool = e.ToolCallID
		m.updateViewportContent()
	case engine.ToolResult:
		m.executingTool = ""
		for idx := len(m.activityHistory) - 1; idx >= 0; idx-- {
			if m.activityHistory[idx].ID == e.ToolCallID || m.activityHistory[idx].Name == e.Name {
				if e.IsError {
					m.activityHistory[idx].Status = "failed"
				} else {
					m.activityHistory[idx].Status = "done"
				}
				m.activityHistory[idx].Output = e.Content
				m.activityHistory[idx].Duration = time.Since(m.activityHistory[idx].StartTime)
				break
			}
		}
		if tools.IsGitRepo() && (strings.Contains(e.Name, "write") || strings.Contains(e.Name, "edit") || strings.Contains(e.Name, "patch") || strings.Contains(e.Name, "bash")) {
			if sum, err := GetTUIGitSummary(); err == nil {
				m.gitSummary = sum
			}
		}
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
		for idx := len(m.activityHistory) - 1; idx >= 0; idx-- {
			if m.activityHistory[idx].ID == e.ToolCallID || m.activityHistory[idx].Name == e.Name {
				m.activityHistory[idx].Status = "failed"
				m.activityHistory[idx].Duration = time.Since(m.activityHistory[idx].StartTime)
				break
			}
		}
	case engine.ToolTimedOut:
		m.executingTool = ""
		for idx := len(m.activityHistory) - 1; idx >= 0; idx-- {
			if m.activityHistory[idx].ID == e.ToolCallID || m.activityHistory[idx].Name == e.Name {
				m.activityHistory[idx].Status = "failed"
				m.activityHistory[idx].Duration = time.Since(m.activityHistory[idx].StartTime)
				break
			}
		}
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
	if m.currentMode == gitMode {
		return RenderGitReview(m.gitSummary, m.gitSelectedIdx, m.gitCommitPrompt, m.gitCommitInput, m.width, m.height)
	}
	if m.currentMode == activityMode {
		return RenderActivityInspector(m.activityHistory, m.activitySelected, m.width, m.height)
	}
	if m.currentMode == sessionsMode {
		return m.renderSessionPickerView()
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

	// Top header with project path, git branch, dirty/clean badge, session ID
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	cwd, _ := os.Getwd()
	projName := filepath.Base(cwd)

	branchStr := "git:no-repo"
	statusBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("● clean")
	if m.gitSummary.Branch != "" {
		branchStr = "⎇ " + m.gitSummary.Branch
		if !m.gitSummary.Clean {
			addCount := 0
			delCount := 0
			for _, f := range m.gitSummary.Files {
				addCount += f.Additions
				delCount += f.Deletions
			}
			statusBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).
				Render(fmt.Sprintf("● dirty (+%d -%d)", addCount, delCount))
		}
	}

	shortID := m.sessionID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	header := fmt.Sprintf("%s  %s  %s  %s  %s",
		titleStyle.Render("LetsGO Code"),
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45")).Render("📁 "+projName),
		lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(branchStr),
		statusBadge,
		sessionStyle.Render("Session: "+shortID),
	)

	// Separator line
	separatorWidth := m.width - 2
	if separatorWidth < 10 {
		separatorWidth = 40
	}
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).
		Render(strings.Repeat("─", separatorWidth))

	s.WriteString(header + "\n" + separator + "\n")

	// Viewport for scrollable messages
	s.WriteString(m.viewport.View())

	// Streaming & Tool execution indicators
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

	// 2-tier real-time bottom status bar
	// Tier 1: Mode badge, Model badge, Token counters, Cost
	modeBg := "82"
	if m.agentMode == "plan" {
		modeBg = "214"
	}
	modeBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color(modeBg)).Render(" " + strings.ToUpper(m.agentMode) + " ")
	modelBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render("◈ " + config.AppConfig.Model)
	tokenStr := fmt.Sprintf("Tokens: %d", m.totalTokens)
	costStr := fmt.Sprintf("Cost: $%.4f", m.totalCost)
	tier1 := fmt.Sprintf("%s  %s  %s  %s",
		modeBadge,
		modelBadge,
		dimStyle.Render(tokenStr),
		dimStyle.Render(costStr),
	)

	// Tier 2: Keyboard shortcuts hint bar
	hintPill := func(key, desc string) string {
		k := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render(key)
		d := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(desc)
		return k + " " + d
	}
	tier2 := fmt.Sprintf("%s   %s   %s   %s   %s   %s",
		hintPill("[Tab]", "Mode"),
		hintPill("[Ctrl+S]", "Models"),
		hintPill("[Ctrl+D]", "Git"),
		hintPill("[Ctrl+B]", "Tasks"),
		hintPill("[/]", "Commands"),
		hintPill("[@]", "Files"),
	)

	s.WriteString("\n" + tier1 + "\n" + tier2 + "\n")

	// Floating Autocomplete popup above textarea
	if m.autocomplete.Active && len(m.autocomplete.Items) > 0 {
		s.WriteString(m.renderAutocompletePopup() + "\n")
	} else if m.showSuggestions && len(m.suggestions) > 0 {
		s.WriteString(m.renderSuggestions())
	}

	if m.pendingApproval != nil {
		s.WriteString(m.renderPermissionDialog() + "\n")
	} else {
		s.WriteString(m.textarea.View() + "\n")
	}
	return s.String()
}

func (m model) renderPermissionDialog() string {
	if m.pendingApproval == nil {
		return ""
	}
	w := m.width - 4
	if w < 20 {
		w = 60
	}
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(0, 1).
		Width(w)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	toolStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("⚠️  TOOL APPROVAL REQUIRED") + "\n\n")
	content.WriteString(descStyle.Render("Tool: ") + toolStyle.Render(m.pendingApproval.toolName) + "\n")

	if cmd, ok := m.pendingApproval.input["command"].(string); ok {
		content.WriteString(descStyle.Render("Command: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render(cmd) + "\n")
	} else if path, ok := m.pendingApproval.input["path"].(string); ok {
		content.WriteString(descStyle.Render("Target: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Render(path) + "\n")
		if oldStr, ok := m.pendingApproval.input["old_string"].(string); ok {
			newStr, _ := m.pendingApproval.input["new_string"].(string)
			if m.showDiff {
				content.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("- "+strings.ReplaceAll(oldStr, "\n", "\n- ")) + "\n")
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("+ "+strings.ReplaceAll(newStr, "\n", "\n+ ")) + "\n")
			} else {
				content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Diff hidden (press [d] to preview)") + "\n")
			}
		}
	} else if patch, ok := m.pendingApproval.input["patch"].(string); ok {
		if m.showDiff {
			lines := strings.Split(patch, "\n")
			for i, l := range lines {
				if i > 15 {
					content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("... (diff truncated)") + "\n")
					break
				}
				if strings.HasPrefix(l, "+") {
					content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(l) + "\n")
				} else if strings.HasPrefix(l, "-") {
					content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(l) + "\n")
				} else {
					content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(l) + "\n")
				}
			}
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Diff hidden (press [d] to preview)") + "\n")
		}
	} else if m.pendingApproval.inputStr != "" {
		content.WriteString(descStyle.Render("Params: ") + lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(m.pendingApproval.inputStr) + "\n")
	}

	content.WriteString("\n" + keyStyle.Render("[y]") + descStyle.Render(" Allow once   ") +
		keyStyle.Render("[c]") + descStyle.Render(" Always this chat   ") +
		keyStyle.Render("[p]") + descStyle.Render(" Always this project   ") +
		keyStyle.Render("[n]") + descStyle.Render(" Deny   ") +
		keyStyle.Render("[d]") + descStyle.Render(" Toggle Diff"))

	return borderStyle.Render(content.String())
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

	// Pump permission approvals into the bubbletea program.
	go func() {
		if m.driver.approvalCh != nil {
			for req := range m.driver.approvalCh {
				if program != nil {
					program.Send(tuiApprovalMsg{req: req})
				}
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

// updateAutocomplete refreshes floating autocomplete suggestions for / and @ triggers
func (m *model) updateAutocomplete() {
	val := m.textarea.Value()
	m.autocomplete.Active = false
	m.autocomplete.Items = nil

	if strings.HasPrefix(val, "/") && !strings.Contains(val, " ") {
		query := strings.ToLower(val)
		for _, cmd := range AvailableSlashCommands {
			if strings.HasPrefix(strings.ToLower(cmd.Name), query) {
				m.autocomplete.Items = append(m.autocomplete.Items, AutocompleteItem{
					Type:        "cmd",
					Icon:        "⚡",
					Title:       cmd.Name,
					Description: cmd.Description,
					Value:       cmd.Name + " ",
				})
			}
		}
		if len(m.autocomplete.Items) > 0 {
			m.autocomplete.Active = true
			m.autocomplete.Trigger = "/"
			m.autocomplete.Query = query
			if m.autocomplete.SelectedIndex >= len(m.autocomplete.Items) {
				m.autocomplete.SelectedIndex = 0
			}
		}
		return
	}

	if idx := strings.LastIndex(val, "@"); idx != -1 {
		prefix := val[idx+1:]
		if !strings.Contains(prefix, " ") {
			suggs := GetMentionSuggestions(prefix)
			for _, sug := range suggs {
				m.autocomplete.Items = append(m.autocomplete.Items, AutocompleteItem{
					Type:        "mention",
					Icon:        "📄",
					Title:       "@" + sug,
					Description: "Context file",
					Value:       val[:idx] + "@" + sug + " ",
				})
			}
			if len(m.autocomplete.Items) > 0 {
				m.autocomplete.Active = true
				m.autocomplete.Trigger = "@"
				m.autocomplete.Query = prefix
				if m.autocomplete.SelectedIndex >= len(m.autocomplete.Items) {
					m.autocomplete.SelectedIndex = 0
				}
			}
			return
		}
	}
}

// renderAutocompletePopup renders the interactive floating popup suggestion box above the textarea
func (m model) renderAutocompletePopup() string {
	if !m.autocomplete.Active || len(m.autocomplete.Items) == 0 {
		return ""
	}
	var b strings.Builder
	boxWidth := m.width - 4
	if boxWidth < 30 {
		boxWidth = 60
	}
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(boxWidth)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	b.WriteString(titleStyle.Render("⚡ Suggestions [↑/↓ navigate, Tab/Enter select, Esc dismiss]") + "\n")

	limit := 6
	if limit > len(m.autocomplete.Items) {
		limit = len(m.autocomplete.Items)
	}

	for i := 0; i < limit; i++ {
		item := m.autocomplete.Items[i]
		isSelected := i == m.autocomplete.SelectedIndex
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		if isSelected {
			cursor = "▸ "
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("238"))
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s\n",
			cursor,
			item.Icon,
			itemStyle.Render(item.Title),
			descStyle.Render(item.Description),
		))
	}
	return boxStyle.Render(b.String())
}

// renderSessionPickerView renders the interactive session switcher overlay
func (m model) renderSessionPickerView() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("236")).Padding(0, 1)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	title := fmt.Sprintf("📂 SESSION PICKER  |  Total Saved: %d", len(m.savedSessions))
	s.WriteString(headerStyle.Render(title) + "\n\n")

	if len(m.savedSessions) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Padding(2, 2)
		s.WriteString(emptyStyle.Render("No saved sessions found."))
		s.WriteString("\n\n" + hintStyle.Render("Press [Esc] to return to chat"))
		return s.String()
	}

	sectionTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	s.WriteString(sectionTitle.Render("Saved Conversations (↑↓ to navigate, Enter to resume, Esc to cancel)") + "\n\n")

	for i, sess := range m.savedSessions {
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if i == m.sessionSelected {
			cursor = "▸ "
			itemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("237"))
		}

		shortID := sess.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		titleStr := sess.Name
		if titleStr == "" {
			titleStr = "Untitled Session"
		}
		timeStr := sess.UpdatedAt.Format("2006-01-02 15:04")
		idBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("[" + shortID + "]")
		timeBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(" + timeStr + ")")

		s.WriteString(fmt.Sprintf("%s%s %s %s\n",
			cursor,
			idBadge,
			itemStyle.Render(titleStr),
			timeBadge,
		))
	}

	s.WriteString("\n" + hintStyle.Render("Press [Enter] to resume selected session | [Esc] to return to chat"))
	return s.String()
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