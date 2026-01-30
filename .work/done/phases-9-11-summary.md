# Phases 9-11 Summary

## Phase 9: Path Argument
**Goal:** Accept repo path as CLI argument instead of requiring `cd` into repo.

**Usage:** `gitree [path]` - defaults to `.` if omitted, supports `~` expansion.

**Files:** `cmd/gitree/main.go` (parse args, validate), `internal/tui/app.go` (store path)

**Validation:**
- Expand `~` to home dir
- Convert to absolute path
- Check `.git` exists
- Error on non-existent/non-git paths

---

## Phase 10: UI Polish
**Goal:** Table-based layout with header, column alignment, and improved footer.

**Layout:**
```
gitree                                                    repo-name
───────────────────────────────────────────────────────────────────
     Message                              Author       Date   Hash
───────────────────────────────────────────────────────────────────
  ○  [main] Add feature...               Danilo K  12m ago  6144c
───────────────────────────────────────────────────────────────────
● watching   8/8 commits   1 branch                [b]ranch [q]uit
```

**Column Order:** cursor | graph | badges+message | author(10) | date(10) | hash(5)

**Files:** `layout.go` (header/footer), `list/list.go` (columns), `styles.go` (new styles), `app.go` (wire up counts)

**Key Changes:**
- Badge moves to message prefix: `[main] commit msg`
- Author/Date/Hash right-aligned, fixed width
- Footer shows: watching status, X/Y commits, Z branches

---

## Phase 11: Graph Rendering
**Goal:** Replace simple graph with DAG visualization showing lanes, merge connections, continuous branch lines.

**Current → Target:**
```
Current:              Target:
  ●   commit            │     commit
  ●   merge           ●─┐   merge
╯●   commit             │ │   commit
                      ●─┘   merge point
```

**Core Algorithm:**
1. Build commit graph (parents/children)
2. Assign lanes top-to-bottom (HEAD=lane 0, branches get new lanes)
3. Render per-row: active lanes (│), nodes (●), connections (┐┘└┌─)

**Data Structures:**
```go
CommitNode { Hash, Parents, Children, Lane, Row }
GraphState { Commits, HashToNode, MaxLanes, Rows }
```

**Character Selection:** Based on connections above/below/left/right + isNode

**Lane Colors:** Cycle through 6 colors (pink, cyan, green, orange, purple, yellow)

**Files:**
- NEW: `graph/lanes.go`, `graph/render.go`, `graph/chars.go`
- MOD: `graph/graph.go`, `list/list.go`

**Sub-phases:** 11a (lane assignment) → 11b (basic render) → 11c (merge/fork connections) → 11d (colors/polish)

**Edge Cases:** Octopus merges, wide graphs (10-15 lane cap), detached commits, linear repos

**Performance:** Lazy render visible rows, cache lane assignments, incremental updates

---

## Dependencies
Phase 10 depends on Phase 9 (needs repo path for header display).
Phase 11 is independent but integrates with Phase 10's column layout.

## Total Scope
- Phase 9: ~40 LOC
- Phase 10: ~150 LOC
- Phase 11: ~450 LOC
