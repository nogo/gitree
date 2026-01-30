# Phase 20: Git Layer Tests

## Goal
Test the git reader implementation with fixture repositories.

## Context
- `internal/git/reader.go` wraps go-git
- Methods: `LoadRepository()`, `LoadCommits()`, `LoadBranches()`, `LoadFileChanges()`, `LoadFileDiff()`
- Need test fixtures (small git repos) or mocks

## Scope
~150 LOC

## Tasks

### 1. Create test fixtures

Option A: In-memory git repo using go-git
```go
// Create temp repo with known commits for testing
func setupTestRepo(t *testing.T) (string, func()) {
    dir := t.TempDir()
    // Initialize repo, create commits
    return dir, func() { /* cleanup */ }
}
```

Option B: Use testdata directory with small real repo
```
internal/git/testdata/
└── test-repo/
    └── .git/
```

**Recommendation:** Option A (in-memory) - more flexible, no git LFS issues

### 2. Test Reader methods

```
internal/git/
├── reader.go
└── reader_test.go    # NEW
```

Test cases:
- `LoadRepository()` - valid repo, invalid path, not a git repo
- `LoadCommits()` - returns commits in order, populates all fields
- `LoadBranches()` - local branches, remote branches
- `LoadFileChanges()` - added/modified/deleted files
- `LoadFileDiff()` - diff content, binary detection

### 3. Error handling tests
- Non-existent path
- Path exists but not git repo
- Corrupt repo (if feasible)

## Acceptance Criteria
- [ ] Reader tests cover happy path for all public methods
- [ ] Error cases return appropriate errors
- [ ] Tests are isolated (don't depend on external repos)

## Files to Read First
- `internal/git/reader.go` - implementation to test
- `internal/domain/types.go` - return types

## Dependencies
- Phase 18 (test infrastructure)

## Notes
- go-git has good in-memory filesystem support
- Keep test repo small (3-5 commits max)
- Don't test go-git itself, test our wrapper logic
