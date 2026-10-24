package gollback_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vardius/gollback/v2"
)

func ExampleRace() {
	// Ask two replicas for the same thing: the first good answer wins, and the other is cancelled.
	answer, err := gollback.Race(context.Background(),
		func(ctx context.Context) (string, error) {
			return "", errors.New("replica a: unavailable")
		},
		func(ctx context.Context) (string, error) {
			return "answer from replica b", nil
		},
	)

	fmt.Println(answer, err)
	// Output: answer from replica b <nil>
}

func ExampleRace_allFail() {
	_, err := gollback.Race(context.Background(),
		func(ctx context.Context) (string, error) { return "", errors.New("replica a: unavailable") },
		func(ctx context.Context) (string, error) { return "", errors.New("replica b: timeout") },
	)

	// Every error, in the order of the fns.
	fmt.Println(err)
	// Output:
	// replica a: unavailable
	// replica b: timeout
}

func ExampleAll() {
	price := func(item string, cents int, err error) gollback.AsyncFunc[int] {
		return func(ctx context.Context) (int, error) { return cents, err }
	}

	prices, errs := gollback.All(context.Background(),
		price("tea", 450, nil),
		price("cake", 0, errors.New("cake: out of stock")),
		price("coffee", 500, nil),
	)

	fmt.Println(prices)
	fmt.Println(errs)
	// Output:
	// [450 0 500]
	// [<nil> cake: out of stock <nil>]
}

func ExampleRetry() {
	calls := 0
	flaky := func(ctx context.Context) (string, error) {
		if calls++; calls < 3 {
			return "", fmt.Errorf("call %d: connection reset", calls)
		}
		return "ok", nil
	}

	res, err := gollback.Retry(context.Background(), 5, flaky)

	fmt.Println(res, err, "after", calls, "calls")
	// Output: ok <nil> after 3 calls
}

func ExampleRetry_timeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// 0 attempts: no limit, so only ctx stops it.
	_, err := gollback.Retry(ctx, 0, func(ctx context.Context) (string, error) {
		time.Sleep(10 * time.Millisecond)
		return "", errors.New("connection refused")
	})

	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	fmt.Println(err)
	// Output:
	// true
	// context deadline exceeded
	// connection refused
}
