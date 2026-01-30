package text

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Styles for stats display
var (
	AdditionsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	DeletionsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// FileStats represents additions/deletions for display.
type FileStats struct {
	Additions int
	Deletions int
}

// Render returns styled "+N -M" string.
func (s FileStats) Render() string {
	return fmt.Sprintf("%s %s",
		AdditionsStyle.Render(fmt.Sprintf("+%d", s.Additions)),
		DeletionsStyle.Render(fmt.Sprintf("-%d", s.Deletions)),
	)
}

// Short returns a compact "+N -M" without styling.
func (s FileStats) Short() string {
	return fmt.Sprintf("+%d -%d", s.Additions, s.Deletions)
}

// OptionLine represents a selectable option in a filter list.
type OptionLine struct {
	Selected bool   // is this the cursor position
	Checked  bool   // is this option checked/selected
	Label    string // display text
	Radio    bool   // true for radio (○/●), false for checkbox (☐/☑)
}

// Render returns the formatted option line.
func (o OptionLine) Render() string {
	cursor := "  "
	if o.Selected {
		cursor = "> "
	}

	var indicator string
	if o.Radio {
		if o.Checked {
			indicator = "●"
		} else {
			indicator = "○"
		}
	} else {
		if o.Checked {
			indicator = "☑"
		} else {
			indicator = "☐"
		}
	}

	return fmt.Sprintf("%s%s %s", cursor, indicator, o.Label)
}

// RelativeTime formats a duration as relative time string.
func RelativeTime(minutes int) string {
	switch {
	case minutes < 60:
		return fmt.Sprintf("%dm ago", minutes)
	case minutes < 60*24:
		return fmt.Sprintf("%dh ago", minutes/60)
	case minutes < 60*24*7:
		return fmt.Sprintf("%dd ago", minutes/(60*24))
	default:
		return fmt.Sprintf("%dw ago", minutes/(60*24*7))
	}
}
