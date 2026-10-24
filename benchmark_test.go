package gollback_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/vardius/gollback/v2"
)

// The fns return at once, so what is measured is gollback's own cost: goroutines, synchronisation, results.

// hedged returns n fns where the first wins at once and the rest wait to be cancelled,
// like the slower replicas of a hedged request.
func hedged(n int) []gollback.AsyncFunc[int] {
	fns := repeat(func(ctx context.Context) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	}, n)
	fns[0] = succeed
	return fns
}

func BenchmarkRace(b *testing.B) {
	for _, n := range []int{1, 4, 16} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			fns := hedged(n)
			b.ReportAllocs()
			for b.Loop() {
				_, _ = gollback.Race(context.Background(), fns...)
			}
		})
	}
}

func BenchmarkRaceAllFail(b *testing.B) {
	fns := repeat(fail, 4)
	b.ReportAllocs()
	for b.Loop() {
		_, _ = gollback.Race(context.Background(), fns...)
	}
}

func BenchmarkRaceParallel(b *testing.B) {
	fns := hedged(4)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = gollback.Race(context.Background(), fns...)
		}
	})
}

func BenchmarkAll(b *testing.B) {
	for _, n := range []int{1, 4, 16} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			fns := repeat(succeed, n)
			b.ReportAllocs()
			for b.Loop() {
				_, _ = gollback.All(context.Background(), fns...)
			}
		})
	}
}

func BenchmarkAllParallel(b *testing.B) {
	fns := repeat(succeed, 4)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = gollback.All(context.Background(), fns...)
		}
	})
}

func BenchmarkRetry(b *testing.B) {
	for _, n := range []int{1, 5} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			calls := 0
			succeedOnLast := func(context.Context) (int, error) {
				if calls++; calls%n == 0 {
					return calls, nil
				}
				return 0, errFailed
			}
			b.ReportAllocs()
			for b.Loop() {
				_, _ = gollback.Retry(context.Background(), n, succeedOnLast)
			}
		})
	}
}
