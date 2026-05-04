# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Fixed

- `FastTimer` is now race-free: snapshots are stored in `atomic.Pointer`, eliminating
  the torn read at the day boundary that previously caused mismatched date/time pairs
  (the README "behavior #3" example no longer occurs).
- `StopFastTimer` now waits for any in-flight background update before returning, so
  stopping the timer cannot republish a cached snapshot after it has been cleared or
  race with an immediate restart.

### Changed

- Module bumped from `go 1.16` to `go 1.19`. The new floor lets the package use
  `atomic.Bool` / `atomic.Pointer` for race-safe internal state.
- `FastTimer` is now built on `atomic.Pointer[snapshot]`. The hot read path costs a
  single atomic load with no measurable per-call regression on either x86_64 or
  Apple Silicon (`DiscardLogger` benchmarks unchanged within noise).
- `ConcurrentFileWriter.closed` is now `atomic.Bool` instead of a `uint32` accessed
  via `atomic.LoadUint32` / `atomic.StoreUint32`.
- All test files now run with `-race` enabled. The historical `//go:build !race`
  guard on `log_test.go` and `log/log_test.go` has been removed.
- Buffered writer constructors share a single internal helper
  (`initBufferedFileWriter`), removing ~30 lines of duplicated initialisation across
  `BufferedFileWriter`, `RotatingFileWriter`, and `TimedRotatingFileWriter`.
- `BufferedFileWriter.Close` now waits for its scheduling goroutine to exit before
  performing the final flush, matching the behaviour of `ConcurrentFileWriter.Close`.
  Concurrent calls are serialised via `sync.Once`.
- `TimedRotatingFileWriter.schedule` collapses its previous two-phase loop with
  `updateLoop` / `flushLoop` labels into a single `select`, with the same observable
  behaviour (writers rate-limit notifications via `w.updated`).
- `Formatter.findParts` is now iterative; the unreachable `*ByteFormatPart` branch
  in `appendBytes` has been deleted.
- `log/log.go` `SetDefaultLogger`, `SetLogFunc`, and `SetLogfFunc` now dispatch
  through `[5]*func` indexed by `Level`, eliminating the previous nested switches.
- `isTimedRotatingFile` rewritten as a single digit/suffix check; previous
  per-character switch (~80 lines) reduced to ~20.
- `utils.go` split into `conv.go` (numeric tables / `writeUintToBuf`),
  `caller.go` (`Caller` + `frameCache` + `runtime.callers` linkname), and
  `timer.go` (`FastTimer` + `now`); `recordPool` and `bufPool` moved next to
  the types that consume them.
- `uintBytes` is now indexed directly by value (`uintBytes[x]` instead of
  `uintBytes[x-2]`); `uintBytes4` extended from 1970-2038 to 1970-2099 (~5 KB
  total table size) so reasonable future timestamps stay on the fast path.
- All `io/ioutil` usages replaced with their `os` equivalents (`os.ReadFile` /
  `os.WriteFile`).

### Added

- `BenchmarkCallerRuntimeCallers` benchmarks the public `runtime.Callers` against the
  linkname'd `Caller`, providing a regression guard for the linkname optimisation.
  Current measurements (Apple M1 Pro): linkname is ~44% faster and zero-alloc, vs.
  ~237 ns/op + 248 B/op + 2 allocs/op for the supported API.
- Comments on `runtime_procPin` / `runtime_procUnpin` and the `callers` linkname
  document why each is necessary and what to do if a future Go version blocks them
  (Go 1.23+ `-checklinkname=0` flag, or fall back to the supported API at the
  documented performance cost).

## v0.2.0

### Fixed

- Avoid panics when formatting years outside the cached `1970-2038` range.
- Avoid panics when formatting records with invalid log levels.
- Avoid panics when creating a handler with a nil formatter; `DefaultFormatter` is used instead.
- Make `BufferedFileWriter.Close` and `ConcurrentFileWriter.Close` idempotent.
- Return `os.ErrClosed` instead of panicking when writing to closed buffered, rotating, or timed rotating file writers.
- Ignore missing old backup files during size-based rotation.
- Prevent timed rotation purge from deleting unrelated files with the same path prefix.
- Make `StartFastTimer` and `StopFastTimer` idempotent.

### Changed

- `ConsoleWriter.Close` and `DiscardWriter.Close` are now no-op operations, so repeated closes and later accidental writes do not panic due to nil internal fields.
- Clarified that logger configuration should be completed before concurrent logging starts.
- Clarified that the fast timer is not race-free.
- Updated CI to test newer Go versions and run race tests for all packages.
- Replace `fastUint2DynamicBytes` with `writeUintToBuf` that writes line numbers directly into the buffer, eliminating heap allocation for line numbers > 1000.
- Set `r.args = nil` before returning `Record` to pool, reducing GC pressure by preventing pooled objects from retaining external references.

## v0.1.0

- Initial public version baseline.
