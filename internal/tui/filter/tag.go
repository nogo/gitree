package filter

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

type TagFilter struct {
	tags         []string
	selected     map[string]bool // tag name → selected
	cursor       int
	scrollOffset int
	width        int
	height       int
}

func NewTagFilter(commits []domain.Commit) TagFilter {
	// Collect unique tags from commits
	tagSet := make(map[string]bool)
	for _, c := range commits {
		for _, tag := range c.Tags {
			tagSet[tag] = true
		}
	}

	// Build sorted tag list
	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	// Initialize selected map (none selected by default for tag filter)
	// Unlike branch/author filters, tag filter shows all until tags are selected
	selected := make(map[string]bool)
	for _, tag := range tags {
		selected[tag] = false
	}

	return TagFilter{
		tags:     tags,
		selected: selected,
	}
}

func (f *TagFilter) SetSize(w, h int) {
	f.width = w
	f.height = h
}

// maxVisibleItems calculates how many items can be displayed
func (f TagFilter) maxVisibleItems() int {
	maxItems := f.height - 11
	if maxItems < 3 {
		maxItems = 3
	}
	return maxItems
}

// adjustScroll keeps cursor within visible range
func (f *TagFilter) adjustScroll() {
	maxVisible := f.maxVisibleItems()
	if len(f.tags) <= maxVisible {
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
func (f TagFilter) Update(msg tea.Msg) (TagFilter, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if f.cursor < len(f.tags)-1 {
				f.cursor++
				f.adjustScroll()
			}
		case "k", "up":
			if f.cursor > 0 {
				f.cursor--
				f.adjustScroll()
			}
		case " ", "space":
			// Toggle current tag
			if f.cursor < len(f.tags) {
				name := f.tags[f.cursor]
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

func (f TagFilter) View() string {
	var lines []string
	lines = append(lines, TitleStyle.Render("Filter tags"))
	lines = append(lines, HintStyle.Render("space=toggle  a=all  n=none"))
	lines = append(lines, "")

	if len(f.tags) == 0 {
		lines = append(lines, HintStyle.Render("  No tags in repository"))
		lines = append(lines, "")
		lines = append(lines, HintStyle.Render("[Esc] Close"))

		content := strings.Join(lines, "\n")
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

	maxVisible := f.maxVisibleItems()
	totalItems := len(f.tags)

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
		tag := f.tags[i]
		checkbox := UncheckedStyle.Render("[ ]")
		if f.selected[tag] {
			checkbox = CheckedStyle.Render("[x]")
		}

		line := fmt.Sprintf("  %s %s", checkbox, tag)
		if i == f.cursor {
			line = fmt.Sprintf("> %s %s", checkbox, tag)
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

// SelectedTags returns the list of selected tag names
func (f TagFilter) SelectedTags() []string {
	var result []string
	for tag, selected := range f.selected {
		if selected {
			result = append(result, tag)
		}
	}
	return result
}

// HasSelection returns true if any tag is selected
func (f TagFilter) HasSelection() bool {
	for _, selected := range f.selected {
		if selected {
			return true
		}
	}
	return false
}

// Reset resets the filter (deselect all tags)
func (f *TagFilter) Reset() {
	for tag := range f.selected {
		f.selected[tag] = false
	}
}

// UpdateTags updates the tag list (e.g., after repo refresh)
func (f *TagFilter) UpdateTags(commits []domain.Commit) {
	// Collect unique tags from commits
	tagSet := make(map[string]bool)
	for _, c := range commits {
		for _, tag := range c.Tags {
			tagSet[tag] = true
		}
	}

	// Build sorted tag list
	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	f.tags = tags

	// Remove tags that no longer exist
	for tag := range f.selected {
		if !tagSet[tag] {
			delete(f.selected, tag)
		}
	}

	// Add new tags as not selected
	for _, tag := range tags {
		if _, exists := f.selected[tag]; !exists {
			f.selected[tag] = false
		}
	}

	// Clamp cursor
	if f.cursor >= len(f.tags) {
		f.cursor = len(f.tags) - 1
	}
	if f.cursor < 0 {
		f.cursor = 0
	}
}

// SelectedCount returns the number of selected tags
func (f TagFilter) SelectedCount() int {
	count := 0
	for _, selected := range f.selected {
		if selected {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of tags
func (f TagFilter) TotalCount() int {
	return len(f.tags)
}
