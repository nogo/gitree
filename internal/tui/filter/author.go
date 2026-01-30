package filter

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

type AuthorEntry struct {
	Name   string
	Emails []string // all emails used by this author (lowercase)
	Count  int
}

type AuthorFilter struct {
	authors  []AuthorEntry
	selected map[string]bool // normalized name → selected
	cursor   int
	width    int
	height   int
}

func NewAuthorFilter(commits []domain.Commit) AuthorFilter {
	// Count commits per author name and collect emails
	authorCounts := make(map[string]int)
	authorEmails := make(map[string]map[string]bool) // name → set of emails

	for _, c := range commits {
		name := normalizeName(c.Author)
		email := strings.ToLower(c.Email)
		authorCounts[name]++
		if authorEmails[name] == nil {
			authorEmails[name] = make(map[string]bool)
		}
		authorEmails[name][email] = true
	}

	// Build sorted author list (by count descending)
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

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Count > authors[j].Count
	})

	// Initialize selected map
	selected := make(map[string]bool)
	for _, a := range authors {
		selected[a.Name] = true
	}

	return AuthorFilter{
		authors:  authors,
		selected: selected,
	}
}

// normalizeName normalizes author name for grouping
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (f *AuthorFilter) SetSize(w, h int) {
	f.width = w
	f.height = h
}

// Update handles input and returns (updated filter, cmd, done, cancelled)
func (f AuthorFilter) Update(msg tea.Msg) (AuthorFilter, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if f.cursor < len(f.authors)-1 {
				f.cursor++
			}
		case "k", "up":
			if f.cursor > 0 {
				f.cursor--
			}
		case " ", "space":
			// Toggle current author
			if f.cursor < len(f.authors) {
				name := f.authors[f.cursor].Name
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

func (f AuthorFilter) View() string {
	var lines []string
	lines = append(lines, TitleStyle.Render("Filter authors"))
	lines = append(lines, HintStyle.Render("space=toggle  a=all  n=none"))
	lines = append(lines, "")

	for i, a := range f.authors {
		checkbox := UncheckedStyle.Render("[ ]")
		if f.selected[a.Name] {
			checkbox = CheckedStyle.Render("[x]")
		}

		// Use display name (capitalize first letter of each word)
		displayName := strings.Title(a.Name)
		label := fmt.Sprintf("%s (%d)", displayName, a.Count)

		line := fmt.Sprintf("  %s %s", checkbox, label)
		if i == f.cursor {
			line = fmt.Sprintf("> %s %s", checkbox, label)
			line = SelectedStyle.Render(line)
		}
		lines = append(lines, line)
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

func (f AuthorFilter) SelectedEmails() []string {
	var result []string
	for _, a := range f.authors {
		if f.selected[a.Name] {
			result = append(result, a.Emails...)
		}
	}
	return result
}

func (f AuthorFilter) AllSelected() bool {
	for _, selected := range f.selected {
		if !selected {
			return false
		}
	}
	return true
}

// Reset resets the filter to show all authors
func (f *AuthorFilter) Reset() {
	for name := range f.selected {
		f.selected[name] = true
	}
}

// UpdateAuthors updates the author list (e.g., after repo refresh)
func (f *AuthorFilter) UpdateAuthors(commits []domain.Commit) {
	// Count commits per author name and collect emails
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

	// Build sorted author list
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

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Count > authors[j].Count
	})

	f.authors = authors

	// Add any new authors as selected
	for _, a := range authors {
		if _, exists := f.selected[a.Name]; !exists {
			f.selected[a.Name] = true
		}
	}

	// Remove authors that no longer exist
	authorSet := make(map[string]bool)
	for _, a := range authors {
		authorSet[a.Name] = true
	}
	for name := range f.selected {
		if !authorSet[name] {
			delete(f.selected, name)
		}
	}

	// Clamp cursor
	if f.cursor >= len(f.authors) {
		f.cursor = len(f.authors) - 1
	}
	if f.cursor < 0 {
		f.cursor = 0
	}
}

// SelectedCount returns the number of selected authors
func (f AuthorFilter) SelectedCount() int {
	count := 0
	for _, selected := range f.selected {
		if selected {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of authors
func (f AuthorFilter) TotalCount() int {
	return len(f.authors)
}
