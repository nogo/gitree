package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/tui/text"
)

func (m Model) renderLayout() string {
	header := m.renderHeader()
	separator := m.renderSeparator()
	columnHeaders := m.renderColumnHeaders()
	content := m.renderContent()
	histogram := m.renderHistogram()
	footer := m.renderFooter()

	if histogram != "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			separator,
			columnHeaders,
			content,
			separator,
			histogram,
			footer,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		separator,
		columnHeaders,
		content,
		separator,
		footer,
	)
}

func (m Model) renderHeader() string {
	title := HeaderStyle.Render("gitree")
	repoName := HeaderDimStyle.Render(filepath.Base(m.repoPath))

	// Calculate spacing to right-align repo name
	titleLen := len("gitree")
	repoLen := len(filepath.Base(m.repoPath))
	spacing := m.width - titleLen - repoLen
	if spacing < 1 {
		spacing = 1
	}

	return title + strings.Repeat(" ", spacing) + repoName
}

func (m Model) renderSeparator() string {
	return SeparatorStyle.Render(strings.Repeat("─", m.width))
}

func (m Model) renderColumnHeaders() string {
	// Use viewport layout for consistent column widths with content
	layout := m.list.ViewportLayout()

	// Build header row: cursor(empty) | Graph | Message | Author | Date | Hash
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", layout.Cursor)) // cursor column (empty)
	b.WriteString(text.Fit("Graph", layout.Graph))
	b.WriteString(" ")
	b.WriteString(text.Fit("Message", layout.Message))
	b.WriteString("  ")
	b.WriteString(text.FitLeft("Author", layout.Author))
	b.WriteString("  ")
	b.WriteString(text.FitLeft("Date", layout.Date))
	b.WriteString("  ")
	b.WriteString(text.FitLeft("Hash", layout.Hash))

	return ColumnHeaderStyle.Render(b.String())
}

func (m Model) renderHistogram() string {
	if !m.HistogramVisible() {
		return ""
	}
	return m.HistogramView()
}

func (m Model) renderContent() string {
	return m.list.View()
}

func (m Model) renderFooter() string {
	// Search input mode - show search box in footer
	if m.SearchInputMode() {
		left := "Search: " + m.SearchInputView()
		right := "[Enter] [Esc]"

		spacing := m.width - len("Search: ") - 40 - len(right) // textinput width is 40
		if spacing < 2 {
			spacing = 2
		}

		return FooterStyle.Render(left + strings.Repeat(" ", spacing) + right)
	}

	// Watch status indicator
	watchStatus := "○"
	if m.watching {
		watchStatus = "●"
	}

	// Commit stats
	filtered := m.FilteredCommitCount()
	total := m.TotalCommitCount()
	commitStats := fmt.Sprintf("%d/%d commits", filtered, total)

	// Filter stats
	var filterParts []string

	// Branch filter status
	if m.BranchFilterActive() {
		branchFiltered := m.FilteredBranchCount()
		branchTotal := m.TotalBranchCount()
		filterParts = append(filterParts, fmt.Sprintf("branch:%d/%d", branchFiltered, branchTotal))
	}

	// Author filter status
	if m.AuthorFilterActive() {
		authorFiltered := m.FilteredAuthorCount()
		authorTotal := m.TotalAuthorCount()
		filterParts = append(filterParts, fmt.Sprintf("author:%d/%d", authorFiltered, authorTotal))
	}

	// Author highlight status
	if m.AuthorHighlightActive() {
		filterParts = append(filterParts, fmt.Sprintf("highlight:%s", m.HighlightedAuthorName()))
	}

	// Time filter status
	if m.TimeFilterActive() {
		filterParts = append(filterParts, m.TimeFilterRange())
	}

	// Search status
	if m.SearchActive() {
		matchCount := m.SearchMatchCount()
		if matchCount > 0 {
			filterParts = append(filterParts, fmt.Sprintf("match %d/%d \"%s\"",
				m.SearchCurrentMatch(), matchCount, m.SearchQuery()))
		} else {
			filterParts = append(filterParts, fmt.Sprintf("no matches \"%s\"", m.SearchQuery()))
		}
	}

	filterStats := ""
	if len(filterParts) > 0 {
		filterStats = "  " + strings.Join(filterParts, " ")
	}

	// Condensed keybindings - context-sensitive
	var keys string
	if m.HistogramFocused() {
		keys = "[←→]nav [+/-]zoom [[]start []]end [enter]apply [esc]back"
	} else if m.SearchActive() && m.SearchMatchCount() > 0 {
		keys = "[n]ext [N]prev [t]ime [c]lear [q]"
	} else {
		keys = "[/]search [a]uthor [b]ranch [t]ime [tab]timeline [c]lear [q]"
	}

	// Build footer with spacing
	left := fmt.Sprintf("%s %s%s", watchStatus, commitStats, filterStats)
	right := keys

	spacing := m.width - len(left) - len(right)
	if spacing < 2 {
		spacing = 2
	}

	return FooterStyle.Render(left + strings.Repeat(" ", spacing) + right)
}
