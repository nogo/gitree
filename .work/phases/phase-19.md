# Phase 19: Graph Algorithm Tests

## Goal
Test the graph rendering logic (lane assignment, character selection).

## Context
- Graph logic in `internal/tui/graph/`
- Complex algorithm: lane assignment, merge rendering, fork detection
- High value to test - bugs here are visual and hard to spot

## Scope
~120 LOC

## Tasks

### 1. Create test file
```
internal/tui/graph/
├── lanes.go
├── render.go
├── chars.go
└── lanes_test.go    # NEW
```

### 2. Test lane assignment

Test cases:
- Linear history (single lane)
- Simple branch and merge
- Multiple parallel branches
- Octopus merge (3+ parents)

### 3. Test character selection

Test cases:
- Node characters (●/○)
- Line characters (│)
- Merge characters (─┐┘└┌)
- Fork characters

## Acceptance Criteria
- [ ] Lane assignment tests pass for linear, branch, merge cases
- [ ] Character selection tests verify correct glyphs
- [ ] Edge cases: empty graph, single commit, detached HEAD

## Files to Read First
- `internal/tui/graph/lanes.go` - lane assignment algorithm
- `internal/tui/graph/render.go` - render logic
- `internal/tui/graph/chars.go` - character definitions

## Dependencies
- Phase 18 (test infrastructure)

## Test Data Pattern
```go
func TestLaneAssignment(t *testing.T) {
    tests := []struct {
        name     string
        commits  []testCommit // simplified commit for testing
        wantLanes map[string]int // hash -> expected lane
    }{
        {
            name: "linear history",
            commits: []testCommit{
                {hash: "c", parents: []string{"b"}},
                {hash: "b", parents: []string{"a"}},
                {hash: "a", parents: nil},
            },
            wantLanes: map[string]int{"c": 0, "b": 0, "a": 0},
        },
        // more cases...
    }
    // ...
}
```

## Notes
- May need to export internal functions for testing or use `_test.go` in same package
- Focus on algorithm correctness, not rendering output
