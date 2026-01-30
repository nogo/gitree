# Gitree - Discovery Document

> A terminal-based git history visualization tool with live updates and time-based navigation.

---

## Problem Statement

Existing git visualization tools (tig, lazygit, git log --graph) lack:
- Intuitive time-based navigation (jump to date ranges, see commit density)
- Modern visual design
- Live updates when repository changes
- Good filtering UX while maintaining graph context

**Goal:** Build a TUI git graph tool focused on visual clarity, time-based exploration, and real-time awareness of repository changes.

---

## Target Users

- Developers who need to understand repository history visually
- Teams doing code review, rebasing, or investigating when changes occurred
- Users who work with medium-to-large repositories (1k - 100k commits)

---

## Core User Workflows

### 1. Explore History
> "I want to see the full picture of how this repository evolved"

- Open tool in repo root
- See graph of all branches and commits
- Scroll through history
- Understand branch relationships at a glance

### 2. Time-Based Navigation
> "What happened last week?" or "Show me commits from Q3"

- See a date histogram (commit density over time)
- Select a time range to filter
- Graph filters to show only commits in selected range

### 3. Investigate a Commit
> "What exactly changed in this commit?"

- Navigate to commit in graph
- Press Enter to expand inline (keeps graph visible)
- See list of changed files with additions/deletions
- View diff of changes

### 4. Filter by Context
> "Show me only what Alice did on the feature branch"

- Filter by branch (show/hide branches)
- Filter by author
- Filters apply while maintaining graph structure
- Clear indication of active filters

### 5. Highlight Contributor
> "Show me everything Alice contributed at a glance"

- Select an author to highlight
- Their commits glow/stand out in the graph (others dim)
- Quickly see contribution patterns
- Toggle between "highlight" (dim others) and "filter" (hide others)

### 6. Monitor Changes
> "I want to know when the repo updates"

- Tool silently updates graph when repo changes
- New commits appear in graph automatically
- No disruptive notifications, just visual update

---

## Current Features (v0.3.0)

| Feature | Status | Keybinding |
|---------|--------|------------|
| Commit graph visualization | Done | - |
| Branch visualization | Done | - |
| Scroll/navigate history | Done | j/k, arrows |
| Live repository watching | Done | - |
| Branch filter | Done | `b` |
| Author filter | Done | `a` |
| Author highlight | Done | `A` |
| Commit search | Done | `/`, `n`/`N` |
| Date histogram | Done | `t`, `Tab` |
| Inline commit expansion | Done | `Enter` |
| File change list | Done | - |
| Diff view | Done | `Enter` on file |

---

## Future Ideas

| Feature | Priority | Notes |
|---------|----------|-------|
| Workspace discovery | Medium | Run from any path, discover git repos in subdirectories |
| Calendar heatmap | Low | GitHub-style activity view |
| Custom color themes | Low | Match terminal theme |
| Stash visualization | Low | Show stashes in graph |

**Descoped** (out of visualization focus):
- Checkout commit
- Interactive rebase prep
- Open in editor

---

## UI Design

### Design Principles

1. **Clean over cluttered** - NOT like tig. Think htop, lazygit, k9s
2. **Information density** - Show useful data, minimize chrome
3. **Subtle structure** - Alignment over borders, spacing over separators
4. **Scannable** - Clear visual hierarchy, eyes flow naturally
5. **Breathing room** - Adequate spacing prevents visual fatigue

### Layout

```
gitree                                                      my-project
─────────────────────────────────────────────────────────────────────────
     Message                                      Author       Date   Hash
─────────────────────────────────────────────────────────────────────────
  ○  [main] feat: add OAuth provider              Alice    29 Jan 15:23  a1b2c
  ●  fix: token refresh on expiry                 Bob      29 Jan 14:13  b2c3d
  ●  wip: testing OAuth flow                      Carol    29 Jan 12:15  c3d4e
● │  changes for gitea certificate                Carol     4 Dec 12:46  i9j0k
●─╯  merge: feature branch                        Alice     4 Nov 14:30  m3n4o
─────────────────────────────────────────────────────────────────────────
● watching   47/1284 commits   12 branches     [b]ranch [c]lear [?] [q]
```

### Inline Expansion (Commit Details)

```
  ○  [main] feat: add OAuth provider              Alice    29 Jan 15:23  a1b2c
  │  ╔════════════════════════════╤════════════════════════════════════╗
  │  ║ Commit: a1b2c3d...         │ Files (5)  +42 -18                  ║
  │  ║ Author: Alice <alice@...>  │ > M src/auth.go           +10 -5   ║
  │  ║ Date:   Jan 29, 2026 15:23 │   A src/oauth.go          +32 -0   ║
  │  ║ Parent: b2c3d4e            │   M internal/config.go    +0 -13   ║
  │  ╚══════ [j/k] file  [Enter] diff  [Esc] close ════════════════════╝
  ●  fix: token refresh on expiry                 Bob      29 Jan 14:13  b2c3d
```

---

## Technical Decisions

### Language: Go

**Rationale:**
- Fast iteration for prototyping
- bubbletea is capable TUI framework
- go-git is pure Go, no C dependencies
- Good performance for target scale

### Libraries

| Purpose | Library |
|---------|---------|
| TUI Framework | bubbletea |
| Styling | lipgloss |
| Git operations | go-git |
| File watching | fsnotify |

### Architecture

```
cmd/gitree/main.go    Entry point, watcher setup
internal/
├── domain/           Core types (Commit, Branch, Repository)
├── git/              go-git reader implementation
├── watcher/          fsnotify file watcher
└── tui/
    ├── app.go        Main bubbletea model
    ├── layout.go     Header/content/footer rendering
    ├── styles.go     Centralized lipgloss styles
    ├── messages.go   Custom tea.Msg types
    ├── list/         Commit list + inline expansion
    ├── diff/         Diff view overlay
    ├── graph/        Graph rendering (lanes, DAG, colors)
    ├── filter/       Branch/author filter modals
    ├── search/       Commit search
    └── histogram/    Date timeline
```

---

## Key Design Decisions

### 1. Graph algorithm
Single-column with lane assignment for parallel branches. Color rotation per lane (6-color palette).

### 2. Date format
Relative for recent, absolute for older:
- < 1 hour: `Xm ago`
- < 24 hours: `Xh ago`
- < 7 days: `X days ago`
- >= 7 days: `D Mon`

### 3. Inline expansion vs modal
Commit details shown inline (expanding between rows) rather than as modal overlay. Keeps graph visible for context.

### 4. Date histogram
Video player timeline metaphor with braille density characters. Selection brackets for time range filtering.

---

## Distribution

### Release Process

Releases automated via GitHub Actions. Push tag to trigger:

```bash
git tag v0.3.0
git push origin v0.3.0
```

### Build Matrix

| OS | Architectures |
|----|---------------|
| macOS | amd64, arm64 |
| Linux | amd64, arm64 |
| Windows | amd64, arm64 |

---

## Documentation

See `.work/ROADMAP.md` for implementation details and `.work/done/` for phase summaries.

---

*Document created: 2025-01-15*
*Last updated: 2026-01-30*
*Status: v0.3.0 complete*
