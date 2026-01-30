package diff

import "github.com/charmbracelet/lipgloss"

var (
	// Container style with border
	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	// Header style
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	// File path in header
	FilePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Stats in header
	AdditionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	DeletionsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	// Diff line styles
	DiffAddedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	DiffDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555"))

	DiffHunkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	DiffContextStyle = lipgloss.NewStyle()

	// Footer style
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// File indicator style
	FileIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	// Loading/error style
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)
