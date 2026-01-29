package git

import (
	"os"
	"testing"
)

func TestLoadCommits(t *testing.T) {
	r := NewReader()

	// Find a git repo to test against
	// Try current directory, then parent directories
	path := findGitRepo()
	if path == "" {
		t.Skip("no git repository found for testing")
	}

	commits, err := r.LoadCommits(path, 10)
	if err != nil {
		t.Fatalf("LoadCommits failed: %v", err)
	}

	if len(commits) == 0 {
		t.Error("expected at least one commit")
	}

	// Verify commit fields are populated
	if len(commits) > 0 {
		c := commits[0]
		if c.Hash == "" {
			t.Error("commit hash should not be empty")
		}
		if c.ShortHash == "" || len(c.ShortHash) != 7 {
			t.Errorf("short hash should be 7 chars, got %q", c.ShortHash)
		}
		if c.Author == "" {
			t.Error("author should not be empty")
		}
		if c.Date.IsZero() {
			t.Error("date should not be zero")
		}
	}
}

func TestLoadBranches(t *testing.T) {
	r := NewReader()

	path := findGitRepo()
	if path == "" {
		t.Skip("no git repository found for testing")
	}

	branches, err := r.LoadBranches(path)
	if err != nil {
		t.Fatalf("LoadBranches failed: %v", err)
	}

	// Most repos have at least one branch
	if len(branches) == 0 {
		t.Log("warning: no branches found")
	}

	// Verify branch fields
	for _, b := range branches {
		if b.Name == "" {
			t.Error("branch name should not be empty")
		}
		if b.HeadHash == "" {
			t.Error("branch head hash should not be empty")
		}
	}
}

func TestLoadRepository(t *testing.T) {
	r := NewReader()

	path := findGitRepo()
	if path == "" {
		t.Skip("no git repository found for testing")
	}

	repo, err := r.LoadRepository(path)
	if err != nil {
		t.Fatalf("LoadRepository failed: %v", err)
	}

	if repo.Path != path {
		t.Errorf("expected path %q, got %q", path, repo.Path)
	}

	if len(repo.Commits) == 0 {
		t.Error("expected commits")
	}
}

func TestLoadCommitsNotGitRepo(t *testing.T) {
	r := NewReader()

	_, err := r.LoadCommits("/tmp", 10)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func findGitRepo() string {
	// Check if current directory or any parent is a git repo
	dir, _ := os.Getwd()
	for dir != "/" && dir != "" {
		if _, err := os.Stat(dir + "/.git"); err == nil {
			return dir
		}
		// Go up one level
		parent := dir[:max(0, len(dir)-len(dir[len(dir)-1:])-1)]
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
