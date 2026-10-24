package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vardius/gollback/v2"
)

func TestFetchFirst(t *testing.T) {
	cancelled := make(chan struct{})
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // answers only once the race cancels it
		close(cancelled)
	}))
	defer slow.Close()
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer broken.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("file contents"))
	}))
	defer good.Close()

	body, err := fetchFirst(t.Context(), http.DefaultClient, slow.URL, broken.URL, good.URL)
	if body != "file contents" || err != nil {
		t.Fatalf("fetchFirst() = %q, %v, want %q, nil", body, err, "file contents")
	}
	<-cancelled // the slow mirror's request was cancelled
}

func TestFetchFirstAllFail(t *testing.T) {
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer broken.Close()

	_, err := fetchFirst(t.Context(), http.DefaultClient, broken.URL+"/a", broken.URL+"/b")
	if err == nil || !strings.Contains(err.Error(), "/a: 503") || !strings.Contains(err.Error(), "/b: 503") {
		t.Errorf("fetchFirst() error = %v, want both mirrors' errors", err)
	}
}

func TestFetchFirstNoMirrors(t *testing.T) {
	if _, err := fetchFirst(t.Context(), http.DefaultClient); !errors.Is(err, gollback.ErrNoCallbacks) {
		t.Errorf("fetchFirst() error = %v, want %v", err, gollback.ErrNoCallbacks)
	}
}
