package gollback

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRace(t *testing.T) {
	r, err := Race(
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

	if err != nil {
		t.Fail()
	}

	if r != 3 {
		t.Fail()
	}
}

func TestRaceTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Race(
		ctx,
		func(ctx context.Context) (interface{}, error) {
			time.Sleep(10 * time.Second)
			return 1, nil
		},
	)

	if err != ctx.Err() {
		t.Fail()
	}
}

func TestAll(t *testing.T) {
	err := errors.New("failed")

	rs, errs := All(
		context.Background(),
		func(ctx context.Context) (interface{}, error) {
			time.Sleep(3 * time.Second)
			return 1, nil
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, err
		},
		func(ctx context.Context) (interface{}, error) {
			return 3, nil
		},
	)

	if !testErrorsEq(errs, []error{nil, err, nil}) {
		t.Fail()
	}

	if !testResponsesEq(rs, []interface{}{1, nil, 3}) {
		t.Fail()
	}
}

func TestRetryTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Will retry infinitely until timeouts by context (after 5 seconds)
	_, err := Retry(ctx, 0, func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("failed")
	})

	if err != ctx.Err() {
		t.Fail()
	}
}

func TestRetryFail(t *testing.T) {
	err := errors.New("failed")

	// Will retry 5 times
	_, e := Retry(context.Background(), 5, func(ctx context.Context) (interface{}, error) {
		return nil, err
	})

	if err != e {
		t.Fail()
	}
}

func TestRetrySuccess(t *testing.T) {
	res, _ := Retry(context.Background(), 5, func(ctx context.Context) (interface{}, error) {
		return "success", nil
	})

	if res != "success" {
		t.Fail()
	}
}

func testErrorsEq(a, b []error) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func testResponsesEq(a, b []interface{}) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
