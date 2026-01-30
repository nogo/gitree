# gitree

A terminal-based git history visualization tool with live updates and time-based navigation.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)

## Features

- **Visual commit graph** - Multi-lane DAG visualization showing branch relationships
- **Live updates** - Graph refreshes automatically when repository changes
- **Inline commit details** - Expand commits to see files and diffs without leaving the graph
- **Filtering** - Filter by branch or author
- **Author highlight** - Dim other commits to focus on one contributor
- **Search** - Find commits by message or hash
- **Date histogram** - Timeline showing commit density, filter by time range
- **Diff view** - View file changes with syntax highlighting
- **Vim-style navigation** - Keyboard-driven with mouse support

## Installation

### Download Binary

Download the latest release from the [releases page](https://github.com/nogo/gitree/releases).

```bash
# Example for macOS Apple Silicon
tar -xzf gitree_0.3.0_darwin_arm64.tar.gz
sudo mv gitree /usr/local/bin/
```

**macOS users:** If you see "cannot be opened because it is from an unidentified developer", run:

```bash
xattr -d com.apple.quarantine /usr/local/bin/gitree
```

### Go Install

```bash
go install github.com/nogo/gitree/cmd/gitree@latest
```

### Build from Source

```bash
git clone https://github.com/nogo/gitree.git
cd gitree
go build -o gitree ./cmd/gitree
```

## Usage

```bash
# Open current directory
gitree

# Open specific repository
gitree /path/to/repo
gitree ~/projects/myrepo
```

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k`, `↑` / `↓` | Navigate up/down |
| `Ctrl+d` / `Ctrl+u` | Page down/up |
| `g` / `G` | Jump to top/bottom |
| `Enter` | Expand commit (show files) |
| `Esc` | Close/collapse |
| `q` | Quit |

### Filtering & Search

| Key | Action |
|-----|--------|
| `b` | Branch filter |
| `a` | Author filter |
| `A` | Author highlight (dims others) |
| `/` | Search commits |
| `n` / `N` | Next/previous match |
| `c` | Clear all filters |

### Timeline

| Key | Action |
|-----|--------|
| `t` | Toggle histogram |
| `Tab` | Focus histogram |
| `h` / `l` | Move selection |
| `[` / `]` | Set range start/end |
| `Enter` | Apply time filter |

### When Expanded

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate files |
| `Enter` | Open diff view |
| `Esc` | Collapse |

### Diff View

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll diff |
| `h` / `l` | Previous/next file |
| `Esc` / `q` | Close |

## Layout

```
gitree                                                    repo-name
───────────────────────────────────────────────────────────────────
     Message                              Author       Date   Hash
───────────────────────────────────────────────────────────────────
│ ○  [main] Add feature...               Alice     12m ago  6144c
│    ╔═══════════════════════╤═══════════════════════════════════╗
│    ║ Commit: 6144c3d...    │ Files (3)  +42 -18                ║
│    ║ Author: Alice <a@...> │ > M src/app.go          +30 -10  ║
│    ║ Date:   Jan 30, 2026  │   A src/new.go          +12 -0   ║
│    ╚═══ [j/k] file  [Enter] diff  [Esc] close ═════════════════╝
●─┘  Merge branch 'feature'              Bob        1h ago  a1b2c
│    Fix authentication bug              Carol      2h ago  b2c3d
───────────────────────────────────────────────────────────────────
                    ⣀⣀⣤⣤⣶⣶⣿⣿⣶⣶⣤⣤⣀⣀
                   [━━━━━━━━━━]
───────────────────────────────────────────────────────────────────
● watching   47/1284 commits   12 branches       [b]ranch [c]lear
```

## Requirements

- Go 1.24 or later
- A terminal with 256-color support

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [go-git](https://github.com/go-git/go-git) - Git operations
- [fsnotify](https://github.com/fsnotify/fsnotify) - File watching

## License

MIT
