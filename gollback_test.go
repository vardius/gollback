package gollback

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRace(t *testing.T) {
	tests := []struct {
		name    string
		fns     []AsyncFunc
		want    interface{}
		wantErr bool
	}{
		{name: "no callback", want: nil, wantErr: true, fns: []AsyncFunc{}},
		{name: "does not return error", want: 1, wantErr: false, fns: []AsyncFunc{
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Second)
				return 1, nil
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Second)
				return 1, nil
			},
			func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("failed")
			},
		}},
		{name: "first non error wins", want: 3, wantErr: false, fns: []AsyncFunc{
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
		}},
		{name: "all functions error out but first one to finish", want: 1, wantErr: false, fns: []AsyncFunc{
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(1 * time.Second)
				return 1, nil
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(2 * time.Second)
				return nil, errors.New("failed")
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, errors.New("failed")
			},
		}},
		{name: "all functions error out but middle one to finish", want: 1, wantErr: false, fns: []AsyncFunc{
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(2 * time.Second)
				return 1, nil
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(1 * time.Second)
				return nil, errors.New("failed")
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, errors.New("failed")
			},
		}},
		{name: "all functions errors but last one to finish", want: 1, wantErr: false, fns: []AsyncFunc{
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(3 * time.Second)
				return 1, nil
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(2 * time.Second)
				return nil, errors.New("failed")
			},
			func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("failed")
			},
		}},
		{name: "all functions errors and return last err", want: nil, wantErr: true, fns: []AsyncFunc{
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(3 * time.Second)
				return nil, errors.New("failed")
			},
			func(ctx context.Context) (interface{}, error) {
				time.Sleep(2 * time.Second)
				return nil, errors.New("failed")
			},
			func(ctx context.Context) (interface{}, error) {
				return nil, errors.New("failed")
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Race(context.Background(), tt.fns...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Race() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Race() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaceNilContext(t *testing.T) {
	assertPanic(t, "nil context provided", func() {
		_, _ = Race(
			nil,
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
	})
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

func TestAllNilContext(t *testing.T) {
	assertPanic(t, "nil context provided", func() {
		_, _ = All(
			nil,
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
	})
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

func TestRetryNilContext(t *testing.T) {
	err := errors.New("failed")

	assertPanic(t, "nil context provided", func() {
		_, _ = Retry(nil, 5, func(ctx context.Context) (interface{}, error) {
			return nil, err
		})
	})
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

// assertPanic asserts if a function panics with an expected message
func assertPanic(t *testing.T, expected interface{}, f func()) {
	didPanic := false
	var message interface{}

	func() {
		defer func() {
			if message = recover(); message != nil {
				didPanic = true
			}
		}()

		// call the target function
		f()
	}()

	if !didPanic {
		t.Fatal("Did not panic")
	}

	if message != expected {
		t.Fatalf("Unexpected panic\nExpected: %q\nActual: %q", expected, message)
	}
}
