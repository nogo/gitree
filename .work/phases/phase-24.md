# Phase 24: Tag Filter

## Goal
Add tag filter modal similar to branch filter.

## Context
- Branch filter exists: `b` key, modal with checkboxes
- Author filter exists: `a` key, same pattern
- Tag filter follows same UX pattern

## Scope
~100 LOC

## Tasks

### 1. Create tag filter component

```go
// internal/tui/filter/tag.go (NEW)
type TagFilter struct {
    tags     []string
    selected map[string]bool
    cursor   int
    // ... similar to branch filter
}
```

### 2. Wire up keybinding

```go
// internal/tui/app.go
case "T": // shift+t for tag filter (t is taken by histogram toggle)
    return m.openTagFilter()
```

### 3. Apply tag filter

Filter logic:
- If no tags selected → show all commits
- If tags selected → show only commits with those tags AND their ancestors

**Question:** Should filtering by tag show:
- A) Only commits with that exact tag
- B) Tag commit + all ancestors (history leading to tag)
- C) Commits between selected tags

**Recommendation:** Option B (tag + ancestors) - most useful for "show me v1.0 history"

### 4. Update footer

Show active tag filter: `tag:2/5` similar to branch filter

### 5. Clear filter integration

`c` key should also clear tag filter.

## Visual Design

```
┌─ Tags ──────────────────────────┐
│ [x] v1.0.0                      │
│ [ ] v0.9.0                      │
│ [x] v0.8.0                      │
│ [ ] beta-1                      │
├─────────────────────────────────┤
│ [a]ll [n]one [Space] toggle     │
│ [Enter] apply  [Esc] cancel     │
└─────────────────────────────────┘
```

## Acceptance Criteria
- [ ] `T` opens tag filter modal
- [ ] Multi-select with Space, all/none shortcuts
- [ ] Filter shows tag commits + ancestors
- [ ] Footer shows `tag:X/Y` when active
- [ ] `c` clears tag filter along with other filters

## Files to Read First
- `internal/tui/filter/branch.go` - pattern to follow
- `internal/tui/filter/author.go` - another example
- `internal/tui/app.go` - filter integration

## Dependencies
- Phase 23 (tag visualization - tags must be loaded)

## Notes
- Reuse filter modal patterns for consistency
- Tags sorted by: semantic version? date? alphabetical?
- Consider grouping: release tags vs other tags
