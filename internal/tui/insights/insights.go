package insights

import "github.com/nogo/gitree/internal/domain"

// InsightsView holds computed statistics and heatmap data for rendering.
type InsightsView struct {
	authorStats []AuthorStats
	fileStats   []FileStats
	summary     Summary
	calendar    CalendarData
	width       int
	height      int
}

// New creates an empty InsightsView with default values.
func New() InsightsView {
	return InsightsView{}
}

// Recalculate computes all statistics from the provided commits and file changes.
// commits should be pointers for calendar computation.
// fileChanges maps commit hash to the list of file changes in that commit.
func (v *InsightsView) Recalculate(commits []*domain.Commit, fileChanges map[string][]domain.FileChange) {
	// Convert pointer slice to value slice for stats functions
	valueCommits := make([]domain.Commit, len(commits))
	for i, c := range commits {
		if c != nil {
			valueCommits[i] = *c
		}
	}

	// Compute statistics with reasonable defaults for topN
	const topAuthors = 10
	const topFiles = 10

	v.authorStats = ComputeAuthorStats(valueCommits, topAuthors)
	v.fileStats = ComputeFileStats(fileChanges, topFiles)
	v.summary = ComputeSummary(valueCommits, v.authorStats, v.fileStats)
	v.calendar = ComputeCalendarData(commits, WeekStartMonday)
}

// SetSize stores the available dimensions for rendering.
func (v *InsightsView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// AuthorStats returns the computed author statistics.
func (v *InsightsView) AuthorStats() []AuthorStats {
	return v.authorStats
}

// FileStats returns the computed file statistics.
func (v *InsightsView) FileStats() []FileStats {
	return v.fileStats
}

// Summary returns the computed repository summary.
func (v *InsightsView) Summary() Summary {
	return v.summary
}

// Calendar returns the computed calendar heatmap data.
func (v *InsightsView) Calendar() CalendarData {
	return v.calendar
}

// Width returns the stored width dimension.
func (v *InsightsView) Width() int {
	return v.width
}

// Height returns the stored height dimension.
func (v *InsightsView) Height() int {
	return v.height
}
