package insights

import (
	"sort"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

// DayCell represents a single day in the calendar heatmap.
type DayCell struct {
	Date  time.Time
	Count int
	Level int // 0-4 for heat intensity
}

// CalendarData holds the computed heatmap data.
type CalendarData struct {
	Cells     [][]DayCell // rows of weeks, 7 columns (days of week)
	StartDate time.Time
	EndDate   time.Time
	MaxCount  int
}

// WeekStartDay determines which day starts the week.
type WeekStartDay int

const (
	WeekStartSunday WeekStartDay = iota
	WeekStartMonday
)

// ComputeCalendarData builds calendar heatmap data from commits.
func ComputeCalendarData(commits []*domain.Commit, weekStart WeekStartDay) CalendarData {
	if len(commits) == 0 {
		return computeEmptyCalendar(weekStart)
	}

	// Count commits per day
	dayCounts := countCommitsByDay(commits)

	// Find date range
	startDate, endDate := findDateRange(dayCounts)

	// Find max count for normalization
	maxCount := findMaxCount(dayCounts)

	// Organize into week rows
	cells := buildWeekRows(dayCounts, startDate, endDate, weekStart)

	return CalendarData{
		Cells:     cells,
		StartDate: startDate,
		EndDate:   endDate,
		MaxCount:  maxCount,
	}
}

// normalizeToDay strips time, keeps only date in UTC.
func normalizeToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// weekdayIndex returns 0-6 column index based on week start preference.
func weekdayIndex(t time.Time, weekStart WeekStartDay) int {
	wd := int(t.Weekday())
	if weekStart == WeekStartMonday {
		// Shift so Monday=0, Sunday=6
		return (wd + 6) % 7
	}
	// Sunday=0, Saturday=6
	return wd
}

// countCommitsByDay aggregates commits into daily buckets.
func countCommitsByDay(commits []*domain.Commit) map[time.Time]int {
	counts := make(map[time.Time]int)
	for _, c := range commits {
		day := normalizeToDay(c.Date)
		counts[day]++
	}
	return counts
}

// findDateRange returns the start and end dates from the commit data.
func findDateRange(dayCounts map[time.Time]int) (time.Time, time.Time) {
	var dates []time.Time
	for d := range dayCounts {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})
	return dates[0], dates[len(dates)-1]
}

// computeEmptyCalendar returns a calendar for the last 52 weeks with zero counts.
func computeEmptyCalendar(weekStart WeekStartDay) CalendarData {
	now := normalizeToDay(time.Now())
	endDate := now
	startDate := now.AddDate(0, 0, -52*7)

	// Align start to week boundary
	startDate = alignToWeekStart(startDate, weekStart)

	cells := buildWeekRows(nil, startDate, endDate, weekStart)

	return CalendarData{
		Cells:     cells,
		StartDate: startDate,
		EndDate:   endDate,
		MaxCount:  0,
	}
}

// alignToWeekStart adjusts date to the start of its week.
func alignToWeekStart(t time.Time, weekStart WeekStartDay) time.Time {
	offset := weekdayIndex(t, weekStart)
	return t.AddDate(0, 0, -offset)
}

// findMaxCount returns the maximum commit count from dayCounts.
func findMaxCount(dayCounts map[time.Time]int) int {
	maxCount := 0
	for _, count := range dayCounts {
		if count > maxCount {
			maxCount = count
		}
	}
	return maxCount
}

// computeLevel returns heat level 0-4 based on count and percentile thresholds.
func computeLevel(count, p25, p50, p75 int) int {
	switch {
	case count == 0:
		return 0
	case count <= p25:
		return 1
	case count <= p50:
		return 2
	case count <= p75:
		return 3
	default:
		return 4
	}
}

// percentile returns the value at the given percentile (0-100).
func percentile(sorted []int, p int) int {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) - 1) * p / 100
	return sorted[idx]
}

// buildWeekRows organizes days into week rows (7 columns).
func buildWeekRows(dayCounts map[time.Time]int, startDate, endDate time.Time, weekStart WeekStartDay) [][]DayCell {
	// Align dates to week boundaries
	alignedStart := alignToWeekStart(startDate, weekStart)
	alignedEnd := endDate

	// Calculate number of weeks
	days := int(alignedEnd.Sub(alignedStart).Hours()/24) + 1
	weeks := (days + 6) / 7

	// Calculate percentile thresholds for level assignment
	var counts []int
	for _, count := range dayCounts {
		if count > 0 {
			counts = append(counts, count)
		}
	}
	sort.Ints(counts)

	p25, p50, p75 := 0, 0, 0
	if len(counts) > 0 {
		p25 = percentile(counts, 25)
		p50 = percentile(counts, 50)
		p75 = percentile(counts, 75)
	}

	// Build rows
	cells := make([][]DayCell, weeks)
	for w := 0; w < weeks; w++ {
		cells[w] = make([]DayCell, 7)
		for d := 0; d < 7; d++ {
			day := alignedStart.AddDate(0, 0, w*7+d)
			count := 0
			if dayCounts != nil {
				count = dayCounts[day]
			}
			cells[w][d] = DayCell{
				Date:  day,
				Count: count,
				Level: computeLevel(count, p25, p50, p75),
			}
		}
	}

	return cells
}
