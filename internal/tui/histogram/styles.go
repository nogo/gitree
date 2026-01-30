package histogram

import "github.com/charmbracelet/lipgloss"

var (
	BarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")) // Cyan

	SelectedBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F1FA8C")) // Yellow

	AxisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")) // Dimmed gray

	FocusedBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62"))

	UnfocusedBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))
)
