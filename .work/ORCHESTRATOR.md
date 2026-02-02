# Gitree Implementation Orchestrator

> Load this file first. It coordinates phased implementation with AI coding agents.

## Project Context

**gitree** = TUI git history visualizer (Go + bubbletea + go-git)

| Aspect | Detail |
|--------|--------|
| Current version | v0.3.0 |
| Architecture | `cmd/gitree/main.go` → `internal/{domain,git,watcher,tui}` |
| Principles | SOLID, Go pragmatic, visualization-focused (not action-focused) |
| Differentiation | Time-based navigation, live watching, visual clarity |

## Current Milestone: v0.4.0 - Stabilization & Polish

**Goals:** Test coverage, performance validation, tag visualization, release check

## Phase Queue

| Phase | Name | Status | Blocks |
|-------|------|--------|--------|
| 18 | Test infrastructure | `done` | - |
| 19 | Graph algorithm tests | `done` | 18 |
| 20 | Git layer tests | `done` | 18 |
| 21 | Performance benchmarks | `done` | 18 |
| 22 | Performance fixes | `done` | 21 (REQUIRED) |
| 22a | UI/rendering fixes | `done` | 22 |
| 23 | Tag visualization | `done` | - |
| 24 | Tag filter | `pending` | 23 |
| 25 | Release check | `pending` | - |

**Status values:** `pending` | `in_progress` | `done` | `skipped`

## Sub-Agent Instructions

### Before Starting a Phase
1. Read the phase file: `.work/phases/phase-{N}.md`
2. Verify blockers are resolved (check status above)
3. Read referenced source files to understand current state
4. Update phase status to `in_progress` in this file

### During Implementation
1. Follow the phase spec exactly - no scope creep
2. Keep changes minimal and focused
3. Run existing tests if any: `go test ./...`
4. Target 50-150 LOC per phase

### After Completing a Phase
1. Run `go build ./...` to verify compilation
2. Run `go test ./...` if tests exist
3. Update phase status to `done` in this file
4. Commit with message: `Phase {N}: {phase name}`
5. Report completion summary

### Commit Format
```
Phase {N}: {Short description}

- Change 1
- Change 2

Co-Authored-By: Claude <noreply@anthropic.com>
```

## File References

| File | Purpose |
|------|---------|
| `.work/DISCOVERY.md` | Product vision, user workflows |
| `.work/ROADMAP.md` | Version milestones, keybindings |
| `.work/phases/phase-{N}.md` | Individual phase specs |
| `.work/done/*.md` | Completed phase summaries |

## Progress Log

| Date | Phase | Notes |
|------|-------|-------|
| - | - | v0.4.0 planning complete |
| 2026-01-30 | 18 | Skipped - 14 tests already exist in git/ and tui/graph/ |
| 2026-01-30 | 19 | Added time_test.go (10 cases) + search_test.go (10 cases) instead of spec |
| 2026-01-30 | 20 | Rewrote reader_test.go with isolated fixtures (12 tests) |
| 2026-01-30 | 21 | Added benchmarks; zed repo (34k commits, 2k refs) takes 13min! Phase 22 required |
| 2026-01-30 | 22 | Fixed O(refs×commits) → O(commits); 556x speedup (778s → 1.4s) |
| 2026-01-30 | 22 | Added virtual scrolling (render visible rows only) + loading indicator |
| 2026-01-30 | 22 | Fixed: Now loads HEAD + local branches (was HEAD only, missed other branches) |
| 2026-01-30 | 22 | Refactored rendering: text/ package, RowLayout, Row struct, dynamic graph width |
| 2026-01-30 | 22a | Started: cursor bug when scrolling up at end of history |
| 2026-02-02 | 23 | Added tag visualization: <tagname> badges with yellow styling |

---

*To implement: Read phase file → Implement → Update status → Commit → Report*
