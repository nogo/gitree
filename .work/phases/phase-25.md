# Phase 25: Release Check

## Goal
Allow users to manually check for new gitree releases.

## Context
- No auto-update or auto-check (user requirement)
- User-triggered only
- Check against GitHub releases API

## Scope
~80 LOC

## Tasks

### 1. Add version info

```go
// cmd/gitree/main.go or internal/version/version.go (NEW)
var (
    Version   = "dev"      // Set via ldflags at build
    GitCommit = "unknown"
)
```

Update build process:
```bash
go build -ldflags "-X main.Version=v0.4.0 -X main.GitCommit=$(git rev-parse --short HEAD)"
```

### 2. Implement release checker

```go
// internal/version/check.go (NEW)
type ReleaseInfo struct {
    Version     string
    URL         string
    PublishedAt time.Time
}

func CheckLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
    // GET https://api.github.com/repos/nogo/gitree/releases/latest
    // Parse response
    // Return latest version info
}

func IsNewer(current, latest string) bool {
    // Semantic version comparison
}
```

### 3. Add CLI flag

```bash
gitree --version          # Show current version
gitree --check-update     # Check for updates (NEW)
```

Output:
```
gitree v0.4.0 (abc1234)
Latest: v0.5.0 available at https://github.com/nogo/gitree/releases/tag/v0.5.0
```

Or if current:
```
gitree v0.4.0 (abc1234)
You're running the latest version.
```

### 4. Optional: In-app check (keybinding)

```go
// internal/tui/app.go
case "u": // 'u' for update check
    return m.checkForUpdates()
```

Display result in footer or modal:
```
● v0.5.0 available - https://github.com/...
```

## Acceptance Criteria
- [ ] `gitree --version` shows version and commit
- [ ] `gitree --check-update` checks GitHub API
- [ ] Graceful handling: no network, rate limited, repo not found
- [ ] No automatic checks ever

## Files to Read First
- `cmd/gitree/main.go` - entry point, flag parsing
- `.goreleaser.yaml` (if exists) - how versions are set

## Dependencies
None - independent feature

## API Details

GitHub releases endpoint:
```
GET https://api.github.com/repos/{owner}/{repo}/releases/latest
```

Response (relevant fields):
```json
{
  "tag_name": "v0.5.0",
  "html_url": "https://github.com/.../releases/tag/v0.5.0",
  "published_at": "2026-02-01T12:00:00Z"
}
```

## Notes
- Use `net/http` standard library
- Set User-Agent header (GitHub requires it)
- Timeout: 5 seconds max
- Cache result for session (don't re-check on every call)
- Handle pre-release versions appropriately
- Consider: use `go-github` library or keep it simple with raw HTTP?
