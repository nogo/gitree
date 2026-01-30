# Phase 23: Tag Visualization

## Goal
Display git tags as badges on commits in the graph view.

## Context
- Branch badges already work: `[main]` `[origin/main]`
- Tags are similar - annotated and lightweight
- Tags should appear alongside branch badges

## Scope
~80 LOC

## Tasks

### 1. Extend domain types

```go
// internal/domain/types.go
type Commit struct {
    // existing fields...
    Tags []string  // NEW: tag names pointing to this commit
}
```

### 2. Load tags in git reader

```go
// internal/git/reader.go
func (r *Reader) loadTags() (map[string][]string, error) {
    // Returns map[commitHash][]tagName
    // Handle both annotated and lightweight tags
}
```

Update `LoadCommits()` to populate `Tags` field.

### 3. Render tag badges

```go
// internal/tui/graph/graph.go or list/list.go
// Tag badge style: different from branch (e.g., yellow bg or different brackets)
// Example: <v1.0.0> or [v1.0.0] with distinct color
```

**Badge order:** HEAD indicator → branch badges → tag badges → message

### 4. Update styles

```go
// internal/tui/styles.go
TagBadge = lipgloss.NewStyle().
    Background(lipgloss.Color("yellow")).
    Foreground(lipgloss.Color("black")).
    Padding(0, 1)
```

## Visual Design

```
  ○  [main] <v1.0.0> feat: release version 1.0    Alice  30 Jan  a1b2c
  ●  fix: bug in auth                              Bob    29 Jan  b2c3d
  ●  <v0.9.0> prepare release                      Alice  28 Jan  c3d4e
```

**Badge styles:**
- Branches: `[name]` with branch color
- Tags: `<name>` with yellow/gold color (stands out as milestone marker)

## Acceptance Criteria
- [ ] Tags appear as badges on commits
- [ ] Both annotated and lightweight tags shown
- [ ] Visual distinction from branch badges
- [ ] Performance: tag loading doesn't slow down large repos

## Files to Read First
- `internal/domain/types.go` - Commit struct
- `internal/git/reader.go` - how branches are loaded (pattern to follow)
- `internal/tui/graph/graph.go` - badge rendering
- `internal/tui/styles.go` - existing badge styles

## Dependencies
None - independent feature

## Notes
- go-git: use `repo.Tags()` iterator
- Handle tag → commit resolution (annotated tags point to tag object, not commit)
- Consider: truncate long tag names? (v1.0.0-beta.1-rc.2...)
