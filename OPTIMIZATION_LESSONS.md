# Optimization Lessons

This document records lessons learned from performance optimization attempts on golog,
backed by actual benchmark data on Apple M1 Pro (darwin/arm64).

## Baseline: commit 503e775 (no optimizations)

| Scenario | Single-thread | Parallel |
|---|---|---|
| DiscardLogger | ~214 ns/op, 0 allocs | ~35 ns/op, 0 allocs |
| MultiLevels | ~1488 ns/op, 0 allocs | ~190 ns/op, 0 allocs |

## Optimization A: `writeUintToBuf` (eliminate heap allocation for line numbers)

Replace `fastUint2DynamicBytes` (returns `[]byte`, may heap-allocate) with
`writeUintToBuf` (writes digits directly into `bytes.Buffer` using stack-allocated array).

**Result: neutral.** No measurable performance change in either single-thread or parallel.
The baseline was already 0 allocs/op — Go's runtime optimizes small slice allocations
onto the stack, so the theoretical heap allocation benefit didn't materialize in practice.

**Lesson:** Verify that the allocation you're eliminating actually causes heap allocation
and measurable overhead. A `make([]byte, 4)` for small line numbers is likely stack-allocated
by the compiler. Use `-gcflags="-m"` to confirm escape analysis before optimizing.

---

## Optimization B: `sync.Map` cache for short file names

Cache `computeShortFile` results in a `sync.Map` to avoid repeated path scanning
in `SourceFormatPart.Format`.

**Result: single-threaded regression ~8%, parallel neutral.**

The path scanning computation (iterate ~20-50 chars) is cheaper than `sync.Map`
internal overhead (atomic operations, dual-map structure). For short paths common
in logging, the cache is slower than the computation it's meant to skip.

**Lesson:** Caching a cheap computation behind an expensive cache mechanism makes
things slower. Always measure — the cache overhead must be meaningfully lower than
the computation it replaces. Path string scanning on typical log files is already fast.

---

## Optimization C: `lruCache(Mutex)` replacing `sync.Map`

Replace `sync.Map` with a bounded LRU cache using `sync.Mutex` to solve the
unbounded growth concern.

**Result: parallel catastrophic regression ~10x.**

`sync.Mutex` serializes all goroutines on every cache access (both reads and writes).
`sync.Map` has lock-free reads — once keys stabilize, concurrent reads are nearly free.
In logging, the set of file paths is small and stable, so nearly all accesses are reads.

**Lesson:** `sync.Mutex`-based caches are catastrophic for read-heavy concurrent access.
`sync.Map` is designed exactly for this pattern (read-mostly, stable key set). If you
need bounded growth, use a concurrent-friendly design:
- `sync.RWMutex` (readers don't block each other)
- Sharded LRU (one mutex per shard, reducing contention)
- `sync.Map` with periodic eviction

Never sacrifice parallel performance for a correctness concern that can be solved
with a concurrent-friendly bounded cache.

---

## Optimization D: `r.args = nil` before pool return

Set `Record.args = nil` before returning the record to `sync.Pool`, preventing
pooled objects from retaining references to external slices.

**Result: neutral in benchmarks (0 allocs).** Existing benchmarks use `Infof("test")`
with no format args, so `args` is always nil/empty — the optimization has nothing to
reduce. Benchmarks with format args (`Infof("count=%d name=%s", i, "test")`) show
2 allocs / 40 B per call, but these are the args slice allocation itself, not
GC pressure from retained references.

**Lesson:** Defensive correctness improvements (preventing pool objects from holding
external references) may not show benchmark improvement but are still worth keeping —
they prevent subtle GC pressure issues in real usage with varied format arguments.
Add targeted benchmarks to at least verify no regression.

---

## Summary

| Optimization | Single-thread | Parallel | Verdict |
|---|---|---|---|
| writeUintToBuf | neutral | neutral | Keep (correctness, no regression) |
| sync.Map cache | -8% | neutral | Reverted (cache slower than computation) |
| lruCache(Mutex) | -6% | **-995%** | Reverted (parallel disaster) |
| r.args = nil | neutral | neutral | Keep (defensive GC correctness) |

**Core principle: always measure before and after. Optimizations that "should" help
can make things worse, especially when adding concurrency primitives to hot paths.**