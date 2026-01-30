# Phases 12-17 Summary: Filtering, Search & Commit Details

> Completed: January 2026

## Overview

These phases added advanced filtering, search capabilities, time-based navigation, and detailed commit inspection with inline expansion and diff viewing.

---

## Phase 12: Author Filter

**Commit:** `e18848f`

Filter commits by author, similar to the branch filter pattern.

### Features
- Modal overlay with author list (sorted by commit count)
- Multi-select with Space, toggle all with `a`/`n`
- Shows commit count per author
- Applies as intersection filter with branch filter
- Keybinding: `a`

### Files
- `internal/tui/filter/author.go` - Author filter modal component
- `internal/tui/app.go` - Keybinding and filter application

---

## Phase 13: Author Highlight

**Commit:** `86c26c1`

Highlight commits by a specific author without hiding others.

### Features
- Modal to select single author to highlight
- Non-matching commits are dimmed (graph + text)
- Visual distinction from filtering (dims vs hides)
- Keybinding: `A` (shift+a)

### Files
- `internal/tui/filter/author_highlight.go` - Highlight modal
- `internal/tui/list/list.go` - Dimming logic in row rendering
- `internal/tui/graph/render.go` - Dimmed graph rendering

---

## Phase 14: Search

**Commit:** `72203f4`

Search commits by message or hash.

### Features
- Search input with `/` keybinding
- Case-insensitive substring match on message and hash
- Match markers (`*`) in commit list
- Navigate matches with `n`/`N`
- Search state shown in footer
- Cursor jumps to matches

### Files
- `internal/tui/search/search.go` - Search component with input and matching
- `internal/tui/list/list.go` - Match indicator rendering
- `internal/tui/app.go` - Search keybindings and state

---

## Phase 15: Date Histogram

**Commit:** `e4c75a2`

Video player-style timeline showing commit density over time.

### Features
- Braille density characters: `⣀⣤⣶⣿`
- Auto-scaling bins: daily/weekly/monthly based on date range
- Selection brackets for time range filtering
- Keyboard navigation: `h`/`l` move, `[`/`]` set range
- Toggle visibility with `t`, focus with `Tab`
- Position: bottom of screen, 3 lines height

### Files
- `internal/tui/histogram/histogram.go` - Histogram component
- `internal/tui/histogram/styles.go` - Histogram styling
- `internal/tui/app.go` - Time filter integration
- `internal/tui/layout.go` - Layout with histogram

---

## Phase 16: File List in Commit Detail

**Commit:** `6425556`

Show changed files when viewing commit details.

### Features
- Inline expansion (not modal) keeps graph visible
- Two-column layout on wide terminals (>= 100 chars)
- Single-column layout on narrow terminals
- File list with status indicators (A/M/D/R)
- Addition/deletion counts per file and total
- File navigation with `j`/`k` when expanded

### Files
- `internal/tui/list/expanded.go` - Inline expanded view rendering
- `internal/tui/list/list.go` - Expansion state and file cursor
- `internal/git/reader.go` - `LoadFileChanges()` method
- `internal/domain/types.go` - `FileChange` type

---

## Phase 17: Diff View

**Commit:** `595a9c5`

View file diffs from the expanded commit view.

### Features
- Full-screen diff overlay
- Syntax highlighting: additions (green), deletions (red)
- Line numbers and hunk headers
- Navigate between files with `h`/`l`
- Scroll diff content with `j`/`k`
- Binary file detection

### Files
- `internal/tui/diff/diff.go` - Diff view component
- `internal/tui/diff/render.go` - Diff rendering with syntax highlighting
- `internal/tui/diff/styles.go` - Diff color styles
- `internal/git/reader.go` - `LoadFileDiff()` method

---

## Post-Phase Refinements

**Commit:** `a421040`

After Phase 17, the inline expanded view was refined:
- Two-column layout (metadata | files) when terminal >= 100 chars wide
- Keyboard and mouse scrolling blocked when expanded
- Loading indicator when files are being fetched
- Help text shows `[j/k] file  [Enter] diff  [Esc] close`

---

## Keybindings Summary

| Key | Action | Phase |
|-----|--------|-------|
| `a` | Author filter | 12 |
| `A` | Author highlight | 13 |
| `/` | Search | 14 |
| `n`/`N` | Next/prev match | 14 |
| `t` | Toggle histogram | 15 |
| `Tab` | Focus histogram | 15 |
| `Enter` | Expand commit | 16 |
| `j`/`k` (expanded) | Navigate files | 16 |
| `Enter` (expanded) | Open diff | 17 |
| `Esc` | Close overlay/collapse | All |
| `c` | Clear all filters | All |

---

## Architecture Notes

### Pattern: Inline Expansion vs Modal
Phase 16 introduced inline expansion instead of modal overlays:
- Keeps graph visible during commit inspection
- Content inserted between rows in the list
- Viewport scrolling disabled when expanded
- Better context awareness than full-screen modal

### Pattern: Async Loading
File changes and diffs are loaded asynchronously:
- `tea.Cmd` returns loading function
- Result delivered via custom `tea.Msg` types
- UI shows loading state while waiting

### Two-Column Layout
When terminal width >= 100 characters:
```
╔════════════════════════════╤════════════════════════════╗
║ Commit: abc123...          │ Files (5)  +42 -18         ║
║ Author: Alice              │ > M src/app.go      +10 -5 ║
║ Date:   Jan 30, 2026       │   A src/new.go      +32 -0 ║
║ Parent: def456             │   D old/file.go     +0 -13 ║
╚══════ [j/k] file  [Enter] diff  [Esc] close ═══════════╝
```
