package list

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/graph"
	"github.com/nogo/gitree/internal/tui/text"
)

type Model struct {
	commits           []domain.Commit
	graph             *graph.Renderer
	layout            RowLayout // base layout for current width/graph
	viewportLayout    RowLayout // dynamic layout based on visible viewport
	cursor            int
	viewOffset        int // first visible row (virtual scrolling)
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

	m.recalculateLayout()
	m.syncViewport()
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

	m.recalculateLayout()
	m.syncViewport()
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ready = true
	m.recalculateLayout()
	m.syncViewport()
}

// recalculateLayout updates the cached layout based on current width and graph
func (m *Model) recalculateLayout() {
	graphLanes := 1
	if m.graph != nil {
		graphLanes = m.graph.Width() / 2 // Width() returns lanes * 2
	}
	m.layout = NewRowLayout(m.width, graphLanes)
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
		return m, nil

	case tea.MouseMsg:
		// Block mouse scrolling when expanded
		if m.expanded {
			return m, nil
		}
		switch msg.Type {
		case tea.MouseWheelUp:
			m.cursorUp(3)
			m.syncViewport()
			return m, nil
		case tea.MouseWheelDown:
			m.cursorDown(3)
			m.syncViewport()
			return m, nil
		case tea.MouseLeft:
			// Click to select row (Y is relative to viewport)
			clickedRow := m.viewOffset + msg.Y
			if clickedRow >= 0 && clickedRow < len(m.commits) {
				m.cursorTo(clickedRow)
				m.syncViewport()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.renderVisibleRows()
}

// renderVisibleRows renders only the rows visible in the viewport
func (m Model) renderVisibleRows() string {
	if len(m.commits) == 0 {
		return ""
	}

	// Calculate how many commit rows we can show
	availableRows := m.height
	if m.expanded {
		// Reserve space for expanded content
		availableRows -= expandedHeight
		if availableRows < 1 {
			availableRows = 1
		}
	}

	endRow := m.viewOffset + availableRows
	if endRow > len(m.commits) {
		endRow = len(m.commits)
	}

	// Calculate dynamic layout based on lanes needed in viewport
	viewportLayout := m.layoutForViewport(m.viewOffset, endRow)

	var rows []string
	for i := m.viewOffset; i < endRow; i++ {
		rows = append(rows, m.renderRowWithLayout(i, m.commits[i], viewportLayout))

		// Insert expanded content after selected row
		if m.expanded && i == m.cursor {
			commit := m.SelectedCommit()
			if commit != nil {
				expandedLines := m.renderExpandedWithLayout(commit, m.expandedFiles, m.fileCursor, m.fileScrollOffset, m.expandedLoading, viewportLayout)
				rows = append(rows, expandedLines...)
			}
		}
	}

	// Pad to fill viewport height with full-width empty rows
	// (ensures old content is cleared in terminals that don't auto-clear)
	emptyRow := strings.Repeat(" ", m.width)
	for len(rows) < m.height {
		rows = append(rows, emptyRow)
	}

	return strings.Join(rows, "\n")
}

// layoutForViewport calculates optimal layout for the visible range
func (m Model) layoutForViewport(start, end int) RowLayout {
	if m.graph == nil {
		return m.layout
	}

	// Get max lanes needed for this viewport
	viewportLanes := m.graph.MaxLanesInRange(start, end)

	// Use at least 1 lane, cap at MaxGraphWidth
	graphWidth := (viewportLanes + 1) * 2
	if graphWidth < 4 {
		graphWidth = 4
	}
	if graphWidth > MaxGraphWidth {
		graphWidth = MaxGraphWidth
	}

	// Tell the graph renderer the display width limit
	// This prevents drawing connections to lanes that would be truncated
	m.graph.SetDisplayWidth(graphWidth)

	// Calculate layout with viewport-specific graph width
	return NewRowLayoutWithGraph(m.width, graphWidth)
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

// syncViewport keeps cursor visible and centered when possible
func (m *Model) syncViewport() {
	if m.height <= 0 || len(m.commits) == 0 {
		return
	}

	middle := m.height / 2

	// Calculate total lines accounting for expansion
	totalLines := len(m.commits)
	if m.expanded {
		totalLines += expandedHeight
	}

	// Scroll up: cursor goes above visible area
	if m.cursor < m.viewOffset {
		m.viewOffset = m.cursor
	}

	// Scroll down: keep cursor visible
	viewNeeded := m.cursor
	if m.expanded {
		viewNeeded = m.cursor + expandedHeight
	}

	if viewNeeded >= m.viewOffset+m.height {
		m.viewOffset = viewNeeded - m.height + 1
	}

	// Try to keep cursor in middle zone
	if m.cursor > m.viewOffset+middle {
		m.viewOffset = m.cursor - middle
	}

	// Clamp to valid range
	maxOffset := totalLines - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.viewOffset = clamp(m.viewOffset, 0, maxOffset)
}

// SelectedCommit returns the currently selected commit
func (m Model) SelectedCommit() *domain.Commit {
	if m.cursor >= 0 && m.cursor < len(m.commits) {
		return &m.commits[m.cursor]
	}
	return nil
}

// GraphWidth returns the calculated graph column width (for external use like headers)
func (m Model) GraphWidth() int {
	return m.ViewportLayout().Graph
}

// Layout returns the base row layout
func (m Model) Layout() RowLayout {
	return m.layout
}

// ViewportLayout returns the dynamic layout based on visible commits
func (m Model) ViewportLayout() RowLayout {
	if m.graph == nil || len(m.commits) == 0 {
		return m.layout
	}

	endRow := m.viewOffset + m.height
	if endRow > len(m.commits) {
		endRow = len(m.commits)
	}

	return m.layoutForViewport(m.viewOffset, endRow)
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
}

// Commits returns the current commit list
func (m Model) Commits() []domain.Commit {
	return m.commits
}

// SetCursor sets the cursor position and syncs viewport
func (m *Model) SetCursor(pos int) {
	m.cursorTo(pos)
	m.syncViewport()
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
	m.syncViewport()
}

// SetExpandedFiles sets the file list for the expanded commit
func (m *Model) SetExpandedFiles(files []domain.FileChange) {
	m.expandedFiles = files
	m.expandedLoading = false
	m.fileCursor = 0
	m.fileScrollOffset = 0
	m.syncViewport()
}

// SetExpandedFilesError handles error loading files
func (m *Model) SetExpandedFilesError() {
	m.expandedFiles = nil
	m.expandedLoading = false
	m.syncViewport()
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
	}
}

func (m Model) renderRow(i int, c domain.Commit) string {
	return m.renderRowWithLayout(i, c, m.layout)
}

// renderRowWithLayout renders a row using a specific layout (for viewport-specific sizing)
func (m Model) renderRowWithLayout(i int, c domain.Commit, layout RowLayout) string {
	selected := i == m.cursor
	isMatch := m.matchIndices != nil && m.matchIndices[i]

	// Build row data
	row := m.buildRowWithLayout(i, c, isMatch, layout)

	// Determine styling
	dimmed := false
	if m.highlightedEmails != nil {
		dimmed = !m.highlightedEmails[strings.ToLower(c.Email)]
	}

	style := RowStyle{
		Selected: selected,
		Dimmed:   dimmed,
		Width:    m.width,
	}

	return row.Render(layout, style)
}

// buildRow creates a Row with all column data for a commit.
func (m Model) buildRow(i int, c domain.Commit, isMatch bool) Row {
	return m.buildRowWithLayout(i, c, isMatch, m.layout)
}

// buildRowWithLayout creates a Row using a specific layout.
func (m Model) buildRowWithLayout(i int, c domain.Commit, isMatch bool, layout RowLayout) Row {
	selected := i == m.cursor

	// Cursor indicator
	cursor := "  "
	if selected && isMatch {
		cursor = ">*"
	} else if selected {
		cursor = "> "
	} else if isMatch {
		cursor = " *"
	}

	// Graph cell
	graphCell := m.graph.RenderGraphCell(i)
	if m.highlightedEmails != nil && !m.highlightedEmails[strings.ToLower(c.Email)] {
		graphCell = m.graph.RenderGraphCellDimmed(i)
	}

	// Message with badges (branches first, then tags)
	badges := m.graph.RenderBranchBadges(c) + m.graph.RenderTagBadges(c)
	badgeLen := text.Width(badges)
	msgAvail := layout.Message - badgeLen
	if msgAvail < 5 {
		msgAvail = 5
	}
	message := badges + text.Truncate(c.Message, msgAvail)

	return Row{
		Cursor:  cursor,
		Graph:   graphCell,
		Message: message,
		Author:  text.Truncate(c.Author, layout.Author),
		Date:    formatRelativeTime(c.Date),
		Hash:    c.ShortHash,
	}
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
