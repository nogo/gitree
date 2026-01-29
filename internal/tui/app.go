package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
)

type Model struct {
	repo   *domain.Repository
	width  int
	height int
	ready  bool
	err    error
}

func NewModel(repo *domain.Repository) Model {
	return Model{repo: repo}
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
	}
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.renderLayout()
}
