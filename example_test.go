package gollback_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	gollback "github.com/vardius/gollback"
)

func Example_race() {
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

func Example_all() {
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

func Example_retryTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g := gollback.New(ctx)

	// Will retry infinitely until timeouts by context (after 5 seconds)
	res, err := g.Retry(0, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})

	fmt.Println(res)
	fmt.Println(err)
	// Output:
	// <nil>
	// context deadline exceeded
}

func Example_retryFiveTimes() {
	g := gollback.New(context.Background())

	// Will retry 5 times
	res, err := g.Retry(5, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})

	fmt.Println(res)
	fmt.Println(err)
	// Output:
	// <nil>
	// failed
}
