package list

import "github.com/charmbracelet/lipgloss"

var (
	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("237")).
				Foreground(lipgloss.Color("255"))

	HashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	AuthorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))

	DateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242"))

	MessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)
