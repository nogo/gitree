package filter

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

type BranchFilter struct {
	branches     []domain.Branch
	selected     map[string]bool // branch name → visible
	cursor       int
	scrollOffset int
	width        int
	height       int
}

func NewBranchFilter(branches []domain.Branch) BranchFilter {
	selected := make(map[string]bool)
	for _, b := range branches {
		selected[b.Name] = true // All visible by default
	}
	return BranchFilter{
		branches: branches,
		selected: selected,
	}
}

func (f *BranchFilter) SetSize(w, h int) {
	f.width = w
	f.height = h
}

// maxVisibleItems calculates how many items can be displayed
func (f BranchFilter) maxVisibleItems() int {
	// Height overhead: title(1) + hint(1) + empty(1) + empty(1) + footer(1) = 5
	// Plus border(2) + padding(2) = 4
	// Plus scroll indicators (2 max)
	// Total overhead = 11
	maxItems := f.height - 11
	if maxItems < 3 {
		maxItems = 3
	}
	return maxItems
}

// adjustScroll keeps cursor within visible range
func (f *BranchFilter) adjustScroll() {
	maxVisible := f.maxVisibleItems()
	if len(f.branches) <= maxVisible {
		f.scrollOffset = 0
		return
	}
	if f.cursor < f.scrollOffset {
		f.scrollOffset = f.cursor
	}
	if f.cursor >= f.scrollOffset+maxVisible {
		f.scrollOffset = f.cursor - maxVisible + 1
	}
}

// Update handles input and returns (updated filter, cmd, done, cancelled)
func (f BranchFilter) Update(msg tea.Msg) (BranchFilter, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if f.cursor < len(f.branches)-1 {
				f.cursor++
				f.adjustScroll()
			}
		case "k", "up":
			if f.cursor > 0 {
				f.cursor--
				f.adjustScroll()
			}
		case " ", "space":
			// Toggle current branch
			if f.cursor < len(f.branches) {
				name := f.branches[f.cursor].Name
				f.selected[name] = !f.selected[name]
			}
		case "a":
			// Select all
			for name := range f.selected {
				f.selected[name] = true
			}
		case "n":
			// Select none
			for name := range f.selected {
				f.selected[name] = false
			}
		case "enter":
			return f, nil, true, false // Done, apply filter
		case "esc":
			return f, nil, false, true // Cancelled
		}
	}
	return f, nil, false, false
}

func (f BranchFilter) View() string {
	var lines []string
	lines = append(lines, TitleStyle.Render("Filter branches"))
	lines = append(lines, HintStyle.Render("space=toggle  a=all  n=none"))
	lines = append(lines, "")

	maxVisible := f.maxVisibleItems()
	totalItems := len(f.branches)

	// Show "more above" indicator
	if f.scrollOffset > 0 {
		lines = append(lines, HintStyle.Render(fmt.Sprintf("  ↑ %d more", f.scrollOffset)))
	}

	// Calculate visible range
	endIdx := f.scrollOffset + maxVisible
	if endIdx > totalItems {
		endIdx = totalItems
	}

	for i := f.scrollOffset; i < endIdx; i++ {
		b := f.branches[i]
		checkbox := UncheckedStyle.Render("[ ]")
		if f.selected[b.Name] {
			checkbox = CheckedStyle.Render("[x]")
		}

		name := b.Name
		if b.IsRemote {
			name = HintStyle.Render(name)
		}

		line := fmt.Sprintf("  %s %s", checkbox, name)
		if i == f.cursor {
			line = fmt.Sprintf("> %s %s", checkbox, name)
			line = SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	// Show "more below" indicator
	if endIdx < totalItems {
		lines = append(lines, HintStyle.Render(fmt.Sprintf("  ↓ %d more", totalItems-endIdx)))
	}

	lines = append(lines, "")
	lines = append(lines, HintStyle.Render("[Enter] Apply  [Esc] Cancel"))

	content := strings.Join(lines, "\n")

	// Calculate inner dimensions
	innerWidth := f.width - 6
	if innerWidth < 30 {
		innerWidth = 30
	}

	return lipgloss.Place(
		f.width, f.height,
		lipgloss.Center, lipgloss.Center,
		FilterStyle.Width(innerWidth).Render(content),
	)
}

func (f BranchFilter) SelectedBranches() []string {
	var result []string
	for name, visible := range f.selected {
		if visible {
			result = append(result, name)
		}
	}
	return result
}

func (f BranchFilter) AllSelected() bool {
	for _, visible := range f.selected {
		if !visible {
			return false
		}
	}
	return true
}

// Reset resets the filter to show all branches
func (f *BranchFilter) Reset() {
	for name := range f.selected {
		f.selected[name] = true
	}
}

// SetSelected sets the selection state for a specific branch
func (f *BranchFilter) SetSelected(name string, selected bool) {
	f.selected[name] = selected
}

// UpdateBranches updates the branch list (e.g., after repo refresh)
func (f *BranchFilter) UpdateBranches(branches []domain.Branch) {
	f.branches = branches
	// Add any new branches as selected
	for _, b := range branches {
		if _, exists := f.selected[b.Name]; !exists {
			f.selected[b.Name] = true
		}
	}
	// Remove branches that no longer exist
	branchSet := make(map[string]bool)
	for _, b := range branches {
		branchSet[b.Name] = true
	}
	for name := range f.selected {
		if !branchSet[name] {
			delete(f.selected, name)
		}
	}
	// Clamp cursor
	if f.cursor >= len(f.branches) {
		f.cursor = len(f.branches) - 1
	}
	if f.cursor < 0 {
		f.cursor = 0
	}
}
