package gollback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vardius/gollback/v2"
)

var errFailed = errors.New("failed")

func succeed(context.Context) (int, error) { return 1, nil }
func fail(context.Context) (int, error)    { return 0, errFailed }

func repeat(fn gollback.AsyncFunc[int], n int) []gollback.AsyncFunc[int] {
	fns := make([]gollback.AsyncFunc[int], n)
	for i := range fns {
		fns[i] = fn
	}
	return fns
}

// TestAllocations guards the allocations per call. Lower a number when an optimisation lands;
// never raise one without a reason in the commit.
func TestAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector changes allocation counts")
	}

	ctx := context.Background()
	one, four := repeat(succeed, 1), repeat(succeed, 4)

	tests := []struct {
		name string
		call func()
		want float64
	}{
		{"Retry, 5 failed attempts", func() { _, _ = gollback.Retry(ctx, 5, fail) }, 0},
		// The two slices it returns. The one fn runs on the caller's goroutine.
		{"All, 1 fn", func() { _, _ = gollback.All(ctx, one...) }, 2},
		// Plus the state the goroutines share, and one closure per goroutine: the go statement's own.
		{"All, 4 fns", func() { _, _ = gollback.All(ctx, four...) }, 6},
		// Its state, the context.WithCancel context, its cancel func and Done channel, and the goroutine's closure.
		{"Race, 1 fn", func() { _, _ = gollback.Race(ctx, one...) }, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(1000, tt.call); got != tt.want {
				t.Errorf("allocates %v times per call, want %v", got, tt.want)
			}
		})
	}
}
