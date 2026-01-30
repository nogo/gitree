package graph

import (
	"fmt"
	"testing"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

// generateLinearCommits creates N commits in a linear chain
func generateLinearCommits(n int) []domain.Commit {
	commits := make([]domain.Commit, n)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	for i := 0; i < n; i++ {
		hash := fmt.Sprintf("%040d", i) // 40-char hash
		var parents []string
		if i < n-1 {
			parents = []string{fmt.Sprintf("%040d", i+1)}
		}

		commits[i] = domain.Commit{
			Hash:      hash,
			ShortHash: hash[:7],
			Author:    "Test Author",
			Email:     "test@example.com",
			Date:      baseTime.Add(-time.Duration(i) * time.Hour),
			Message:   fmt.Sprintf("Commit %d", i),
			Parents:   parents,
		}
	}
	return commits
}

// generateBranchingCommits creates commits with branches and merges
func generateBranchingCommits(n int) []domain.Commit {
	commits := make([]domain.Commit, 0, n)
	baseTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	idx := 0
	for idx < n {
		hash := fmt.Sprintf("%040d", idx)

		// Every 10 commits, create a branch pattern
		if idx > 0 && idx%10 == 0 && idx+3 < n {
			// Merge commit
			prevHash := fmt.Sprintf("%040d", idx-1)
			branchHash := fmt.Sprintf("%040d", idx+1)
			commits = append(commits, domain.Commit{
				Hash:      hash,
				ShortHash: hash[:7],
				Author:    "Test Author",
				Email:     "test@example.com",
				Date:      baseTime.Add(-time.Duration(idx) * time.Hour),
				Message:   fmt.Sprintf("Merge commit %d", idx),
				Parents:   []string{prevHash, branchHash},
			})
			idx++

			// Branch commit
			hash = fmt.Sprintf("%040d", idx)
			rootHash := fmt.Sprintf("%040d", idx+2)
			commits = append(commits, domain.Commit{
				Hash:      hash,
				ShortHash: hash[:7],
				Author:    "Test Author",
				Email:     "test@example.com",
				Date:      baseTime.Add(-time.Duration(idx) * time.Hour),
				Message:   fmt.Sprintf("Branch commit %d", idx),
				Parents:   []string{rootHash},
			})
			idx++
		} else {
			var parents []string
			if idx < n-1 {
				parents = []string{fmt.Sprintf("%040d", idx+1)}
			}

			commits = append(commits, domain.Commit{
				Hash:      hash,
				ShortHash: hash[:7],
				Author:    "Test Author",
				Email:     "test@example.com",
				Date:      baseTime.Add(-time.Duration(idx) * time.Hour),
				Message:   fmt.Sprintf("Commit %d", idx),
				Parents:   parents,
			})
			idx++
		}
	}

	return commits[:n] // Ensure exact count
}

func BenchmarkBuildLayout_Linear(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}

	for _, size := range sizes {
		commits := generateLinearCommits(size)
		b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BuildLayout(commits)
			}
		})
	}
}

func BenchmarkBuildLayout_Branching(b *testing.B) {
	sizes := []int{100, 500, 1000}

	for _, size := range sizes {
		commits := generateBranchingCommits(size)
		b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				BuildLayout(commits)
			}
		})
	}
}

func BenchmarkRenderRow(b *testing.B) {
	sizes := []int{100, 500, 1000}

	for _, size := range sizes {
		commits := generateLinearCommits(size)
		layout := BuildLayout(commits)
		palette := NewColorPalette()
		renderer := NewRowRenderer(layout, palette)

		b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Render middle row (typical viewport behavior)
				renderer.RenderRow(size / 2)
			}
		})
	}
}

func BenchmarkRenderAllRows(b *testing.B) {
	sizes := []int{100, 500}

	for _, size := range sizes {
		commits := generateLinearCommits(size)
		layout := BuildLayout(commits)
		palette := NewColorPalette()
		renderer := NewRowRenderer(layout, palette)

		b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for row := 0; row < size; row++ {
					renderer.RenderRow(row)
				}
			}
		})
	}
}

func BenchmarkRenderViewport(b *testing.B) {
	// Simulate rendering only visible rows (typical 30-50 rows)
	commits := generateBranchingCommits(1000)
	layout := BuildLayout(commits)
	palette := NewColorPalette()
	renderer := NewRowRenderer(layout, palette)

	viewportSizes := []int{20, 40, 60}

	for _, vpSize := range viewportSizes {
		b.Run(fmt.Sprintf("viewport_%d", vpSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Render from middle of list
				start := 500
				for row := start; row < start+vpSize; row++ {
					renderer.RenderRow(row)
				}
			}
		})
	}
}

func BenchmarkBuildLayout_Memory(b *testing.B) {
	commits := generateLinearCommits(1000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		BuildLayout(commits)
	}
}
