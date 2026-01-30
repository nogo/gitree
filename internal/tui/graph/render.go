package graph

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RowRenderer handles rendering graph cells for display
type RowRenderer struct {
	layout *GraphLayout
	colors *ColorPalette
}

// NewRowRenderer creates a renderer for the given layout
func NewRowRenderer(layout *GraphLayout, colors *ColorPalette) *RowRenderer {
	return &RowRenderer{
		layout: layout,
		colors: colors,
	}
}

// RenderRow produces the graph string for a given row index
// Each lane takes 2 characters (symbol + connector/space), total width = MaxLanes * 2
func (r *RowRenderer) RenderRow(row int) string {
	if row < 0 || row >= len(r.layout.Nodes) {
		return strings.Repeat(" ", r.layout.MaxLanes*2)
	}

	node := r.layout.Nodes[row]
	activeLanes := r.layout.ActiveLanesAt(row)

	// Build the graph string with fixed-width cells
	// Format: [cell0][sep0][cell1][sep1]... where each cell is 1 char and each sep is 1 char
	var result strings.Builder

	for lane := range r.layout.MaxLanes {
		// Render the cell for this lane
		var cellChar rune
		var cellColor int

		if lane == node.Lane {
			cellChar = CharNode
			cellColor = lane
		} else if activeLanes[lane] {
			cellChar = CharVertical
			cellColor = lane
		} else {
			cellChar = CharSpace
			cellColor = -1 // no color
		}

		// Check for merge/fork connections that override the cell
		for _, mergeLane := range node.MergeFrom {
			if mergeLane == lane {
				if mergeLane > node.Lane {
					cellChar = CharCornerBR // ┘
				} else {
					cellChar = CharCornerBL // └
				}
				cellColor = lane
			}
		}
		for _, forkLane := range node.ForkTo {
			if forkLane == lane {
				if forkLane > node.Lane {
					cellChar = CharCornerTR // ┐
				} else {
					cellChar = CharCornerTL // ┌
				}
				cellColor = lane
			}
		}

		// Write the cell
		if cellColor >= 0 {
			result.WriteString(r.colorForLane(cellColor).Render(string(cellChar)))
		} else {
			result.WriteRune(cellChar)
		}

		// Write separator (connector or space) - except after last lane
		if lane < r.layout.MaxLanes-1 {
			connector := r.getConnector(row, lane, node)
			result.WriteString(connector)
		}
	}

	// Add trailing space for consistent width (MaxLanes * 2 total)
	// We have MaxLanes cells + (MaxLanes-1) separators = 2*MaxLanes - 1 chars
	// Add 1 space to make it 2*MaxLanes
	result.WriteRune(' ')

	return result.String()
}

// getConnector returns the connector character between lane i and i+1
func (r *RowRenderer) getConnector(row, lane int, node *CommitNode) string {
	// Check if there's a horizontal connection between lane and lane+1
	needsConnection := false
	connectionLane := lane // for color

	// Check merge connections
	for _, mergeLane := range node.MergeFrom {
		if (mergeLane > node.Lane && lane >= node.Lane && lane < mergeLane) ||
			(mergeLane < node.Lane && lane >= mergeLane && lane < node.Lane) {
			needsConnection = true
			connectionLane = mergeLane
			break
		}
	}

	// Check fork connections
	if !needsConnection {
		for _, forkLane := range node.ForkTo {
			if (forkLane > node.Lane && lane >= node.Lane && lane < forkLane) ||
				(forkLane < node.Lane && lane >= forkLane && lane < node.Lane) {
				needsConnection = true
				connectionLane = forkLane
				break
			}
		}
	}

	if needsConnection {
		return r.colorForLane(connectionLane).Render(string(CharHorizontal))
	}
	return " "
}

// colorForLane returns the style for a given lane
func (r *RowRenderer) colorForLane(lane int) lipgloss.Style {
	colors := []lipgloss.Color{
		lipgloss.Color("205"), // pink
		lipgloss.Color("86"),  // cyan
		lipgloss.Color("156"), // green
		lipgloss.Color("221"), // yellow
		lipgloss.Color("213"), // magenta
		lipgloss.Color("81"),  // blue
	}
	return lipgloss.NewStyle().Foreground(colors[lane%len(colors)])
}

// DimmedGraphStyle is used for dimmed graph elements
var DimmedGraphStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))

// RenderRowDimmed produces a dimmed graph string for a given row index
func (r *RowRenderer) RenderRowDimmed(row int) string {
	if row < 0 || row >= len(r.layout.Nodes) {
		return strings.Repeat(" ", r.layout.MaxLanes*2)
	}

	node := r.layout.Nodes[row]
	activeLanes := r.layout.ActiveLanesAt(row)

	var result strings.Builder

	for lane := range r.layout.MaxLanes {
		var cellChar rune
		hasContent := false

		if lane == node.Lane {
			cellChar = CharNode
			hasContent = true
		} else if activeLanes[lane] {
			cellChar = CharVertical
			hasContent = true
		} else {
			cellChar = CharSpace
		}

		// Check for merge/fork connections
		for _, mergeLane := range node.MergeFrom {
			if mergeLane == lane {
				if mergeLane > node.Lane {
					cellChar = CharCornerBR
				} else {
					cellChar = CharCornerBL
				}
				hasContent = true
			}
		}
		for _, forkLane := range node.ForkTo {
			if forkLane == lane {
				if forkLane > node.Lane {
					cellChar = CharCornerTR
				} else {
					cellChar = CharCornerTL
				}
				hasContent = true
			}
		}

		// Write the cell (all dimmed)
		if hasContent {
			result.WriteString(DimmedGraphStyle.Render(string(cellChar)))
		} else {
			result.WriteRune(cellChar)
		}

		// Write separator
		if lane < r.layout.MaxLanes-1 {
			connector := r.getConnectorDimmed(row, lane, node)
			result.WriteString(connector)
		}
	}

	result.WriteRune(' ')
	return result.String()
}

// getConnectorDimmed returns a dimmed connector character
func (r *RowRenderer) getConnectorDimmed(row, lane int, node *CommitNode) string {
	needsConnection := false

	for _, mergeLane := range node.MergeFrom {
		if (mergeLane > node.Lane && lane >= node.Lane && lane < mergeLane) ||
			(mergeLane < node.Lane && lane >= mergeLane && lane < node.Lane) {
			needsConnection = true
			break
		}
	}

	if !needsConnection {
		for _, forkLane := range node.ForkTo {
			if (forkLane > node.Lane && lane >= node.Lane && lane < forkLane) ||
				(forkLane < node.Lane && lane >= forkLane && lane < node.Lane) {
				needsConnection = true
				break
			}
		}
	}

	if needsConnection {
		return DimmedGraphStyle.Render(string(CharHorizontal))
	}
	return " "
}

// displayWidth calculates the display width in runes, excluding ANSI codes
func (r *RowRenderer) displayWidth(s string) int {
	width := 0
	inEscape := false
	for _, ch := range s {
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		width++
	}
	return width
}

// RenderContinuation renders vertical continuation lines for expanded rows
// This draws │ in all active lanes at the given row index
func (r *RowRenderer) RenderContinuation(row int) string {
	if row < 0 || row >= len(r.layout.Nodes) {
		return strings.Repeat(" ", r.layout.MaxLanes*2)
	}

	activeLanes := r.layout.ActiveLanesAt(row)

	var result strings.Builder
	for lane := range r.layout.MaxLanes {
		if activeLanes[lane] {
			result.WriteString(r.colorForLane(lane).Render(string(CharVertical)))
		} else {
			result.WriteRune(CharSpace)
		}

		// Separator space
		if lane < r.layout.MaxLanes-1 {
			result.WriteRune(' ')
		}
	}
	result.WriteRune(' ')

	return result.String()
}
