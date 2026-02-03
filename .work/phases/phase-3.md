# Phase 3: Insights Model & Styles

## Goal
Create the InsightsView model that holds computed stats and heatmap data, plus define styling constants.

## Changes
- `internal/tui/insights/insights.go`: New file with InsightsView model
- `internal/tui/insights/styles.go`: New file with color definitions

## Implementation

### insights.go
1. Define `InsightsView` struct:
   - authorStats []AuthorStats
   - fileStats []FileStats
   - summary Summary
   - calendar CalendarData
   - width, height int
2. Implement `New() InsightsView` - empty initial state
3. Implement `Recalculate(commits []*domain.Commit, fileChanges map[string][]domain.FileChange)`:
   - Call ComputeAuthorStats()
   - Call ComputeFileStats()
   - Call ComputeSummary()
   - Call ComputeCalendarData()
4. Implement `SetSize(width, height int)` - store dimensions for rendering
5. Implement getters: `AuthorStats()`, `FileStats()`, `Summary()`, `Calendar()`

### styles.go
1. Define heat intensity colors (5 levels):
   - Heat0: gray/empty (color "240")
   - Heat1: light green (color "22")
   - Heat2: medium green (color "28")
   - Heat3: dark green (color "34")
   - Heat4: bright green (color "40")
2. Define stats table styles:
   - HeaderStyle: bold
   - NameStyle: default
   - CountStyle: dimmed
   - PercentStyle: cyan
3. Define section title style

## Testing
- Recalculate with commits → all stats populated
- SetSize stores dimensions correctly
- Styles render without panic

## Estimated LOC
~80
