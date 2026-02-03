package insights

import (
	"testing"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

func TestNew(t *testing.T) {
	v := New()

	if v.authorStats != nil {
		t.Error("expected authorStats to be nil")
	}
	if v.fileStats != nil {
		t.Error("expected fileStats to be nil")
	}
	if v.width != 0 || v.height != 0 {
		t.Errorf("expected dimensions to be 0, got width=%d height=%d", v.width, v.height)
	}
}

func TestSetSize(t *testing.T) {
	v := New()
	v.SetSize(100, 50)

	if v.Width() != 100 {
		t.Errorf("expected width 100, got %d", v.Width())
	}
	if v.Height() != 50 {
		t.Errorf("expected height 50, got %d", v.Height())
	}
}

func TestRecalculate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	commits := []*domain.Commit{
		{Hash: "aaa", Author: "Alice", Email: "alice@example.com", Date: now},
		{Hash: "bbb", Author: "Alice", Email: "alice@example.com", Date: yesterday},
		{Hash: "ccc", Author: "Bob", Email: "bob@example.com", Date: yesterday},
	}

	fileChanges := map[string][]domain.FileChange{
		"aaa": {{Path: "main.go", Additions: 10, Deletions: 5}},
		"bbb": {{Path: "main.go", Additions: 20, Deletions: 10}},
		"ccc": {{Path: "util.go", Additions: 5, Deletions: 2}},
	}

	v := New()
	v.Recalculate(commits, fileChanges)

	// Verify author stats populated
	authors := v.AuthorStats()
	if len(authors) != 2 {
		t.Errorf("expected 2 authors, got %d", len(authors))
	}
	if len(authors) > 0 && authors[0].Email != "alice@example.com" {
		t.Errorf("expected top author alice@example.com, got %s", authors[0].Email)
	}

	// Verify file stats populated
	files := v.FileStats()
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
	if len(files) > 0 && files[0].Path != "main.go" {
		t.Errorf("expected top file main.go, got %s", files[0].Path)
	}

	// Verify summary populated
	summary := v.Summary()
	if summary.TotalCommits != 3 {
		t.Errorf("expected 3 commits, got %d", summary.TotalCommits)
	}
	if summary.TotalAuthors != 2 {
		t.Errorf("expected 2 authors, got %d", summary.TotalAuthors)
	}
	if summary.TotalFiles != 2 {
		t.Errorf("expected 2 files, got %d", summary.TotalFiles)
	}

	// Verify calendar populated
	calendar := v.Calendar()
	if len(calendar.Cells) == 0 {
		t.Error("expected calendar cells to be populated")
	}
}

func TestRecalculateEmpty(t *testing.T) {
	v := New()
	v.Recalculate(nil, nil)

	if len(v.AuthorStats()) != 0 {
		t.Errorf("expected 0 authors, got %d", len(v.AuthorStats()))
	}
	if len(v.FileStats()) != 0 {
		t.Errorf("expected 0 files, got %d", len(v.FileStats()))
	}
	if v.Summary().TotalCommits != 0 {
		t.Errorf("expected 0 commits, got %d", v.Summary().TotalCommits)
	}
}

