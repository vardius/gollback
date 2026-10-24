package gollback_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vardius/gollback/v2"
)

var (
	errA = errors.New("a failed")
	errB = errors.New("b failed")
)

// waitForLosers lets the fns that lost a race finish: Race returns without them,
// and a synctest bubble must not end while they run.
func waitForLosers() {
	time.Sleep(24 * time.Hour)
}

// after returns a func that waits d, ignoring ctx, then returns v and err.
func after(d time.Duration, v int, err error) gollback.AsyncFunc[int] {
	return func(context.Context) (int, error) {
		time.Sleep(d)
		return v, err
	}
}

func TestRace(t *testing.T) {
	tests := []struct {
		name    string
		fns     []gollback.AsyncFunc[int]
		want    int
		wantErr error
		elapsed time.Duration
	}{
		{
			name:    "first to succeed wins",
			fns:     []gollback.AsyncFunc[int]{after(3*time.Second, 1, nil), after(0, 0, errA), after(time.Second, 3, nil)},
			want:    3,
			elapsed: time.Second,
		},
		{
			name:    "a failure does not end the race",
			fns:     []gollback.AsyncFunc[int]{after(0, 0, errA), after(2*time.Second, 2, nil)},
			want:    2,
			elapsed: 2 * time.Second,
		},
		{
			name:    "a single fn",
			fns:     []gollback.AsyncFunc[int]{after(time.Second, 7, nil)},
			want:    7,
			elapsed: time.Second,
		},
		{
			name:    "every fn fails",
			fns:     []gollback.AsyncFunc[int]{after(2*time.Second, 1, errA), after(time.Second, 2, errB)},
			wantErr: errors.Join(errA, errB),
			elapsed: 2 * time.Second,
		},
		{
			name:    "no fns",
			wantErr: gollback.ErrNoCallbacks,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				start := time.Now()
				got, err := gollback.Race(t.Context(), tt.fns...)

				if fmt.Sprint(err) != fmt.Sprint(tt.wantErr) {
					t.Errorf("Race() error = %v, want %v", err, tt.wantErr)
				}
				if got != tt.want {
					t.Errorf("Race() = %d, want %d", got, tt.want)
				}
				if elapsed := time.Since(start); elapsed != tt.elapsed {
					t.Errorf("Race() returned after %v, want %v", elapsed, tt.elapsed)
				}
				waitForLosers()
			})
		})
	}
}

func TestRaceJoinsErrorsInOrderOfFns(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		_, err := gollback.Race(t.Context(), after(2*time.Second, 0, errA), after(time.Second, 0, errB))

		if !errors.Is(err, errA) || !errors.Is(err, errB) {
			t.Fatalf("Race() error = %v, want both %v and %v", err, errA, errB)
		}
		if want := "a failed\nb failed"; err.Error() != want {
			t.Errorf("Race() error = %q, want %q: fns order, not the order they failed in", err, want)
		}
	})
}

func TestRaceCancelsLosers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		loserErr := make(chan error, 1)
		loser := func(ctx context.Context) (int, error) {
			<-ctx.Done()
			loserErr <- ctx.Err()
			return 0, ctx.Err()
		}

		got, err := gollback.Race(t.Context(), loser, after(time.Second, 1, nil))
		if got != 1 || err != nil {
			t.Fatalf("Race() = %d, %v, want 1, nil", got, err)
		}
		if err := <-loserErr; !errors.Is(err, context.Canceled) {
			t.Errorf("loser's ctx error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestRaceDoesNotWaitForLosers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var loserDone atomic.Bool
		stubborn := func(context.Context) (int, error) {
			time.Sleep(time.Hour) // ignores ctx
			loserDone.Store(true)
			return 0, nil
		}

		start := time.Now()
		got, err := gollback.Race(t.Context(), stubborn, after(time.Second, 1, nil))
		if got != 1 || err != nil {
			t.Fatalf("Race() = %d, %v, want 1, nil", got, err)
		}
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Errorf("Race() returned after %v, want 1s", elapsed)
		}
		if loserDone.Load() {
			t.Error("Race() waited for the loser")
		}
		waitForLosers()
		if !loserDone.Load() {
			t.Error("the loser never finished")
		}
	})
}

func TestRaceParentDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		start := time.Now()
		got, err := gollback.Race(ctx, after(time.Hour, 1, nil), after(time.Hour, 2, nil))

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Race() error = %v, want %v", err, context.DeadlineExceeded)
		}
		if got != 0 {
			t.Errorf("Race() = %d, want the zero value", got)
		}
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Errorf("Race() returned after %v, want 1s", elapsed)
		}
		waitForLosers()
	})
}

func TestRacePassesParentValues(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(t.Context(), key{}, "value")

	got, err := gollback.Race(ctx, func(ctx context.Context) (string, error) {
		return ctx.Value(key{}).(string), nil
	})
	if got != "value" || err != nil {
		t.Errorf("Race() = %q, %v, want %q, nil", got, err, "value")
	}
}

func TestAll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		results, errs := gollback.All(t.Context(),
			after(3*time.Second, 1, nil),
			after(0, 2, errB), // its result is dropped: it failed
			after(time.Second, 3, nil),
		)

		if want := []int{1, 0, 3}; !slices.Equal(results, want) {
			t.Errorf("All() results = %v, want %v", results, want)
		}
		if want := []error{nil, errB, nil}; !slices.Equal(errs, want) {
			t.Errorf("All() errs = %v, want %v", errs, want)
		}
		if elapsed := time.Since(start); elapsed != 3*time.Second {
			t.Errorf("All() returned after %v, want 3s: as long as the slowest fn", elapsed)
		}
	})
}

func TestAllNoFns(t *testing.T) {
	results, errs := gollback.All[int](t.Context())
	if len(results) != 0 || len(errs) != 0 {
		t.Errorf("All() = %v, %v, want empty slices", results, errs)
	}
}

func TestAllRunsConcurrently(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Each fn waits for the other, so All only returns if they run at the same time.
		ping, pong := make(chan struct{}), make(chan struct{})
		results, _ := gollback.All(t.Context(),
			func(context.Context) (int, error) { ping <- struct{}{}; <-pong; return 1, nil },
			func(context.Context) (int, error) { <-ping; pong <- struct{}{}; return 2, nil },
		)
		if want := []int{1, 2}; !slices.Equal(results, want) {
			t.Errorf("All() results = %v, want %v", results, want)
		}
	})
}

func TestAllPassesCtx(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(t.Context(), key{}, 7)
	read := func(ctx context.Context) (int, error) { return ctx.Value(key{}).(int), nil }

	results, _ := gollback.All(ctx, read, read)
	if want := []int{7, 7}; !slices.Equal(results, want) {
		t.Errorf("All() results = %v, want %v", results, want)
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name      string
		attempts  int
		succeedOn int // the call that succeeds; 0 for none
		want      int
		wantErr   string
		wantCalls int
	}{
		{name: "succeeds on the first call", attempts: 3, succeedOn: 1, want: 1, wantCalls: 1},
		{name: "succeeds on the third call", attempts: 5, succeedOn: 3, want: 3, wantCalls: 3},
		{name: "succeeds on the last call", attempts: 3, succeedOn: 3, want: 3, wantCalls: 3},
		{name: "never succeeds", attempts: 3, wantErr: "call 3 failed", wantCalls: 3},
		{name: "one attempt", attempts: 1, wantErr: "call 1 failed", wantCalls: 1},
		{name: "zero attempts means no limit", attempts: 0, succeedOn: 100, want: 100, wantCalls: 100},
		{name: "negative attempts means no limit", attempts: -1, succeedOn: 50, want: 50, wantCalls: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			got, err := gollback.Retry(t.Context(), tt.attempts, func(context.Context) (int, error) {
				calls++
				if calls == tt.succeedOn {
					return calls, nil
				}
				return calls, fmt.Errorf("call %d failed", calls)
			})

			if msg := errString(err); msg != tt.wantErr {
				t.Errorf("Retry() error = %v, want %q", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Retry() = %d, want %d", got, tt.want)
			}
			if calls != tt.wantCalls {
				t.Errorf("fn called %d times, want %d", calls, tt.wantCalls)
			}
		})
	}
}

func TestRetryStopsWhenCtxIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	calls := 0
	_, err := gollback.Retry(ctx, 0, func(context.Context) (int, error) {
		calls++
		if calls == 3 {
			cancel()
		}
		return 0, errA
	})

	if !errors.Is(err, context.Canceled) || !errors.Is(err, errA) {
		t.Errorf("Retry() error = %v, want %v joined with %v", err, context.Canceled, errA)
	}
	if calls != 3 {
		t.Errorf("fn called %d times, want 3", calls)
	}
}

func TestRetryCtxDoneBeforeFirstCall(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := gollback.Retry(ctx, 5, func(context.Context) (int, error) {
		t.Fatal("fn called after ctx was done")
		return 0, nil
	})
	if err != context.Canceled {
		t.Errorf("Retry() error = %v, want %v itself", err, context.Canceled)
	}
}

func TestRetryDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 2500*time.Millisecond)
		defer cancel()

		calls := 0
		_, err := gollback.Retry(ctx, 0, func(context.Context) (int, error) {
			calls++
			time.Sleep(time.Second)
			return 0, errA
		})

		if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, errA) {
			t.Errorf("Retry() error = %v, want %v joined with %v", err, context.DeadlineExceeded, errA)
		}
		if calls != 3 {
			t.Errorf("fn called %d times, want 3: at 0s, 1s and 2s", calls)
		}
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
