# Phase 1: Stats Computation

## Goal
Create data aggregation functions that compute author and file statistics from a commit list.

## Changes
- `internal/tui/insights/stats.go`: New file with stats types and computation

## Implementation
1. Create `internal/tui/insights/` package
2. Define `AuthorStats` struct:
   - Name, Email string
   - Commits int
   - Additions, Deletions int
3. Define `FileStats` struct:
   - Path string
   - ChangeCount int
   - Additions, Deletions int
4. Implement `ComputeAuthorStats(commits []*domain.Commit) []AuthorStats`:
   - Aggregate commits by author email
   - Sort by commit count descending
   - Return top N (configurable, default 10)
5. Implement `ComputeFileStats(commits []*domain.Commit, files map[string][]domain.FileChange) []FileStats`:
   - Aggregate file changes across commits
   - Sort by change count descending
   - Return top N
6. Implement `ComputeSummary(commits, authorStats, fileStats)`:
   - Total commits, total authors, date range
   - Total additions/deletions

## Testing
- Create stats with 5 commits from 3 authors → top author has most commits
- Empty commit list → empty stats, no panic
- Single author → returns that author with 100%

## Estimated LOC
~80
