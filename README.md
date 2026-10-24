# ⚙️ gollback

[![Test](https://github.com/vardius/gollback/actions/workflows/test.yml/badge.svg)](https://github.com/vardius/gollback/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vardius/gollback/v2.svg)](https://pkg.go.dev/github.com/vardius/gollback/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/gollback/v2)](https://goreportcard.com/report/github.com/vardius/gollback/v2)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

Run Go functions concurrently and get typed results back.

- **Race** returns the first result without an error and cancels the rest.
- **All** waits for every function and returns each result and error, in order.
- **Retry** calls a function until it succeeds, runs out of attempts, or its context is done.
- **Typed.** Generic functions return a `T`, not an `interface{}`: no type assertions.
- **Lean.** `Retry` allocates nothing, `All` runs the first function on your goroutine, and tests guard the allocation counts.

📖 **[Documentation](https://vardius.github.io/gollback/)** · [API reference](https://pkg.go.dev/github.com/vardius/gollback/v2) · [Examples](examples) · [Migrating from v1](MIGRATION.md)

## Install

```sh
go get github.com/vardius/gollback/v2
```

Requires Go 1.26 or newer.

## Race

The first function to return a nil error wins, and the context the others were given is cancelled.
Race doesn't wait for them. If every function fails, it returns all their errors, joined in order.

```go
user, err := gollback.Race(ctx,
	func(ctx context.Context) (User, error) { return primary.User(ctx, id) },
	func(ctx context.Context) (User, error) { return replica.User(ctx, id) },
)
```

## All

Every function runs at once; All returns when the last one does. `results[i]` and `errs[i]` are what the i-th function returned,
and a failed function's result is the zero value.

```go
prices, errs := gollback.All(ctx,
	func(ctx context.Context) (Price, error) { return shopA.Price(ctx, sku) },
	func(ctx context.Context) (Price, error) { return shopB.Price(ctx, sku) },
)
```

## Retry

Calls the function until it succeeds, it has been called `attempts` times, or ctx is done. Below 1 attempt means no limit.
When ctx ends it, the error is ctx's joined with the function's last one.

```go
cfg, err := gollback.Retry(ctx, 5, func(ctx context.Context) (Config, error) {
	return store.Load(ctx, "app")
})
```

Retry calls again at once. To back off between attempts, wait in the function: see [examples/retry](examples/retry/main.go).

## Benchmarks

Time and allocations per call, on an Apple M4 Max with Go 1.27.0, median of 5 runs. Both versions ran on the same harness,
with functions that return at once; in a race, one wins and the others wait to be cancelled.
With 16 callers at once, the time is wall time per call across all of them, so it measures throughput.
Run v2's own suite with `go test -bench .`.

| Scenario | v1 | v2 |
| --- | ---: | ---: |
| Race, 1 function | 762 ns · 13 allocs | **348 ns · 5 allocs** |
| Race, 4 functions | 1.95 µs · 22 allocs | **1.35 µs · 8 allocs** |
| Race, 4 functions, all fail | 2.19 µs · 22 allocs | **1.08 µs · 11 allocs** |
| Race, 4 functions, 16 callers at once | 446 ns · 22 allocs | **284 ns · 8 allocs** |
| All, 1 function | 285 ns · 5 allocs | **15.2 ns · 2 allocs** |
| All, 4 functions | 950 ns · 11 allocs | **719 ns · 6 allocs** |
| All, 4 functions, 16 callers at once | 313 ns · 11 allocs | **191 ns · 6 allocs** |
| Retry, 5 failed attempts | 24.6 ns · 0 allocs | **13.4 ns · 0 allocs** |

Beyond its own state, each goroutine costs one small allocation, the runtime's, for a `go` statement's arguments.
Handing work to pooled goroutines would avoid it, but a prototype that did was slower with many callers at once.

## Upgrading from v1

The functions are the same, now generic:

```go
// v1
res, err := gollback.Race(ctx,
	func(ctx context.Context) (interface{}, error) { return primary.User(ctx, id) },
	func(ctx context.Context) (interface{}, error) { return replica.User(ctx, id) },
)
user := res.(User)

// v2
user, err := gollback.Race(ctx,
	func(ctx context.Context) (User, error) { return primary.User(ctx, id) },
	func(ctx context.Context) (User, error) { return replica.User(ctx, id) },
)
```

[MIGRATION.md](MIGRATION.md) maps every v1 call to its v2 form and lists the behaviour changes.

## Contributing

Bug reports and pull requests are welcome: see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

MIT, see [LICENSE.md](LICENSE.md).
