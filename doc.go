// Package gollback runs functions concurrently and collects what they return.
//
// Race returns the first result that comes without an error and cancels the rest,
// All waits for every function and returns each one's result and error in order,
// and Retry calls a function until it succeeds.
//
//	user, err := gollback.Race(ctx,
//		func(ctx context.Context) (User, error) { return primary.User(ctx, id) },
//		func(ctx context.Context) (User, error) { return replica.User(ctx, id) },
//	)
//
// The functions are AsyncFuncs: they take a context.Context and return a value and an error.
// The type parameter is the value's type, so results come back typed, with no assertions.
//
// # Race
//
// Race starts every function in its own goroutine. The first to return a nil error wins:
// Race returns its result and cancels the context the others were given. It doesn't wait for them,
// so a function that ignores its context keeps running after Race returns, and its result is dropped.
// If every function fails, Race returns their errors joined with errors.Join, in the order of the functions.
// If the context passed to Race is done first, Race returns its error, joined with the errors of
// the functions that had failed by then.
//
// # All
//
// All runs every function concurrently and returns when the last one does. results[i] and errs[i]
// are what the i-th function returned; a failed function's result is the zero value.
// The first function runs on the calling goroutine, the others on goroutines of their own.
// All doesn't cancel anything: to stop the others when one fails, have them watch a context you cancel.
//
// # Retry
//
// Retry calls a function until it returns a nil error, it has been called the given number of times,
// or the context is done. A number below 1 means no limit. Calls follow each other with no pause:
// to back off between attempts, wait in the function, and return early when its context is done.
//
// # Cost
//
// Retry allocates nothing. All allocates the two slices it returns and, for more than one function,
// the state its goroutines share. Race allocates its state, the context it cancels, that context's
// cancel func, and its channel. Starting a goroutine with arguments costs one small allocation; the package pays it rather
// than funnel work through a shared queue, which is slower once many callers race at the same time.
package gollback
