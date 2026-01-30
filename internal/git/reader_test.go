package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// testRepo holds a temporary git repository for testing
type testRepo struct {
	path   string
	repo   *git.Repository
	hashes []string // commit hashes in order (newest first after setup)
}

// setupTestRepo creates an isolated git repository with known commits
func setupTestRepo(t *testing.T) *testRepo {
	t.Helper()

	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	tr := &testRepo{path: dir, repo: repo}

	// Commit 1: Initial commit with README
	writeFile(t, dir, "README.md", "# Test Repo\n")
	wt.Add("README.md")
	hash1, err := wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Commit 2: Add a source file
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	wt.Add("main.go")
	hash2, err := wt.Commit("Add main.go", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "test@example.com",
			When:  time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Commit 3: Modify README and delete nothing (simple modify)
	writeFile(t, dir, "README.md", "# Test Repo\n\nUpdated content.\n")
	wt.Add("README.md")
	hash3, err := wt.Commit("Update README", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Another Dev",
			Email: "dev@example.com",
			When:  time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Store hashes newest-first (as LoadCommits returns them)
	tr.hashes = []string{hash3.String(), hash2.String(), hash1.String()}

	return tr
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func TestLoadCommits(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	commits, err := r.LoadCommits(tr.path, 10)
	if err != nil {
		t.Fatalf("LoadCommits failed: %v", err)
	}

	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}

	// Verify newest commit (index 0)
	c := commits[0]
	if c.Hash != tr.hashes[0] {
		t.Errorf("expected hash %s, got %s", tr.hashes[0], c.Hash)
	}
	if c.ShortHash != tr.hashes[0][:7] {
		t.Errorf("expected short hash %s, got %s", tr.hashes[0][:7], c.ShortHash)
	}
	if c.Message != "Update README" {
		t.Errorf("expected message 'Update README', got %q", c.Message)
	}
	if c.Author != "Another Dev" {
		t.Errorf("expected author 'Another Dev', got %q", c.Author)
	}
	if c.Email != "dev@example.com" {
		t.Errorf("expected email 'dev@example.com', got %q", c.Email)
	}

	// Verify parent relationship
	if len(c.Parents) != 1 || c.Parents[0] != tr.hashes[1] {
		t.Errorf("expected parent %s, got %v", tr.hashes[1], c.Parents)
	}

	// Verify oldest commit has no parents
	oldest := commits[2]
	if len(oldest.Parents) != 0 {
		t.Errorf("initial commit should have no parents, got %v", oldest.Parents)
	}
}

func TestLoadCommits_Limit(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	commits, err := r.LoadCommits(tr.path, 2)
	if err != nil {
		t.Fatalf("LoadCommits failed: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("expected 2 commits with limit, got %d", len(commits))
	}
}

func TestLoadCommits_NotGitRepo(t *testing.T) {
	r := NewReader()
	dir := t.TempDir() // Empty dir, not a git repo

	_, err := r.LoadCommits(dir, 10)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestLoadCommits_NonexistentPath(t *testing.T) {
	r := NewReader()

	_, err := r.LoadCommits("/nonexistent/path/12345", 10)
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadBranches(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	branches, err := r.LoadBranches(tr.path)
	if err != nil {
		t.Fatalf("LoadBranches failed: %v", err)
	}

	// Should have at least master/main branch
	if len(branches) == 0 {
		t.Fatal("expected at least one branch")
	}

	// Find the default branch
	found := false
	for _, b := range branches {
		if b.Name == "master" || b.Name == "main" {
			found = true
			if b.HeadHash == "" {
				t.Error("branch head hash should not be empty")
			}
			if b.IsRemote {
				t.Error("local branch should not be marked as remote")
			}
		}
	}
	if !found {
		t.Errorf("expected master or main branch, got: %v", branches)
	}
}

func TestLoadRepository(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	repo, err := r.LoadRepository(tr.path)
	if err != nil {
		t.Fatalf("LoadRepository failed: %v", err)
	}

	if repo.Path != tr.path {
		t.Errorf("expected path %q, got %q", tr.path, repo.Path)
	}

	if len(repo.Commits) != 3 {
		t.Errorf("expected 3 commits, got %d", len(repo.Commits))
	}

	if len(repo.Branches) == 0 {
		t.Error("expected at least one branch")
	}

	// HEAD should be set
	if repo.HEAD == "" {
		t.Error("HEAD should not be empty")
	}
}

func TestLoadFileChanges(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	// Test commit that added main.go (hash index 1)
	changes, err := r.LoadFileChanges(tr.path, tr.hashes[1])
	if err != nil {
		t.Fatalf("LoadFileChanges failed: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 file change, got %d", len(changes))
	}

	fc := changes[0]
	if fc.Path != "main.go" {
		t.Errorf("expected path 'main.go', got %q", fc.Path)
	}
	if fc.Status != 0 { // FileAdded = 0
		t.Errorf("expected FileAdded status, got %d", fc.Status)
	}
	if fc.Additions == 0 {
		t.Error("expected additions > 0")
	}
}

func TestLoadFileChanges_InitialCommit(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	// Initial commit (oldest, index 2)
	changes, err := r.LoadFileChanges(tr.path, tr.hashes[2])
	if err != nil {
		t.Fatalf("LoadFileChanges failed: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 file in initial commit, got %d", len(changes))
	}

	if changes[0].Path != "README.md" {
		t.Errorf("expected README.md, got %q", changes[0].Path)
	}
}

func TestLoadFileChanges_ModifiedFile(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	// Commit that modified README (newest, index 0)
	changes, err := r.LoadFileChanges(tr.path, tr.hashes[0])
	if err != nil {
		t.Fatalf("LoadFileChanges failed: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 file change, got %d", len(changes))
	}

	fc := changes[0]
	if fc.Path != "README.md" {
		t.Errorf("expected README.md, got %q", fc.Path)
	}
	if fc.Status != 1 { // FileModified = 1
		t.Errorf("expected FileModified status (1), got %d", fc.Status)
	}
}

func TestLoadFileDiff(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	// Get diff for README.md in the modify commit
	diff, isBinary, err := r.LoadFileDiff(tr.path, tr.hashes[0], "README.md")
	if err != nil {
		t.Fatalf("LoadFileDiff failed: %v", err)
	}

	if isBinary {
		t.Error("README.md should not be binary")
	}

	if diff == "" {
		t.Error("diff should not be empty")
	}

	// Diff should contain the added line
	if !contains(diff, "Updated content") {
		t.Errorf("diff should contain 'Updated content', got: %s", diff)
	}
}

func TestLoadFileDiff_AddedFile(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	// Get diff for main.go when it was added
	diff, isBinary, err := r.LoadFileDiff(tr.path, tr.hashes[1], "main.go")
	if err != nil {
		t.Fatalf("LoadFileDiff failed: %v", err)
	}

	if isBinary {
		t.Error("main.go should not be binary")
	}

	if diff == "" {
		t.Error("diff should not be empty for added file")
	}
}

func TestLoadFileDiff_FileNotFound(t *testing.T) {
	tr := setupTestRepo(t)
	r := NewReader()

	diff, _, err := r.LoadFileDiff(tr.path, tr.hashes[0], "nonexistent.txt")
	if err != nil {
		t.Fatalf("LoadFileDiff should not error for missing file: %v", err)
	}

	if diff != "" {
		t.Errorf("diff should be empty for nonexistent file, got: %s", diff)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
