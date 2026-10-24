# Changelog

## v2.0.0 (2026-10-24)

The same three functions, now generic. See [MIGRATION.md](MIGRATION.md) for the before and after of every call.

### Added

- A type parameter on `AsyncFunc`, `Race`, `All` and `Retry`: results come back typed, with no type assertions.
- Runnable examples in `examples/`, each with a test: racing mirrors, checking URLs at once, retrying with backoff.

### Changed

- Module path is `github.com/vardius/gollback/v2`; Go 1.26 or newer is required.
- `Race` returns every error, joined in the order of the functions, when they all fail.
- `Race` returns ctx's error joined with the errors of the functions that had failed when ctx is done first,
  and calls no function when ctx is done before the call.
- `Retry` returns ctx's error joined with the function's last error when ctx is done first.
- A function that calls `runtime.Goexit`, as `t.FailNow` does, counts as failed in `Race`, and in `All`
  for every function but the first, which runs on the calling goroutine.
- A failed call's result is dropped: `Race` and `Retry` return the zero value with an error,
  and `All` leaves `results[i]` zero when `errs[i]` is set.
- `All` runs the first function on the calling goroutine.
- Faster, with fewer allocations. Racing one function went from 795 ns and 13 allocations to 349 ns and 4;
  `All` of one function from 293 ns to 18 ns; `Retry` from 25 ns to 14 ns, still allocating nothing.
  Allocation counts are guarded by tests.

### Removed

- The nil-context check and its "nil context provided" panic.
