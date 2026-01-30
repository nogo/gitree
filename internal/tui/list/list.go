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
	commits           []domain.Commit
	graph             *graph.Renderer
	viewport          viewport.Model
	cursor            int
	width             int
	height            int
	ready             bool
	highlightedEmails map[string]bool // emails to highlight (nil = no highlight)
	matchIndices      map[int]bool    // indices of search matches (nil = no search)

	// Expansion state
	expanded         bool                 // whether a commit is expanded
	expandedFiles    []domain.FileChange  // files for expanded commit
	expandedLoading  bool                 // loading files
	fileCursor       int                  // cursor within file list
	fileScrollOffset int                  // scroll offset for file list
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

	// Collapse expansion on repo change
	m.expanded = false
	m.expandedFiles = nil
	m.expandedLoading = false

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

	// Collapse expansion on filter change
	m.expanded = false
	m.expandedFiles = nil
	m.expandedLoading = false

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
		m.viewport.SetContent(m.renderList())
		m.syncViewport()
		// Don't pass navigation keys to viewport - we handle scrolling ourselves
		return m, nil

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			m.cursorUp(3)
			m.viewport.SetContent(m.renderList())
			m.syncViewport()
			return m, nil
		case tea.MouseWheelDown:
			m.cursorDown(3)
			m.viewport.SetContent(m.renderList())
			m.syncViewport()
			return m, nil
		case tea.MouseLeft:
			// Click to select row (Y is relative to viewport)
			clickedRow := m.viewport.YOffset + msg.Y
			if clickedRow >= 0 && clickedRow < len(m.commits) {
				m.cursorTo(clickedRow)
				m.viewport.SetContent(m.renderList())
				m.syncViewport()
			}
			return m, nil
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

// syncViewport keeps cursor in the middle zone when scrolling
func (m *Model) syncViewport() {
	offset := m.viewport.YOffset
	middle := m.height / 2

	// Calculate total lines accounting for expansion
	totalLines := len(m.commits)
	cursorLine := m.cursor
	if m.expanded {
		totalLines += expandedHeight
		// Cursor line stays the same, but content after it is shifted
	}

	// Scroll up: only if cursor goes above visible area
	if cursorLine < offset {
		offset = cursorLine
	}

	// Scroll down: keep cursor in upper half once it passes middle
	// If expanded, we want to see the expanded content
	viewNeeded := cursorLine
	if m.expanded {
		viewNeeded = cursorLine + expandedHeight
	}

	if viewNeeded > offset+m.height-1 {
		offset = viewNeeded - m.height + 1
	}

	if cursorLine > offset+middle {
		offset = cursorLine - middle
	}

	// Clamp to valid range
	maxOffset := totalLines - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset = clamp(offset, 0, maxOffset)

	m.viewport.SetYOffset(offset)
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

// SetHighlightedEmails sets which author emails to highlight (nil = no highlight)
func (m *Model) SetHighlightedEmails(emails []string) {
	if len(emails) == 0 {
		m.highlightedEmails = nil
	} else {
		m.highlightedEmails = make(map[string]bool)
		for _, e := range emails {
			m.highlightedEmails[strings.ToLower(e)] = true
		}
	}
	if m.ready {
		m.viewport.SetContent(m.renderList())
	}
}

// SetMatchIndices sets which commit indices are search matches (nil = no search)
func (m *Model) SetMatchIndices(indices []int) {
	if len(indices) == 0 {
		m.matchIndices = nil
	} else {
		m.matchIndices = make(map[int]bool)
		for _, i := range indices {
			m.matchIndices[i] = true
		}
	}
	if m.ready {
		m.viewport.SetContent(m.renderList())
	}
}

// Commits returns the current commit list
func (m Model) Commits() []domain.Commit {
	return m.commits
}

// SetCursor sets the cursor position and syncs viewport
func (m *Model) SetCursor(pos int) {
	m.cursorTo(pos)
	if m.ready {
		m.viewport.SetContent(m.renderList())
		m.syncViewport()
	}
}

// GraphWidth returns the current graph column width
func (m Model) GraphWidth() int {
	if m.graph == nil {
		return 2 // minimum
	}
	return m.graph.Width()
}

// Expansion methods

// IsExpanded returns whether a commit is currently expanded
func (m Model) IsExpanded() bool {
	return m.expanded
}

// Expand expands the currently selected commit
func (m *Model) Expand() {
	m.expanded = true
	m.expandedLoading = true
	m.expandedFiles = nil
	m.fileCursor = 0
	m.fileScrollOffset = 0
}

// Collapse collapses the expanded commit
func (m *Model) Collapse() {
	m.expanded = false
	m.expandedLoading = false
	m.expandedFiles = nil
	m.fileCursor = 0
	m.fileScrollOffset = 0
	if m.ready {
		m.viewport.SetContent(m.renderList())
		m.syncViewport()
	}
}

// SetExpandedFiles sets the file list for the expanded commit
func (m *Model) SetExpandedFiles(files []domain.FileChange) {
	m.expandedFiles = files
	m.expandedLoading = false
	m.fileCursor = 0
	m.fileScrollOffset = 0
	if m.ready {
		m.viewport.SetContent(m.renderList())
		m.syncViewport()
	}
}

// SetExpandedFilesError handles error loading files
func (m *Model) SetExpandedFilesError() {
	m.expandedFiles = nil
	m.expandedLoading = false
	if m.ready {
		m.viewport.SetContent(m.renderList())
		m.syncViewport()
	}
}

// FileCursor returns the current file cursor position
func (m Model) FileCursor() int {
	return m.fileCursor
}

// ExpandedFiles returns the files for the expanded commit
func (m Model) ExpandedFiles() []domain.FileChange {
	return m.expandedFiles
}

// HasExpandedFiles returns true if there are files in the expanded view
func (m Model) HasExpandedFiles() bool {
	return len(m.expandedFiles) > 0
}

// SelectedFile returns the selected file in expanded view
func (m Model) SelectedFile() *domain.FileChange {
	if m.fileCursor >= 0 && m.fileCursor < len(m.expandedFiles) {
		return &m.expandedFiles[m.fileCursor]
	}
	return nil
}

// FileCursorUp moves file cursor up
func (m *Model) FileCursorUp() {
	if m.fileCursor > 0 {
		m.fileCursor--
		// Adjust scroll offset
		if m.fileCursor < m.fileScrollOffset {
			m.fileScrollOffset = m.fileCursor
		}
		if m.ready {
			m.viewport.SetContent(m.renderList())
		}
	}
}

// FileCursorDown moves file cursor down
func (m *Model) FileCursorDown() {
	if m.fileCursor < len(m.expandedFiles)-1 {
		m.fileCursor++
		// Adjust scroll offset
		if m.fileCursor >= m.fileScrollOffset+maxVisibleFiles {
			m.fileScrollOffset = m.fileCursor - maxVisibleFiles + 1
		}
		if m.ready {
			m.viewport.SetContent(m.renderList())
		}
	}
}

func (m Model) renderList() string {
	var rows []string
	for i, c := range m.commits {
		rows = append(rows, m.renderRow(i, c))

		// Insert expanded content after selected row
		if m.expanded && i == m.cursor {
			commit := m.SelectedCommit()
			if commit != nil {
				expandedLines := m.renderExpanded(commit, m.expandedFiles, m.fileCursor, m.fileScrollOffset)
				rows = append(rows, expandedLines...)
			}
		}
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderRow(i int, c domain.Commit) string {
	selected := i == m.cursor
	isMatch := m.matchIndices != nil && m.matchIndices[i]

	// Determine if this commit should be dimmed (highlight active but not matching)
	dimmed := false
	if m.highlightedEmails != nil {
		email := strings.ToLower(c.Email)
		dimmed = !m.highlightedEmails[email]
	}

	// Cursor indicator (2 chars): combines cursor (>) and match marker (*)
	cursor := "  "
	if selected && isMatch {
		cursor = ">*"
	} else if selected {
		cursor = "> "
	} else if isMatch {
		cursor = " *"
	}

	// Graph cell from renderer (dynamic width)
	graphCell := m.graph.RenderGraphCell(i)
	if dimmed {
		graphCell = m.graph.RenderGraphCellDimmed(i)
	}
	graphWidth := m.graph.Width()

	// Branch badges
	badges := m.graph.RenderBranchBadges(c)

	// Fixed width columns (rune-aware padding)
	authorWidth := 12
	authorTrunc := truncate(c.Author, authorWidth)
	authorLen := len([]rune(authorTrunc))
	author := strings.Repeat(" ", authorWidth-authorLen) + authorTrunc // right-align

	dateStr := formatRelativeTime(c.Date)
	dateLen := len(dateStr)
	date := strings.Repeat(" ", 10-dateLen) + dateStr // right-align

	hash := c.ShortHash // full 7 chars

	// Message width: total - cursor(2) - graph(dynamic) - space(1) - spacing(2) - author(12) - spacing(2) - date(10) - spacing(2) - hash(7)
	// = width - 38 - graphWidth
	msgWidth := m.width - 38 - graphWidth
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
		if dimmed {
			hash = DimmedHashStyle.Render(hash)
			author = DimmedAuthorStyle.Render(author)
			date = DimmedDateStyle.Render(date)
			msgDisplay = DimmedMessageStyle.Render(msgDisplay)
		} else {
			hash = HashStyle.Render(hash)
			author = AuthorStyle.Render(author)
			date = DateStyle.Render(date)
		}
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
