package list

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     now.Add(-30 * time.Second),
			expected: "now",
		},
		{
			name:     "minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5m ago",
		},
		{
			name:     "59 minutes ago",
			time:     now.Add(-59 * time.Minute),
			expected: "59m ago",
		},
		{
			name:     "hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3h ago",
		},
		{
			name:     "23 hours ago",
			time:     now.Add(-23 * time.Hour),
			expected: "23h ago",
		},
		{
			name:     "days ago",
			time:     now.Add(-2 * 24 * time.Hour),
			expected: "2d ago",
		},
		{
			name:     "6 days ago",
			time:     now.Add(-6 * 24 * time.Hour),
			expected: "6d ago",
		},
		{
			name:     "weeks ago",
			time:     now.Add(-14 * 24 * time.Hour),
			expected: "2w ago",
		},
		{
			name:     "4 weeks ago",
			time:     now.Add(-28 * 24 * time.Hour),
			expected: "4w ago",
		},
		{
			name: "over 30 days shows date",
			time: now.Add(-60 * 24 * time.Hour),
			expected: func() string {
				t := now.Add(-60 * 24 * time.Hour)
				if t.Year() != now.Year() {
					return t.Format("Jan 2 '06")
				}
				return t.Format("Jan 2")
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatRelativeTime(tc.time)
			if got != tc.expected {
				t.Errorf("formatRelativeTime() = %q, want %q", got, tc.expected)
			}
		})
	}
}
