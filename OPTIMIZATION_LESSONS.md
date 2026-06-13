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

## Optimization G: on-demand `Caller()` (skip the stack walk when unused)

`Caller()` resolves the file and line via a runtime stack walk. Optimization E
established it is already near-optimal (the linkname + `//go:noescape` makes it
zero-alloc) and that the cost is irreducible *if you call it*. So instead of
making `Caller()` faster, avoid calling it when the result is discarded.

A formatter only needs the source location when its format contains `%s`/`%S`.
`Formatter.needsCaller` is set at parse time; `Logger.needsCaller` is the OR over
its handlers' formatters; each level method calls `Caller(1)` only when the flag
is set. The `log` package gains `*NoCaller` dispatch variants selected by
`SetDefaultLogger`.

**Result: ~70% faster on source-less formats; default (`%s`) path unchanged.**
Measured on linux/amd64 (Intel Xeon Platinum 8559C, Go 1.24.4), `-count=8`:

| Benchmark | Time/op | Notes |
|---|---|---|
| `DiscardLogger` (`%s`, renders source) | 200.8 ns | unchanged vs master (201.5 ns) |
| `DiscardLoggerNoSource` (no `%s`/`%S`) | 59.09 ns | new fast path |
| `Caller` (microbench) | 113.2 ns | unchanged — we skip it, not speed it up |

The ~110 ns `Caller()` is ~54% of a `DiscardLogger` call, so removing it for
source-less formats roughly cuts the per-call cost into a third. Formats that do
render the source (including `DefaultFormatter`) take the exact same path as
before, so there is no regression for the common case.

**Lesson:** the cheapest work is the work you don't do. When an expensive hot-path
computation is already optimal, look for callers that discard its result and gate
it behind a flag computed once at configuration time.

---

## Optimization H: bound pooled buffer capacity + fully reset pooled records

Two `sync.Pool` memory-safety changes. (1) `bufPool` buffers whose capacity
exceeds 64 KiB are dropped instead of pooled, so one megabyte-long message cannot
grow a pooled buffer and pin that capacity for the process lifetime. (2) Clear
`Record.message` and `Record.file` (alongside the existing `r.args = nil`,
Optimization D) before returning a record to `recordPool`.

**Result: neutral in benchmarks (within noise, 0 allocs).** The existing
benchmarks log short, fixed messages, so neither the oversized-buffer guard nor
the extra field clears have anything to reclaim — exactly as with Optimization D.

**Lesson:** unbounded `sync.Pool` retention is a real leak in production (rare huge
records permanently inflate pooled capacity) even when microbenchmarks show
nothing. Keep defensive pool hygiene; the guard is a single `Cap()` comparison on
a path already dominated by I/O.

---

## Optimization I: pack digit lookup tables into contiguous backing arrays

`uintBytes2` / `uintBytes4` / `uintBytes` held one separately heap-allocated
`[]byte` per entry (~1200 tiny slices). Point each table into one of three fixed
backing arrays (`init()` writes digits in place via the shared `writeFixedUint`).
The rendered bytes are identical.

**Result: neutral on throughput (within noise); not a speed optimization.** The
benefit is memory layout — contiguous digit bytes (better cache locality) and ~3
GC-scanned objects instead of ~1200. It is kept because it is strictly better on
allocation count and locality at zero output or speed cost.

**Lesson:** "fewer, larger allocations" is worth doing for GC/locality hygiene
even when wall-clock benchmarks are flat — but label it honestly as a memory-layout
change, not a throughput win, and prove output is byte-identical.

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
| on-demand `Caller()` (source-less formats) | **+70%** (201→59 ns) | n/a | Keep (default path unchanged) |
| bound bufPool + reset record fields | neutral | neutral | Keep (defensive pool hygiene) |
| pack digit tables into backing arrays | neutral | neutral | Keep (GC/locality, byte-identical) |

**Core principle: always measure before and after. Optimizations that "should" help
can make things worse, especially when adding concurrency primitives to hot paths.**
