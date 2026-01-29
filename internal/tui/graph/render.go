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
// Each lane takes 2 characters (symbol + space), total width = MaxLanes * 2
func (r *RowRenderer) RenderRow(row int) string {
	if row < 0 || row >= len(r.layout.Nodes) {
		return strings.Repeat(" ", r.layout.MaxLanes*2)
	}

	node := r.layout.Nodes[row]
	activeLanes := r.computeActiveLanes(row)

	// Build cells for each lane position
	cells := make([]string, r.layout.MaxLanes)

	// First pass: render base cells (nodes and passthroughs)
	for lane := 0; lane < r.layout.MaxLanes; lane++ {
		if lane == node.Lane {
			// This is where the commit node goes
			cells[lane] = r.colorForLane(lane).Render(string(CharNode))
		} else if activeLanes[lane] {
			// Lane passes through this row
			cells[lane] = r.colorForLane(lane).Render(string(CharVertical))
		} else {
			cells[lane] = " "
		}
	}

	// Second pass: handle merge connections (lanes coming INTO this node from right)
	// MergeFrom contains lanes that were targeting this commit and merge here
	if len(node.MergeFrom) > 0 {
		r.renderMergeConnections(cells, node, activeLanes)
	}

	// Third pass: handle fork connections (new lanes starting FROM this node going right)
	// ForkTo contains new lanes allocated for non-primary parents
	if len(node.ForkTo) > 0 {
		r.renderForkConnections(cells, node, activeLanes)
	}

	// Join cells with spaces between
	var result strings.Builder
	for i, cell := range cells {
		result.WriteString(cell)
		if i < len(cells)-1 {
			// Check if we need horizontal connector between this cell and next
			connector := r.getConnector(row, i, node)
			result.WriteString(connector)
		}
	}

	// Pad to consistent width
	rendered := result.String()
	displayWidth := r.displayWidth(rendered)
	targetWidth := r.layout.MaxLanes * 2
	if displayWidth < targetWidth {
		rendered += strings.Repeat(" ", targetWidth-displayWidth)
	}

	return rendered
}

// renderMergeConnections handles visual connections for merges
func (r *RowRenderer) renderMergeConnections(cells []string, node *CommitNode, activeLanes map[int]bool) {
	for _, mergeLane := range node.MergeFrom {
		if mergeLane > node.Lane {
			// Merge coming from right - use ┘ at the merge lane
			cells[mergeLane] = r.colorForLane(mergeLane).Render(string(CharCornerBR))
		} else if mergeLane < node.Lane {
			// Merge coming from left - use └ at the merge lane
			cells[mergeLane] = r.colorForLane(mergeLane).Render(string(CharCornerBL))
		}
	}
}

// renderForkConnections handles visual connections for forks (new parent lanes)
func (r *RowRenderer) renderForkConnections(cells []string, node *CommitNode, activeLanes map[int]bool) {
	for _, forkLane := range node.ForkTo {
		if forkLane > node.Lane {
			// Fork going to right - use ┐ at the fork lane
			cells[forkLane] = r.colorForLane(forkLane).Render(string(CharCornerTR))
		} else if forkLane < node.Lane {
			// Fork going to left - use ┌ at the fork lane
			cells[forkLane] = r.colorForLane(forkLane).Render(string(CharCornerTL))
		}
	}
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

// computeActiveLanes determines which lanes are active at a given row
func (r *RowRenderer) computeActiveLanes(row int) map[int]bool {
	active := make(map[int]bool)

	// A lane is active at this row if:
	// 1. A commit at or before this row has a parent at or after this row in that lane
	// 2. The lane was created by a fork and the target parent hasn't been reached yet

	// Simple approach: trace from beginning
	activeLanes := make(map[int]string) // lane -> target hash
	var freeLanes []int

	for i := 0; i <= row; i++ {
		node := r.layout.Nodes[i]

		// Find lanes targeting this node
		var targetingLanes []int
		for lane, hash := range activeLanes {
			if hashMatch(hash, node.Hash) {
				targetingLanes = append(targetingLanes, lane)
			}
		}
		sortInts(targetingLanes)

		// Assign lane
		var assignedLane int
		if len(targetingLanes) > 0 {
			assignedLane = targetingLanes[0]
			for _, lane := range targetingLanes[1:] {
				delete(activeLanes, lane)
				freeLanes = insertSorted(freeLanes, lane)
			}
		} else {
			if len(freeLanes) > 0 {
				assignedLane = freeLanes[0]
				freeLanes = freeLanes[1:]
			} else {
				assignedLane = len(activeLanes)
			}
		}

		// Update for parents
		if len(node.Parents) == 0 {
			delete(activeLanes, assignedLane)
			freeLanes = insertSorted(freeLanes, assignedLane)
		} else {
			activeLanes[assignedLane] = node.Parents[0]
			for _, parentHash := range node.Parents[1:] {
				var newLane int
				if len(freeLanes) > 0 {
					newLane = freeLanes[0]
					freeLanes = freeLanes[1:]
				} else {
					maxLane := -1
					for l := range activeLanes {
						if l > maxLane {
							maxLane = l
						}
					}
					newLane = maxLane + 1
				}
				activeLanes[newLane] = parentHash
			}
		}
	}

	// Convert to bool map
	for lane := range activeLanes {
		active[lane] = true
	}
	// Include current node's lane
	active[r.layout.Nodes[row].Lane] = true

	return active
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
