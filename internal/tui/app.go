package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui/detail"
	"github.com/nogo/gitree/internal/tui/filter"
	"github.com/nogo/gitree/internal/tui/list"
	"github.com/nogo/gitree/internal/watcher"
)

type Model struct {
	repo               *domain.Repository
	repoPath           string
	list               list.Model
	detail             detail.Model
	branchFilter       filter.BranchFilter
	authorFilter       filter.AuthorFilter
	authorHighlight    filter.AuthorHighlight
	watcher            *watcher.Watcher
	watching           bool
	showDetail         bool
	showBranchFilter   bool
	showAuthorFilter   bool
	showAuthorHighlight bool
	branchFilterActive bool
	authorFilterActive bool
	width              int
	height             int
	ready              bool
	err                error
}

func NewModel(repo *domain.Repository, repoPath string, w *watcher.Watcher) Model {
	return Model{
		repo:            repo,
		repoPath:        repoPath,
		list:            list.New(repo),
		detail:          detail.New(),
		branchFilter:    filter.NewBranchFilter(repo.Branches),
		authorFilter:    filter.NewAuthorFilter(repo.Commits),
		authorHighlight: filter.NewAuthorHighlight(repo.Commits),
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
			// Reapply filter if active
			if m.branchFilterActive || m.authorFilterActive {
				m.applyFilter()
			} else {
				m.list.SetRepo(msg.Repo)
			}
			// Reapply highlight if active
			if m.authorHighlight.IsActive() {
				m.applyHighlight()
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Handle detail view keys
		if m.showDetail {
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				m.showDetail = false
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

		case "c":
			// Clear all filters and highlight
			m.branchFilter.Reset()
			m.authorFilter.Reset()
			m.authorHighlight.Reset()
			m.branchFilterActive = false
			m.authorFilterActive = false
			m.list.SetHighlightedEmails(nil)
			m.list.SetRepo(m.repo)
			return m, nil

		case "esc":
			// No action when not in detail/filter view
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Header(1) + separator(1) + column headers(1) + separator(1) + footer(1) = 5 lines
		contentHeight := msg.Height - 5
		if contentHeight < 1 {
			contentHeight = 1
		}
		m.list.SetSize(msg.Width, contentHeight)
		m.detail.SetSize(msg.Width, msg.Height)
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
	branchAllSelected := m.branchFilter.AllSelected()
	authorAllSelected := m.authorFilter.AllSelected()

	m.branchFilterActive = !branchAllSelected
	m.authorFilterActive = !authorAllSelected

	if branchAllSelected && authorAllSelected {
		// No filtering needed
		m.list.SetRepo(m.repo)
		return
	}

	// Start with all commits
	filtered := m.repo.Commits

	// Apply branch filter if active
	if !branchAllSelected {
		selectedBranches := m.branchFilter.SelectedBranches()
		filtered = m.filterCommitsByBranch(filtered, selectedBranches)
	}

	// Apply author filter if active
	if !authorAllSelected {
		selectedEmails := m.authorFilter.SelectedEmails()
		filtered = m.filterCommitsByAuthor(filtered, selectedEmails)
	}

	m.list.SetFilteredCommits(filtered, m.repo)
}

func (m *Model) applyHighlight() {
	emails := m.authorHighlight.HighlightedEmails()
	m.list.SetHighlightedEmails(emails)
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
