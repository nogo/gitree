package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLayout() string {
	header := m.renderHeader()
	separator := m.renderSeparator()
	columnHeaders := m.renderColumnHeaders()
	content := m.renderContent()
	footer := m.renderFooter()

	// Header(1) + separator(1) + column headers(1) + separator(1) + footer(1) = 5 lines
	contentHeight := m.height - 5
	if contentHeight < 0 {
		contentHeight = 0
	}

	content = lipgloss.NewStyle().
		Height(contentHeight).
		Render(content)

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
	// Get actual graph width from list
	graphWidth := m.list.GraphWidth()

	// Match column layout from list: cursor(2) + graph(dynamic) + space(1) + message(flex) + spacing(2) + author(12) + spacing(2) + date(10) + spacing(2) + hash(7)
	// Total fixed: 2 + graphWidth + 1 + 2 + 12 + 2 + 10 + 2 + 7 = 38 + graphWidth
	msgWidth := m.width - 38 - graphWidth
	if msgWidth < 10 {
		msgWidth = 10
	}

	// Prefix: cursor(2) + graph(graphWidth) + space(1)
	prefix := strings.Repeat(" ", 2+graphWidth+1)
	header := fmt.Sprintf("%s%-*s  %12s  %10s  %7s",
		prefix, msgWidth, "Message", "Author", "Date", "Hash")

	return ColumnHeaderStyle.Render(header)
}

func (m Model) renderContent() string {
	return m.list.View()
}

func (m Model) renderFooter() string {
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

	filterStats := ""
	if len(filterParts) > 0 {
		filterStats = "  " + strings.Join(filterParts, " ")
	}

	// Condensed keybindings
	keys := "[a]uthor [b]ranch [c]lear [q]"

	// Build footer with spacing
	left := fmt.Sprintf("%s %s%s", watchStatus, commitStats, filterStats)
	right := keys

	spacing := m.width - len(left) - len(right)
	if spacing < 2 {
		spacing = 2
	}

	return FooterStyle.Render(left + strings.Repeat(" ", spacing) + right)
}
