package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/diff"
	"github.com/nogo/gitree/internal/tui/filtering"
	"github.com/nogo/gitree/internal/tui/histogram"
	"github.com/nogo/gitree/internal/tui/insights"
	"github.com/nogo/gitree/internal/tui/list"
	"github.com/nogo/gitree/internal/tui/search"
	"github.com/nogo/gitree/internal/watcher"
)

// Spinner frames for loading animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinnerTick returns a command that sends a tick after a delay
func spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

type Model struct {
	repo     *domain.Repository
	repoPath string
	reader   domain.GitReader
	list     list.Model
	diffView diff.DiffView
	filters  *filtering.Manager
	search   search.Search
	histogram           histogram.Histogram
	insights            insights.InsightsView
	watcher             *watcher.Watcher
	watching            bool
	showDiff            bool
	showBranchFilter    bool
	showAuthorFilter    bool
	showAuthorHighlight bool
	showTagFilter       bool
	showHelp            bool
	showInsights        bool
	insightsLoading     bool
	spinnerFrame        int
	width               int
	height              int
	ready               bool
	err                 error
}

func NewModel(repo *domain.Repository, repoPath string, w *watcher.Watcher, reader domain.GitReader) Model {
	return Model{
		repo:      repo,
		repoPath:  repoPath,
		reader:    reader,
		list:      list.New(repo),
		diffView:  diff.New(),
		filters:   filtering.New(repo),
		search:    search.New(),
		histogram: histogram.New(repo.Commits, 80), // default width, will resize
		insights:  insights.New(),
		watcher:   w,
		watching:  w != nil,
	}
}

func (m Model) Init() tea.Cmd {
	if m.watcher != nil {
		return m.watchForChanges()
	}
	return nil
}

// watchForChanges returns a command that waits for watcher signal
func (m Model) watchForChanges() tea.Cmd {
	return func() tea.Msg {
		<-m.watcher.Changes()
		return RepoChangedMsg{}
	}
}

// reloadRepo returns a command that reloads repository data
func (m Model) reloadRepo() tea.Cmd {
	reader := m.reader
	repoPath := m.repoPath
	return func() tea.Msg {
		repo, err := reader.LoadRepository(repoPath)
		return RepoLoadedMsg{Repo: repo, Err: err}
	}
}

// loadFileDiff returns a command that loads the diff for the current file
func (m Model) loadFileDiff() tea.Cmd {
	commit := m.list.SelectedCommit()
	if commit == nil {
		return nil
	}
	reader := m.reader
	filePath := m.diffView.CurrentFile()
	fileIndex := m.diffView.FileIndex()
	repoPath := m.repoPath
	commitHash := commit.Hash

	return func() tea.Msg {
		diff, isBinary, err := reader.LoadFileDiff(repoPath, commitHash, filePath)
		return DiffLoadedMsg{
			FilePath:  filePath,
			Diff:      diff,
			IsBinary:  isBinary,
			FileIndex: fileIndex,
			Err:       err,
		}
	}
}

// loadExpandedFiles returns a command that loads files for the expanded commit
func (m Model) loadExpandedFiles() tea.Cmd {
	commit := m.list.SelectedCommit()
	if commit == nil {
		return nil
	}
	reader := m.reader
	hash := commit.Hash
	path := m.repoPath
	return func() tea.Msg {
		files, err := reader.LoadFileChanges(path, hash)
		return ExpandedFilesLoadedMsg{Files: files, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle help overlay
	if m.showHelp {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc", "h", "q", "enter", " ":
				m.showHelp = false
			}
			return m, nil
		}
		return m, nil
	}

	// Handle branch filter overlay first
	if m.showBranchFilter {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			bf := m.filters.BranchFilter()
			var done, cancelled bool
			*bf, _, done, cancelled = bf.Update(keyMsg)
			if done {
				cmd := m.applyFilter()
				m.showBranchFilter = false
				return m, cmd
			}
			if cancelled {
				m.showBranchFilter = false
			}
			return m, nil
		}
		return m, nil
	}

	// Handle author filter overlay
	if m.showAuthorFilter {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			af := m.filters.AuthorFilter()
			var done, cancelled bool
			*af, _, done, cancelled = af.Update(keyMsg)
			if done {
				cmd := m.applyFilter()
				m.showAuthorFilter = false
				return m, cmd
			}
			if cancelled {
				m.showAuthorFilter = false
			}
			return m, nil
		}
		return m, nil
	}

	// Handle author highlight overlay
	if m.showAuthorHighlight {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			ah := m.filters.AuthorHighlight()
			var done, cancelled bool
			*ah, _, done, cancelled = ah.Update(keyMsg)
			if done {
				m.applyHighlight()
				m.showAuthorHighlight = false
			}
			if cancelled {
				m.showAuthorHighlight = false
			}
			return m, nil
		}
		return m, nil
	}

	// Handle tag filter overlay
	if m.showTagFilter {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			tf := m.filters.TagFilter()
			var done, cancelled bool
			*tf, _, done, cancelled = tf.Update(keyMsg)
			if done {
				cmd := m.applyFilter()
				m.showTagFilter = false
				return m, cmd
			}
			if cancelled {
				m.showTagFilter = false
			}
			return m, nil
		}
		return m, nil
	}

	// Handle search input mode
	if m.search.IsInputMode() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			var done, cancelled bool
			m.search, cmd, done, cancelled = m.search.Update(keyMsg)
			if done {
				m.executeSearch()
			}
			if cancelled {
				// Input cancelled, search state preserved
			}
			return m, cmd
		}
		return m, nil
	}

	// Handle histogram focus mode
	if m.histogram.IsFocused() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "tab":
				// Switch focus back to list
				m.histogram.SetFocused(false)
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				var selectionChanged bool
				m.histogram, _, selectionChanged = m.histogram.Update(keyMsg)
				if selectionChanged {
					cmd := m.applyTimeFilter()
					return m, cmd
				}
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case RepoChangedMsg:
		// Repo changed, trigger reload and re-arm watcher
		return m, tea.Batch(
			m.reloadRepo(),
			m.watchForChanges(),
		)

	case RepoLoadedMsg:
		if msg.Err == nil {
			m.repo = msg.Repo
			m.filters.UpdateRepo(msg.Repo)
			// Recalculate histogram
			m.histogram.Recalculate(msg.Repo.Commits, m.width)
			// Reapply filters if any active
			var cmd tea.Cmd
			if m.filters.BranchFilterActive() || m.filters.AuthorFilterActive() || m.filters.TagFilterActive() || m.filters.TimeFilterActive() {
				cmd = m.applyAllFilters()
			} else {
				m.list.SetRepo(msg.Repo)
				// Still need to reload insights if visible
				if m.showInsights {
					m.insightsLoading = true
					cmd = tea.Batch(m.loadInsights(), spinnerTick())
				}
			}
			// Reapply highlight if active
			if m.filters.AuthorHighlightActive() {
				m.applyHighlight()
			}
			// Re-execute search if active
			if m.search.IsActive() {
				m.executeSearch()
			}
			return m, cmd
		}
		return m, nil

	case ExpandedFilesLoadedMsg:
		if msg.Err == nil {
			m.list.SetExpandedFiles(msg.Files)
		} else {
			m.list.SetExpandedFilesError()
		}
		return m, nil

	case DiffLoadedMsg:
		if msg.Err == nil {
			m.diffView.SetDiff(msg.Diff, msg.IsBinary)
		} else {
			m.diffView.SetDiff("", false)
		}
		return m, nil

	case InsightsLoadedMsg:
		m.insightsLoading = false
		m.insights.SetSize(m.width, m.insightsContentHeight())
		m.insights.Recalculate(msg.Commits, msg.FileChanges)
		return m, nil

	case SpinnerTickMsg:
		if m.insightsLoading {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, spinnerTick()
		}
		return m, nil

	case tea.KeyMsg:
		// Handle diff view keys
		if m.showDiff {
			switch msg.String() {
			case "q", "esc":
				m.diffView.Hide()
				m.showDiff = false
				return m, nil
			case "h", "left":
				// Previous file
				if m.diffView.PrevFile() {
					return m, m.loadFileDiff()
				}
				return m, nil
			case "l", "right":
				// Next file
				if m.diffView.NextFile() {
					return m, m.loadFileDiff()
				}
				return m, nil
			}
			m.diffView, _ = m.diffView.Update(msg)
			return m, nil
		}

		// Handle expanded commit view keys
		if m.list.IsExpanded() {
			switch msg.String() {
			case "esc":
				m.list.Collapse()
				return m, nil
			case "enter":
				// Open diff for selected file
				if m.list.HasExpandedFiles() {
					m.diffView.Show(m.list.ExpandedFiles(), m.list.FileCursor())
					m.diffView.SetSize(m.width, m.height)
					m.showDiff = true
					return m, m.loadFileDiff()
				}
				return m, nil
			case "j", "down":
				// Navigate within file list
				m.list.FileCursorDown()
				return m, nil
			case "k", "up":
				// Navigate within file list
				m.list.FileCursorUp()
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			// Block all other keys (including commit navigation) when expanded
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedCommit()
			if selected != nil {
				m.list.Expand()
				return m, m.loadExpandedFiles()
			}
			return m, nil

		case "b":
			m.filters.BranchFilter().SetSize(m.width, m.height)
			m.showBranchFilter = true
			return m, nil

		case "a":
			m.filters.AuthorFilter().SetSize(m.width, m.height)
			m.showAuthorFilter = true
			return m, nil

		case "A":
			m.filters.AuthorHighlight().SetSize(m.width, m.height)
			m.showAuthorHighlight = true
			return m, nil

		case "t":
			m.filters.TagFilter().SetSize(m.width, m.height)
			m.showTagFilter = true
			return m, nil

		case "h":
			m.showHelp = true
			return m, nil

		case "/":
			m.search.Activate()
			return m, nil

		case "n":
			// Next search match
			if m.search.IsActive() && m.search.MatchCount() > 0 {
				m.search.NextMatch()
				m.jumpToCurrentMatch()
			}
			return m, nil

		case "N":
			// Previous search match
			if m.search.IsActive() && m.search.MatchCount() > 0 {
				m.search.PrevMatch()
				m.jumpToCurrentMatch()
			}
			return m, nil

		case "r":
			// Toggle range/histogram visibility
			m.histogram.Toggle()
			m.recalculateListHeight()
			m.insights.SetSize(m.width, m.insightsContentHeight())
			return m, nil

		case "i":
			// Toggle insights view
			m.showInsights = !m.showInsights
			if m.showInsights {
				m.insightsLoading = true
				return m, tea.Batch(m.loadInsights(), spinnerTick())
			}
			return m, nil

		case "tab":
			// Switch focus to histogram (if visible)
			if m.histogram.IsVisible() {
				m.histogram.SetFocused(true)
			}
			return m, nil

		case "c":
			// Clear all filters, highlight, and search
			m.filters.Reset()
			m.histogram.Reset()
			m.search.Clear()
			m.list.SetHighlightedEmails(nil)
			m.list.SetMatchIndices(nil)
			m.list.SetRepo(m.repo)
			// Recalculate histogram with all commits
			m.histogram.Recalculate(m.repo.Commits, m.width)
			// Reload insights if visible
			if m.showInsights {
				m.insightsLoading = true
				return m, tea.Batch(m.loadInsights(), spinnerTick())
			}
			return m, nil

		case "esc":
			// No action when not in detail/filter view
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Recalculate histogram with new width
		m.histogram.Recalculate(m.repo.Commits, msg.Width)
		m.recalculateListHeight()
		m.diffView.SetSize(msg.Width, msg.Height)
		m.filters.BranchFilter().SetSize(msg.Width, msg.Height)
		m.filters.AuthorFilter().SetSize(msg.Width, msg.Height)
		m.filters.AuthorHighlight().SetSize(msg.Width, msg.Height)
		m.filters.TagFilter().SetSize(msg.Width, msg.Height)
		m.insights.SetSize(msg.Width, m.insightsContentHeight())
	}

	// Route updates to list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) applyFilter() tea.Cmd {
	m.filters.UpdateFilterActive()
	return m.applyAllFilters()
}

func (m *Model) applyHighlight() {
	emails := m.filters.HighlightedEmails()
	m.list.SetHighlightedEmails(emails)
}

func (m *Model) recalculateListHeight() {
	// Header(1) + separator(1) + column headers(1) + separator(1) + footer(1) = 5 lines
	// Plus histogram height if visible
	histHeight := m.histogram.Height()
	contentHeight := m.height - 5 - histHeight
	if contentHeight < 1 {
		contentHeight = 1
	}
	m.list.SetSize(m.width, contentHeight)
}

func (m *Model) insightsContentHeight() int {
	// Header(1) + separator(1) + separator(1) + footer(1) = 4 lines (no column headers)
	// Plus histogram height if visible
	histHeight := m.histogram.Height()
	contentHeight := m.height - 4 - histHeight
	if contentHeight < 1 {
		contentHeight = 1
	}
	return contentHeight
}

func (m *Model) applyTimeFilter() tea.Cmd {
	start, end, hasSelection := m.histogram.SelectedRange()
	if !hasSelection {
		m.filters.ClearTimeFilter()
		// Reapply other filters without time constraint
		return m.applyAllFilters()
	}

	m.filters.SetTimeFilter(start, end, true)
	return m.applyAllFilters()
}

func (m *Model) applyAllFilters() tea.Cmd {
	result := m.filters.ApplyFilters()
	if result.IsFiltered {
		m.list.SetFilteredCommits(result.Commits, m.repo)
	} else {
		m.list.SetRepo(m.repo)
	}

	// Re-execute search if active
	if m.search.IsActive() {
		m.executeSearch()
	}

	// Reload insights if visible
	if m.showInsights {
		m.insightsLoading = true
		return tea.Batch(m.loadInsights(), spinnerTick())
	}
	return nil
}

// loadInsights returns a command that loads insights data asynchronously
func (m Model) loadInsights() tea.Cmd {
	// Capture values for the closure
	commits := m.list.Commits()
	reader := m.reader
	repoPath := m.repoPath

	return func() tea.Msg {
		// Convert to pointer slice for insights
		commitPtrs := make([]*domain.Commit, len(commits))
		for i := range commits {
			commitPtrs[i] = &commits[i]
		}

		// Limit file change loading for performance on large repos
		// Author stats use all commits (fast - just metadata)
		// File stats only need recent commits to be meaningful
		const maxFileChangeCommits = 200
		fileChangeLimit := len(commits)
		if fileChangeLimit > maxFileChangeCommits {
			fileChangeLimit = maxFileChangeCommits
		}

		// Load file changes in parallel for better performance
		type result struct {
			hash  string
			files []domain.FileChange
		}
		resultChan := make(chan result, fileChangeLimit)

		// Use worker pool to limit concurrency
		const maxWorkers = 20
		sem := make(chan struct{}, maxWorkers)

		for i := 0; i < fileChangeLimit; i++ {
			go func(commit domain.Commit) {
				sem <- struct{}{}        // acquire
				defer func() { <-sem }() // release

				files, err := reader.LoadFileChanges(repoPath, commit.Hash)
				if err == nil {
					resultChan <- result{hash: commit.Hash, files: files}
				} else {
					resultChan <- result{hash: commit.Hash, files: nil}
				}
			}(commits[i])
		}

		// Collect results
		fileChanges := make(map[string][]domain.FileChange)
		for i := 0; i < fileChangeLimit; i++ {
			r := <-resultChan
			if r.files != nil {
				fileChanges[r.hash] = r.files
			}
		}

		return InsightsLoadedMsg{
			Commits:     commitPtrs,
			FileChanges: fileChanges,
		}
	}
}


func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.showBranchFilter {
		return m.filters.BranchFilter().View()
	}
	if m.showAuthorFilter {
		return m.filters.AuthorFilter().View()
	}
	if m.showAuthorHighlight {
		return m.filters.AuthorHighlight().View()
	}
	if m.showTagFilter {
		return m.filters.TagFilter().View()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	if m.showInsights {
		return m.renderInsightsLayout()
	}
	if m.showDiff {
		return m.renderWithDiff()
	}
	return m.renderLayout()
}

// renderWithDiff shows the diff view as a centered overlay
func (m Model) renderWithDiff() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.diffView.View(),
	)
}

// Watching returns whether the watcher is active
func (m Model) Watching() bool {
	return m.watching
}

// BranchFilterActive returns whether a branch filter is currently applied
func (m Model) BranchFilterActive() bool {
	return m.filters.BranchFilterActive()
}

// AuthorFilterActive returns whether an author filter is currently applied
func (m Model) AuthorFilterActive() bool {
	return m.filters.AuthorFilterActive()
}

// FilteredBranchCount returns the number of branches in filter
func (m Model) FilteredBranchCount() int {
	return m.filters.SelectedBranchCount()
}

// TotalBranchCount returns the total number of branches
func (m Model) TotalBranchCount() int {
	return m.filters.TotalBranchCount()
}

// FilteredCommitCount returns the number of commits currently displayed
func (m Model) FilteredCommitCount() int {
	return m.list.CommitCount()
}

// TotalCommitCount returns the total number of commits in the repo
func (m Model) TotalCommitCount() int {
	return len(m.repo.Commits)
}

// FilteredAuthorCount returns the number of authors in filter
func (m Model) FilteredAuthorCount() int {
	return m.filters.SelectedAuthorCount()
}

// TotalAuthorCount returns the total number of authors
func (m Model) TotalAuthorCount() int {
	return m.filters.TotalAuthorCount()
}

// TagFilterActive returns whether a tag filter is currently applied
func (m Model) TagFilterActive() bool {
	return m.filters.TagFilterActive()
}

// FilteredTagCount returns the number of tags in filter
func (m Model) FilteredTagCount() int {
	return m.filters.SelectedTagCount()
}

// TotalTagCount returns the total number of tags
func (m Model) TotalTagCount() int {
	return m.filters.TotalTagCount()
}

// AuthorHighlightActive returns whether an author is currently highlighted
func (m Model) AuthorHighlightActive() bool {
	return m.filters.AuthorHighlightActive()
}

// HighlightedAuthorName returns the name of the highlighted author
func (m Model) HighlightedAuthorName() string {
	return m.filters.HighlightedAuthorName()
}

// SearchActive returns whether search is active
func (m Model) SearchActive() bool {
	return m.search.IsActive()
}

// SearchInputMode returns whether search input is active
func (m Model) SearchInputMode() bool {
	return m.search.IsInputMode()
}

// SearchQuery returns the current search query
func (m Model) SearchQuery() string {
	return m.search.Query()
}

// SearchMatchCount returns the number of search matches
func (m Model) SearchMatchCount() int {
	return m.search.MatchCount()
}

// SearchCurrentMatch returns the current match number (1-indexed)
func (m Model) SearchCurrentMatch() int {
	return m.search.CurrentMatch() + 1
}

// SearchInputView returns the search input view
func (m Model) SearchInputView() string {
	return m.search.InputView()
}

// executeSearch runs the search and updates the view
func (m *Model) executeSearch() {
	// Search on currently displayed commits (may be filtered)
	commits := m.list.Commits()
	m.search.Execute(commits)
	m.list.SetMatchIndices(m.search.Matches())
	m.jumpToCurrentMatch()
}

// jumpToCurrentMatch moves cursor to the current search match
func (m *Model) jumpToCurrentMatch() {
	idx := m.search.CurrentMatchCommitIndex()
	if idx >= 0 {
		m.list.SetCursor(idx)
	}
}

// TimeFilterActive returns whether a time filter is currently applied
func (m Model) TimeFilterActive() bool {
	return m.filters.TimeFilterActive()
}

// TimeFilterRange returns the formatted time range
func (m Model) TimeFilterRange() string {
	return m.filters.TimeFilterRange()
}

// HistogramVisible returns whether histogram is visible
func (m Model) HistogramVisible() bool {
	return m.histogram.IsVisible()
}

// HistogramFocused returns whether histogram is focused
func (m Model) HistogramFocused() bool {
	return m.histogram.IsFocused()
}

// HistogramView returns the histogram rendered view
func (m Model) HistogramView() string {
	return m.histogram.View()
}

// HistogramHeight returns the histogram height
func (m Model) HistogramHeight() int {
	return m.histogram.Height()
}

// SpinnerFrame returns the current spinner animation frame
func (m Model) SpinnerFrame() string {
	return spinnerFrames[m.spinnerFrame]
}

// InsightsLoading returns whether insights are currently loading
func (m Model) InsightsLoading() bool {
	return m.insightsLoading
}

// ApplyInitialFilters applies filters from CLI arguments
func (m *Model) ApplyInitialFilters(branch, author, tag string) {
	needsApply := false

	// Apply branch filter
	if branch != "" {
		m.filters.SetBranchFilter(branch)
		needsApply = true
	}

	// Apply author filter
	if author != "" {
		m.filters.SetAuthorFilter(author)
		needsApply = true
	}

	// Apply tag filter
	if tag != "" {
		m.filters.SetTagFilter(tag)
		needsApply = true
	}

	if needsApply {
		m.filters.UpdateFilterActive()
		result := m.filters.ApplyFilters()
		if result.IsFiltered {
			m.list.SetFilteredCommits(result.Commits, m.repo)
		}
	}
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	help := `Keyboard Shortcuts

 Navigation
   j/↓  k/↑       Move cursor
   Ctrl+d/u      Page down/up
   g/G           Jump to first/last
   Enter         Expand commit

 Filters
   a             Author filter
   b             Branch filter
   t             Tag filter
   A             Author highlight
   r             Range (histogram)
   c             Clear all filters

 Search
   /             Start search
   n/N           Next/prev match

 Histogram (when focused)
   h/l ←/→       Move selection
   Space         Toggle selection
   [/]           Set start/end
   +/-           Zoom in/out
   Enter         Apply filter
   Tab           Return to list

 General
   i             Insights view
   h             This help
   q             Quit

Press any key to close`

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		style.Render(help),
	)
}
