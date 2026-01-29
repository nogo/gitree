package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	BadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("252"))

	OriginBadgeStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("22")).
				Foreground(lipgloss.Color("252"))

	HeadBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("161")).
			Foreground(lipgloss.Color("255")).
			Bold(true)
)

func (r *Renderer) badgeStyle(ref string) lipgloss.Style {
	if ref == "HEAD" {
		return HeadBadgeStyle
	}
	if strings.Contains(ref, "origin") || strings.Contains(ref, "/") {
		return OriginBadgeStyle
	}
	// Use branch color for local branches
	color := r.colors.ForBranch(ref)
	return BadgeStyle.Background(color.GetForeground()).Foreground(lipgloss.Color("0"))
}
