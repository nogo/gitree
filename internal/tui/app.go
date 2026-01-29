package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/git"
	"github.com/nogo/gitree/internal/tui/detail"
	"github.com/nogo/gitree/internal/tui/list"
	"github.com/nogo/gitree/internal/watcher"
)

type Model struct {
	repo       *domain.Repository
	repoPath   string
	list       list.Model
	detail     detail.Model
	watcher    *watcher.Watcher
	watching   bool
	showDetail bool
	width      int
	height     int
	ready      bool
	err        error
}

func NewModel(repo *domain.Repository, repoPath string, w *watcher.Watcher) Model {
	return Model{
		repo:     repo,
		repoPath: repoPath,
		list:     list.New(repo),
		detail:   detail.New(),
		watcher:  w,
		watching: w != nil,
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
			m.list.SetRepo(msg.Repo)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if !m.showDetail {
				return m, tea.Quit
			}
			// Close detail view instead of quitting
			m.showDetail = false
			return m, nil

		case "enter":
			if !m.showDetail {
				// Open detail for selected commit
				selected := m.list.SelectedCommit()
				if selected != nil {
					m.detail.SetCommit(selected)
					m.detail.SetSize(m.width, m.height)
					m.showDetail = true
				}
				return m, nil
			}

		case "esc":
			if m.showDetail {
				m.showDetail = false
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Header and footer each take 1 line
		contentHeight := msg.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}
		m.list.SetSize(msg.Width, contentHeight)
		m.detail.SetSize(msg.Width, msg.Height)
	}

	// Route updates to active component
	var cmd tea.Cmd
	if m.showDetail {
		m.detail, cmd = m.detail.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
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
