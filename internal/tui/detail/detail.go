package detail

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
)

type Model struct {
	commit *domain.Commit
	width  int
	height int
}

func New() Model {
	return Model{}
}

func (m *Model) SetCommit(c *domain.Commit) {
	m.commit = c
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	if m.commit == nil {
		return ""
	}

	c := m.commit

	// Format detail sections
	header := m.renderHeader(c)
	meta := m.renderMeta(c)
	message := m.renderMessage(c)
	parents := m.renderParents(c)
	refs := m.renderRefs(c)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		meta,
		"",
		message,
		"",
		parents,
	)

	if refs != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			refs,
		)
	}

	// Account for border and padding in dimensions
	innerWidth := m.width - 6
	innerHeight := m.height - 4
	if innerWidth < 20 {
		innerWidth = 20
	}
	if innerHeight < 10 {
		innerHeight = 10
	}

	return DetailStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(content)
}

func (m Model) renderHeader(c *domain.Commit) string {
	return HeaderStyle.Render("Commit " + c.Hash)
}

func (m Model) renderMeta(c *domain.Commit) string {
	author := fmt.Sprintf("%s  %s <%s>",
		LabelStyle.Render("Author:"),
		c.Author,
		c.Email,
	)
	date := fmt.Sprintf("%s  %s",
		LabelStyle.Render("Date:"),
		c.Date.Format("Mon Jan 2 15:04:05 2006 -0700"),
	)
	return author + "\n" + date
}

func (m Model) renderMessage(c *domain.Commit) string {
	return MessageStyle.Render(strings.TrimSpace(c.FullMessage))
}

func (m Model) renderParents(c *domain.Commit) string {
	label := LabelStyle.Render("Parents:")
	if len(c.Parents) == 0 {
		return fmt.Sprintf("%s  (none - initial commit)", label)
	}

	// Show short hashes for parents
	shortParents := make([]string, len(c.Parents))
	for i, p := range c.Parents {
		if len(p) > 7 {
			shortParents[i] = ParentStyle.Render(p[:7])
		} else {
			shortParents[i] = ParentStyle.Render(p)
		}
	}
	return fmt.Sprintf("%s  %s", label, strings.Join(shortParents, ", "))
}

func (m Model) renderRefs(c *domain.Commit) string {
	if len(c.BranchRefs) == 0 {
		return ""
	}
	label := LabelStyle.Render("Refs:")
	return fmt.Sprintf("%s  %s", label, strings.Join(c.BranchRefs, ", "))
}
