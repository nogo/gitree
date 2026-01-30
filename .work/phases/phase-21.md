# Phase 21: Performance Benchmarks

## Goal
Establish baseline performance metrics to identify if optimization is needed.

## Context
- Unknown performance ceiling currently
- Target: repos with 1k-100k commits
- Key operations: load commits, render graph, scroll viewport

## Scope
~100 LOC

## Tasks

### 1. Create benchmark file

```
internal/git/
├── reader.go
├── reader_test.go
└── reader_bench_test.go    # NEW
```

### 2. Benchmark commit loading

```go
func BenchmarkLoadCommits(b *testing.B) {
    // Setup: create repo with N commits or use real large repo
    sizes := []int{100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("commits_%d", size), func(b *testing.B) {
            // benchmark LoadCommits
        })
    }
}
```

### 3. Benchmark graph rendering

```
internal/tui/graph/
└── render_bench_test.go    # NEW
```

```go
func BenchmarkLaneAssignment(b *testing.B) {
    // Benchmark with varying commit counts
}

func BenchmarkRenderGraph(b *testing.B) {
    // Benchmark full graph render
}
```

### 4. Memory profiling (optional)

```go
func BenchmarkMemoryUsage(b *testing.B) {
    b.ReportAllocs()
    // ...
}
```

### 5. Document results

Create `.work/BENCHMARKS.md` with:
- Test environment (machine specs)
- Baseline numbers
- Performance targets
- Go/no-go decision for Phase 22

## Acceptance Criteria
- [ ] `go test -bench=. ./...` runs benchmarks
- [ ] Baseline numbers documented
- [ ] Clear decision: Phase 22 needed or skip

## Files to Read First
- `internal/git/reader.go` - commit loading
- `internal/tui/graph/lanes.go` - graph algorithm

## Dependencies
- Phase 18 (test infrastructure)

## Benchmark Targets (suggested)

| Operation | Target | Concern if |
|-----------|--------|------------|
| Load 10k commits | < 2s | > 5s |
| Render 10k graph | < 500ms | > 2s |
| Scroll (render visible) | < 16ms | > 50ms |

## Notes
- Use `testing.B` for benchmarks
- Run with `-benchmem` for allocation stats
- May need to generate large test repo programmatically
- Real-world test: clone linux kernel repo (1M+ commits)
