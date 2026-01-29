package detail

import "github.com/charmbracelet/lipgloss"

var (
	DetailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	MessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	ParentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
)
