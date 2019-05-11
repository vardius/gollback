package gollback

import (
	"context"
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

func getCallbacks(b *testing.B) []AsyncFunc {
	cbs := make([]AsyncFunc, b.N)
	for n := 0; n < b.N; n++ {
		cbs[n] = func(ctx context.Context) (interface{}, error) {
			return n, nil
		}
	}

	return cbs
}
