package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// createBenchRepo creates a repo with N commits for benchmarking
func createBenchRepo(b *testing.B, numCommits int) string {
	b.Helper()

	dir, err := os.MkdirTemp("", "gitree-bench-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}

	repo, err := git.PlainInit(dir, false)
	if err != nil {
		os.RemoveAll(dir)
		b.Fatalf("failed to init repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		os.RemoveAll(dir)
		b.Fatalf("failed to get worktree: %v", err)
	}

	// Create commits
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	for i := 0; i < numCommits; i++ {
		filename := fmt.Sprintf("file%d.txt", i%10) // Rotate through 10 files
		content := fmt.Sprintf("Content for commit %d\n", i)

		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			os.RemoveAll(dir)
			b.Fatalf("failed to write file: %v", err)
		}

		if _, err := wt.Add(filename); err != nil {
			os.RemoveAll(dir)
			b.Fatalf("failed to add file: %v", err)
		}

		_, err := wt.Commit(fmt.Sprintf("Commit %d", i), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Bench Author",
				Email: "bench@example.com",
				When:  baseTime.Add(time.Duration(i) * time.Hour),
			},
		})
		if err != nil {
			os.RemoveAll(dir)
			b.Fatalf("failed to commit: %v", err)
		}
	}

	return dir
}

func BenchmarkLoadCommits(b *testing.B) {
	sizes := []int{50, 100, 200}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
			// Setup: create repo with N commits
			dir := createBenchRepo(b, size)
			defer os.RemoveAll(dir)

			r := NewReader()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := r.LoadCommits(dir, 0) // 0 = no limit
				if err != nil {
					b.Fatalf("LoadCommits failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkLoadCommits_WithLimit(b *testing.B) {
	// Create a larger repo, benchmark with limit
	dir := createBenchRepo(b, 200)
	defer os.RemoveAll(dir)

	r := NewReader()
	limits := []int{10, 50, 100}

	for _, limit := range limits {
		b.Run(fmt.Sprintf("limit_%d", limit), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := r.LoadCommits(dir, limit)
				if err != nil {
					b.Fatalf("LoadCommits failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkLoadRepository(b *testing.B) {
	dir := createBenchRepo(b, 100)
	defer os.RemoveAll(dir)

	r := NewReader()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.LoadRepository(dir)
		if err != nil {
			b.Fatalf("LoadRepository failed: %v", err)
		}
	}
}

func BenchmarkLoadFileChanges(b *testing.B) {
	dir := createBenchRepo(b, 50)
	defer os.RemoveAll(dir)

	r := NewReader()

	// Get a commit hash to benchmark
	commits, err := r.LoadCommits(dir, 1)
	if err != nil || len(commits) == 0 {
		b.Fatalf("failed to get commits: %v", err)
	}
	hash := commits[0].Hash

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.LoadFileChanges(dir, hash)
		if err != nil {
			b.Fatalf("LoadFileChanges failed: %v", err)
		}
	}
}

func BenchmarkLoadCommits_Memory(b *testing.B) {
	dir := createBenchRepo(b, 100)
	defer os.RemoveAll(dir)

	r := NewReader()
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := r.LoadCommits(dir, 0)
		if err != nil {
			b.Fatalf("LoadCommits failed: %v", err)
		}
	}
}
