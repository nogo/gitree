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

	FileSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true)

	FileAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")) // Green

	FileModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Yellow

	FileDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")) // Red

	FileRenamedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("51")) // Cyan

	AdditionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")) // Green

	DeletionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red

	FilePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)
