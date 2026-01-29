package graph

import "github.com/charmbracelet/lipgloss"

type ColorPalette struct {
	branchColors map[string]lipgloss.Style
	palette      []lipgloss.Style
	nextIdx      int
}

func NewColorPalette() *ColorPalette {
	return &ColorPalette{
		branchColors: make(map[string]lipgloss.Style),
		palette: []lipgloss.Style{
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")), // pink (main)
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")),  // cyan
			lipgloss.NewStyle().Foreground(lipgloss.Color("221")), // yellow
			lipgloss.NewStyle().Foreground(lipgloss.Color("156")), // green
			lipgloss.NewStyle().Foreground(lipgloss.Color("213")), // magenta
			lipgloss.NewStyle().Foreground(lipgloss.Color("81")),  // blue
		},
	}
}

func (p *ColorPalette) ForBranch(name string) lipgloss.Style {
	if style, ok := p.branchColors[name]; ok {
		return style
	}
	// Assign next color in palette
	style := p.palette[p.nextIdx%len(p.palette)]
	p.branchColors[name] = style
	p.nextIdx++
	return style
}

func (p *ColorPalette) Default() lipgloss.Style {
	return p.palette[0]
}
