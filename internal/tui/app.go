package tui

import (
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
	repo         *domain.Repository
	repoPath     string
	list         list.Model
	detail       detail.Model
	branchFilter filter.BranchFilter
	watcher      *watcher.Watcher
	watching     bool
	showDetail   bool
	showFilter   bool
	filterActive bool
	width        int
	height       int
	ready        bool
	err          error
}

func NewModel(repo *domain.Repository, repoPath string, w *watcher.Watcher) Model {
	return Model{
		repo:         repo,
		repoPath:     repoPath,
		list:         list.New(repo),
		detail:       detail.New(),
		branchFilter: filter.NewBranchFilter(repo.Branches),
		watcher:      w,
		watching:     w != nil,
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
	// Handle filter overlay first
	if m.showFilter {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var done, cancelled bool
			m.branchFilter, _, done, cancelled = m.branchFilter.Update(keyMsg)
			if done {
				m.applyFilter()
				m.showFilter = false
			}
			if cancelled {
				m.showFilter = false
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
			// Reapply filter if active
			if m.filterActive {
				m.applyFilter()
			} else {
				m.list.SetRepo(msg.Repo)
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
			m.showFilter = true
			return m, nil

		case "c":
			// Clear filter
			m.branchFilter.Reset()
			m.filterActive = false
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
	}

	// Route updates to list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) applyFilter() {
	if m.branchFilter.AllSelected() {
		// No filtering needed
		m.filterActive = false
		m.list.SetRepo(m.repo)
		return
	}

	m.filterActive = true
	selectedBranches := m.branchFilter.SelectedBranches()
	filtered := m.filterCommits(selectedBranches)
	m.list.SetFilteredCommits(filtered, m.repo)
}

func (m Model) filterCommits(branchNames []string) []domain.Commit {
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
	for _, c := range m.repo.Commits {
		if reachable[c.Hash] {
			result = append(result, c)
		}
	}

	return result
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.showFilter {
		return m.branchFilter.View()
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

// FilterActive returns whether a filter is currently applied
func (m Model) FilterActive() bool {
	return m.filterActive
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
