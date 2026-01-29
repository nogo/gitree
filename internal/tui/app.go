package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/list"
)

type Model struct {
	repo   *domain.Repository
	list   list.Model
	width  int
	height int
	ready  bool
	err    error
}

func NewModel(repo *domain.Repository) Model {
	return Model{
		repo: repo,
		list: list.New(repo),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
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
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.renderLayout()
}
