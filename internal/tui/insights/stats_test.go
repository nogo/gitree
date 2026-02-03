package insights

import (
	"testing"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

func TestComputeAuthorStats(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		commits        []domain.Commit
		topN           int
		expectedLen    int
		expectedFirst  string // email of top author
		expectedCount  int    // commit count of top author
	}{
		{
			name:        "empty commit list",
			commits:     []domain.Commit{},
			topN:        10,
			expectedLen: 0,
		},
		{
			name: "single author",
			commits: []domain.Commit{
				{Hash: "aaa", Author: "Alice", Email: "alice@example.com", Date: now},
				{Hash: "bbb", Author: "Alice", Email: "alice@example.com", Date: now},
			},
			topN:          10,
			expectedLen:   1,
			expectedFirst: "alice@example.com",
			expectedCount: 2,
		},
		{
			name: "multiple authors sorted by commit count",
			commits: []domain.Commit{
				{Hash: "a1", Author: "Alice", Email: "alice@example.com", Date: now},
				{Hash: "b1", Author: "Bob", Email: "bob@example.com", Date: now},
				{Hash: "b2", Author: "Bob", Email: "bob@example.com", Date: now},
				{Hash: "b3", Author: "Bob", Email: "bob@example.com", Date: now},
				{Hash: "c1", Author: "Carol", Email: "carol@example.com", Date: now},
				{Hash: "c2", Author: "Carol", Email: "carol@example.com", Date: now},
			},
			topN:          10,
			expectedLen:   3,
			expectedFirst: "bob@example.com",
			expectedCount: 3,
		},
		{
			name: "topN limits results",
			commits: []domain.Commit{
				{Hash: "a1", Author: "Alice", Email: "alice@example.com", Date: now},
				{Hash: "a2", Author: "Alice", Email: "alice@example.com", Date: now},
				{Hash: "b1", Author: "Bob", Email: "bob@example.com", Date: now},
				{Hash: "c1", Author: "Carol", Email: "carol@example.com", Date: now},
			},
			topN:          2,
			expectedLen:   2,
			expectedFirst: "alice@example.com",
			expectedCount: 2,
		},
		{
			name: "topN zero returns all",
			commits: []domain.Commit{
				{Hash: "a1", Author: "Alice", Email: "alice@example.com", Date: now},
				{Hash: "b1", Author: "Bob", Email: "bob@example.com", Date: now},
				{Hash: "c1", Author: "Carol", Email: "carol@example.com", Date: now},
			},
			topN:        0,
			expectedLen: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeAuthorStats(tc.commits, tc.topN)

			if len(result) != tc.expectedLen {
				t.Errorf("expected %d authors, got %d", tc.expectedLen, len(result))
			}

			if tc.expectedLen > 0 && tc.expectedFirst != "" {
				if result[0].Email != tc.expectedFirst {
					t.Errorf("expected top author %s, got %s", tc.expectedFirst, result[0].Email)
				}
				if result[0].Commits != tc.expectedCount {
					t.Errorf("expected top author commits %d, got %d", tc.expectedCount, result[0].Commits)
				}
			}
		})
	}
}

func TestComputeFileStats(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string][]domain.FileChange
		topN          int
		expectedLen   int
		expectedFirst string // path of top file
		expectedCount int    // change count of top file
	}{
		{
			name:        "empty file map",
			files:       map[string][]domain.FileChange{},
			topN:        10,
			expectedLen: 0,
		},
		{
			name: "single file changed multiple times",
			files: map[string][]domain.FileChange{
				"commit1": {{Path: "main.go", Additions: 10, Deletions: 5}},
				"commit2": {{Path: "main.go", Additions: 20, Deletions: 10}},
			},
			topN:          10,
			expectedLen:   1,
			expectedFirst: "main.go",
			expectedCount: 2,
		},
		{
			name: "multiple files sorted by change count",
			files: map[string][]domain.FileChange{
				"commit1": {
					{Path: "main.go", Additions: 10, Deletions: 5},
					{Path: "util.go", Additions: 5, Deletions: 2},
				},
				"commit2": {
					{Path: "main.go", Additions: 20, Deletions: 10},
				},
				"commit3": {
					{Path: "main.go", Additions: 15, Deletions: 8},
				},
			},
			topN:          10,
			expectedLen:   2,
			expectedFirst: "main.go",
			expectedCount: 3,
		},
		{
			name: "topN limits results",
			files: map[string][]domain.FileChange{
				"commit1": {
					{Path: "a.go", Additions: 10, Deletions: 5},
					{Path: "b.go", Additions: 5, Deletions: 2},
					{Path: "c.go", Additions: 3, Deletions: 1},
				},
			},
			topN:        2,
			expectedLen: 2,
		},
		{
			name: "additions and deletions are summed",
			files: map[string][]domain.FileChange{
				"commit1": {{Path: "main.go", Additions: 10, Deletions: 5}},
				"commit2": {{Path: "main.go", Additions: 20, Deletions: 10}},
			},
			topN:          10,
			expectedLen:   1,
			expectedFirst: "main.go",
			expectedCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeFileStats(tc.files, tc.topN)

			if len(result) != tc.expectedLen {
				t.Errorf("expected %d files, got %d", tc.expectedLen, len(result))
			}

			if tc.expectedLen > 0 && tc.expectedFirst != "" {
				if result[0].Path != tc.expectedFirst {
					t.Errorf("expected top file %s, got %s", tc.expectedFirst, result[0].Path)
				}
				if result[0].ChangeCount != tc.expectedCount {
					t.Errorf("expected top file change count %d, got %d", tc.expectedCount, result[0].ChangeCount)
				}
			}
		})
	}

	// Additional test to verify additions/deletions are summed correctly
	t.Run("verify additions deletions sum", func(t *testing.T) {
		files := map[string][]domain.FileChange{
			"commit1": {{Path: "main.go", Additions: 10, Deletions: 5}},
			"commit2": {{Path: "main.go", Additions: 20, Deletions: 10}},
		}
		result := ComputeFileStats(files, 10)
		if result[0].Additions != 30 {
			t.Errorf("expected additions 30, got %d", result[0].Additions)
		}
		if result[0].Deletions != 15 {
			t.Errorf("expected deletions 15, got %d", result[0].Deletions)
		}
	})
}

func TestComputeSummary(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	tests := []struct {
		name         string
		commits      []domain.Commit
		authorStats  []AuthorStats
		fileStats    []FileStats
		wantCommits  int
		wantAuthors  int
		wantFiles    int
		wantFirst    time.Time
		wantLast     time.Time
		wantAdds     int
		wantDels     int
	}{
		{
			name:        "empty inputs",
			commits:     []domain.Commit{},
			authorStats: []AuthorStats{},
			fileStats:   []FileStats{},
			wantCommits: 0,
			wantAuthors: 0,
			wantFiles:   0,
		},
		{
			name: "full statistics",
			commits: []domain.Commit{
				{Hash: "aaa", Date: now},
				{Hash: "bbb", Date: yesterday},
				{Hash: "ccc", Date: lastWeek},
			},
			authorStats: []AuthorStats{
				{Name: "Alice", Commits: 2},
				{Name: "Bob", Commits: 1},
			},
			fileStats: []FileStats{
				{Path: "main.go", Additions: 100, Deletions: 50},
				{Path: "util.go", Additions: 30, Deletions: 10},
			},
			wantCommits: 3,
			wantAuthors: 2,
			wantFiles:   2,
			wantFirst:   lastWeek,
			wantLast:    now,
			wantAdds:    130,
			wantDels:    60,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeSummary(tc.commits, tc.authorStats, tc.fileStats)

			if result.TotalCommits != tc.wantCommits {
				t.Errorf("TotalCommits = %d, want %d", result.TotalCommits, tc.wantCommits)
			}
			if result.TotalAuthors != tc.wantAuthors {
				t.Errorf("TotalAuthors = %d, want %d", result.TotalAuthors, tc.wantAuthors)
			}
			if result.TotalFiles != tc.wantFiles {
				t.Errorf("TotalFiles = %d, want %d", result.TotalFiles, tc.wantFiles)
			}
			if result.TotalAdditions != tc.wantAdds {
				t.Errorf("TotalAdditions = %d, want %d", result.TotalAdditions, tc.wantAdds)
			}
			if result.TotalDeletions != tc.wantDels {
				t.Errorf("TotalDeletions = %d, want %d", result.TotalDeletions, tc.wantDels)
			}

			if len(tc.commits) > 0 {
				if !result.FirstCommit.Equal(tc.wantFirst) {
					t.Errorf("FirstCommit = %v, want %v", result.FirstCommit, tc.wantFirst)
				}
				if !result.LastCommit.Equal(tc.wantLast) {
					t.Errorf("LastCommit = %v, want %v", result.LastCommit, tc.wantLast)
				}
			}
		})
	}
}
