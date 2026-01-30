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

	// Dimmed styles for non-highlighted commits
	DimmedHashStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("239"))

	DimmedAuthorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("239"))

	DimmedDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("239"))

	DimmedMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("239"))
)
