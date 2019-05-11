package gollback

import (
	"context"
)

// AsyncFunc represents asynchronous function
type AsyncFunc func(ctx context.Context) (interface{}, error)

// Gollback provides set of utility methods to easily manage asynchronous functions
type Gollback interface {
	Race(fns ...AsyncFunc) (interface{}, error)
	All(fns ...AsyncFunc) ([]interface{}, []error)
}

type gollback struct {
	gollbacks []AsyncFunc
	ctx       context.Context
	cancel    context.CancelFunc
}

type response struct {
	res   interface{}
	err   error
	index int
}

// Race method returns a response as soon as one of the callbacks in an iterable resolves with the value that is not an error,
// otherwise last error is returned
func (p *gollback) Race(fns ...AsyncFunc) (interface{}, error) {
	out := make(chan *response, 1)

	for i, fn := range fns {
		go func(index int, f AsyncFunc) {
			for {
				select {
				case <-p.ctx.Done():
					return
				default:
					var r response
					r.res, r.err = f(p.ctx)

					if p.ctx.Err() != nil {
						return
					}

					if r.err == nil || index == len(fns)-1 {
						p.cancel()
						out <- &r
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
func (p *gollback) All(fns ...AsyncFunc) ([]interface{}, []error) {
	out := make(chan *response, len(fns))

	for i, fn := range fns {
		go func(index int, f AsyncFunc) {
			for {
				select {
				case <-p.ctx.Done():
					return
				default:
					var r response
					r.res, r.err = f(p.ctx)

					if p.ctx.Err() != nil {
						return
					}

					r.index = index

					out <- &r

					return
				}
			}
		}(i, fn)
	}

	rs := make([]interface{}, len(fns))
	errs := make([]error, len(fns))

	for i := 0; i < len(fns); i++ {
		r := <-out

		rs[r.index] = r.res
		errs[r.index] = r.err
	}

	p.cancel()

	return rs, errs
}

// New creates new gollback
func New(ctx context.Context) Gollback {
	if ctx == nil {
		ctx = context.Background()
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)

	return &gollback{
		ctx:    ctxWithCancel,
		cancel: cancel,
	}
}
