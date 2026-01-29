package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderLayout() string {
	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// Header and footer each take 1 line
	contentHeight := m.height - 2
	if contentHeight < 0 {
		contentHeight = 0
	}

	content = lipgloss.NewStyle().
		Height(contentHeight).
		Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m Model) renderHeader() string {
	title := "gitree"
	return HeaderStyle.Width(m.width).Render(title)
}

func (m Model) renderContent() string {
	return m.list.View()
}

func (m Model) renderFooter() string {
	var parts []string

	// Watch status
	watchStatus := "○"
	if m.watching {
		watchStatus = "●"
	}
	parts = append(parts, watchStatus)

	// Filter status
	if m.filterActive {
		filterStatus := fmt.Sprintf("branch:%d/%d", m.FilteredBranchCount(), m.TotalBranchCount())
		parts = append(parts, filterStatus)
	}

	// Key hints
	hints := "[j/k] nav  [b]ranch  [c]lear  [enter] details  [q]uit"
	parts = append(parts, hints)

	return FooterStyle.Width(m.width).Render(fmt.Sprintf("%s  %s", parts[0], joinParts(parts[1:])))
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "  "
		}
		result += p
	}
	return result
}
