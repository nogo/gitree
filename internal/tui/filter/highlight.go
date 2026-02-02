package filter

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

// AuthorHighlight provides single-select author highlighting
type AuthorHighlight struct {
	authors      []AuthorEntry
	cursor       int
	scrollOffset int
	selectedName string // empty = no highlight (None option)
	width        int
	height       int
}

// NewAuthorHighlight creates a highlight selector from commits
func NewAuthorHighlight(commits []domain.Commit) AuthorHighlight {
	// Reuse author extraction logic from AuthorFilter
	authorCounts := make(map[string]int)
	authorEmails := make(map[string]map[string]bool)

	for _, c := range commits {
		name := normalizeName(c.Author)
		email := strings.ToLower(c.Email)
		authorCounts[name]++
		if authorEmails[name] == nil {
			authorEmails[name] = make(map[string]bool)
		}
		authorEmails[name][email] = true
	}

	var authors []AuthorEntry
	for name, count := range authorCounts {
		var emails []string
		for email := range authorEmails[name] {
			emails = append(emails, email)
		}
		authors = append(authors, AuthorEntry{
			Name:   name,
			Emails: emails,
			Count:  count,
		})
	}

	// Sort by count descending
	for i := 0; i < len(authors); i++ {
		for j := i + 1; j < len(authors); j++ {
			if authors[j].Count > authors[i].Count {
				authors[i], authors[j] = authors[j], authors[i]
			}
		}
	}

	return AuthorHighlight{
		authors:      authors,
		cursor:       0, // Start on "None"
		selectedName: "",
	}
}

func (h *AuthorHighlight) SetSize(w, height int) {
	h.width = w
	h.height = height
}

// totalItems returns total item count including the "None" option
func (h AuthorHighlight) totalItems() int {
	return len(h.authors) + 1 // +1 for "None" option
}

// maxVisibleItems calculates how many items can be displayed
func (h AuthorHighlight) maxVisibleItems() int {
	// Height overhead: title(1) + hint(1) + empty(1) + empty(1) + footer(1) = 5
	// Plus border(2) + padding(2) = 4
	// Plus scroll indicators (2 max)
	// Total overhead = 11
	maxItems := h.height - 11
	if maxItems < 3 {
		maxItems = 3
	}
	return maxItems
}

// adjustScroll keeps cursor within visible range
func (h *AuthorHighlight) adjustScroll() {
	maxVisible := h.maxVisibleItems()
	total := h.totalItems()
	if total <= maxVisible {
		h.scrollOffset = 0
		return
	}
	if h.cursor < h.scrollOffset {
		h.scrollOffset = h.cursor
	}
	if h.cursor >= h.scrollOffset+maxVisible {
		h.scrollOffset = h.cursor - maxVisible + 1
	}
}

// Update handles input and returns (updated, cmd, done, cancelled)
func (h AuthorHighlight) Update(msg tea.Msg) (AuthorHighlight, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			// cursor 0 = None, 1..N = authors
			if h.cursor < len(h.authors) {
				h.cursor++
				h.adjustScroll()
			}
		case "k", "up":
			if h.cursor > 0 {
				h.cursor--
				h.adjustScroll()
			}
		case "enter", " ", "space":
			// Select current option
			if h.cursor == 0 {
				h.selectedName = "" // None
			} else {
				h.selectedName = h.authors[h.cursor-1].Name
			}
			return h, nil, true, false // Done, apply
		case "esc":
			return h, nil, false, true // Cancelled
		}
	}
	return h, nil, false, false
}

func (h AuthorHighlight) View() string {
	var lines []string
	lines = append(lines, TitleStyle.Render("Highlight Author"))
	lines = append(lines, HintStyle.Render("j/k=move  Enter=select"))
	lines = append(lines, "")

	maxVisible := h.maxVisibleItems()
	total := h.totalItems()

	// Show "more above" indicator
	if h.scrollOffset > 0 {
		lines = append(lines, HintStyle.Render(fmt.Sprintf("  ↑ %d more", h.scrollOffset)))
	}

	// Calculate visible range
	endIdx := h.scrollOffset + maxVisible
	if endIdx > total {
		endIdx = total
	}

	// Render visible items (index 0 = None, index 1..N = authors)
	for i := h.scrollOffset; i < endIdx; i++ {
		if i == 0 {
			// "None" option
			radio := RadioOffStyle.Render("( )")
			if h.selectedName == "" {
				radio = RadioOnStyle.Render("(•)")
			}
			line := fmt.Sprintf("  %s None (show all normally)", radio)
			if h.cursor == 0 {
				line = fmt.Sprintf("> %s None (show all normally)", radio)
				line = SelectedStyle.Render(line)
			}
			lines = append(lines, line)
		} else {
			// Author option (index i corresponds to authors[i-1])
			a := h.authors[i-1]
			radio := RadioOffStyle.Render("( )")
			if h.selectedName == a.Name {
				radio = RadioOnStyle.Render("(•)")
			}

			displayName := strings.Title(a.Name)
			label := fmt.Sprintf("%s (%d)", displayName, a.Count)

			line := fmt.Sprintf("  %s %s", radio, label)
			if h.cursor == i {
				line = fmt.Sprintf("> %s %s", radio, label)
				line = SelectedStyle.Render(line)
			}
			lines = append(lines, line)
		}
	}

	// Show "more below" indicator
	if endIdx < total {
		lines = append(lines, HintStyle.Render(fmt.Sprintf("  ↓ %d more", total-endIdx)))
	}

	lines = append(lines, "")
	lines = append(lines, HintStyle.Render("[Enter] Apply  [Esc] Cancel"))

	content := strings.Join(lines, "\n")

	innerWidth := h.width - 6
	if innerWidth < 30 {
		innerWidth = 30
	}

	return lipgloss.Place(
		h.width, h.height,
		lipgloss.Center, lipgloss.Center,
		FilterStyle.Width(innerWidth).Render(content),
	)
}

// HighlightedEmails returns emails for the highlighted author (empty if None)
func (h AuthorHighlight) HighlightedEmails() []string {
	if h.selectedName == "" {
		return nil
	}
	for _, a := range h.authors {
		if a.Name == h.selectedName {
			return a.Emails
		}
	}
	return nil
}

// HighlightedName returns the display name of highlighted author (empty if None)
func (h AuthorHighlight) HighlightedName() string {
	if h.selectedName == "" {
		return ""
	}
	return strings.Title(h.selectedName)
}

// IsActive returns true if an author is highlighted
func (h AuthorHighlight) IsActive() bool {
	return h.selectedName != ""
}

// Reset clears the highlight
func (h *AuthorHighlight) Reset() {
	h.selectedName = ""
	h.cursor = 0
}

// UpdateAuthors refreshes the author list after repo reload
func (h *AuthorHighlight) UpdateAuthors(commits []domain.Commit) {
	authorCounts := make(map[string]int)
	authorEmails := make(map[string]map[string]bool)

	for _, c := range commits {
		name := normalizeName(c.Author)
		email := strings.ToLower(c.Email)
		authorCounts[name]++
		if authorEmails[name] == nil {
			authorEmails[name] = make(map[string]bool)
		}
		authorEmails[name][email] = true
	}

	var authors []AuthorEntry
	for name, count := range authorCounts {
		var emails []string
		for email := range authorEmails[name] {
			emails = append(emails, email)
		}
		authors = append(authors, AuthorEntry{
			Name:   name,
			Emails: emails,
			Count:  count,
		})
	}

	// Sort by count descending
	for i := 0; i < len(authors); i++ {
		for j := i + 1; j < len(authors); j++ {
			if authors[j].Count > authors[i].Count {
				authors[i], authors[j] = authors[j], authors[i]
			}
		}
	}

	h.authors = authors

	// Verify selected author still exists
	if h.selectedName != "" {
		found := false
		for _, a := range authors {
			if a.Name == h.selectedName {
				found = true
				break
			}
		}
		if !found {
			h.selectedName = ""
			h.cursor = 0
		}
	}

	// Clamp cursor
	if h.cursor > len(h.authors) {
		h.cursor = len(h.authors)
	}
}

// Radio button styles
var (
	RadioOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("156"))

	RadioOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)
