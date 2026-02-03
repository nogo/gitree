# Implementation Orchestrator

> Load this file to implement phases. For milestone planning, start with `.work/DISCOVERY.md`.

## Project Context

<!-- Customize this section for your project -->

**Project:** gitree - TUI git history visualizer

| Aspect | Detail |
|--------|--------|
| Language | Go (bubbletea + go-git) |
| Architecture | `cmd/gitree/main.go` → `internal/{domain,git,tui,watcher}` |
| Build | `go build ./...` |
| Test | `go test ./...` |

---

## Execution Strategy

### Use Subagents to Save Context

**CRITICAL:** Each phase should be implemented using the Task tool with a subagent. This preserves main conversation context and allows parallel execution of independent phases.

```
Task tool with subagent_type="general-purpose":
- Prompt includes: phase spec content, relevant file paths, build/test commands
- Subagent reads files, implements, verifies, reports back
- Main agent updates ORCHESTRATOR.md status and commits
```

**When to use subagents:**
| Scenario | Approach |
|----------|----------|
| Single phase implementation | One subagent per phase |
| Independent phases (no blockers) | Launch multiple subagents in parallel |
| Exploration before coding | Use `subagent_type="Explore"` |
| Complex debugging | Subagent for focused investigation |

**Subagent prompt template:**
```
Implement Phase {N}: {Name}

## Phase Spec
{paste full phase spec content}

## Key Files to Read First
- {file1}
- {file2}

## After Implementation
1. Run: go build ./...
2. Run: go test ./...
3. Report: files changed, LOC added, any issues

## Constraints
- Follow SOLID principles (see below)
- Match existing code patterns
- No scope creep
```

---

## Go Pragmatic & SOLID Principles

### SOLID for Go

| Principle | Go Application |
|-----------|----------------|
| **S**ingle Responsibility | One struct/file = one job. `stats.go` computes stats, `render_stats.go` renders them. |
| **O**pen/Closed | Use interfaces for extension. Add new renderers without modifying existing ones. |
| **L**iskov Substitution | If it implements the interface, it works. No surprising behavior. |
| **I**nterface Segregation | Small interfaces. `type Renderer interface { View() string }` not mega-interfaces. |
| **D**ependency Inversion | Depend on interfaces, not concretions. Pass `domain.GitReader` not `*git.Reader`. |

### Go Pragmatic Rules

**Do:**
- Accept interfaces, return structs
- Make zero values useful
- Errors are values - handle them explicitly
- Group related code in same file (type + methods)
- Use table-driven tests
- Prefer composition over inheritance

**Don't:**
- Don't create interfaces until you need them (accept concrete types initially)
- Don't over-abstract for "future flexibility"
- Don't use getters/setters for public fields
- Don't panic in library code
- Don't use init() unless absolutely necessary

**Code Style:**
```go
// Good: Small interface, defined where used
type statsCalculator interface {
    Compute(commits []*domain.Commit) Stats
}

// Good: Constructor returns concrete type
func NewInsightsView() InsightsView { ... }

// Good: Method accepts interface
func (v *InsightsView) Recalculate(reader domain.GitReader) { ... }

// Bad: Premature interface
type InsightsViewInterface interface { ... } // Don't do this upfront
```

**Package Design:**
```
internal/tui/insights/
  insights.go      # Main type + New() + public methods
  stats.go         # Stats computation (separate concern)
  heatmap.go       # Heatmap computation (separate concern)
  render.go        # View() rendering (separate concern)
  styles.go        # Style constants (separate concern)
```

---

## Current Milestone

<!-- Updated by discovery process -->

**Version:** 0.5.0

**Theme:** Insights Mode - See patterns in your git history, not just commits

**Goals:**
- Add statistics dashboard showing top authors and most-changed files
- Add GitHub-style calendar heatmap for commit activity visualization
- Create toggleable Insights Mode that replaces graph view
- Integrate with existing filters (branch/author/tag/time all apply globally)
- Responsive layout: side-by-side on wide terminals, stacked on narrow

---

## Phase Queue

<!-- Updated by discovery process -->

| Phase | Name | Status | Blocks |
|-------|------|--------|--------|
| 1 | Stats Computation | done | - |
| 2 | Heatmap Data | done | - |
| 3 | Insights Model & Styles | done | 1, 2 |
| 4 | Stats Rendering | pending | 3 |
| 5 | Heatmap Rendering & Layout | pending | 3 |
| 6 | App Integration | pending | 4, 5 |

**Status values:** `pending` | `in_progress` | `done` | `skipped`

---

## How to Select Next Phase

```
1. Find ALL rows with status = `pending` and empty `Blocks` column
2. These phases can run IN PARALLEL - launch subagents for each
3. For phases with `Blocks`:
   - All blocking phases must be `done` before starting
4. If all phases `done` → milestone complete, run DISCOVERY.md
```

**Parallel execution example:**
```
Current queue:
| 1 | Stats Computation | pending | - |      ← Ready
| 2 | Heatmap Data      | pending | - |      ← Ready
| 3 | Model & Styles    | pending | 1, 2 |   ← Blocked

Action: Launch subagents for Phase 1 AND Phase 2 simultaneously
```

---

## Implementation Workflow

### Before Starting a Phase

```
1. Read phase spec: .work/phases/phase-{N}.md
2. Verify blockers resolved (check Phase Queue above)
3. Update phase status to `in_progress` (use Edit tool)
4. Launch subagent with phase spec (see Execution Strategy above)
```

### Subagent Implementation (runs inside subagent)

The subagent should:

1. **Read first, code second**
   - Read all files mentioned in phase spec
   - Grep for similar patterns in codebase
   - Understand existing conventions before writing

2. **Apply SOLID + Go Pragmatic**
   - Single responsibility per file
   - Small interfaces, concrete returns
   - Match existing code style exactly

3. **Implement incrementally**
   - Write one file at a time
   - Run `go build ./...` after each file
   - Fix errors before moving to next file

4. **Verify before reporting**
   - `go build ./...` passes
   - `go test ./...` passes
   - `go vet ./...` clean

**If stuck:**
- Re-read the phase spec
- Search for similar implementations: `Grep "pattern"`
- Check if blocking phase has relevant code
- Report back to main agent with specific question

### After Subagent Completes

Main agent should:

```
1. Review subagent report
2. Verify build/test if not done by subagent
3. Update status to `done` in this file
4. Commit with proper format
5. Log completion in Progress Log
6. Check for newly unblocked phases → launch next subagents
```

---

## Quality Checklist

Before marking phase `done`:

- [ ] Code compiles/builds successfully
- [ ] Tests pass
- [ ] No linter warnings
- [ ] Changes match phase spec exactly (no extras)
- [ ] Status updated in Phase Queue
- [ ] Commit created with proper format
- [ ] Progress Log entry added

---

## Error Recovery

### Build Fails
1. Read error message carefully
2. Fix syntax/type errors
3. Re-run build command

### Tests Fail
1. Read failing test output
2. Understand what's expected vs actual
3. Fix code or update test if spec changed
4. Re-run test command

### Scope Creep Detected
1. Stop implementation
2. Compare changes to phase spec
3. Remove anything not in spec
4. If spec is incomplete, consult user

### Blocked by Missing Code
1. Check if blocking phase exists
2. If yes, mark current phase blocked, move to next
3. If no, phase spec is incomplete - consult user

---

## Commit Format

```
Phase {N}: {Short description}

- Change 1
- Change 2

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Example:**
```
Phase 12: Add user authentication

- Add login handler with JWT tokens
- Add auth middleware for protected routes
- Add logout handler

Co-Authored-By: Claude <noreply@anthropic.com>
```

---

## Tool Usage

### Reading Code
| Tool | Use For |
|------|---------|
| `Read` | Phase specs, specific source files |
| `Glob` | Find files by pattern |
| `Grep` | Search patterns in code |

### Writing Code
| Tool | Use For |
|------|---------|
| `Edit` | Modify existing files (preferred) |
| `Write` | Create new files (only if needed) |

### Running Commands
| Command | Purpose |
|---------|---------|
| `go build ./...` | Verify compilation |
| `go test ./...` | Run tests |
| `go vet ./...` | Check for issues |

---

## Resuming Work

If continuing a previous session:

```
1. Read this file (ORCHESTRATOR.md)
2. Check Phase Queue for `in_progress` phase
3. If found: Read that phase spec, continue work
4. If not found: Select next `pending` phase
5. If all done: Milestone complete
```

---

## File References

| File | Purpose |
|------|---------|
| `.work/DISCOVERY.md` | Milestone planning (start here for new milestone) |
| `.work/ROADMAP.md` | Future ideas |
| `CHANGELOG.md` | What's been built |
| `.work/phases/phase-{N}.md` | Individual phase specs |
| `.work/done/*.md` | Completed phase summaries |

---

## Progress Log

<!-- Updated during implementation -->

| Date | Phase | Notes |
|------|-------|-------|
| 2026-02-03 | Planning | v0.5.0 Insights Mode - 6 phases defined |
| 2026-02-03 | Phase 1 | Stats Computation complete (149 LOC + 285 LOC tests) |
| 2026-02-03 | Phase 2 | Heatmap Data complete (203 LOC + 357 LOC tests) |
| 2026-02-03 | Phase 3 | Insights Model & Styles complete (104 LOC + 143 LOC tests) |

---

*Workflow: Select phase → Read spec → Implement → Verify → Update status → Commit*
