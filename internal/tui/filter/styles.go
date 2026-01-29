package filter

import "github.com/charmbracelet/lipgloss"

var (
	FilterStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255"))

	CheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("156"))

	UncheckedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	HintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)
