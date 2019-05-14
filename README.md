Vardius - gollback
================
[![Build Status](https://travis-ci.org/vardius/gollback.svg?branch=master)](https://travis-ci.org/vardius/gollback)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/gollback)](https://goreportcard.com/report/github.com/vardius/gollback)
[![codecov](https://codecov.io/gh/vardius/gollback/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/gollback)
[![](https://godoc.org/github.com/vardius/gollback?status.svg)](http://godoc.org/github.com/vardius/gollback)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/gollback/blob/master/LICENSE.md)

gollback - Go asynchronous simple function utilities, for managing execution of closures and callbacks

ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/gollback/issues) to manage them.

HOW TO USE
==================================================

1. [GoDoc](http://godoc.org/github.com/vardius/gollback)

## Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  gollback git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/gollback
BenchmarkRace-4    	  500000	      2961 ns/op	     668 B/op	       5 allocs/op
BenchmarkAll-4     	 5000000	       273 ns/op	      40 B/op	       1 allocs/op
BenchmarkRetry-4   	200000000	         6.68 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/vardius/gollback	32.418s
```

## Race
> Race method returns a response as soon as one of the callbacks in an iterable resolves with the value that is not an error, otherwise last error is returne
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

        "github.com/vardius/gollback"
)

func main() {
	g := gollback.New(context.Background())

	r, err := g.Race(
		func(ctx context.Context) (interface{}, error) {
			time.Sleep(3 * time.Second)
			return 1, nil
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failed")
		},
		func(ctx context.Context) (interface{}, error) {
			return 3, nil
		},
	)
}
```

## All
> All method returns when all of the callbacks passed as an iterable have finished, returned responses and errors are ordered according to callback order
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

        "github.com/vardius/gollback"
)

func main() {
	g := gollback.New(context.Background())

	rs, errs := g.All(
		func(ctx context.Context) (interface{}, error) {
			time.Sleep(3 * time.Second)
			return 1, nil
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, errors.New("failed")
		},
		func(ctx context.Context) (interface{}, error) {
			return 3, nil
		},
	)
}
```

## Retry
> Retry method retries callback given amount of times until it executes without an error, when retries = 0 it will retry infinitely
```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

        "github.com/vardius/gollback"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g := gollback.New(ctx)

	// Will retry infinitely until timeouts by context (after 5 seconds)
	res, err := g.Retry(0, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})

	// Will retry 5 times or will timeout by context (after 5 seconds)
	res, e := g.Retry(5, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})
}
```

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
