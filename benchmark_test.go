package gollback

import (
	"context"
	"errors"
	"testing"
)

func BenchmarkRace(b *testing.B) {
	cbs := getCallbacks(b)

	b.ResetTimer()

	Race(context.Background(), cbs...)
}

func BenchmarkAll(b *testing.B) {
	cbs := getCallbacks(b)

	b.ResetTimer()

	All(context.Background(), cbs...)
}

func BenchmarkRetry(b *testing.B) {
	err := errors.New("failed")

	b.ResetTimer()

	Retry(context.Background(), b.N, func(ctx context.Context) (interface{}, error) {
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
