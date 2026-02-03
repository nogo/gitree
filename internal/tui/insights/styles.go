package insights

import "github.com/charmbracelet/lipgloss"

// Heat intensity colors (5 levels, 0-4)
var (
	Heat0 = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray/empty
	Heat1 = lipgloss.NewStyle().Foreground(lipgloss.Color("22"))  // Light green
	Heat2 = lipgloss.NewStyle().Foreground(lipgloss.Color("28"))  // Medium green
	Heat3 = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // Dark green
	Heat4 = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))  // Bright green
)

// HeatStyles provides indexed access to heat level styles.
var HeatStyles = []lipgloss.Style{Heat0, Heat1, Heat2, Heat3, Heat4}

// Stats table styles
var (
	HeaderStyle  = lipgloss.NewStyle().Bold(true)
	NameStyle    = lipgloss.NewStyle()
	CountStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241")) // Dimmed
	PercentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))  // Cyan
)

// Section title style
var SectionTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("205"))
