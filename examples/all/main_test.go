package main

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	got := check(t.Context(), http.DefaultClient, srv.URL+"/", srv.URL+"/missing", "http://127.0.0.1:0/")

	want := []string{srv.URL + "/: 200 OK", srv.URL + "/missing: 404 Not Found"}
	if !slices.Equal(got[:2], want) {
		t.Errorf("check() = %q, want it to start with %q", got, want)
	}
	if !strings.HasPrefix(got[2], "http://127.0.0.1:0/: ") || !strings.Contains(got[2], "connect") {
		t.Errorf("check()[2] = %q, want the connection error", got[2])
	}
}
