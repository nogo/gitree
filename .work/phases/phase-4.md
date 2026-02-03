# Phase 4: Stats Rendering

## Goal
Implement rendering functions for author stats, file stats, and summary line.

## Changes
- `internal/tui/insights/render_stats.go`: New file with stats rendering functions

## Implementation
1. Implement `renderAuthorStats(stats []AuthorStats, width, height int) string`:
   - Title: "TOP AUTHORS" (styled)
   - For each author (up to height-2 rows):
     - Format: `Name        123  45%`
     - Name left-aligned, truncated to fit
     - Count right-aligned
     - Percentage right-aligned
   - Use lipgloss for column alignment
2. Implement `renderFileStats(stats []FileStats, width, height int) string`:
   - Title: "TOP FILES"
   - For each file (up to height-2 rows):
     - Format: `path/to/file.go   89`
     - Path left-aligned, truncated with "..." if needed
     - Count right-aligned
3. Implement `renderSummary(summary Summary, width int) string`:
   - Single line format: `847 commits · 12 authors · +12,431/-8,291 lines`
   - Or if narrow: `847 commits · +12k/-8k`
   - Center-aligned
4. Helper: `formatCount(n int) string` - returns "12,431" or "12k" for large numbers
5. Helper: `formatPercent(n, total int) string` - returns "45%"

## Testing
- 5 authors render as 5 rows
- Long file path truncates with "..."
- Summary shows correct totals
- Narrow width shows compact format

## Estimated LOC
~80
