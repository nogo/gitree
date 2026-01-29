# gitree

A terminal-based git history visualization tool with live updates.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)

## Features

- **Visual commit graph** - See branch relationships at a glance with multi-lane DAG visualization
- **Live updates** - Graph refreshes automatically when repository changes
- **Branch filtering** - Filter commits by branch with reachability-based filtering
- **Commit details** - View full commit info in an overlay panel
- **Vim-style navigation** - Keyboard-driven with mouse support
- **Clean design** - Minimal chrome, maximum information density

## Installation

### Download Binary

Download the latest release from the [releases page](https://github.com/nogo/gitree/releases).

```bash
# Example for macOS Apple Silicon
tar -xzf gitree_0.2.0_darwin_arm64.tar.gz
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

| Key | Action |
|-----|--------|
| `j` / `k`, `↑` / `↓` | Navigate up/down |
| `Ctrl+d` / `Ctrl+u` | Page down/up |
| `PgDn` / `PgUp` | Page down/up |
| `g` / `G` | Jump to top/bottom |
| `Home` / `End` | Jump to top/bottom |
| `Enter` | Open commit detail |
| `Esc` | Close overlay |
| `b` | Open branch filter |
| `c` | Clear filter |
| `q`, `Ctrl+c` | Quit |
| Mouse wheel | Scroll |
| Mouse click | Select commit |

## Layout

```
gitree                                                    repo-name
───────────────────────────────────────────────────────────────────
     Message                              Author       Date   Hash
───────────────────────────────────────────────────────────────────
│ ○  [main] Add feature...               Danilo K  12m ago  6144c
●─┘  Merge branch 'feature'              Alice     1h ago   a1b2c
│    Fix authentication bug              Bob       2h ago   b2c3d
───────────────────────────────────────────────────────────────────
● watching   47/1284 commits   12 branches          [b]ranch [q]uit
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
