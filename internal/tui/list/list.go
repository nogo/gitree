package list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/graph"
)

type Model struct {
	commits  []domain.Commit
	graph    *graph.Renderer
	viewport viewport.Model
	cursor   int
	width    int
	height   int
	ready    bool
}

func New(repo *domain.Repository) Model {
	return Model{
		commits: repo.Commits,
		graph:   graph.NewRenderer(repo.Commits, repo.Branches, repo.HEAD),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	if !m.ready {
		m.viewport = viewport.New(w, h)
		m.ready = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = h
	}
	m.viewport.SetContent(m.renderList())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.cursorDown()
			m.viewport.SetContent(m.renderList())
			m.ensureCursorVisible()
		case "k", "up":
			m.cursorUp()
			m.viewport.SetContent(m.renderList())
			m.ensureCursorVisible()
		case "g":
			m.cursor = 0
			m.viewport.SetContent(m.renderList())
			m.viewport.GotoTop()
		case "G":
			m.cursor = len(m.commits) - 1
			m.viewport.SetContent(m.renderList())
			m.viewport.GotoBottom()
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.viewport.View()
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) cursorDown() {
	if m.cursor < len(m.commits)-1 {
		m.cursor++
	}
}

func (m *Model) ensureCursorVisible() {
	// Ensure cursor row is visible in viewport
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	} else if m.cursor >= m.viewport.YOffset+m.height {
		m.viewport.SetYOffset(m.cursor - m.height + 1)
	}
}

func (m Model) renderList() string {
	var rows []string
	for i, c := range m.commits {
		rows = append(rows, m.renderRow(i, c))
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderRow(i int, c domain.Commit) string {
	selected := i == m.cursor

	// Graph cell from renderer
	graphCell := m.graph.RenderGraphCell(i)

	// Branch badges
	badges := m.graph.RenderBranchBadges(c)

	// Fixed width columns
	hash := truncate(c.ShortHash, 7)
	author := truncate(c.Author, 12)
	date := formatRelativeTime(c.Date)

	// Message with badges prepended
	msgWidth := m.width - 50
	if msgWidth < 10 {
		msgWidth = 10
	}
	// Account for badge width in message truncation
	badgeLen := len(stripAnsi(badges))
	msgAvail := msgWidth - badgeLen
	if msgAvail < 5 {
		msgAvail = 5
	}
	msg := badges + truncate(c.Message, msgAvail)

	// Apply styles to individual parts (non-selected rows)
	if !selected {
		hash = HashStyle.Render(hash)
		author = AuthorStyle.Render(author)
		date = DateStyle.Render(date)
		msg = MessageStyle.Render(msg)
	}

	row := fmt.Sprintf("%s %s  %-12s  %-8s  %s",
		graphCell, hash, author, date, msg)

	if selected {
		return SelectedRowStyle.Width(m.width).Render(row)
	}
	return row
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
