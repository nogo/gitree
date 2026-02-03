# Discovery Session

> Load this file to start planning the next milestone.

## Instructions for AI

You are helping plan the next development milestone. Your role:

1. **Explore** - Read codebase to understand current state
2. **Discuss** - Present options, challenge assumptions
3. **Decide** - Help prioritize and select features
4. **Generate** - Create phase specs and update ORCHESTRATOR.md

---

## Discovery Process

### STEP 1: Context Review

**Do this immediately after loading this file:**

```
1. Read `.work/ROADMAP.md` (future ideas)
2. Read `CHANGELOG.md` (what's built)
3. Explore codebase structure (Glob for source files)
4. Read key files to understand patterns (entry point, core types, main logic)
```

**Then summarize to user:**
> "I've reviewed the codebase. Current structure: [X packages/modules], [Y files]. Latest version added [Z]. The roadmap has [N] ideas across [M] categories. Ready to discuss?"

### STEP 2: Discussion Loop

Present roadmap options and drive decisions:

1. **What problem are we solving?** - Which user pain point matters most?
2. **What's the smallest useful increment?** - MVP for this milestone?
3. **What can we cut?** - Nice-to-have vs essential?
4. **What are the risks?** - Technical debt, breaking changes?
5. **What's the theme?** - Single sentence for this milestone?

**Challenge vague ideas.** If user is unclear, ask:
- What specifically? (which feature, which behavior?)
- Where does it go? (which component, which view?)
- What's the interaction model? (keys, clicks, commands?)

### STEP 3: Milestone Definition

Agree on:
- **Version number** (semver based on scope)
- **Theme** (one sentence)
- **Goals** (3-5 bullet points)
- **Features** (3-8 items)
- **Success criteria** (how to know it's done)

### STEP 4: Phase Generation

Break milestone into phases:

1. **Explore dependencies** - Use `Grep` to find where new code connects
2. **Order by dependency** - What must exist before other things work?
3. **Size each phase** - Target 50-150 LOC, single responsibility
4. **Write phase specs** - Use template below, write to `.work/phases/`

**Determine phase numbers:**
```
1. Read existing files: Glob .work/phases/phase-*.md
2. Find highest number
3. New phases start at highest + 1
```

### STEP 5: Orchestrator Update

Update `.work/ORCHESTRATOR.md`:
- Set `Current Milestone` section (version, theme, goals)
- Populate `Phase Queue` table
- Add planning date to `Progress Log`

---

## Project Identity

<!-- Customize this section for your project -->

**Project:** gitree - TUI git history visualizer

| Aspect | Value |
|--------|-------|
| Language | Go (bubbletea + go-git) |
| Focus | Visualization (not git actions) |
| Differentiators | Time-based navigation, live watching, visual clarity |

### Design Principles

1. **Clean over cluttered** - NOT like tig. Think htop, lazygit, k9s
2. **Information density** - Show useful data, minimize chrome
3. **Visualization focus** - See history, don't manipulate it
4. **Keyboard-first** - Vim-style navigation, discoverable shortcuts

### What this project does NOT do (by design)

- Checkout commits/branches
- Interactive rebase
- Staging/committing
- Push/pull (visualization only)

---

## Tool Usage

### Reading Context
| Tool | Use For |
|------|---------|
| `Read` | Specific files |
| `Glob` | Find files by pattern |
| `Grep` | Search for patterns in code |

### Writing Output
| Tool | Use For |
|------|---------|
| `Write` | New phase spec files |
| `Edit` | Update ORCHESTRATOR.md sections |

### Exploring Code
```
# Find source files (adjust pattern for your language)
Glob src/**/*.ts
Glob internal/**/*.go
Glob **/*.py

# Find where a type/function is used
Grep "TypeName" --type {lang}

# Find existing patterns to follow
Grep "function.*Pattern" --type {lang}
```

---

## Output Artifacts

### Phase Spec Template

Write to `.work/phases/phase-{N}.md`:

```markdown
# Phase {N}: {Name}

## Goal
{One sentence describing the outcome}

## Changes
- `{file}`: {what changes}
- `{file}`: {what changes}

## Implementation
1. {Step with specific detail}
2. {Step with specific detail}

## Testing
- {How to verify it works}

## Estimated LOC
{50-150}
```

### Example Phase Spec

```markdown
# Phase 12: User Authentication

## Goal
Add login/logout functionality with session management.

## Changes
- `src/auth/login.ts`: Add login handler with credential validation
- `src/auth/session.ts`: Add session creation and token generation
- `src/middleware/auth.ts`: Add authentication middleware
- `src/routes/index.ts`: Register auth routes

## Implementation
1. Create login handler that validates credentials against database
2. Generate JWT token on successful login
3. Create middleware that validates token on protected routes
4. Add logout handler that invalidates session
5. Register routes: POST /login, POST /logout

## Testing
- Login with valid credentials → returns token
- Login with invalid credentials → returns 401
- Access protected route with token → succeeds
- Access protected route without token → returns 401

## Estimated LOC
~120
```

---

## Reference Files

| File | Purpose |
|------|---------|
| `CHANGELOG.md` | What's been built |
| `.work/ROADMAP.md` | Future ideas to discuss |
| `.work/ORCHESTRATOR.md` | Implementation coordination |
| `.work/done/*.md` | Completed phase summaries |

---

## Session Start

After completing Step 1 (context review), begin with:

> "I've reviewed the codebase and documentation. [Summarize current state]. The roadmap includes [categories]. What direction interests you for the next milestone?"

Then follow steps 2-5.
