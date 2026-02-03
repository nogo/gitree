package insights

import (
	"strings"
	"testing"
)

func TestFormatCount(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234, "1,234"},
		{9999, "9,999"},
		{10000, "10k"},
		{12431, "12k"},
		{100000, "100k"},
		{999999, "999k"},
	}

	for _, tt := range tests {
		got := formatCount(tt.n)
		if got != tt.want {
			t.Errorf("formatCount(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		n, total int
		want     string
	}{
		{0, 0, "0%"},
		{0, 100, "0%"},
		{50, 100, "50%"},
		{45, 100, "45%"},
		{1, 3, "33%"},
		{100, 100, "100%"},
	}

	for _, tt := range tests {
		got := formatPercent(tt.n, tt.total)
		if got != tt.want {
			t.Errorf("formatPercent(%d, %d) = %q, want %q", tt.n, tt.total, got, tt.want)
		}
	}
}

func TestRenderAuthorStats(t *testing.T) {
	stats := []AuthorStats{
		{Name: "Alice", Email: "alice@example.com", Commits: 45},
		{Name: "Bob", Email: "bob@example.com", Commits: 30},
		{Name: "Charlie", Email: "charlie@example.com", Commits: 15},
		{Name: "Diana", Email: "diana@example.com", Commits: 7},
		{Name: "Eve", Email: "eve@example.com", Commits: 3},
	}

	result := renderAuthorStats(stats, 40, 10)

	// Check title is present
	if !strings.Contains(result, "TOP AUTHORS") {
		t.Error("expected 'TOP AUTHORS' title")
	}

	// Check all 5 authors are rendered
	lines := strings.Split(result, "\n")
	if len(lines) != 6 { // 1 title + 5 authors
		t.Errorf("expected 6 lines (title + 5 authors), got %d", len(lines))
	}

	// Check names are present
	for _, s := range stats {
		if !strings.Contains(result, s.Name) {
			t.Errorf("expected author name %q in output", s.Name)
		}
	}
}

func TestRenderAuthorStatsLimitedHeight(t *testing.T) {
	stats := []AuthorStats{
		{Name: "Alice", Commits: 45},
		{Name: "Bob", Commits: 30},
		{Name: "Charlie", Commits: 15},
		{Name: "Diana", Commits: 7},
		{Name: "Eve", Commits: 3},
	}

	// Height 4 = title + 2 rows max
	result := renderAuthorStats(stats, 40, 4)
	lines := strings.Split(result, "\n")

	if len(lines) != 3 { // 1 title + 2 authors
		t.Errorf("expected 3 lines with height=4, got %d", len(lines))
	}

	// Should have Alice and Bob but not Charlie
	if !strings.Contains(result, "Alice") {
		t.Error("expected Alice in limited output")
	}
	if !strings.Contains(result, "Bob") {
		t.Error("expected Bob in limited output")
	}
	if strings.Contains(result, "Charlie") {
		t.Error("Charlie should not be in limited output")
	}
}

func TestRenderAuthorStatsEmpty(t *testing.T) {
	result := renderAuthorStats(nil, 40, 10)
	if result != "" {
		t.Errorf("expected empty string for nil stats, got %q", result)
	}

	result = renderAuthorStats([]AuthorStats{}, 40, 10)
	if result != "" {
		t.Errorf("expected empty string for empty stats, got %q", result)
	}
}

func TestRenderFileStats(t *testing.T) {
	stats := []FileStats{
		{Path: "cmd/main.go", ChangeCount: 89},
		{Path: "internal/tui/app.go", ChangeCount: 45},
		{Path: "internal/domain/commit.go", ChangeCount: 23},
		{Path: "README.md", ChangeCount: 12},
		{Path: "go.mod", ChangeCount: 5},
	}

	result := renderFileStats(stats, 50, 10)

	// Check title is present
	if !strings.Contains(result, "TOP FILES") {
		t.Error("expected 'TOP FILES' title")
	}

	// Check all 5 files are rendered
	lines := strings.Split(result, "\n")
	if len(lines) != 6 { // 1 title + 5 files
		t.Errorf("expected 6 lines (title + 5 files), got %d", len(lines))
	}
}

func TestRenderFileStatsLongPath(t *testing.T) {
	stats := []FileStats{
		{Path: "internal/tui/insights/very/long/path/to/some/deeply/nested/file.go", ChangeCount: 42},
	}

	// Narrow width should truncate with "..."
	result := renderFileStats(stats, 30, 5)

	if !strings.Contains(result, "...") {
		t.Error("expected long path to be truncated with '...'")
	}
}

func TestRenderFileStatsEmpty(t *testing.T) {
	result := renderFileStats(nil, 40, 10)
	if result != "" {
		t.Errorf("expected empty string for nil stats, got %q", result)
	}
}

func TestRenderSummary(t *testing.T) {
	summary := Summary{
		TotalCommits:   847,
		TotalAuthors:   12,
		TotalAdditions: 12431,
		TotalDeletions: 8291,
	}

	// Wide width should show full format
	result := renderSummary(summary, 60)
	if !strings.Contains(result, "847 commits") {
		t.Error("expected commit count in summary")
	}
	if !strings.Contains(result, "12 authors") {
		t.Error("expected author count in full summary")
	}
	if !strings.Contains(result, "+12k/-8k") || !strings.Contains(result, "+12,431/-8,291") {
		// Should have either formatted version
		if !strings.Contains(result, "+") || !strings.Contains(result, "-") {
			t.Error("expected additions/deletions in summary")
		}
	}
}

func TestRenderSummaryNarrow(t *testing.T) {
	summary := Summary{
		TotalCommits:   847,
		TotalAuthors:   12,
		TotalAdditions: 12431,
		TotalDeletions: 8291,
	}

	// Narrow width should show compact format
	result := renderSummary(summary, 25)
	if !strings.Contains(result, "commits") {
		t.Error("expected 'commits' in narrow summary")
	}
	// Should not have "authors" in narrow format
	if strings.Contains(result, "authors") {
		t.Error("narrow summary should not include authors")
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
	}

	for _, tt := range tests {
		got := truncateWithEllipsis(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateWithEllipsis(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

func TestTruncatePathWithEllipsis(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"file.go", 10, "file.go"},
		{"path/to/file.go", 15, "path/to/file.go"},
		{"very/long/path/to/file.go", 15, "...h/to/file.go"},
		{"abc", 2, "bc"},
		{"abc", 0, ""},
	}

	for _, tt := range tests {
		got := truncatePathWithEllipsis(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncatePathWithEllipsis(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}
