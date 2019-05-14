package gollback

import (
	"context"
	"errors"
	"testing"
)

func BenchmarkRace(b *testing.B) {
	g := New(context.Background())
	cbs := getCallbacks(b)

	b.ResetTimer()

	g.Race(cbs...)
}

func BenchmarkAll(b *testing.B) {
	g := New(context.Background())
	cbs := getCallbacks(b)

	b.ResetTimer()

	g.All(cbs...)
}

func BenchmarkRetry(b *testing.B) {
	g := New(context.Background())
	err := errors.New("failed")

	b.ResetTimer()

	g.Retry(b.N, func(ctx context.Context) (interface{}, error) {
		return nil, err
	})
}

func getCallbacks(b *testing.B) []AsyncFunc {
	cbs := make([]AsyncFunc, b.N)
	for n := 0; n < b.N; n++ {
		cbs[n] = func(ctx context.Context) (interface{}, error) {
			return n, nil
		}
	}

	return cbs
}
