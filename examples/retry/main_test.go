package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetWithBackoff(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("done"))
	}))
	defer srv.Close()

	body, err := getWithBackoff(t.Context(), http.DefaultClient, srv.URL, 5, time.Millisecond)
	if body != "done" || err != nil {
		t.Fatalf("getWithBackoff() = %q, %v, want %q, nil", body, err, "done")
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("server got %d requests, want 3", n)
	}
}

func TestGetWithBackoffGivesUp(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "busy", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := getWithBackoff(t.Context(), http.DefaultClient, srv.URL, 3, time.Millisecond)
	if err == nil || !strings.HasSuffix(err.Error(), "503 Service Unavailable") {
		t.Errorf("getWithBackoff() error = %v, want the last 503", err)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("server got %d requests, want 3", n)
	}
}

func TestGetWithBackoffStopsWaitingWhenCtxIsDone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	// With an hour's pause, only ctx can end the wait.
	_, err := getWithBackoff(ctx, http.DefaultClient, srv.URL, 5, time.Hour)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("getWithBackoff() error = %v, want %v", err, context.DeadlineExceeded)
	}
}
