package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/detail"
	"github.com/nogo/gitree/internal/tui/list"
)

type Model struct {
	repo       *domain.Repository
	list       list.Model
	detail     detail.Model
	showDetail bool
	width      int
	height     int
	ready      bool
	err        error
}

func NewModel(repo *domain.Repository) Model {
	return Model{
		repo:   repo,
		list:   list.New(repo),
		detail: detail.New(),
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
