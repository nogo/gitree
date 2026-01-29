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
	// Match column layout from list: cursor(2) + graph(5) + space(1) + message(flex) + space(2) + author(10) + space(2) + date(10) + space(2) + hash(5)
	// Total fixed: 2 + 5 + 1 + 2 + 10 + 2 + 10 + 2 + 5 = 39
	msgWidth := m.width - 39
	if msgWidth < 10 {
		msgWidth = 10
	}

	// 8 spaces = cursor(2) + graph(5) + space(1)
	header := fmt.Sprintf("        %-*s  %10s  %10s  %5s",
		msgWidth, "Message", "Author", "Date", "Hash")

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

	// Branch stats
	branchCount := m.TotalBranchCount()
	branchStats := fmt.Sprintf("%d branches", branchCount)

	// Condensed keybindings
	keys := "[b]ranch [c]lear [?] [q]"

	// Build footer with spacing
	left := fmt.Sprintf("%s %s  %s", watchStatus, commitStats, branchStats)
	right := keys

	spacing := m.width - len(left) - len(right)
	if spacing < 2 {
		spacing = 2
	}

	return FooterStyle.Render(left + strings.Repeat(" ", spacing) + right)
}
