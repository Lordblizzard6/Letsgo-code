package ui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	messages      []Message
	currentInput  string
	isStreaming   bool
	toolConfirm   *ToolConfirm
	width         int
	height        int
	historyOffset int
}

type Message struct {
	Role    string
	Content string
}

type ToolConfirm struct {
	ToolName  string
	Args      string
	Confirmed *bool
}

type StreamMsg struct {
	Content string
	Done    bool
}

type InputMsg struct {
	Text string
}

func InitialModel() Model {
	return Model{
		messages:     make([]Message, 0),
		currentInput: "",
		isStreaming:  false,
		width:        80,
		height:       24,
	}
}

func (m Model) Init() bubbletea.Cmd {
	return nil
}

func (m Model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case bubbletea.KeyMsg:
		switch msg.Type {
		case bubbletea.KeyCtrlC:
			return m, bubbletea.Quit

		case bubbletea.KeyEnter:
			if m.currentInput != "" && !m.isStreaming {
				return m, func() bubbletea.Msg {
					return InputMsg{Text: m.currentInput}
				}
			}

		case bubbletea.KeyBackspace:
			if len(m.currentInput) > 0 {
				m.currentInput = m.currentInput[:len(m.currentInput)-1]
			}

		case bubbletea.KeyUp:
			if m.historyOffset < len(m.messages) {
				m.historyOffset++
			}

		case bubbletea.KeyDown:
			if m.historyOffset > 0 {
				m.historyOffset--
			}

		default:
			if !m.isStreaming && len(msg.String()) == 1 {
				m.currentInput += msg.String()
			}
		}

	case StreamMsg:
		if len(m.messages) > 0 {
			m.messages[len(m.messages)-1].Content += msg.Content
		}
		if msg.Done {
			m.isStreaming = false
		}

	case InputMsg:
		m.messages = append(m.messages, Message{
			Role:    "user",
			Content: msg.Text,
		})
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: "",
		})
		m.currentInput = ""
		m.isStreaming = true
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("🤖 MyCLI - Assistant"))
	b.WriteString("\n\n")

	// Historial de mensajes
	historyHeight := m.height - 10
	startIdx := 0
	if len(m.messages) > historyHeight {
		startIdx = len(m.messages) - historyHeight + m.historyOffset
	}

	for i := startIdx; i < len(m.messages); i++ {
		msg := m.messages[i]
		if msg.Role == "user" {
			b.WriteString(userStyle.Render("❯ " + msg.Content))
		} else {
			b.WriteString(assistantStyle.Render("  " + msg.Content))
		}
		b.WriteString("\n")
	}

	// Input
	b.WriteString("\n")
	b.WriteString(inputPromptStyle.Render("❯ "))
	b.WriteString(inputStyle.Render(m.currentInput))
	if m.isStreaming {
		b.WriteString(spinnerStyle.Render(" ⏳"))
	}
	b.WriteString("\n")

	// Ayuda
	b.WriteString(helpStyle.Render("Ctrl+C: salir | Enter: enviar | ↑↓: scroll"))

	return b.String()
}

// Styles
var (
	headerStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	userStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	assistantStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	inputPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	inputStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	spinnerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	helpStyle        = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
)
