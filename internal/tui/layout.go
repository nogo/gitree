package tui

import (
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
	return FooterStyle.Width(m.width).Render("[j/k] navigate  [g/G] top/bottom  [q]uit")
}
