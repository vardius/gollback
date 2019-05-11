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
