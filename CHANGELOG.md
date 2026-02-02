# Changelog

## [0.4.0] - 2026-02-02

### Added
- **Tag visualization** - Tags displayed as yellow `<tag>` badges on commits
- **Tag filter** - Filter by tag with `T` key, shows tag commits + ancestors
- **Help overlay** - Press `?` to see all keybindings
- **Version info** - `--version` flag shows version and git commit
- **Update checker** - `--check-update` flag checks GitHub for new releases
- **CLI filter flags** - `-b`/`--branch` and `-a`/`--author` for initial filters
- **Year in dates** - Date column now shows year for older commits

### Changed
- **556x faster loading** - Fixed O(refs×commits) bug, 34k commits now load in 1.5s instead of 13 minutes
- **Virtual scrolling** - Only visible rows rendered, navigation is instant regardless of repo size
- **Dynamic graph width** - Graph column adjusts based on visible commits
- **Refactored rendering** - New `text/` package, `RowLayout` struct, cleaner architecture

### Fixed
- Graph color leak between commits
- Incomplete branch rendering on complex histories
- Rendering artifacts when scrolling quickly
- Filter overlays now scroll for large lists

## [0.3.0] - 2026-01-30

### Added
- **Author filter** - Filter commits by author (`a` key)
- **Author highlight** - Dim non-matching commits to focus on one author (`A` key)
- **Search** - Find commits by message or hash (`/` key, `n`/`N` to navigate)
- **Date histogram** - Timeline showing commit density with time range filtering (`t` to toggle, `Tab` to focus)
- **Inline commit expansion** - View commit details and file list without leaving the graph (`Enter` to expand)
- **Diff view** - View file changes with syntax highlighting (`Enter` on file)
- Two-column layout for expanded view on wide terminals (>= 100 chars)

### Changed
- Commit details now shown inline (expanding between rows) instead of modal overlay
- Graph remains visible when viewing commit details
- Keyboard and mouse scrolling disabled when commit is expanded

## [0.2.0] - 2026-01-29

### Added
- Path argument support (`gitree /path/to/repo`)
- Column headers with proper alignment
- Multi-lane DAG visualization with lane assignment algorithm
- Topological sort for correct commit ordering

### Fixed
- UTF-8 rendering issues
- Lane assignment for rebased/cherry-picked commits
- Merge branch label positioning

## [0.1.0] - 2026-01-29

### Added
- Initial release
- Visual commit graph with branch badges and 6-color palette
- Scrollable commit list with virtual viewport
- Vim-style navigation (j/k, g/G) and mouse support
- Commit detail overlay
- Branch filter modal with reachability-based filtering
- Live repository watching with fsnotify
- 100ms debounce for rapid changes
- Visual indicator in footer (● watching / ○ not watching)
