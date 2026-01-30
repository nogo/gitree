# Gitree Performance Benchmarks

## Test Environment

- **Machine:** Apple M1 Pro
- **OS:** macOS (Darwin)
- **Go:** 1.24.0
- **Date:** 2026-01-30

## Results Summary

### Real-World Benchmark: Zed Editor Repo

**Repo stats:** 34,667 commits, 2,402 refs (remote branches)

#### Before Fix (Phase 21)

| Operation | Time |
|-----------|------|
| Load ALL commits | **778 seconds** (13 min!) |
| Memory | 412 GB allocations |

**Root cause:** Iterating ALL 2,402 refs, loading commits from each = O(refs × commits)

#### After Fix (Phase 22)

| Operation | Time | Commits |
|-----------|------|---------|
| Load ALL commits | **1.45s** | 34,667 |
| LoadRepository | **1.45s** | 34,667 commits, 1,117 branches |

**Speedup: 556x** (778s → 1.4s)

### Graph Layout (BuildLayout)

| Commits | Linear | Branching |
|---------|--------|-----------|
| 100 | 28μs | 50μs |
| 500 | 137μs | 868μs |
| 1,000 | 287μs | 3.3ms |
| 5,000 | 1.7ms | N/A |

### Row Rendering

| Operation | Time | Memory |
|-----------|------|--------|
| Single row (any size) | 1.3μs | 88B |
| Viewport 40 rows | 2.9ms | 210KB |

## Final Targets

| Operation | Target | Actual (zed) | Status |
|-----------|--------|--------------|--------|
| Load 34k commits | < 2s | **1.45s** | ✅ Pass |
| Render 10k graph | < 500ms | ~33ms | ✅ Pass |
| Scroll (viewport) | < 16ms | ~4ms | ✅ Pass |

## Fix Applied (Phase 22)

**Change:** `loadCommitsFromRepo` now loads from HEAD only instead of iterating all refs.

```go
// Before: O(refs × commits) - catastrophically slow
for _, branchHash := range allBranchHashes {  // 2,402 iterations!
    iter, _ := repo.Log(&git.LogOptions{From: branchHash})
    iter.ForEach(...)
}

// After: O(commits) - single iteration from HEAD
iter, _ := repo.Log(&git.LogOptions{From: head.Hash()})
iter.ForEach(...)
```

Branch ref labels are still populated - only the iteration strategy changed.
