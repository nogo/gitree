# Phase 22a: UI/Rendering Fixes

## Status: IN PROGRESS

## Goal
Fix UI bugs discovered during performance testing with large repositories.

## Known Issues

### Issue 1: Cursor Bug - Multiple cursors appear when scrolling ✅ PARTIAL FIX
**Symptom:** After scrolling up a lot (especially near end of history), multiple rows show `>` cursor indicator and header disappears.

**Root cause:** Zed editor's built-in terminal has rendering quirks - it doesn't properly clear/redraw lines when content changes rapidly (e.g., Page Up/Down).

**Fix applied:**
- All rows now padded to full terminal width (overwrites old content)
- Empty padding rows are full-width spaces instead of empty strings
- Expanded view reserves space correctly to prevent overflow

**Known limitation:**
- Page Up/Down in **Zed's terminal only** may still show brief rendering glitches
- Workaround: continue scrolling (one more keystroke fixes it)
- Works correctly in iTerm2, Terminal.app, Kitty, Alacritty, etc.

**Files changed:**
- `internal/tui/list/list.go` - Full-width padding for rows and empty lines
- `internal/tui/list/row.go` - Non-selected rows now padded to full width

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

- [x] Cursor appears on exactly ONE row at all times (except Zed terminal quirk)
- [x] Scrolling works correctly at all list boundaries
- [x] No rendering artifacts when scrolling quickly (except Zed terminal Page Up/Down)
- [x] Dynamic graph width working
- [x] Column headers align with content
- [x] Expanded commit view doesn't push header off screen

**Note:** Zed editor's terminal has known rendering issues with rapid screen updates. Works correctly in standard terminals.

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
