# Phase 2: Heatmap Data

## Goal
Create calendar heatmap data structure with daily bucketing logic.

## Changes
- `internal/tui/insights/heatmap.go`: New file with calendar data types and computation

## Implementation
1. Define `DayCell` struct:
   - Date time.Time
   - Count int
   - Level int (0-4 for heat intensity)
2. Define `CalendarData` struct:
   - Cells [][]DayCell (7 columns for days of week, N rows for weeks)
   - StartDate, EndDate time.Time
   - MaxCount int (for normalization)
3. Implement `ComputeCalendarData(commits []*domain.Commit) CalendarData`:
   - Find date range from commits (or default to last 52 weeks)
   - Create daily buckets for entire range (sparse = 0 count)
   - Iterate commits, increment day bucket by commit date
   - Calculate heat levels based on percentiles:
     - Level 0: 0 commits
     - Level 1: 1-25th percentile
     - Level 2: 25-50th percentile
     - Level 3: 50-75th percentile
     - Level 4: 75-100th percentile
   - Organize into week rows (Mon-Sun or Sun-Sat configurable)
4. Helper: `normalizeToDay(t time.Time) time.Time` - strips time, keeps date
5. Helper: `weekdayIndex(t time.Time) int` - returns 0-6 column index

## Testing
- 7 commits on same day → single cell with count 7
- Commits spanning 2 weeks → 2 rows of cells
- No commits → empty calendar with 0-count cells for default range

## Estimated LOC
~80
