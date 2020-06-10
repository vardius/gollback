⚙️ gollback
================
[![Build Status](https://travis-ci.org/vardius/gollback.svg?branch=master)](https://travis-ci.org/vardius/gollback)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/gollback)](https://goreportcard.com/report/github.com/vardius/gollback)
[![codecov](https://codecov.io/gh/vardius/gollback/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/gollback)
[![](https://godoc.org/github.com/vardius/gollback?status.svg)](https://pkg.go.dev/github.com/vardius/gollback)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/gollback/blob/master/LICENSE.md)

<img align="right" height="180px" src="https://github.com/vardius/gorouter/blob/master/website/src/static/img/logo.png?raw=true" alt="logo" />

gollback - Go asynchronous simple function utilities, for managing execution of closures and callbacks

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/gollback/issues) to manage them.

## 📚 Documentation

For __examples__ **visit [godoc#pkg-examples](http://godoc.org/github.com/vardius/gollback#pkg-examples)**

For **GoDoc** reference, **visit [pkg.go.dev](https://pkg.go.dev/github.com/vardius/gollback)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  gollback git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/gollback
BenchmarkRace-4    	  566022	      2608 ns/op	     663 B/op	       5 allocs/op
BenchmarkAll-4     	 5052489	       241 ns/op	      42 B/op	       1 allocs/op
BenchmarkRetry-4   	206430384	         5.93 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/vardius/gollback	31.741s
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
	r, err := gollback.Race(
		context.Background(),
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
	rs, errs := gollback.All(
		context.Background(),
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

	// Will retry infinitely until timeouts by context (after 5 seconds)
	res, err := gollback.Retry(ctx, 0, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})

	// Will retry 5 times or will timeout by context (after 5 seconds)
	res, err := gollback.Retry(ctx, 5, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})
}
```

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package
