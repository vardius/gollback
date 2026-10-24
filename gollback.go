package gollback

import (
	"context"
	"errors"
	"sync"
)

// ErrNoCallbacks is what Race returns when it is given no functions to run.
var ErrNoCallbacks = errors.New("gollback: no callbacks to run")

// AsyncFunc is a function that Race, All and Retry call. It should return once ctx is done.
type AsyncFunc[T any] func(ctx context.Context) (T, error)

// Race calls each fn in its own goroutine and returns the first result that comes with a nil error.
// Once one succeeds, the context passed to the others is cancelled, and Race returns without waiting for them.
//
// If every fn fails, Race returns their errors joined, in the order of fns.
// If ctx is done first, Race returns ctx's error. With no fns, it returns ErrNoCallbacks.
func Race[T any](ctx context.Context, fns ...AsyncFunc[T]) (T, error) {
	var zero T
	if len(fns) == 0 {
		return zero, ErrNoCallbacks
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r := &race[T]{ctx: ctx, cancel: cancel, fns: fns, pending: len(fns)}
	for i := range fns {
		go r.run(i)
	}

	// Closed by the winner, by the last fn to fail, or by the parent.
	<-ctx.Done()

	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.won:
		return r.value, nil
	case r.pending == 0:
		return zero, errors.Join(r.errs...)
	default:
		return zero, ctx.Err()
	}
}

type race[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	fns    []AsyncFunc[T]

	mu      sync.Mutex
	pending int // fns still running
	won     bool
	value   T
	errs    []error // made at the first failure, so a race that is won allocates none
}

func (r *race[T]) run(i int) {
	v, err := r.fns[i](r.ctx)

	r.mu.Lock()
	r.pending--
	decided := false
	switch {
	case r.won:
	case err == nil:
		r.won, r.value, decided = true, v, true
	default:
		if r.errs == nil {
			r.errs = make([]error, len(r.fns))
		}
		r.errs[i] = err
		decided = r.pending == 0
	}
	r.mu.Unlock()

	if decided {
		r.cancel()
	}
}

// All calls every fn concurrently and waits for all of them to return.
// results[i] and errs[i] are what fns[i] returned: results[i] is the zero value when errs[i] is not nil.
//
// The first fn runs on the calling goroutine and each of the others on its own.
func All[T any](ctx context.Context, fns ...AsyncFunc[T]) (results []T, errs []error) {
	results = make([]T, len(fns))
	errs = make([]error, len(fns))
	if len(fns) == 0 {
		return results, errs
	}

	var a *all[T]
	if len(fns) > 1 {
		a = &all[T]{ctx: ctx, fns: fns, results: results, errs: errs}
		a.wg.Add(len(fns) - 1)
		for i := 1; i < len(fns); i++ {
			go a.run(i)
		}
	}
	results[0], errs[0] = settle(fns[0](ctx))
	if a != nil {
		a.wg.Wait()
	}

	return results, errs
}

type all[T any] struct {
	ctx     context.Context
	fns     []AsyncFunc[T]
	results []T
	errs    []error
	wg      sync.WaitGroup
}

func (a *all[T]) run(i int) {
	defer a.wg.Done()
	a.results[i], a.errs[i] = settle(a.fns[i](a.ctx))
}

// settle drops the result of a call that failed.
func settle[T any](v T, err error) (T, error) {
	if err != nil {
		var zero T
		return zero, err
	}
	return v, nil
}

// Retry calls fn until it returns a nil error, it has been called attempts times, or ctx is done.
// An attempts below 1 means no limit: fn is called until it succeeds or ctx is done.
// Calls follow one another with no pause; to wait between attempts, wait in fn.
//
// When fn never succeeds, Retry returns its last error. When ctx is done first, it returns ctx's error,
// joined with fn's last error if there was one.
func Retry[T any](ctx context.Context, attempts int, fn AsyncFunc[T]) (T, error) {
	var zero T
	var last error
	for n := 0; attempts < 1 || n < attempts; n++ {
		// ctx.Err, not ctx.Done: Done makes a cancellable context allocate its channel.
		if err := ctx.Err(); err != nil {
			if last == nil {
				return zero, err
			}
			return zero, errors.Join(err, last)
		}

		v, err := fn(ctx)
		if err == nil {
			return v, nil
		}
		last = err
	}

	return zero, last
}
