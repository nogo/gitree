package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui/detail"
	"github.com/nogo/gitree/internal/tui/diff"
	"github.com/nogo/gitree/internal/tui/filter"
	"github.com/nogo/gitree/internal/tui/histogram"
	"github.com/nogo/gitree/internal/tui/list"
	"github.com/nogo/gitree/internal/tui/search"
	"github.com/nogo/gitree/internal/watcher"
)

type Model struct {
	repo                *domain.Repository
	repoPath            string
	list                list.Model
	detail              detail.Model
	diffView            diff.DiffView
	branchFilter        filter.BranchFilter
	authorFilter        filter.AuthorFilter
	authorHighlight     filter.AuthorHighlight
	search              search.Search
	histogram           histogram.Histogram
	watcher             *watcher.Watcher
	watching            bool
	showDetail          bool
	showDiff            bool
	showBranchFilter    bool
	showAuthorFilter    bool
	showAuthorHighlight bool
	branchFilterActive  bool
	authorFilterActive  bool
	timeFilterActive    bool
	timeFilterStart     time.Time
	timeFilterEnd       time.Time
	width               int
	height              int
	ready               bool
	err                 error
}

func NewModel(repo *domain.Repository, repoPath string, w *watcher.Watcher) Model {
	d := detail.New()
	d.SetRepoPath(repoPath)
	return Model{
		repo:            repo,
		repoPath:        repoPath,
		list:            list.New(repo),
		detail:          d,
		diffView:        diff.New(),
		branchFilter:    filter.NewBranchFilter(repo.Branches),
		authorFilter:    filter.NewAuthorFilter(repo.Commits),
		authorHighlight: filter.NewAuthorHighlight(repo.Commits),
		search:          search.New(),
		histogram:       histogram.New(repo.Commits, 80), // default width, will resize
		watcher:         w,
		watching:        w != nil,
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
	return func() tea.Msg {
		reader := git.NewReader()
		repo, err := reader.LoadRepository(m.repoPath)
		return RepoLoadedMsg{Repo: repo, Err: err}
	}
}

// loadFileDiff returns a command that loads the diff for the current file
func (m Model) loadFileDiff() tea.Cmd {
	commit := m.detail.Commit()
	if commit == nil {
		return nil
	}
	filePath := m.diffView.CurrentFile()
	fileIndex := m.diffView.FileIndex()
	repoPath := m.repoPath
	commitHash := commit.Hash

	return func() tea.Msg {
		reader := git.NewReader()
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle branch filter overlay first
	if m.showBranchFilter {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var done, cancelled bool
			m.branchFilter, _, done, cancelled = m.branchFilter.Update(keyMsg)
			if done {
				m.applyFilter()
				m.showBranchFilter = false
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
			var done, cancelled bool
			m.authorFilter, _, done, cancelled = m.authorFilter.Update(keyMsg)
			if done {
				m.applyFilter()
				m.showAuthorFilter = false
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
			var done, cancelled bool
			m.authorHighlight, _, done, cancelled = m.authorHighlight.Update(keyMsg)
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
					m.applyTimeFilter()
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
			m.branchFilter.UpdateBranches(msg.Repo.Branches)
			m.authorFilter.UpdateAuthors(msg.Repo.Commits)
			m.authorHighlight.UpdateAuthors(msg.Repo.Commits)
			// Recalculate histogram
			m.histogram.Recalculate(msg.Repo.Commits, m.width)
			// Reapply filters if any active
			if m.branchFilterActive || m.authorFilterActive || m.timeFilterActive {
				m.applyAllFilters()
			} else {
				m.list.SetRepo(msg.Repo)
			}
			// Reapply highlight if active
			if m.authorHighlight.IsActive() {
				m.applyHighlight()
			}
			// Re-execute search if active
			if m.search.IsActive() {
				m.executeSearch()
			}
		}
		return m, nil

	case detail.FileChangesLoadedMsg:
		if msg.Err == nil {
			m.detail.SetFiles(msg.Files)
		} else {
			m.detail.SetFilesError()
		}
		return m, nil

	case DiffLoadedMsg:
		if msg.Err == nil {
			m.diffView.SetDiff(msg.Diff, msg.IsBinary)
		} else {
			m.diffView.SetDiff("", false)
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

		// Handle detail view keys
		if m.showDetail {
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				m.showDetail = false
				return m, nil
			case "enter":
				// Open diff for selected file
				if m.detail.HasFiles() {
					m.diffView.Show(m.detail.Files(), m.detail.FileCursor())
					m.diffView.SetSize(m.width, m.height)
					m.showDiff = true
					return m, m.loadFileDiff()
				}
				return m, nil
			}
			m.detail, _ = m.detail.Update(msg)
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			selected := m.list.SelectedCommit()
			if selected != nil {
				m.detail.SetCommit(selected)
				m.detail.SetSize(m.width, m.height)
				m.showDetail = true
				return m, m.detail.LoadFilesCmd()
			}
			return m, nil

		case "b":
			m.branchFilter.SetSize(m.width, m.height)
			m.showBranchFilter = true
			return m, nil

		case "a":
			m.authorFilter.SetSize(m.width, m.height)
			m.showAuthorFilter = true
			return m, nil

		case "A":
			m.authorHighlight.SetSize(m.width, m.height)
			m.showAuthorHighlight = true
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

		case "t":
			// Toggle histogram visibility
			m.histogram.Toggle()
			m.recalculateListHeight()
			return m, nil

		case "tab":
			// Switch focus to histogram (if visible)
			if m.histogram.IsVisible() {
				m.histogram.SetFocused(true)
			}
			return m, nil

		case "c":
			// Clear all filters, highlight, and search
			m.branchFilter.Reset()
			m.authorFilter.Reset()
			m.authorHighlight.Reset()
			m.histogram.Reset()
			m.search.Clear()
			m.branchFilterActive = false
			m.authorFilterActive = false
			m.timeFilterActive = false
			m.timeFilterStart = time.Time{}
			m.timeFilterEnd = time.Time{}
			m.list.SetHighlightedEmails(nil)
			m.list.SetMatchIndices(nil)
			m.list.SetRepo(m.repo)
			// Recalculate histogram with all commits
			m.histogram.Recalculate(m.repo.Commits, m.width)
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
		m.detail.SetSize(msg.Width, msg.Height)
		m.diffView.SetSize(msg.Width, msg.Height)
		m.branchFilter.SetSize(msg.Width, msg.Height)
		m.authorFilter.SetSize(msg.Width, msg.Height)
		m.authorHighlight.SetSize(msg.Width, msg.Height)
	}

	// Route updates to list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) applyFilter() {
	m.branchFilterActive = !m.branchFilter.AllSelected()
	m.authorFilterActive = !m.authorFilter.AllSelected()
	m.applyAllFilters()
}

func (m *Model) applyHighlight() {
	emails := m.authorHighlight.HighlightedEmails()
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

func (m *Model) applyTimeFilter() {
	start, end, hasSelection := m.histogram.SelectedRange()
	if !hasSelection {
		m.timeFilterActive = false
		m.timeFilterStart = time.Time{}
		m.timeFilterEnd = time.Time{}
		// Reapply other filters without time constraint
		m.applyAllFilters()
		return
	}

	m.timeFilterActive = true
	m.timeFilterStart = start
	m.timeFilterEnd = end
	m.applyAllFilters()
}

func (m *Model) applyAllFilters() {
	// Start with all commits
	filtered := m.repo.Commits

	// Apply branch filter
	if !m.branchFilter.AllSelected() {
		selectedBranches := m.branchFilter.SelectedBranches()
		filtered = m.filterCommitsByBranch(filtered, selectedBranches)
	}

	// Apply author filter
	if !m.authorFilter.AllSelected() {
		selectedEmails := m.authorFilter.SelectedEmails()
		filtered = m.filterCommitsByAuthor(filtered, selectedEmails)
	}

	// Apply time filter
	if m.timeFilterActive {
		filtered = m.filterCommitsByTime(filtered, m.timeFilterStart, m.timeFilterEnd)
	}

	if len(filtered) == len(m.repo.Commits) {
		m.list.SetRepo(m.repo)
	} else {
		m.list.SetFilteredCommits(filtered, m.repo)
	}

	// Re-execute search if active
	if m.search.IsActive() {
		m.executeSearch()
	}
}

func (m Model) filterCommitsByTime(commits []domain.Commit, start, end time.Time) []domain.Commit {
	var result []domain.Commit
	for _, c := range commits {
		if !c.Date.Before(start) && c.Date.Before(end) {
			result = append(result, c)
		}
	}
	return result
}

func (m Model) filterCommitsByBranch(commits []domain.Commit, branchNames []string) []domain.Commit {
	if len(branchNames) == 0 {
		return []domain.Commit{}
	}

	// Build set of selected branch names
	branchSet := make(map[string]bool)
	for _, name := range branchNames {
		branchSet[name] = true
	}

	// Find head hashes for selected branches
	var headHashes []string
	for _, b := range m.repo.Branches {
		if branchSet[b.Name] {
			headHashes = append(headHashes, b.HeadHash)
		}
	}

	// Build hash → commit map for parent lookup
	commitMap := make(map[string]*domain.Commit)
	for i := range m.repo.Commits {
		commitMap[m.repo.Commits[i].Hash] = &m.repo.Commits[i]
	}

	// BFS to find all reachable commits
	reachable := make(map[string]bool)
	queue := headHashes

	for len(queue) > 0 {
		hash := queue[0]
		queue = queue[1:]

		if reachable[hash] {
			continue
		}
		reachable[hash] = true

		if commit, ok := commitMap[hash]; ok {
			for _, parentHash := range commit.Parents {
				if !reachable[parentHash] {
					queue = append(queue, parentHash)
				}
			}
		}
	}

	// Filter commits maintaining order
	var result []domain.Commit
	for _, c := range commits {
		if reachable[c.Hash] {
			result = append(result, c)
		}
	}

	return result
}

func (m Model) filterCommitsByAuthor(commits []domain.Commit, emails []string) []domain.Commit {
	if len(emails) == 0 {
		return []domain.Commit{}
	}

	// emails are already normalized to lowercase from AuthorFilter
	emailSet := make(map[string]bool)
	for _, e := range emails {
		emailSet[e] = true
	}

	var result []domain.Commit
	for _, c := range commits {
		// Normalize commit email for comparison
		if emailSet[strings.ToLower(c.Email)] {
			result = append(result, c)
		}
	}

	return result
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.showBranchFilter {
		return m.branchFilter.View()
	}
	if m.showAuthorFilter {
		return m.authorFilter.View()
	}
	if m.showAuthorHighlight {
		return m.authorHighlight.View()
	}
	if m.showDiff {
		return m.renderWithDiff()
	}
	if m.showDetail {
		return m.renderWithDetail()
	}
	return m.renderLayout()
}

// renderWithDetail shows the detail view as a centered overlay
func (m Model) renderWithDetail() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		m.detail.View(),
	)
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
	return m.branchFilterActive
}

// AuthorFilterActive returns whether an author filter is currently applied
func (m Model) AuthorFilterActive() bool {
	return m.authorFilterActive
}

// FilteredBranchCount returns the number of branches in filter
func (m Model) FilteredBranchCount() int {
	return len(m.branchFilter.SelectedBranches())
}

// TotalBranchCount returns the total number of branches
func (m Model) TotalBranchCount() int {
	return len(m.repo.Branches)
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
	return m.authorFilter.SelectedCount()
}

// TotalAuthorCount returns the total number of authors
func (m Model) TotalAuthorCount() int {
	return m.authorFilter.TotalCount()
}

// AuthorHighlightActive returns whether an author is currently highlighted
func (m Model) AuthorHighlightActive() bool {
	return m.authorHighlight.IsActive()
}

// HighlightedAuthorName returns the name of the highlighted author
func (m Model) HighlightedAuthorName() string {
	return m.authorHighlight.HighlightedName()
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
	return m.timeFilterActive
}

// TimeFilterRange returns the formatted time range
func (m Model) TimeFilterRange() string {
	if !m.timeFilterActive {
		return ""
	}
	return m.timeFilterStart.Format("Jan 2") + " - " + m.timeFilterEnd.Format("Jan 2")
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
