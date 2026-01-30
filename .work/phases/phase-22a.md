# Phase 22a: UI/Rendering Fixes

## Status: IN PROGRESS

## Goal
Fix UI bugs discovered during performance testing with large repositories.

## Known Issues

### Issue 1: Cursor Bug - Multiple cursors appear when scrolling ❌ OPEN
**Symptom:** After scrolling up a lot (especially near end of history), multiple rows show `>` cursor indicator.

**Reproduction:**
1. Open large repo (e.g., zed with 43k commits)
2. Go to earliest commits (G or End key)
3. Scroll up several times (k key)
4. Multiple `>` cursors appear on different rows

**Suspected causes:**
- Viewport sync issue at list boundaries
- Off-by-one error in cursor position vs viewOffset
- ANSI escape codes leaking between rows
- Graph rendering producing `>` like characters

**Investigation needed:**
- [ ] Add logging to track cursor vs viewOffset values
- [ ] Check syncViewport() boundary conditions
- [ ] Verify graph characters don't include `>`
- [ ] Test with minimal repo to isolate

**Files to investigate:**
- `internal/tui/list/list.go` - syncViewport(), renderVisibleRows()
- `internal/tui/graph/render.go` - RenderRow()

---

### Issue 2: Graph toggle (nice to have)
**Request:** Press `g` to toggle graph visibility for repos with many lanes.

**Status:** Deferred - dynamic graph width partially addresses this.

---

## Completed Work (from Phase 22)

### Rendering Architecture Refactor ✅
- Created `text/` package for ANSI-aware string utilities
- Created `RowLayout` struct for column width management
- Created `Row` struct to separate data from rendering
- Replaced scattered `fmt.Sprintf` with structured rendering
- Updated to Go 1.22+ features (slices.Sort, range-over-int)

### Dynamic Graph Width ✅
- Graph column adjusts based on visible viewport
- Early history (few lanes) → narrow graph → more message space
- Headers sync with content layout

---

## Acceptance Criteria

- [ ] Cursor appears on exactly ONE row at all times
- [ ] Scrolling works correctly at all list boundaries
- [ ] No rendering artifacts when scrolling quickly
- [x] Dynamic graph width working
- [x] Column headers align with content

---

## Files Changed

- `internal/tui/text/text.go` - NEW: ANSI utilities
- `internal/tui/text/stats.go` - NEW: FileStats struct
- `internal/tui/list/layout.go` - NEW: RowLayout
- `internal/tui/list/row.go` - NEW: Row struct
- `internal/tui/list/list.go` - Refactored rendering
- `internal/tui/list/expanded.go` - Use layout param
- `internal/tui/graph/graph.go` - MaxLanesInRange()
- `internal/tui/graph/lanes.go` - MaxLaneAt(), slices.Sort
- `internal/tui/graph/render.go` - range-over-int
- `internal/tui/layout.go` - ViewportLayout()
