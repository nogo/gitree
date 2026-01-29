package tui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	HeaderDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	ColumnHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)
