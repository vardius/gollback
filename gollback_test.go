package gollback

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	g := New(nil)

	if g == nil {
		t.Fail()
	}
}

func TestNewWithContext(t *testing.T) {
	g := New(context.Background())

	if g == nil {
		t.Fail()
	}
}

func TestRace(t *testing.T) {
	g := New(context.Background())

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

	if err != nil {
		t.Fail()
	}

	if r != 3 {
		t.Fail()
	}
}

func TestAll(t *testing.T) {
	g := New(context.Background())
	err := errors.New("failed")

	rs, errs := g.All(
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
