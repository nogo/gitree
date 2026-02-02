package filtering

import (
	"strings"
	"time"

	"github.com/nogo/gitree/internal/domain"
	"github.com/nogo/gitree/internal/tui/filter"
)

// Manager encapsulates all filtering state and logic
type Manager struct {
	branchFilter    filter.BranchFilter
	authorFilter    filter.AuthorFilter
	authorHighlight filter.AuthorHighlight
	tagFilter       filter.TagFilter

	branchFilterActive bool
	authorFilterActive bool
	tagFilterActive    bool
	timeFilterActive   bool
	timeFilterStart    time.Time
	timeFilterEnd      time.Time

	repo *domain.Repository
}

// Result contains the output of applying all filters
type Result struct {
	Commits    []domain.Commit
	IsFiltered bool
}

// New creates a new FilterManager for the given repository
func New(repo *domain.Repository) *Manager {
	return &Manager{
		branchFilter:    filter.NewBranchFilter(repo.Branches),
		authorFilter:    filter.NewAuthorFilter(repo.Commits),
		authorHighlight: filter.NewAuthorHighlight(repo.Commits),
		tagFilter:       filter.NewTagFilter(repo.Commits),
		repo:            repo,
	}
}

// UpdateRepo updates the filter manager when the repository changes
func (m *Manager) UpdateRepo(repo *domain.Repository) {
	m.repo = repo
	m.branchFilter.UpdateBranches(repo.Branches)
	m.authorFilter.UpdateAuthors(repo.Commits)
	m.authorHighlight.UpdateAuthors(repo.Commits)
	m.tagFilter.UpdateTags(repo.Commits)
}

// ApplyFilters applies all active filters and returns the result
func (m *Manager) ApplyFilters() Result {
	filtered := m.repo.Commits

	// Apply branch filter
	if !m.branchFilter.AllSelected() {
		selectedBranches := m.branchFilter.SelectedBranches()
		filtered = m.filterCommitsByBranch(filtered, selectedBranches)
	}

	// Apply tag filter (if any tags selected)
	if m.tagFilter.HasSelection() {
		selectedTags := m.tagFilter.SelectedTags()
		filtered = m.filterCommitsByTag(filtered, selectedTags)
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

	return Result{
		Commits:    filtered,
		IsFiltered: len(filtered) != len(m.repo.Commits),
	}
}

// SetTimeFilter sets the time filter range
func (m *Manager) SetTimeFilter(start, end time.Time, active bool) {
	m.timeFilterActive = active
	m.timeFilterStart = start
	m.timeFilterEnd = end
}

// ClearTimeFilter clears the time filter
func (m *Manager) ClearTimeFilter() {
	m.timeFilterActive = false
	m.timeFilterStart = time.Time{}
	m.timeFilterEnd = time.Time{}
}

// Reset clears all filters
func (m *Manager) Reset() {
	m.branchFilter.Reset()
	m.authorFilter.Reset()
	m.authorHighlight.Reset()
	m.tagFilter.Reset()
	m.branchFilterActive = false
	m.authorFilterActive = false
	m.tagFilterActive = false
	m.timeFilterActive = false
	m.timeFilterStart = time.Time{}
	m.timeFilterEnd = time.Time{}
}

// UpdateFilterActive updates filter active state based on selection
func (m *Manager) UpdateFilterActive() {
	m.branchFilterActive = !m.branchFilter.AllSelected()
	m.authorFilterActive = !m.authorFilter.AllSelected()
	m.tagFilterActive = m.tagFilter.HasSelection()
}

// BranchFilter returns a pointer to the branch filter for UI updates
func (m *Manager) BranchFilter() *filter.BranchFilter {
	return &m.branchFilter
}

// AuthorFilter returns a pointer to the author filter for UI updates
func (m *Manager) AuthorFilter() *filter.AuthorFilter {
	return &m.authorFilter
}

// AuthorHighlight returns a pointer to the author highlight for UI updates
func (m *Manager) AuthorHighlight() *filter.AuthorHighlight {
	return &m.authorHighlight
}

// TagFilter returns a pointer to the tag filter for UI updates
func (m *Manager) TagFilter() *filter.TagFilter {
	return &m.tagFilter
}

// BranchFilterActive returns whether a branch filter is applied
func (m *Manager) BranchFilterActive() bool {
	return m.branchFilterActive
}

// AuthorFilterActive returns whether an author filter is applied
func (m *Manager) AuthorFilterActive() bool {
	return m.authorFilterActive
}

// TagFilterActive returns whether a tag filter is applied
func (m *Manager) TagFilterActive() bool {
	return m.tagFilterActive
}

// TimeFilterActive returns whether a time filter is applied
func (m *Manager) TimeFilterActive() bool {
	return m.timeFilterActive
}

// TimeFilterRange returns the formatted time range
func (m *Manager) TimeFilterRange() string {
	if !m.timeFilterActive {
		return ""
	}
	return m.timeFilterStart.Format("Jan 2") + " - " + m.timeFilterEnd.Format("Jan 2")
}

// SelectedBranchCount returns the number of selected branches
func (m *Manager) SelectedBranchCount() int {
	return len(m.branchFilter.SelectedBranches())
}

// TotalBranchCount returns the total number of branches
func (m *Manager) TotalBranchCount() int {
	return len(m.repo.Branches)
}

// SelectedAuthorCount returns the number of selected authors
func (m *Manager) SelectedAuthorCount() int {
	return m.authorFilter.SelectedCount()
}

// TotalAuthorCount returns the total number of authors
func (m *Manager) TotalAuthorCount() int {
	return m.authorFilter.TotalCount()
}

// SelectedTagCount returns the number of selected tags
func (m *Manager) SelectedTagCount() int {
	return m.tagFilter.SelectedCount()
}

// TotalTagCount returns the total number of tags
func (m *Manager) TotalTagCount() int {
	return m.tagFilter.TotalCount()
}

// HighlightedEmails returns the emails of the highlighted author
func (m *Manager) HighlightedEmails() []string {
	return m.authorHighlight.HighlightedEmails()
}

// AuthorHighlightActive returns whether an author is highlighted
func (m *Manager) AuthorHighlightActive() bool {
	return m.authorHighlight.IsActive()
}

// HighlightedAuthorName returns the name of the highlighted author
func (m *Manager) HighlightedAuthorName() string {
	return m.authorHighlight.HighlightedName()
}

// filterCommitsByBranch filters commits to only those reachable from selected branches
func (m *Manager) filterCommitsByBranch(commits []domain.Commit, branchNames []string) []domain.Commit {
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

// filterCommitsByAuthor filters commits by author email
func (m *Manager) filterCommitsByAuthor(commits []domain.Commit, emails []string) []domain.Commit {
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

// filterCommitsByTime filters commits by time range
func (m *Manager) filterCommitsByTime(commits []domain.Commit, start, end time.Time) []domain.Commit {
	var result []domain.Commit
	for _, c := range commits {
		if !c.Date.Before(start) && c.Date.Before(end) {
			result = append(result, c)
		}
	}
	return result
}

// filterCommitsByTag filters commits to those with selected tags + their ancestors
func (m *Manager) filterCommitsByTag(commits []domain.Commit, tagNames []string) []domain.Commit {
	if len(tagNames) == 0 {
		return commits // No tags selected = show all
	}

	// Build set of selected tags
	tagSet := make(map[string]bool)
	for _, tag := range tagNames {
		tagSet[tag] = true
	}

	// Find commits with selected tags
	var headHashes []string
	for _, c := range m.repo.Commits {
		for _, tag := range c.Tags {
			if tagSet[tag] {
				headHashes = append(headHashes, c.Hash)
				break
			}
		}
	}

	// Build hash → commit map for parent lookup
	commitMap := make(map[string]*domain.Commit)
	for i := range m.repo.Commits {
		commitMap[m.repo.Commits[i].Hash] = &m.repo.Commits[i]
	}

	// BFS to find all reachable commits (tag commits + ancestors)
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
