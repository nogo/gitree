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

// SetRepo updates the list with new repository data
func (m *Model) SetRepo(repo *domain.Repository) {
	// Preserve cursor position if possible
	oldCursor := m.cursor
	m.commits = repo.Commits
	m.graph = graph.NewRenderer(repo.Commits, repo.Branches, repo.HEAD)

	// Clamp cursor to new bounds
	if m.cursor >= len(m.commits) {
		m.cursor = len(m.commits) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// Keep same position if possible
	if oldCursor < len(m.commits) {
		m.cursor = oldCursor
	}

	// Update viewport content
	if m.ready {
		m.viewport.SetContent(m.renderList())
	}
}

// SetFilteredCommits updates the list with filtered commits while using
// the original repo's branches/HEAD for graph context
func (m *Model) SetFilteredCommits(commits []domain.Commit, repo *domain.Repository) {
	oldCursor := m.cursor
	m.commits = commits
	m.graph = graph.NewRenderer(commits, repo.Branches, repo.HEAD)

	// Clamp cursor to new bounds
	if m.cursor >= len(m.commits) {
		m.cursor = len(m.commits) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	if oldCursor < len(m.commits) {
		m.cursor = oldCursor
	}

	if m.ready {
		m.viewport.SetContent(m.renderList())
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
		// Vertical movement
		case "j", "down":
			m.cursorDown(1)
		case "k", "up":
			m.cursorUp(1)

		// Page movement
		case "ctrl+d", "pgdown":
			m.cursorDown(m.height / 2)
		case "ctrl+u", "pgup":
			m.cursorUp(m.height / 2)

		// Jump to edges
		case "g", "home":
			m.cursorTo(0)
		case "G", "end":
			m.cursorTo(len(m.commits) - 1)
		}
		m.syncViewport()
		m.viewport.SetContent(m.renderList())

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.cursorUp(3)
			m.syncViewport()
			m.viewport.SetContent(m.renderList())
		case tea.MouseWheelDown:
			m.cursorDown(3)
			m.syncViewport()
			m.viewport.SetContent(m.renderList())
		case tea.MouseLeft:
			// Click to select row (Y is relative to viewport)
			clickedRow := m.viewport.YOffset + msg.Y
			if clickedRow >= 0 && clickedRow < len(m.commits) {
				m.cursorTo(clickedRow)
				m.viewport.SetContent(m.renderList())
			}
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

func (m *Model) cursorUp(n int) {
	m.cursor = max(0, m.cursor-n)
}

func (m *Model) cursorDown(n int) {
	m.cursor = min(len(m.commits)-1, m.cursor+n)
}

func (m *Model) cursorTo(n int) {
	m.cursor = clamp(n, 0, len(m.commits)-1)
}

// syncViewport keeps cursor visible in viewport
func (m *Model) syncViewport() {
	// If cursor above visible area, scroll up
	if m.cursor < m.viewport.YOffset {
		m.viewport.SetYOffset(m.cursor)
	}
	// If cursor below visible area, scroll down
	if m.cursor >= m.viewport.YOffset+m.height {
		m.viewport.SetYOffset(m.cursor - m.height + 1)
	}
}

// SelectedCommit returns the currently selected commit
func (m Model) SelectedCommit() *domain.Commit {
	if m.cursor >= 0 && m.cursor < len(m.commits) {
		return &m.commits[m.cursor]
	}
	return nil
}

// CommitCount returns the number of commits in the list
func (m Model) CommitCount() int {
	return len(m.commits)
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

	// Cursor indicator (2 chars)
	cursor := "  "
	if selected {
		cursor = "> "
	}

	// Graph cell from renderer (dynamic width)
	graphCell := m.graph.RenderGraphCell(i)
	graphWidth := m.graph.Width()

	// Branch badges
	badges := m.graph.RenderBranchBadges(c)

	// Fixed width columns (rune-aware padding)
	authorTrunc := truncate(c.Author, 10)
	authorLen := len([]rune(authorTrunc))
	author := strings.Repeat(" ", 10-authorLen) + authorTrunc // right-align

	dateStr := formatRelativeTime(c.Date)
	dateLen := len(dateStr)
	date := strings.Repeat(" ", 10-dateLen) + dateStr // right-align

	hash := c.ShortHash[:5]

	// Message width: total - cursor(2) - graph(dynamic) - space(1) - spacing(2) - author(10) - spacing(2) - date(10) - spacing(2) - hash(5)
	// = width - 34 - graphWidth
	msgWidth := m.width - 34 - graphWidth
	if msgWidth < 10 {
		msgWidth = 10
	}

	// Account for badge width in message truncation
	badgeLen := runeLen(badges)
	msgAvail := msgWidth - badgeLen
	if msgAvail < 5 {
		msgAvail = 5
	}
	msg := badges + truncate(c.Message, msgAvail)

	// Pad message to fixed width for alignment
	msgLen := runeLen(msg)
	msgDisplay := msg + strings.Repeat(" ", msgWidth-msgLen)
	if msgLen > msgWidth {
		msgDisplay = msg
	}

	// Apply styles to individual parts (non-selected rows)
	if !selected {
		hash = HashStyle.Render(hash)
		author = AuthorStyle.Render(author)
		date = DateStyle.Render(date)
	}

	// New column order: cursor | graph | message | author | date | hash
	row := fmt.Sprintf("%s%s %s  %s  %s  %s",
		cursor, graphCell, msgDisplay, author, date, hash)

	if selected {
		return SelectedRowStyle.Width(m.width).Render(row)
	}
	return row
}

// truncate truncates a string to max runes, adding ellipsis if needed
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

// runeLen returns the display width of a string (rune count, excluding ANSI codes)
func runeLen(s string) int {
	count := 0
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
		count++
	}
	return count
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

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
