package gollback

import (
	"context"
	"sync"
)

// AsyncFunc represents asynchronous function
type AsyncFunc func(ctx context.Context) (interface{}, error)

type response struct {
	res interface{}
	err error
}

// Race method returns a response as soon as one of the callbacks in an iterable executes without an error,
// otherwise last error is returned
// will panic if context is nil
func Race(ctx context.Context, fns ...AsyncFunc) (interface{}, error) {
	if ctx == nil {
		panic("nil context provided")
	}

	out := make(chan *response, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, fn := range fns {
		go func(index int, f AsyncFunc) {
			c := make(chan *response, 1)

			go func() {
				defer close(c)
				var r response
				r.res, r.err = f(ctx)

				c <- &r
			}()

			for {
				select {
				case <-ctx.Done():
					if index == len(fns)-1 {
						out <- &response{
							err: ctx.Err(),
						}
					}

					return
				case r := <-c:
					if r.err == nil || index == len(fns)-1 {
						out <- r
					}

					return
				}
			}
		}(i, fn)
	}

	r := <-out

	return r.res, r.err
}

// All method returns when all of the callbacks passed as an iterable have finished,
// returned responses and errors are ordered according to callback order
// will panic if context is nil
func All(ctx context.Context, fns ...AsyncFunc) ([]interface{}, []error) {
	if ctx == nil {
		panic("nil context provided")
	}

	rs := make([]interface{}, len(fns))
	errs := make([]error, len(fns))

	var wg sync.WaitGroup
	wg.Add(len(fns))

	for i, fn := range fns {
		go func(index int, f AsyncFunc) {
			defer wg.Done()

			var r response
			r.res, r.err = f(ctx)

			rs[index] = r.res
			errs[index] = r.err
		}(i, fn)
	}

	wg.Wait()

	return rs, errs
}

// Retry method retries callback given amount of times until it executes without an error,
// when retries = 0 it will retry infinitely
// will panic if context is nil
func Retry(ctx context.Context, retires int, fn AsyncFunc) (interface{}, error) {
	if ctx == nil {
		panic("nil context provided")
	}

	i := 1

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			var r response
			r.res, r.err = fn(ctx)

			if r.err == nil || i == retires {
				return r.res, r.err
			}

			i++
		}
	}
}
