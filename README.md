Vardius - gollback
================
[![Build Status](https://travis-ci.org/vardius/gollback.svg?branch=master)](https://travis-ci.org/vardius/gollback)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/gollback)](https://goreportcard.com/report/github.com/vardius/gollback)
[![codecov](https://codecov.io/gh/vardius/gollback/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/gollback)
[![](https://godoc.org/github.com/vardius/gollback?status.svg)](http://godoc.org/github.com/vardius/gollback)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/gollback/blob/master/LICENSE.md)

gollback - Go asynchronous function simple utilities, for managing execution of closure, callbacks.

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
BenchmarkRace-4          5000000               240 ns/op               0 B/op          0 allocs/op
BenchmarkAll-4           1000000              2387 ns/op             464 B/op          2 allocs/op
PASS
ok      github.com/vardius/gollback     10.572s
```

## Race example
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

	fmt.Println(r)
	fmt.Println(err)
	// Output:
	// 3
	// <nil>
}
```

## All example
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

	fmt.Println(rs)
	fmt.Println(errs)
	// Output:
	// [1 <nil> 3]
	// [<nil> failed <nil>]
}
```

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
