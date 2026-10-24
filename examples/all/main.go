// Command all checks several URLs at once and prints each one's status, in the order given.
//
//	go run ./examples/all https://go.dev https://pkg.go.dev https://example.invalid
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vardius/gollback/v2"
)

// check requests every URL concurrently and reports on each, whether or not the others failed.
func check(ctx context.Context, client *http.Client, urls ...string) []string {
	checks := make([]gollback.AsyncFunc[string], len(urls))
	for i, url := range urls {
		checks[i] = func(ctx context.Context) (string, error) {
			return status(ctx, client, url)
		}
	}

	statuses, errs := gollback.All(ctx, checks...)

	report := make([]string, len(urls))
	for i, url := range urls {
		if errs[i] != nil {
			report[i] = fmt.Sprintf("%s: %v", url, errs[i])
			continue
		}
		report[i] = fmt.Sprintf("%s: %s", url, statuses[i])
	}
	return report
}

func status(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	return resp.Status, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, line := range check(ctx, http.DefaultClient, os.Args[1:]...) {
		fmt.Println(line)
	}
}
