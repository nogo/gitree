package insights

import (
	"sort"
	"time"

	"github.com/nogo/gitree/internal/domain"
)

// AuthorStats holds aggregated statistics for a single author.
type AuthorStats struct {
	Name      string
	Email     string
	Commits   int
	Additions int
	Deletions int
}

// FileStats holds aggregated statistics for a single file.
type FileStats struct {
	Path        string
	ChangeCount int
	Additions   int
	Deletions   int
}

// Summary holds overall repository statistics.
type Summary struct {
	TotalCommits  int
	TotalAuthors  int
	TotalFiles    int
	FirstCommit   time.Time
	LastCommit    time.Time
	TotalAdditions int
	TotalDeletions int
}

// ComputeAuthorStats aggregates commit statistics by author email.
// Returns authors sorted by commit count descending, limited to topN results.
func ComputeAuthorStats(commits []domain.Commit, topN int) []AuthorStats {
	if len(commits) == 0 {
		return nil
	}

	// Aggregate by email
	byEmail := make(map[string]*AuthorStats)
	for _, c := range commits {
		stats, ok := byEmail[c.Email]
		if !ok {
			stats = &AuthorStats{
				Name:  c.Author,
				Email: c.Email,
			}
			byEmail[c.Email] = stats
		}
		stats.Commits++
	}

	// Convert to slice
	result := make([]AuthorStats, 0, len(byEmail))
	for _, s := range byEmail {
		result = append(result, *s)
	}

	// Sort by commit count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Commits > result[j].Commits
	})

	// Limit to topN
	if topN > 0 && len(result) > topN {
		result = result[:topN]
	}

	return result
}

// ComputeFileStats aggregates file change statistics across commits.
// Returns files sorted by change count descending, limited to topN results.
func ComputeFileStats(files map[string][]domain.FileChange, topN int) []FileStats {
	if len(files) == 0 {
		return nil
	}

	// Aggregate by file path
	byPath := make(map[string]*FileStats)
	for _, changes := range files {
		for _, fc := range changes {
			stats, ok := byPath[fc.Path]
			if !ok {
				stats = &FileStats{Path: fc.Path}
				byPath[fc.Path] = stats
			}
			stats.ChangeCount++
			stats.Additions += fc.Additions
			stats.Deletions += fc.Deletions
		}
	}

	// Convert to slice
	result := make([]FileStats, 0, len(byPath))
	for _, s := range byPath {
		result = append(result, *s)
	}

	// Sort by change count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].ChangeCount > result[j].ChangeCount
	})

	// Limit to topN
	if topN > 0 && len(result) > topN {
		result = result[:topN]
	}

	return result
}

// ComputeSummary computes overall repository statistics.
func ComputeSummary(commits []domain.Commit, authorStats []AuthorStats, fileStats []FileStats) Summary {
	var s Summary

	s.TotalCommits = len(commits)
	s.TotalAuthors = len(authorStats)
	s.TotalFiles = len(fileStats)

	// Calculate date range
	if len(commits) > 0 {
		s.FirstCommit = commits[0].Date
		s.LastCommit = commits[0].Date

		for _, c := range commits {
			if c.Date.Before(s.FirstCommit) {
				s.FirstCommit = c.Date
			}
			if c.Date.After(s.LastCommit) {
				s.LastCommit = c.Date
			}
		}
	}

	// Sum additions/deletions from file stats
	for _, fs := range fileStats {
		s.TotalAdditions += fs.Additions
		s.TotalDeletions += fs.Deletions
	}

	return s
}
