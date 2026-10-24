# Migrating from v1 to v2

v2 keeps the three functions and what they do. They are now generic, so results come back typed
instead of as `interface{}`, and a few edge cases behave differently: they are listed below.

## 1. Update the import path

```sh
go get github.com/vardius/gollback/v2
```

```go
import "github.com/vardius/gollback/v2"
```

v2 needs Go 1.26 or newer.

## 2. Before and after

```go
// v1
res, err := gollback.Race(ctx,
	func(ctx context.Context) (interface{}, error) { return primary.User(ctx, id) },
	func(ctx context.Context) (interface{}, error) { return replica.User(ctx, id) },
)
if err != nil {
	return err
}
user := res.(User)
```

```go
// v2
user, err := gollback.Race(ctx,
	func(ctx context.Context) (User, error) { return primary.User(ctx, id) },
	func(ctx context.Context) (User, error) { return replica.User(ctx, id) },
)
if err != nil {
	return err
}
```

The type parameter is inferred from the functions, so calls don't name it. A slice of functions does:
`[]gollback.AsyncFunc[User]`.

## 3. Every call

| v1 | v2 |
| --- | --- |
| `type AsyncFunc func(ctx context.Context) (interface{}, error)` | `type AsyncFunc[T any] func(ctx context.Context) (T, error)` |
| `Race(ctx, fns...) (interface{}, error)` | `Race(ctx, fns...) (T, error)` |
| `All(ctx, fns...) ([]interface{}, []error)` | `All(ctx, fns...) ([]T, []error)` |
| `Retry(ctx, retries, fn) (interface{}, error)` | `Retry(ctx, attempts, fn) (T, error)` |
| `ErrNoCallbacks` | `ErrNoCallbacks` |

Functions that returned different types into one call need a common type in v2: an interface they all
implement, or `any` if you really want v1's behaviour back.

## 4. Behaviour changes

- **Race, when every function fails,** returns all of their errors joined with `errors.Join`, in the order of
  the functions. v1 returned one of them. Check for a particular error with `errors.Is` or `errors.As`.
- **Retry, when ctx is done,** returns ctx's error joined with the function's last error, so you can see why
  the attempts failed. `errors.Is(err, context.DeadlineExceeded)` still works; comparing with `==` doesn't.
  If ctx is done before the first call, the error is ctx's alone, as before.
- **A failed call's result is dropped.** v1's Retry returned the last attempt's result beside its error,
  and v1's All kept whatever a failed function returned. v2 returns the zero value with every error.
- **All runs the first function on the calling goroutine,** and the others on goroutines of their own,
  as before. A panic in the first function now reaches the caller, where it can be recovered.
- **A nil context is no longer checked.** v1 panicked with "nil context provided". v2's Race panics in
  `context.WithCancel`, and All and Retry hand the nil context to your functions. Pass `context.Background()`
  or `context.TODO()` instead; `go vet` and staticcheck flag a nil context.
