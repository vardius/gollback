//go:build !race

package gollback_test

// raceEnabled is true when built with -race, which changes allocation counts.
const raceEnabled = false
