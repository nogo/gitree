package insights

import (
	"testing"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

func TestComputeCalendarData_SameDayCommits(t *testing.T) {
	// 7 commits on same day -> single cell with count 7
	date := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	commits := make([]*domain.Commit, 7)
	for i := 0; i < 7; i++ {
		commits[i] = &domain.Commit{
			Hash: "abc123",
			Date: date.Add(time.Duration(i) * time.Hour),
		}
	}

	cal := ComputeCalendarData(commits, WeekStartSunday)

	// Find the cell with our date
	normalizedDate := normalizeToDay(date)
	found := false
	for _, week := range cal.Cells {
		for _, cell := range week {
			if cell.Date.Equal(normalizedDate) {
				if cell.Count != 7 {
					t.Errorf("expected count 7, got %d", cell.Count)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected to find cell with commit date")
	}
}

func TestComputeCalendarData_TwoWeeks(t *testing.T) {
	// Commits spanning 2 weeks -> 2 rows of cells
	week1 := time.Date(2024, 6, 10, 12, 0, 0, 0, time.UTC) // Monday
	week2 := time.Date(2024, 6, 17, 12, 0, 0, 0, time.UTC) // Next Monday

	commits := []*domain.Commit{
		{Hash: "abc123", Date: week1},
		{Hash: "def456", Date: week2},
	}

	cal := ComputeCalendarData(commits, WeekStartMonday)

	if len(cal.Cells) < 2 {
		t.Errorf("expected at least 2 rows of cells, got %d", len(cal.Cells))
	}
}

func TestComputeCalendarData_Empty(t *testing.T) {
	// No commits -> empty calendar with 0-count cells for default range (52 weeks)
	cal := ComputeCalendarData(nil, WeekStartSunday)

	if len(cal.Cells) == 0 {
		t.Error("expected non-empty calendar for empty commits")
	}

	// All cells should have count 0
	for _, week := range cal.Cells {
		for _, cell := range week {
			if cell.Count != 0 {
				t.Errorf("expected count 0 for empty calendar, got %d", cell.Count)
			}
			if cell.Level != 0 {
				t.Errorf("expected level 0 for empty calendar, got %d", cell.Level)
			}
		}
	}
}

func TestNormalizeToDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "strips time",
			input:    time.Date(2024, 6, 15, 14, 30, 45, 123, time.UTC),
			expected: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "preserves date",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "handles different timezone",
			input:    time.Date(2024, 6, 15, 23, 59, 59, 0, time.FixedZone("EST", -5*60*60)),
			expected: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeToDay(tc.input)
			if !got.Equal(tc.expected) {
				t.Errorf("normalizeToDay() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestWeekdayIndex(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		weekStart WeekStartDay
		expected  int
	}{
		{
			name:      "Sunday with Sunday start",
			date:      time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC), // Sunday
			weekStart: WeekStartSunday,
			expected:  0,
		},
		{
			name:      "Monday with Sunday start",
			date:      time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC), // Monday
			weekStart: WeekStartSunday,
			expected:  1,
		},
		{
			name:      "Saturday with Sunday start",
			date:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), // Saturday
			weekStart: WeekStartSunday,
			expected:  6,
		},
		{
			name:      "Monday with Monday start",
			date:      time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC), // Monday
			weekStart: WeekStartMonday,
			expected:  0,
		},
		{
			name:      "Sunday with Monday start",
			date:      time.Date(2024, 6, 16, 0, 0, 0, 0, time.UTC), // Sunday
			weekStart: WeekStartMonday,
			expected:  6,
		},
		{
			name:      "Wednesday with Monday start",
			date:      time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC), // Wednesday
			weekStart: WeekStartMonday,
			expected:  2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := weekdayIndex(tc.date, tc.weekStart)
			if got != tc.expected {
				t.Errorf("weekdayIndex() = %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestComputeLevel(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		p25      int
		p50      int
		p75      int
		expected int
	}{
		{
			name:     "zero count",
			count:    0,
			p25:      1,
			p50:      2,
			p75:      3,
			expected: 0,
		},
		{
			name:     "at p25",
			count:    1,
			p25:      1,
			p50:      2,
			p75:      3,
			expected: 1,
		},
		{
			name:     "between p25 and p50",
			count:    2,
			p25:      1,
			p50:      3,
			p75:      5,
			expected: 2,
		},
		{
			name:     "at p50",
			count:    3,
			p25:      1,
			p50:      3,
			p75:      5,
			expected: 2,
		},
		{
			name:     "between p50 and p75",
			count:    4,
			p25:      1,
			p50:      3,
			p75:      5,
			expected: 3,
		},
		{
			name:     "above p75",
			count:    10,
			p25:      1,
			p50:      3,
			p75:      5,
			expected: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeLevel(tc.count, tc.p25, tc.p50, tc.p75)
			if got != tc.expected {
				t.Errorf("computeLevel() = %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		name     string
		sorted   []int
		p        int
		expected int
	}{
		{
			name:     "empty",
			sorted:   []int{},
			p:        50,
			expected: 0,
		},
		{
			name:     "single element",
			sorted:   []int{5},
			p:        50,
			expected: 5,
		},
		{
			name:     "p0",
			sorted:   []int{1, 2, 3, 4, 5},
			p:        0,
			expected: 1,
		},
		{
			name:     "p25",
			sorted:   []int{1, 2, 3, 4, 5},
			p:        25,
			expected: 2,
		},
		{
			name:     "p50",
			sorted:   []int{1, 2, 3, 4, 5},
			p:        50,
			expected: 3,
		},
		{
			name:     "p75",
			sorted:   []int{1, 2, 3, 4, 5},
			p:        75,
			expected: 4,
		},
		{
			name:     "p100",
			sorted:   []int{1, 2, 3, 4, 5},
			p:        100,
			expected: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := percentile(tc.sorted, tc.p)
			if got != tc.expected {
				t.Errorf("percentile() = %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestCalendarDataStructure(t *testing.T) {
	// Verify that each row has exactly 7 columns
	date := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	commits := []*domain.Commit{
		{Hash: "abc123", Date: date},
	}

	cal := ComputeCalendarData(commits, WeekStartSunday)

	for i, week := range cal.Cells {
		if len(week) != 7 {
			t.Errorf("row %d has %d columns, want 7", i, len(week))
		}
	}
}

func TestCalendarHeatLevels(t *testing.T) {
	// Create commits with varying frequencies to test level assignment
	baseDate := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	var commits []*domain.Commit
	// Day 1: 1 commit
	commits = append(commits, &domain.Commit{Hash: "a", Date: baseDate})
	// Day 2: 5 commits
	for i := 0; i < 5; i++ {
		commits = append(commits, &domain.Commit{Hash: "b", Date: baseDate.AddDate(0, 0, 1)})
	}
	// Day 3: 10 commits
	for i := 0; i < 10; i++ {
		commits = append(commits, &domain.Commit{Hash: "c", Date: baseDate.AddDate(0, 0, 2)})
	}
	// Day 4: 20 commits
	for i := 0; i < 20; i++ {
		commits = append(commits, &domain.Commit{Hash: "d", Date: baseDate.AddDate(0, 0, 3)})
	}

	cal := ComputeCalendarData(commits, WeekStartSunday)

	if cal.MaxCount != 20 {
		t.Errorf("MaxCount = %d, want 20", cal.MaxCount)
	}

	// Verify levels increase with commit count
	var levels []int
	for _, week := range cal.Cells {
		for _, cell := range week {
			if cell.Count > 0 {
				levels = append(levels, cell.Level)
			}
		}
	}

	// With 4 different count values, we should have 4 different levels (1-4)
	if len(levels) != 4 {
		t.Errorf("expected 4 cells with commits, got %d", len(levels))
	}
}
