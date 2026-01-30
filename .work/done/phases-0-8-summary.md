# gitree Implementation Summary (Phases 0-8)

## Project
TUI git history visualizer using Go, bubbletea, lipgloss, and go-git.

## Architecture
```
cmd/gitree/main.go          → Entry point, watcher setup
internal/
├── domain/                  → Types (Commit, Branch, Repository)
├── git/                     → go-git reader implementation
├── watcher/                 → fsnotify file watcher
└── tui/
    ├── app.go              → Main bubbletea model
    ├── layout.go           → Header/content/footer rendering
    ├── styles.go           → lipgloss styles
    ├── messages.go         → Custom tea.Msg types
    ├── list/               → Commit list component (viewport)
    ├── detail/             → Commit detail overlay
    ├── graph/              → Graph rendering (nodes, colors, badges)
    └── filter/             → Branch filter modal
```

## Phases Completed

### Phase 0: Setup
- Go module `github.com/nogo/gitree`
- Domain types: `Commit`, `Branch`, `Repository`
- Interfaces: `GitReader`, `RepositoryWatcher`

### Phase 1: Git Layer
- `git.Reader` using go-git
- `LoadRepository()`, `LoadCommits()`, `LoadBranches()`
- Populates: Hash, ShortHash, Author, Email, Date, Message, FullMessage, Parents, BranchRefs

### Phase 2: TUI Shell
- bubbletea app with Elm architecture
- Header/content/footer layout
- Alt screen, quit handling (q/Ctrl+C)
- Window resize support

### Phase 3: Commit List
- `bubbles/viewport` for virtual scrolling
- Row format: graph | hash | author | date | message
- j/k navigation, g/G for top/bottom
- Cursor highlighting

### Phase 4: Graph Rendering
- Single-column graph (●/○ nodes, │ lines, ╯ merges)
- 6-color palette rotation for branches
- Inline branch badges `[main]` `[origin/main]`
- Badge styles: local (colored), remote (green bg), HEAD (red bg)

### Phase 5: Navigation
- Page movement: Ctrl+d/u, PgUp/PgDown
- Mouse: wheel scroll, click to select
- `>` cursor indicator
- `tea.WithMouseCellMotion()`

### Phase 6: Detail View
- Centered overlay panel on Enter
- Shows: full hash, author, email, date, full message, parents, refs
- Close with Esc or q

### Phase 7: File Watcher
- fsnotify watching .git/HEAD, refs/heads, refs/remotes
- 100ms debouncing
- Async reload via `RepoChangedMsg` → `RepoLoadedMsg`
- Footer shows ● watching / ○ not watching
- Graceful degradation if watcher fails

### Phase 8: Branch Filter
- Modal overlay (b to open)
- Checkbox list with j/k, space toggle, a=all, n=none
- BFS parent walk for commit reachability
- Filter persists after repo refresh
- Footer shows `branch:X/Y` when active
- c to clear filter

## Key Bindings
| Key | Action |
|-----|--------|
| j/k, ↑/↓ | Navigate |
| Ctrl+d/u, PgDn/PgUp | Page |
| g/G, Home/End | Top/bottom |
| Enter | Open detail |
| Esc | Close overlay |
| b | Branch filter |
| c | Clear filter |
| q, Ctrl+c | Quit |
| Mouse wheel | Scroll |
| Mouse click | Select |

## Dependencies
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/bubbles/viewport`
- `github.com/charmbracelet/lipgloss`
- `github.com/go-git/go-git/v5`
- `github.com/fsnotify/fsnotify`

## Commits
```
6144c5b Phase 8: Branch filter
573643b Phase 7: File watcher
741c508 Phase 6: Detail view
210f66f Phase 5: Navigation
f69eeb9 Phase 4: Graph rendering
6c35b09 Phase 3: Commit list
f92fae4 Phase 2: TUI shell
f101477 Phase 0+1: Setup + git layer
```
