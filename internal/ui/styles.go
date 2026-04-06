package ui

import "github.com/charmbracelet/lipgloss"

// Styles adicionales para la UI
var (
	// Colores
	ColorPrimary   = lipgloss.Color("12")
	ColorSecondary = lipgloss.Color("7")
	ColorSuccess   = lipgloss.Color("10")
	ColorError     = lipgloss.Color("9")
	ColorWarning   = lipgloss.Color("11")
	ColorMuted     = lipgloss.Color("8")

	// Estilos de contenedores
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1).
			Foreground(ColorError)

	ToolCallStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(1)

	// Estilos de texto
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(ColorSecondary)

	CodeStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(ColorMuted)

	StatusStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)
)
