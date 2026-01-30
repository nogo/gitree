package search

import (
	"testing"

	"github.com/nogo/gitree/internal/domain"
)

func TestSearchCommits(t *testing.T) {
	commits := []domain.Commit{
		{Hash: "abc1234567890", ShortHash: "abc1234", Message: "Fix bug in parser", FullMessage: "Fix bug in parser\n\nDetailed description here"},
		{Hash: "def5678901234", ShortHash: "def5678", Message: "Add new feature", FullMessage: "Add new feature"},
		{Hash: "ghi9012345678", ShortHash: "ghi9012", Message: "Update README", FullMessage: "Update README"},
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedFirst int // index of first match, -1 if no matches
	}{
		{
			// Note: empty query handling is done by Search.Execute, not searchCommits
			name:          "empty query matches all (caller should filter)",
			query:         "",
			expectedCount: 3,
			expectedFirst: 0,
		},
		{
			name:          "match message case insensitive",
			query:         "FIX BUG",
			expectedCount: 1,
			expectedFirst: 0,
		},
		{
			name:          "match short hash",
			query:         "def5678",
			expectedCount: 1,
			expectedFirst: 1,
		},
		{
			name:          "match full hash prefix",
			query:         "ghi901234",
			expectedCount: 1,
			expectedFirst: 2,
		},
		{
			name:          "match full message content",
			query:         "detailed description",
			expectedCount: 1,
			expectedFirst: 0,
		},
		{
			name:          "match multiple commits",
			query:         "e", // present in all messages
			expectedCount: 3,
			expectedFirst: 0,
		},
		{
			name:          "no matches",
			query:         "nonexistent",
			expectedCount: 0,
			expectedFirst: -1,
		},
		{
			name:          "partial word match",
			query:         "feat",
			expectedCount: 1,
			expectedFirst: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := searchCommits(commits, tc.query)

			if len(matches) != tc.expectedCount {
				t.Errorf("searchCommits() returned %d matches, want %d", len(matches), tc.expectedCount)
			}

			if tc.expectedFirst >= 0 && len(matches) > 0 {
				if matches[0] != tc.expectedFirst {
					t.Errorf("first match index = %d, want %d", matches[0], tc.expectedFirst)
				}
			}
		})
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	// Note: containsIgnoreCase expects substr to already be lowercase
	// (caller is responsible for lowercasing the search query)
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "world", true},
		{"Hello World", "foo", false},
		{"", "", true},
		{"abc", "", true},
		{"", "abc", false},
		{"UPPERCASE", "uppercase", true},
		{"MixedCase", "mixedcase", true},
	}

	for _, tc := range tests {
		got := containsIgnoreCase(tc.s, tc.substr)
		if got != tc.expected {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tc.s, tc.substr, got, tc.expected)
		}
	}
}
