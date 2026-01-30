# Phase 18: Test Infrastructure

## Goal
Set up test framework and add foundational tests for domain logic.

## Context
- No tests currently exist in codebase
- Domain layer (`internal/domain/`) has pure types and logic
- Need CI-friendly test setup

## Scope
~100 LOC

## Tasks

### 1. Create test file structure
```
internal/
├── domain/
│   └── types_test.go      # NEW
└── git/
    └── reader_test.go     # NEW (placeholder)
```

### 2. Domain tests (`internal/domain/types_test.go`)

Test cases for any helper methods on domain types:
- `Commit` type validation
- `Branch` type validation
- Any filtering/sorting helpers if they exist

### 3. Add Makefile targets (if Makefile exists) or document commands

```makefile
test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
```

## Acceptance Criteria
- [ ] `go test ./...` runs without error
- [ ] At least 3 meaningful test cases exist
- [ ] Test files follow Go conventions (`*_test.go`)

## Files to Read First
- `internal/domain/types.go` - understand what to test
- `go.mod` - verify module name for imports

## Dependencies
None - this is foundational.

## Notes
- Use table-driven tests (Go idiom)
- No external test frameworks needed (standard `testing` package)
- Mock git repo not needed for domain tests
