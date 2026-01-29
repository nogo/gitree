# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Multi-lane DAG graph visualization with proper merge/fork connections
- Topological sort for correct commit ordering
- Lane assignment algorithm for branch visualization
- Column headers with proper alignment

### Fixed
- Column header alignment and increased column widths
- Lane assignment for rebased/cherry-picked commits
- Graph alignment and merge branch labels
- Viewport scrolling to keep cursor centered

## [0.1.0] - 2025-01-29

Initial MVP release.

### Added

#### Core
- Commit graph visualization with color-coded branches
- Branch visualization with inline badges (`[main]`, `[origin/main]`)
- 6-color palette rotation for branch differentiation

#### Navigation
- Vim-style navigation (`j`/`k`, `g`/`G`)
- Arrow keys and Page Up/Down support
- Mouse wheel scrolling and click-to-select
- Cursor indicator for selected commit

#### Views
- Commit list with virtual scrolling via viewport
- Commit detail overlay showing full hash, author, email, date, message, parents, and refs
- Branch filter modal with checkbox list

#### Live Updates
- File watcher on `.git/HEAD`, `refs/heads`, `refs/remotes`
- 100ms debouncing for rapid changes
- Graceful degradation if watcher fails
- Visual indicator in footer (● watching / ○ not watching)

#### Filtering
- Branch filter with BFS parent walk for commit reachability
- Filter persists after repository refresh
- Clear filter with `c` key

### Technical
- Built with Go, bubbletea, lipgloss, go-git, fsnotify
- Elm architecture for predictable state management
- Alt screen mode for clean terminal experience
