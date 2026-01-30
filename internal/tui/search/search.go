package search

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nogo/gitree/internal/domain"
)

type Search struct {
	active       bool
	inputMode    bool // true when typing in search box
	query        string
	matches      []int // indices into commits
	currentMatch int   // index into matches (-1 if no matches)
	textInput    textinput.Model
}

func New() Search {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Width = 40
	return Search{
		textInput:    ti,
		currentMatch: -1,
	}
}

// Activate starts search input mode
func (s *Search) Activate() {
	s.active = true
	s.inputMode = true
	s.textInput.SetValue("")
	s.textInput.Focus()
}

// IsActive returns whether search mode is active (has results or input open)
func (s Search) IsActive() bool {
	return s.active
}

// IsInputMode returns whether user is typing in search box
func (s Search) IsInputMode() bool {
	return s.inputMode
}

// Query returns the current search query
func (s Search) Query() string {
	return s.query
}

// Matches returns indices of matching commits
func (s Search) Matches() []int {
	return s.matches
}

// CurrentMatch returns the current match index (-1 if none)
func (s Search) CurrentMatch() int {
	return s.currentMatch
}

// CurrentMatchCommitIndex returns the commit index of the current match
func (s Search) CurrentMatchCommitIndex() int {
	if s.currentMatch >= 0 && s.currentMatch < len(s.matches) {
		return s.matches[s.currentMatch]
	}
	return -1
}

// MatchCount returns the number of matches
func (s Search) MatchCount() int {
	return len(s.matches)
}

// Update handles input during search input mode
func (s Search) Update(msg tea.Msg) (Search, tea.Cmd, bool, bool) {
	// done = execute search, cancelled = cancel input
	if !s.inputMode {
		return s, nil, false, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Execute search
			s.query = s.textInput.Value()
			s.inputMode = false
			return s, nil, true, false
		case tea.KeyEsc:
			// Cancel input, keep previous results if any
			s.inputMode = false
			if s.query == "" {
				s.active = false
			}
			return s, nil, false, true
		}
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd, false, false
}

// Execute runs the search on the given commits
func (s *Search) Execute(commits []domain.Commit) {
	if s.query == "" {
		s.matches = nil
		s.currentMatch = -1
		s.active = false
		return
	}

	s.matches = searchCommits(commits, s.query)
	if len(s.matches) > 0 {
		s.currentMatch = 0
	} else {
		s.currentMatch = -1
	}
}

// NextMatch moves to the next match (wraps around)
func (s *Search) NextMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.currentMatch = (s.currentMatch + 1) % len(s.matches)
}

// PrevMatch moves to the previous match (wraps around)
func (s *Search) PrevMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.currentMatch--
	if s.currentMatch < 0 {
		s.currentMatch = len(s.matches) - 1
	}
}

// Clear resets the search state
func (s *Search) Clear() {
	s.active = false
	s.inputMode = false
	s.query = ""
	s.matches = nil
	s.currentMatch = -1
	s.textInput.SetValue("")
}

// InputView returns the text input view for rendering in footer
func (s Search) InputView() string {
	return s.textInput.View()
}

// searchCommits finds commits matching the query (case-insensitive)
func searchCommits(commits []domain.Commit, query string) []int {
	query = strings.ToLower(query)
	var matches []int
	for i, c := range commits {
		if containsIgnoreCase(c.Message, query) ||
			containsIgnoreCase(c.FullMessage, query) ||
			containsIgnoreCase(c.Hash, query) ||
			containsIgnoreCase(c.ShortHash, query) {
			matches = append(matches, i)
		}
	}
	return matches
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), substr)
}
