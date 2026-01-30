# Gitree Implementation Roadmap

## Goal
TUI git history visualization with graph, live updates, and time-based navigation.

## Tech Stack
| Purpose | Library |
|---------|---------|
| TUI | bubbletea |
| Styling | lipgloss |
| Git | go-git |
| File watch | fsnotify |

## Architecture
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

## Completed Versions

### v0.1.0 - MVP
- Commit graph visualization
- Branch visualization with color-coded badges
- Scroll/navigate history (vim + arrows + mouse)
- Commit detail view
- Live repository watching
- Basic branch filter

### v0.2.0 - Usability & Polish
- Path argument support (`gitree /path/to/repo`)
- Column headers and alignment
- Proper DAG graph with lane management

### v0.3.0 - Filtering & Exploration
- Author filter (`a`)
- Author highlight (`A`)
- Commit search (`/`, `n`/`N`)
- Date histogram/timeline (`t`, `Tab`)
- Inline commit expansion with file list (`Enter`)
- Diff view (`Enter` on file)

## Current Version

### v0.4.0 - Stabilization & Polish (In Progress)

| Phase | Feature | Status |
|-------|---------|--------|
| 18 | Test infrastructure | pending |
| 19 | Graph algorithm tests | pending |
| 20 | Git layer tests | pending |
| 21 | Performance benchmarks | pending |
| 22 | Performance fixes | pending (conditional) |
| 23 | Tag visualization | pending |
| 24 | Tag filter | pending |
| 25 | Release check | pending |

**Goals:**
- Test coverage for core logic
- Performance validation for large repos
- Tag visualization (`T` for filter)
- Manual release check (`--check-update`)

See `.work/phases/phase-{18-25}.md` for implementation details.

## Keybindings

| Key | Action |
|-----|--------|
| `j`/`k`, `↑`/`↓` | Navigate commits |
| `Ctrl+d`/`Ctrl+u` | Page down/up |
| `g`/`G` | Jump to top/bottom |
| `Enter` | Expand commit (show files) |
| `b` | Branch filter |
| `a` | Author filter |
| `A` | Author highlight |
| `/` | Search commits |
| `n`/`N` | Next/prev search match |
| `t` | Toggle histogram |
| `Tab` | Focus histogram |
| `c` | Clear all filters |
| `T` | Tag filter (v0.4.0) |
| `u` | Check for updates (v0.4.0) |
| `Esc` | Close/collapse |
| `q` | Quit |

### When Expanded
| Key | Action |
|-----|--------|
| `j`/`k` | Navigate files |
| `Enter` | Open diff |
| `Esc` | Collapse |

### In Diff View
| Key | Action |
|-----|--------|
| `j`/`k` | Scroll diff |
| `h`/`l` | Prev/next file |
| `Esc`/`q` | Close diff |

## Future Ideas (Post v0.4.0)

**After user testing:**

### Multi-Repo & Workflow
| Feature | Description |
|---------|-------------|
| Workspace discovery | Scan directory for repos, repo switcher UI |
| Git commands | Execute fetch, pull, etc. from within gitree |
| Rebasing visualization | Show rebase state, todo, conflicts |
| Reflog view | Recovery scenarios, undo history |

### Visualization & Insights
| Feature | Description |
|---------|-------------|
| Commit comparison | Select two commits, see diff between them |
| Branch comparison | Visual diff between branches |
| Statistics dashboard | Author contributions, file churn, frequency |
| Blame integration | Jump from diff to blame view |
| Calendar heatmap | GitHub-style activity visualization |

### Performance & Compatibility
| Feature | Description |
|---------|-------------|
| Large repo optimizations | Lazy loading, commit limiting, caching |
| Shallow clone support | Handle repos cloned with --depth |
| Worktree support | Handle multiple worktrees |
| Submodule awareness | Show submodule status in commits |

### Polish
| Feature | Description |
|---------|-------------|
| Stash visualization | Show stashes in graph |
| Custom color themes | Match terminal theme |
| Export as image | SVG/PNG for docs and PRs |

## Documentation

| File | Purpose |
|------|---------|
| `.work/ORCHESTRATOR.md` | Load-first coordination for AI implementation |
| `.work/phases/phase-{N}.md` | Individual phase specifications |
| `.work/done/*.md` | Completed phase summaries |

**Completed summaries:**
- `phases-0-8-summary.md` - MVP (v0.1.0)
- `phases-9-11-summary.md` - Usability & Polish (v0.2.0)
- `phases-12-17-summary.md` - Filtering & Exploration (v0.3.0)
