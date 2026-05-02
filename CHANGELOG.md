# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

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

- Initial public version baseline before the current unreleased fixes.
