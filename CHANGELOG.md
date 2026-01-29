# Changelog

## 2026-01-29

### Foundation
- Set up Go project structure with domain types (Commit, Branch, Repository)
- Implemented git layer using go-git for loading commits and branches
- Created TUI shell with bubbletea and Elm architecture

### Commit List
- Added scrollable commit list with virtual viewport
- Implemented graph rendering with branch badges and 6-color palette
- Added Vim-style navigation (j/k, g/G) and mouse support

### Views & Overlays
- Added commit detail overlay showing full hash, author, date, message, parents, refs
- Added branch filter modal with checkbox list and reachability-based filtering

### Live Updates
- Implemented file watcher on `.git/HEAD` and refs directories
- Added 100ms debounce for rapid changes
- Visual indicator in footer (● watching / ○ not watching)

### Path Argument & UI Polish
- Added support for `gitree [path]` to open any repository
- Added column headers with proper alignment
- Fixed UTF-8 rendering issues

### Graph Improvements
- Replaced simple graph with multi-lane DAG visualization
- Added lane assignment algorithm for proper branch layout
- Implemented topological sort for correct commit ordering
- Fixed lane assignment for rebased/cherry-picked commits
- Improved merge branch label positioning

### Documentation
- Added README with installation, usage, and key bindings
- Added this changelog
