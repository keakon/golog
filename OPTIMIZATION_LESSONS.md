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

## Optimization E: `runtime.callers` linkname vs `runtime.Callers`

Replace the linkname'd `runtime.callers` (with `//go:noescape`) with the public
`runtime.Callers` API, exchanging "fragile internal symbol" for "supported boundary".

**Result: ~44% slower and 2 allocs / 248 B per call.**

| Variant | Time/op | Alloc/op | Allocs/op |
|---|---|---|---|
| `Caller` (linkname'd, `//go:noescape`) | 132.8 ns | 0 B | 0 |
| `runtime.Callers` (public API) | 237 ns | 248 B | 2 |

Measured on Apple M1 Pro, Go 1.26.

The win is not the validation overhead inside `runtime.Callers` (negligible). It is
that the public API cannot be annotated `//go:noescape` from outside the runtime
package, so the compiler must conservatively assume `pc []uintptr` escapes. That
forces the `[1]uintptr{}` array onto the heap and adds the two observed allocations.

`BenchmarkCallerRuntimeCallers` is kept as a regression guard. If a future Go
release narrows the gap (improved escape analysis, intrinsic recognition of fixed-
size slice arguments, or runtime API changes), the linkname can be retired with the
benchmark as evidence.

**Lesson:** `//go:linkname` is sometimes the only way to express a property the
public API does not surface (here: `//go:noescape`). Treat the linkname as a
load-bearing pillar, not a quick hack — document the alternative and the price,
then track the gap with a benchmark so the decision can be revisited.

---

## Optimization F: `atomic.Pointer[snapshot]` for FastTimer

Replace per-field reads of `FastTimer.date` / `.time` (plus a `bool isRunning`) with a
single `atomic.Pointer[fastTimerSnapshot]` storing date+time as an immutable pair.
Update path stores a fresh pointer once per second; read path is one atomic load.

**Result: neutral (within measurement noise), unlocks `-race` testing, and fixes a
documented torn-read at the day boundary.**

| Benchmark | Before | After | Delta |
|---|---|---|---|
| `DiscardLogger` | 214 ns/op | 212 ns/op | ≈ 0 |
| `DiscardLoggerParallel` | 35 ns/op | 32–40 ns/op | ≈ 0 |
| `MultiLevels` | 1488 ns/op | 1513 ns/op | +1.7% (noise band) |

On x86_64 the atomic load compiles to the same `MOV` as a plain load (aligned
acquire is implicit). On ARM64 it costs a few extra cycles for `LDAR`, fully
amortised by the data-dependent string copies that follow.

The orthogonal benefit: with the snapshot stored as a single value, readers cannot
observe a half-updated FastTimer at the day boundary. The README's behaviour #3
("date and time may belong to different days") goes away at zero cost.

**Lesson:** When a field group must update together, store it as a single
immutable value behind `atomic.Pointer`. You get torn-read freedom *and* race-test
coverage for the price of one extra acquire load — frequently zero on x86 and
sub-nanosecond on ARM. Avoid `sync.Mutex` / `sync.RWMutex` for read-mostly
hot-path data: the lock cost dwarfs the work.

---

## Summary

| Optimization | Single-thread | Parallel | Verdict |
|---|---|---|---|
| writeUintToBuf | neutral | neutral | Keep (correctness, no regression) |
| sync.Map cache | -8% | neutral | Reverted (cache slower than computation) |
| lruCache(Mutex) | -6% | -90% (~10x slower) | Reverted (parallel disaster) |
| r.args = nil | neutral | neutral | Keep (defensive GC correctness) |
| `runtime.Callers` (public) instead of linkname | -44% + 2 allocs | same | Reverted (linkname stays load-bearing) |
| `atomic.Pointer[snapshot]` for FastTimer | neutral | neutral | Keep (race-safe, fixes torn read) |

**Core principle: always measure before and after. Optimizations that "should" help
can make things worse, especially when adding concurrency primitives to hot paths.**
